# Shortcuts Tasks API

Typed, HTTP-exposed background tasks designed to be called from iOS Shortcuts.

Each task is a small Python function with a typed input model. The API exposes
`POST /run/<task-name>` for kicking off work and `GET /tasks/{task_id}` for
polling when running asynchronously. Locally, tasks run synchronously ("eager"
mode) so Shortcuts gets the result immediately.


## Overview

- FastAPI app registers all tasks in `functions/` via a simple decorator.
- Celery executes tasks. In development, Celery runs in eager mode (sync).
- In production, set a Redis URL and run a Celery worker for async execution.
- iOS Shortcuts calls the API with "Get Contents of URL" actions and parses
  simple JSON responses.


## Endpoints

- `POST /run/<task-name>`: Enqueue or run a task.
  - Dev (eager): returns `{ task_id, status: "done", result: {...} }`.
  - Prod (async): returns `{ task_id, status: "pending"|"started" }`.
- `GET /tasks/{task_id}`: Poll task status.
  - Returns `{ status: "pending"|"running"|"done"|"error", ... }`.
- `GET /healthz`: Simple health check.


## Available tasks

### travel-time
Returns OSRM-estimated travel duration (seconds) between two coordinates.

Input JSON fields:
- `start_lon_lat`: string, longitude,latitude (e.g. `"-122.4194,37.7749"`).
- `end_lon_lat`: string, longitude,latitude (e.g. `"-118.2437,34.0522"`).
- `profile` (optional): `"driving"` | `"walking"` | `"cycling"` (default `driving`).
- `max_retries` (optional), `timeout_seconds` (optional), `retry_pause_seconds` (optional).

Response on success (dev/eager mode):
```
{
  "task_id": "...",
  "status": "done",
  "result": { "duration": 21874.5, "profile": "driving" }
}
```

Note: Coordinates are `lon,lat` (not `lat,lon`).


## Using from iOS Shortcuts

Eager (local dev) and async (production) flows differ only in whether you need
to poll for completion.

1) Kick off the task
- Add a Shortcuts action: "Get Contents of URL".
- URL: `https://<your-host>/run/travel-time`
- Method: `POST`
- Request Body: JSON with the fields described above.

2a) Dev/eager mode (default locally)
- The response includes `result` immediately; extract with
  "Get Dictionary Value" → `result` → then `duration`.

2b) Async mode (production)
- The initial response returns a `task_id`.
- Create a Repeat/Until loop:
  - Wait 0.5–1.0s
  - "Get Contents of URL" → `https://<your-host>/tasks/{task_id}`
  - If `status` is `done`, extract `result` and break.
  - If `status` is `error`, show an alert and stop.


## Local development

Prereqs: Python 3.11+, `pip`.

1) Install dependencies
```
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

2) Enable eager mode for local (already default)
- Create `.env` with:
```
CELERY_TASK_ALWAYS_EAGER=true
```

3) Run the API
```
fastapi dev app.py
```
- Open Swagger at `http://127.0.0.1:8000/docs`.
- Test via curl:
```
curl -sS -X POST \
  -H 'Content-Type: application/json' \
  -d '{"start_lon_lat":"-122.4194,37.7749","end_lon_lat":"-118.2437,34.0522"}' \
  http://127.0.0.1:8000/run/travel-time | jq
```


## Production / async execution

1) Set environment
- `REDIS_URL=redis://:password@host:port/0`
- `CELERY_TASK_ALWAYS_EAGER=false` (or unset)

2) Run a Celery worker
```
bash ./worker.sh
```

3) Run the API (e.g., `uvicorn` or `fastapi run`)
- `POST /run/<task>` returns `{ task_id, status }`.
- Poll `GET /tasks/{task_id}` until `status` is `done`.

Notes:
- In non-eager mode, requests will hang without a running worker.
- OSRM is a public service; be mindful of rate limits and availability.


## Adding a new task

1) Create a module in `functions/` with a Pydantic input model and a function
decorated with `@define_task("your-task", YourInputModel)`.

2) Import the function in `functions/__init__.py` so the FastAPI app registers it.

Example skeleton:
```python
# functions/example_task.py
from pydantic import BaseModel
from tasks_core import define_task

class ExampleInput(BaseModel):
    name: str

@define_task("hello", ExampleInput)
def hello(inp: ExampleInput, progress):
    progress(100)
    return {"message": f"Hello, {inp.name}!"}

# functions/__init__.py
from .example_task import hello  # ensure import for registration
```

After saving, restart dev server (or rely on auto-reload) and you should see a
new `POST /run/hello` endpoint in `/docs`.


## Troubleshooting

- 400 validation errors: check your JSON matches the task input schema.
- No result in production: ensure `REDIS_URL` is set and the Celery worker is
  running.
- OSRM errors: verify coordinate order is `lon,lat` and consider retrying.


---

This project is optimized for personal use with iOS Shortcuts, but the API is
generic and can be called from any HTTP client.
