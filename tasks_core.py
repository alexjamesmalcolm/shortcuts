from __future__ import annotations
from typing import Callable, Dict, Type, Optional
from fastapi import APIRouter, Body, HTTPException
from pydantic import BaseModel, ValidationError
from celery import Celery
from celery.result import AsyncResult
from config import celery_config_from_env

# Create a shared Celery app
celery = Celery("app")
celery.conf.update(celery_config_from_env())


# Registry of declared tasks
class _TaskMeta:
    def __init__(self, name: str, input_model: Type[BaseModel], celery_name: str):
        self.name = name
        self.input_model = input_model
        self.celery_name = celery_name


TASKS: Dict[str, _TaskMeta] = {}


# Public decorator
def define_task(name: str, input_model: Type[BaseModel]):
    """
    Usage:
        @define_task("resize", ResizeInput)
        def resize(inp: ResizeInput, progress: Callable[[int, dict|None], None]) -> dict:
            ...
            return {...}
    """

    def wrapper(user_func: Callable[..., dict]):
        celery_task_name = f"tasks.{name}"

        @celery.task(bind=True, name=celery_task_name)
        def _celery_task(self, payload: dict):
            # Validate payload into typed model
            data = input_model.model_validate(payload)

            # Give user a simple progress reporter
            def progress(pct: int, meta: Optional[dict] = None):
                meta = meta or {}
                self.update_state(
                    state="PROGRESS",
                    meta={"progress": max(0, min(100, int(pct))), **meta},
                )

            # Run user function
            return user_func(data, progress)

        TASKS[name] = _TaskMeta(
            name=name, input_model=input_model, celery_name=celery_task_name
        )
        return user_func

    return wrapper


def build_router(prefix: str = "/run") -> APIRouter:
    """
    Creates a router that exposes one POST endpoint per declared task:
      POST {prefix}/{task_name}
      body := <input_model JSON>

    Returns:
      - 202 + {task_id, status} if enqueued (or 200-ish "done" in eager)
    """
    router = APIRouter()

    for name, meta in list(TASKS.items()):
        # We capture values as defaults to avoid late-binding in the loop:
        task_name = name
        input_model = meta.input_model
        celery_name = meta.celery_name

        # Endpoint factory to keep type hints per route
        def _make_endpoint(
            task_name=task_name, input_model=input_model, celery_name=celery_name
        ):
            async def endpoint(payload: dict = Body(...)):
                # Support both raw body and nested under common keys
                body = payload
                if isinstance(payload, dict):
                    if "payload" in payload and isinstance(payload["payload"], dict):
                        body = payload["payload"]
                    elif "input" in payload and isinstance(payload["input"], dict):
                        body = payload["input"]

                # Validate request body into the declared Pydantic model
                try:
                    data = input_model.model_validate(body)
                except ValidationError as e:
                    # Return a 400 with the full Pydantic error payload
                    # Keep structure as a list of issues like FastAPI's `detail` field
                    raise HTTPException(status_code=400, detail=e.errors())
                # Enqueue with the validated dict using task object
                task = celery.tasks[celery_name]
                res = task.apply_async(args=(data.model_dump(),))
                # In eager mode, task already ran
                if res.successful():
                    return {"task_id": res.id, "status": "done", "result": res.result}
                return {"task_id": res.id, "status": res.state.lower()}

            endpoint.__name__ = f"post_{task_name}"
            return endpoint

        router.add_api_route(
            path=f"{prefix}/{task_name}",
            endpoint=_make_endpoint(),
            methods=["POST"],
            status_code=202,
            name=f"run_{task_name}",
            summary=f"Run task '{task_name}'",
            response_model=None,
        )

    return router


def task_status(task_id: str):
    """Helper for a generic poll endpoint."""
    res = AsyncResult(task_id, app=celery)
    state = res.state  # PENDING/STARTED/PROGRESS/SUCCESS/FAILURE/RETRY
    if state == "SUCCESS":
        return {"status": "done", "result": res.result}
    if state == "PROGRESS":
        meta = res.info or {}
        return {
            "status": "running",
            "progress": meta.get("progress"),
            **{k: v for k, v in meta.items() if k != "progress"},
        }
    if state in {"PENDING", "STARTED"}:
        return {"status": state.lower()}
    if state in {"FAILURE", "RETRY"}:
        return {"status": "error", "error": str(res.info)}
    return {"status": state.lower()}
