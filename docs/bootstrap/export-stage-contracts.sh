#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

usage() {
  cat <<'EOF'
Usage:

  ./docs/bootstrap/export-stage-contracts.sh [all|1-project|2-iam|3-gke|4-ci ...]

Examples:

  ./docs/bootstrap/export-stage-contracts.sh
  ./docs/bootstrap/export-stage-contracts.sh 3-gke
  ./docs/bootstrap/export-stage-contracts.sh 1-project 2-iam

Behavior:

- exports `terraform output -json` into `config/stages/<stage>/outputs/current.json`
- skips stages that do not have readable Terraform state yet
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

stage_output_path() {
  local stage_name
  stage_name="${1#infra/}"
  printf '%s/config/stages/%s/outputs/current.json\n' "$ROOT_DIR" "$stage_name"
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

for stage in "${stages[@]}"; do
  target="$(stage_output_path "$stage")"

  if ! terraform -chdir="$stage" state pull >/dev/null 2>&1; then
    log "skipping ${stage}; no readable state yet"
    continue
  fi

  mkdir -p "$(dirname "$target")"
  log "exporting ${stage} outputs to ${target#$ROOT_DIR/}"
  terraform -chdir="$stage" output -json > "$target"
done

log "stage contract export complete"
