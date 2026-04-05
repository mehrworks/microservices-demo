#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SKIP_EXPORT="false"

usage() {
  cat <<'EOF'
Usage:

  ./docs/bootstrap/staged-infra-cycle.sh [--skip-export] [all|1-project|2-iam|3-gke|4-ci ...]

Examples:

  ./docs/bootstrap/staged-infra-cycle.sh
  ./docs/bootstrap/staged-infra-cycle.sh 3-gke
  ./docs/bootstrap/staged-infra-cycle.sh --skip-export 1-project 2-iam

Behavior:

- runs `validate-infra-contracts.sh`
- then runs `export-stage-contracts.sh` unless `--skip-export` is set
EOF
}

log() {
  printf '==> %s\n' "$*"
}

declare -a stages

while [[ $# -gt 0 ]]; do
  case "$1" in
    --skip-export)
      SKIP_EXPORT="true"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      stages+=("$1")
      shift
      ;;
  esac
done

cd "$ROOT_DIR"

if [[ ${#stages[@]} -eq 0 ]]; then
  stages=(all)
fi

log "running staged infra validation"
./docs/bootstrap/validate-infra-contracts.sh "${stages[@]}"

if [[ "$SKIP_EXPORT" == "true" ]]; then
  log "skipping stage contract export"
  exit 0
fi

log "exporting stage contracts when state exists"
./docs/bootstrap/export-stage-contracts.sh "${stages[@]}"

log "staged infra cycle complete"
