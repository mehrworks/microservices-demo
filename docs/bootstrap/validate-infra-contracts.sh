#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

usage() {
  cat <<'EOF'
Usage:

  ./docs/bootstrap/validate-infra-contracts.sh [all|1-project|2-iam|3-gke|4-ci ...]

Examples:

  ./docs/bootstrap/validate-infra-contracts.sh
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

need_cmd terraform

cd "$ROOT_DIR"

if [[ "${1:-all}" == "-h" || "${1:-all}" == "--help" ]]; then
  usage
  exit 0
fi

declare -a stages

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

for stage in "${stages[@]}"; do
  log "initializing ${stage}"
  terraform -chdir="$stage" init -backend=false -input=false

  log "validating ${stage}"
  terraform -chdir="$stage" validate

  log "planning ${stage}"
  terraform -chdir="$stage" plan -input=false -lock=false -no-color
done

log "checking diff hygiene"
git diff --check

log "infra contract validation complete"
