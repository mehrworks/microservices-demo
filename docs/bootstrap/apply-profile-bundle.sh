#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PROFILE_ROOT="$ROOT_DIR/config/profiles"
FORCE="false"
DRY_RUN="false"

usage() {
  cat <<'EOF'
Usage:

  ./docs/bootstrap/apply-profile-bundle.sh [--force] [--dry-run] PROFILE

Examples:

  ./docs/bootstrap/apply-profile-bundle.sh sandbox-public
  ./docs/bootstrap/apply-profile-bundle.sh --dry-run org-hardened
  ./docs/bootstrap/apply-profile-bundle.sh --force adopt-existing-foundation

Behavior:

- reads `config/profiles/<profile>/bundle.yaml`
- copies referenced dataset example files into local `config/datasets/*/overrides.yaml`
- refuses to overwrite existing `overrides.yaml` files unless `--force` is set
EOF
}

log() {
  printf '==> %s\n' "$*"
}

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --force)
      FORCE="true"
      shift
      ;;
    --dry-run)
      DRY_RUN="true"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      PROFILE_NAME="$1"
      shift
      if [[ $# -gt 0 ]]; then
        die "unexpected extra arguments"
      fi
      break
      ;;
  esac
done

[[ -n "${PROFILE_NAME:-}" ]] || die "profile name is required"

PROFILE_DIR="$PROFILE_ROOT/$PROFILE_NAME"
BUNDLE_FILE="$PROFILE_DIR/bundle.yaml"

[[ -d "$PROFILE_DIR" ]] || die "profile directory not found: $PROFILE_DIR"
[[ -f "$BUNDLE_FILE" ]] || die "bundle file not found: $BUNDLE_FILE"

cd "$ROOT_DIR"

./docs/bootstrap/validate-profile-bundles.sh "$PROFILE_NAME" >/dev/null

python3 - "$ROOT_DIR" "$BUNDLE_FILE" "$FORCE" "$DRY_RUN" <<'PY'
from pathlib import Path
import shutil
import sys

root = Path(sys.argv[1])
bundle = Path(sys.argv[2])
force = sys.argv[3].lower() == "true"
dry_run = sys.argv[4].lower() == "true"

lines = bundle.read_text().splitlines()
datasets = {}
in_datasets = False

for raw in lines:
    line = raw.rstrip()
    if not line or line.lstrip().startswith("#"):
        continue
    if line == "datasets:":
        in_datasets = True
        continue
    if not in_datasets:
        continue
    if not line.startswith("  "):
        break
    key, value = line.strip().split(":", 1)
    datasets[key.strip()] = value.strip()

if not datasets:
    raise SystemExit(f"error: no datasets found in {bundle}")

for dataset, rel_source in datasets.items():
    source = root / rel_source
    target = root / "config" / "datasets" / dataset / "overrides.yaml"
    if not source.is_file():
        raise SystemExit(f"error: source file not found for {dataset}: {source}")
    if target.exists() and not force:
        raise SystemExit(f"error: target already exists for {dataset}: {target}. Re-run with --force to overwrite.")
    print(f"{dataset}: {source.relative_to(root)} -> {target.relative_to(root)}")
    if not dry_run:
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(source, target)
PY

if [[ "$DRY_RUN" == "true" ]]; then
  log "dry run complete"
else
  log "profile bundle applied"
fi
