import os


def celery_config_from_env():
    """
    Returns a dict of Celery settings from environment variables.

    DEV (no broker):
      CELERY_TASK_ALWAYS_EAGER=true
      (No REDIS_URL required)

    PROD (Railway):
      REDIS_URL=redis://:password@host:port/0
      CELERY_TASK_ALWAYS_EAGER=false (or unset)
    """
    eager = os.getenv("CELERY_TASK_ALWAYS_EAGER", "false").lower() == "true"
    redis_url = os.getenv("REDIS_URL")

    cfg = {
        # Eager mode (sync, no external services)
        "task_always_eager": eager,
        "task_eager_propagates": True,  # raise errors in dev
        # Serialization
        "task_serializer": "json",
        "accept_content": ["json"],
        "result_serializer": "json",
        # Optional: visibility timeouts, acks, etc.
        "task_acks_late": True,
        "worker_prefetch_multiplier": 1,
    }

    # Only wire broker/backends if NOT eager
    if not eager:
        if not redis_url:
            raise RuntimeError("REDIS_URL is required in non-eager mode.")
        cfg.update(
            {
                "broker_url": redis_url,
                "result_backend": redis_url,
                # You can tune time limits as needed
                "task_time_limit": 660,
                "task_soft_time_limit": 600,
            }
        )

    return cfg
