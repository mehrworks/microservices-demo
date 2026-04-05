#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TEMPLATE_DIR="$ROOT_DIR/config/backends"
OUTPUT_DIR="$TEMPLATE_DIR"
BUCKET=""
PREFIX_ROOT="microservices-demo"

usage() {
  cat <<'EOF'
Usage:

  ./docs/bootstrap/render-tfbackends.sh --bucket BUCKET [--prefix-root PREFIX] [--output-dir DIR]

Examples:

  ./docs/bootstrap/render-tfbackends.sh --bucket my-tf-state-bucket
  ./docs/bootstrap/render-tfbackends.sh --bucket my-tf-state-bucket --prefix-root envs/dev/microservices-demo
  ./docs/bootstrap/render-tfbackends.sh --bucket my-tf-state-bucket --output-dir /tmp/msdemo-backends

Behavior:

- reads tracked `config/backends/*.tfbackend.example` files
- writes real `*.tfbackend` files with a concrete bucket and prefix root
- keeps generated files out of git when written into `config/backends/`
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

while [[ $# -gt 0 ]]; do
  case "$1" in
    --bucket)
      BUCKET="${2:-}"
      shift 2
      ;;
    --prefix-root)
      PREFIX_ROOT="${2:-}"
      shift 2
      ;;
    --output-dir)
      OUTPUT_DIR="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "unknown argument '$1'"
      ;;
  esac
done

[[ -n "$BUCKET" ]] || die "--bucket is required"

need_cmd python3
mkdir -p "$OUTPUT_DIR"

shopt -s nullglob
templates=("$TEMPLATE_DIR"/*.tfbackend.example)
(( ${#templates[@]} > 0 )) || die "no backend example templates found in $TEMPLATE_DIR"

for template in "${templates[@]}"; do
  stage_name="$(basename "$template" .tfbackend.example)"
  output_path="$OUTPUT_DIR/${stage_name}.tfbackend"
  prefix="$PREFIX_ROOT/infra/$stage_name"

  log "rendering ${stage_name}.tfbackend"
  python3 - "$template" "$output_path" "$BUCKET" "$prefix" <<'PY'
from pathlib import Path
import sys

template_path = Path(sys.argv[1])
output_path = Path(sys.argv[2])
bucket = sys.argv[3]
prefix = sys.argv[4]

text = template_path.read_text()
text = text.replace("REPLACE_WITH_TF_STATE_BUCKET", bucket)
text = text.replace(f'microservices-demo/infra/{template_path.stem.replace(".tfbackend", "")}', prefix)
output_path.write_text(text)
PY
done

log "backend rendering complete"
