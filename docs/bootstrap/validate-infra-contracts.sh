#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WORKSPACE=""

usage() {
  cat <<'EOF'
Usage:

  ./docs/bootstrap/validate-infra-contracts.sh [--workspace NAME] [all|1-project|2-iam|3-gke|4-ci ...]

Examples:

  ./docs/bootstrap/validate-infra-contracts.sh
  ./docs/bootstrap/validate-infra-contracts.sh --workspace sandbox-public
  ./docs/bootstrap/validate-infra-contracts.sh 3-gke
  ./docs/bootstrap/validate-infra-contracts.sh 1-project 2-iam

See also:

  ./docs/bootstrap/staged-infra-cycle.sh
EOF
}

log() {
  printf '==> %s\n' "$*"
}

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "missing required command: $1"
}

validate_contract_artifacts() {
  python3 - "$ROOT_DIR" <<'PY'
from pathlib import Path
import json
import sys

root = Path(sys.argv[1])

schema_files = [
    "config/contracts/job-control/schema/submit-job-request.schema.json",
    "config/contracts/job-control/schema/submit-job-response.schema.json",
    "config/contracts/job-control/schema/worker-identity.schema.json",
    "config/contracts/job-control/schema/lease.schema.json",
    "config/contracts/job-control/schema/claim-next-request.schema.json",
    "config/contracts/job-control/schema/claim-next-response.schema.json",
    "config/contracts/job-control/schema/status-update.schema.json",
    "config/contracts/job-control/schema/job-record.schema.json",
    "config/contracts/job-control/schema/job-store-record.schema.json",
    "config/contracts/job-control/schema/job-store-update.schema.json",
    "config/contracts/job-control/schema/result-access-record.schema.json",
    "config/contracts/job-control/schema/cancel-job-response.schema.json",
    "config/contracts/job-control/schema/result-reference.schema.json",
    "config/contracts/job-control/schema/error-envelope.schema.json",
    "config/contracts/job-control/job-control.openapi.json",
]

examples = {
    "config/contracts/job-control/submit-job-request.example.json": ["job_type", "submitted_by", "payload"],
    "config/contracts/job-control/submit-job-response.example.json": ["job_id", "state", "job_ref"],
    "config/contracts/job-control/worker-identity.example.json": ["worker_id", "auth_mode", "auth_subject"],
    "config/contracts/job-control/lease.example.json": ["lease_id", "issued_at", "lease_expires_at", "lease_duration_seconds"],
    "config/contracts/job-control/claim-next-request.example.json": ["worker"],
    "config/contracts/job-control/claim-next-response.example.json": ["worker_id", "poll_after_seconds", "job"],
    "config/contracts/job-control/status-update.example.json": ["job_id", "worker_id", "state", "updated_at"],
    "config/contracts/job-control/job-record.example.json": ["job_id", "job_type", "state", "submitted_at", "updated_at", "payload"],
    "config/contracts/job-control/job-store-record.example.json": ["job", "revision", "etag"],
    "config/contracts/job-control/job-store-update.example.json": ["job_id", "expected_revision"],
    "config/contracts/job-control/result-access-record.example.json": ["result_ref", "retrieval_mode", "available"],
    "config/contracts/job-control/cancel-job-response.example.json": ["job_id", "cancel_requested", "state"],
    "config/contracts/job-control/result-reference.example.json": ["kind", "uri", "content_type"],
    "config/contracts/job-control/error-envelope.example.json": ["error_code", "message"],
}

for rel in schema_files:
    path = root / rel
    if not path.is_file():
        raise SystemExit(f"error: missing contract schema {rel}")
    data = json.loads(path.read_text())
    if rel.endswith("job-control.openapi.json"):
        required_paths = {
            "/v1/jobs",
            "/v1/jobs/{job_id}",
            "/v1/jobs/{job_id}:cancel",
            "/v1/worker/claim",
            "/v1/jobs/{job_id}/status",
        }
        missing = sorted(required_paths.difference(data.get("paths", {}).keys()))
        if missing:
            raise SystemExit(f"error: missing job-control API paths: {', '.join(missing)}")
        schemes = set(data.get("components", {}).get("securitySchemes", {}).keys())
        missing_schemes = {"operatorBearer", "workerBearer"}.difference(schemes)
        if missing_schemes:
            raise SystemExit(f"error: missing job-control security schemes: {', '.join(sorted(missing_schemes))}")

for rel, required_keys in examples.items():
    path = root / rel
    if not path.is_file():
        raise SystemExit(f"error: missing contract example {rel}")
    data = json.loads(path.read_text())
    missing = [key for key in required_keys if key not in data]
    if missing:
        raise SystemExit(f"error: missing keys in {rel}: {', '.join(missing)}")

print("validated hybrid contract artifacts")
PY
}

resolve_stage() {
  case "$1" in
    1-project|2-iam|3-gke|4-ci)
      printf 'infra/%s\n' "$1"
      ;;
    infra/1-project|infra/2-iam|infra/3-gke|infra/4-ci)
      printf '%s\n' "$1"
      ;;
    *)
      die "unknown stage '$1'"
      ;;
  esac
}

ensure_workspace() {
  local stage="$1"

  [[ -n "$WORKSPACE" ]] || return 0

  if terraform -chdir="$stage" workspace select "$WORKSPACE" >/dev/null 2>&1; then
    log "selected workspace ${WORKSPACE} for ${stage}"
    return 0
  fi

  terraform -chdir="$stage" workspace new "$WORKSPACE" >/dev/null
  log "created workspace ${WORKSPACE} for ${stage}"
}

need_cmd terraform
need_cmd python3
need_cmd go

cd "$ROOT_DIR"

if [[ "${1:-all}" == "-h" || "${1:-all}" == "--help" ]]; then
  usage
  exit 0
fi

declare -a stages

while [[ $# -gt 0 ]]; do
  case "$1" in
    --workspace)
      WORKSPACE="${2:-}"
      [[ -n "$WORKSPACE" ]] || die "--workspace requires a value"
      shift 2
      ;;
    *)
      break
      ;;
  esac
done

if [[ $# -eq 0 || "$1" == "all" ]]; then
  stages=(
    "infra/1-project"
    "infra/2-iam"
    "infra/3-gke"
    "infra/4-ci"
  )
else
  for stage in "$@"; do
    stages+=("$(resolve_stage "$stage")")
  done
fi

log "checking terraform formatting"
terraform fmt -check -recursive infra

log "validating profile bundles"
./docs/bootstrap/validate-profile-bundles.sh

log "validating contract artifacts"
validate_contract_artifacts

log "validating frontend job-control package"
(cd "$ROOT_DIR/src/frontend" && go test ./jobcontrol/... >/dev/null)

log "validating job-control reference package"
(cd "$ROOT_DIR/tools/jobcontrolref" && go test ./... >/dev/null)

for stage in "${stages[@]}"; do
  log "initializing ${stage}"
  terraform -chdir="$stage" init -backend=false -input=false

  ensure_workspace "$stage"

  log "validating ${stage}"
  terraform -chdir="$stage" validate

  log "planning ${stage}"
  terraform -chdir="$stage" plan -input=false -lock=false -no-color
done

log "checking diff hygiene"
git diff --check

log "infra contract validation complete"
