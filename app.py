from fastapi import FastAPI, HTTPException
from tasks_core import build_router, task_status

app = FastAPI(title="Typed Tasks API")

# Register routes for all tasks you've defined under functions/
# Importing functions.* modules populates the task registry via decorators
import functions  # noqa: F401

# Mount the auto-generated endpoints under /run/<task-name>
app.include_router(build_router(prefix="/run"))


# Generic poll endpoint
@app.get("/tasks/{task_id}")
def get_task(task_id: str):
    if not task_id:
        raise HTTPException(400, "task_id required")
    return task_status(task_id)


# Optional health check
@app.get("/healthz")
def healthz():
    return {"ok": True}
