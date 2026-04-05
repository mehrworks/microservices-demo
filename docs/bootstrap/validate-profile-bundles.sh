#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PROFILE_ROOT="$ROOT_DIR/config/profiles"

usage() {
  cat <<'EOF'
Usage:

  ./docs/bootstrap/validate-profile-bundles.sh [all|PROFILE ...]

Examples:

  ./docs/bootstrap/validate-profile-bundles.sh
  ./docs/bootstrap/validate-profile-bundles.sh sandbox-public org-hardened
EOF
}

log() {
  printf '==> %s\n' "$*"
}

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

cd "$ROOT_DIR"

if [[ "${1:-all}" == "-h" || "${1:-all}" == "--help" ]]; then
  usage
  exit 0
fi

declare -a profile_names

if [[ $# -eq 0 || "$1" == "all" ]]; then
  while IFS= read -r bundle; do
    profile_names+=("$(basename "$(dirname "$bundle")")")
  done < <(find "$PROFILE_ROOT" -mindepth 2 -maxdepth 2 -type f -name 'bundle.yaml' | sort)
else
  profile_names=("$@")
fi

(( ${#profile_names[@]} > 0 )) || die "no profiles found"

python3 - "$ROOT_DIR" "${profile_names[@]}" <<'PY'
from pathlib import Path
import sys

root = Path(sys.argv[1])
profiles = sys.argv[2:]
required = ["project", "iam", "gke", "ci"]

def parse_bundle(path: Path):
    data = {"datasets": {}}
    section = None
    for raw in path.read_text().splitlines():
        line = raw.rstrip()
        if not line or line.lstrip().startswith("#"):
            continue
        if not line.startswith("  "):
            section = None
            key, value = line.split(":", 1)
            key = key.strip()
            value = value.strip()
            if key == "datasets":
                section = "datasets"
            else:
                data[key] = value
            continue
        if section == "datasets":
            key, value = line.strip().split(":", 1)
            data["datasets"][key.strip()] = value.strip()
    return data

for name in profiles:
    bundle = root / "config" / "profiles" / name / "bundle.yaml"
    if not bundle.is_file():
        raise SystemExit(f"error: bundle not found for profile '{name}': {bundle}")

    data = parse_bundle(bundle)

    if data.get("name") != name:
        raise SystemExit(f"error: profile name mismatch in {bundle}: expected '{name}', got '{data.get('name')}'")

    if not data.get("description"):
        raise SystemExit(f"error: missing description in {bundle}")

    datasets = data.get("datasets", {})
    missing = [key for key in required if key not in datasets]
    if missing:
        raise SystemExit(f"error: missing dataset mappings in {bundle}: {', '.join(missing)}")

    extra = [key for key in datasets if key not in required]
    if extra:
        raise SystemExit(f"error: unexpected dataset mappings in {bundle}: {', '.join(extra)}")

    for dataset, rel_path in datasets.items():
        target = root / rel_path
        if not target.is_file():
            raise SystemExit(f"error: missing referenced file for {name}/{dataset}: {target}")
        expected_prefix = root / "config" / "datasets" / dataset
        if expected_prefix not in target.parents:
            raise SystemExit(f"error: {name}/{dataset} must point under {expected_prefix.relative_to(root)}")

    print(f"validated profile bundle: {name}")
PY

log "profile bundle validation complete"
