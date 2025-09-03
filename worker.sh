#!/usr/bin/env bash
set -euo pipefail
# Railway sets REDIS_URL; eager mode should be false/unset in prod.
exec celery -A tasks.celery worker --loglevel=INFO
