#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
OPERATOR_TOKEN="${JOB_CONTROL_OPERATOR_TOKEN:-}"
WORKER_TOKEN="${JOB_CONTROL_WORKER_TOKEN:-}"
JOB_TYPE="${JOB_CONTROL_JOB_TYPE:-local-analysis}"
WORKER_ID="${JOB_CONTROL_WORKER_ID:-worker:static-bearer}"
WORKER_SUBJECT="${JOB_CONTROL_WORKER_SUBJECT:-$WORKER_ID}"

usage() {
  cat <<'EOF'
Usage:

  JOB_CONTROL_OPERATOR_TOKEN=... JOB_CONTROL_WORKER_TOKEN=... \
    ./docs/bootstrap/job-control-smoke.sh

Optional environment variables:

  BASE_URL=http://127.0.0.1:8080
  JOB_CONTROL_JOB_TYPE=local-analysis

This helper exercises the dormant internal operator/worker simulation path:

1. submit a job
2. fetch the job record
3. claim the job as the simulated worker
4. post a succeeded status update with a synthetic local-file result
5. fetch the result-access record
6. request cancellation to show the endpoint behavior after a terminal update

This helper remains a manual simulation path. It does not exercise lease renewal
or stale-worker recovery logic automatically.
EOF
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    printf 'error: missing required command: %s\n' "$1" >&2
    exit 1
  }
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

[[ -n "$OPERATOR_TOKEN" ]] || {
  printf 'error: JOB_CONTROL_OPERATOR_TOKEN must be set\n' >&2
  exit 1
}

[[ -n "$WORKER_TOKEN" ]] || {
  printf 'error: JOB_CONTROL_WORKER_TOKEN must be set\n' >&2
  exit 1
}

need_cmd curl
need_cmd python3

submit_payload=$(python3 - <<'PY'
import json
import os
print(json.dumps({
    "job_type": os.environ.get("JOB_CONTROL_JOB_TYPE", "local-analysis"),
    "payload": {"input": "demo", "mode": "smoke"}
}))
PY
)

printf '==> submit job\n'
submit_response=$(curl -fsS -X POST "$BASE_URL/internal/jobs" \
  -H "Authorization: Bearer $OPERATOR_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$submit_payload")
printf '%s\n' "$submit_response"

JOB_ID=$(python3 - <<'PY' "$submit_response"
import json
import sys
print(json.loads(sys.argv[1])["job_id"])
PY
)

printf '==> get job\n'
curl -fsS "$BASE_URL/internal/jobs/$JOB_ID" \
  -H "Authorization: Bearer $OPERATOR_TOKEN"
printf '\n'

printf '==> claim job as simulated worker\n'
claim_payload=$(python3 - <<'PY' "$WORKER_ID" "$WORKER_SUBJECT"
import json
import sys
print(json.dumps({
    "worker_id": sys.argv[1],
    "auth_mode": "bearer",
    "auth_subject": sys.argv[2]
}))
PY
)
claim_response=$(curl -fsS -X POST "$BASE_URL/internal/worker/claim" \
  -H "Authorization: Bearer $WORKER_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$claim_payload")
printf '%s\n' "$claim_response"

printf '==> mark job succeeded\n'
status_payload=$(python3 - <<'PY' "$JOB_ID"
import json
import sys
job_id = sys.argv[1]
print(json.dumps({
    "state": "succeeded",
    "progress_percent": 100,
    "status_message": "manual smoke complete",
    "result_ref": {
        "kind": "local-file",
        "uri": f"file:///tmp/{job_id}.json",
        "content_type": "application/json",
        "size_bytes": 128
    }
}))
PY
)

curl -fsS -X POST "$BASE_URL/internal/jobs/$JOB_ID/status" \
  -H "Authorization: Bearer $WORKER_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$status_payload"
printf '\n'

printf '==> get result access\n'
curl -fsS "$BASE_URL/internal/jobs/$JOB_ID/result" \
  -H "Authorization: Bearer $OPERATOR_TOKEN"
printf '\n'

printf '==> request cancel after terminal state\n'
curl -fsS -X POST "$BASE_URL/internal/jobs/$JOB_ID:cancel" \
  -H "Authorization: Bearer $OPERATOR_TOKEN"
printf '\n'

printf '==> smoke flow complete for %s\n' "$JOB_ID"
