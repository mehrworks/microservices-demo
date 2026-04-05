#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WORKSPACE=""
OUTPUT_PATH="$ROOT_DIR/config/gke-exposure/overrides.yaml"

usage() {
  cat <<'EOF'
Usage:

  ./docs/bootstrap/render-gke-proof-bind.sh [--workspace NAME] [--output PATH]

Examples:

  ./docs/bootstrap/render-gke-proof-bind.sh
  ./docs/bootstrap/render-gke-proof-bind.sh --workspace org-constrained
  ./docs/bootstrap/render-gke-proof-bind.sh --workspace default --output /tmp/gke-proof.yaml

Behavior:

- reads exported stage outputs for `1-project` and `3-gke`
- renders a local `config/gke-exposure/overrides.yaml` proof bind
- chooses the proof URL based on the staged `proof_contract`
EOF
}

log() {
  printf '==> %s\n' "$*"
}

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

sanitize_workspace() {
  printf '%s' "$1" | tr -c 'A-Za-z0-9._-' '_'
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --workspace)
      WORKSPACE="${2:-}"
      [[ -n "$WORKSPACE" ]] || die "--workspace requires a value"
      shift 2
      ;;
    --output)
      OUTPUT_PATH="${2:-}"
      [[ -n "$OUTPUT_PATH" ]] || die "--output requires a value"
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

cd "$ROOT_DIR"

if [[ -z "$WORKSPACE" ]]; then
  WORKSPACE="$(terraform -chdir="infra/3-gke" workspace show 2>/dev/null || printf 'default')"
fi

WORKSPACE_FILE="$(sanitize_workspace "$WORKSPACE")"
PROJECT_OUTPUT="$ROOT_DIR/config/stages/1-project/outputs/${WORKSPACE_FILE}.json"
GKE_OUTPUT="$ROOT_DIR/config/stages/3-gke/outputs/${WORKSPACE_FILE}.json"

[[ -f "$PROJECT_OUTPUT" ]] || die "missing stage output snapshot: ${PROJECT_OUTPUT#$ROOT_DIR/}; run ./docs/bootstrap/export-stage-contracts.sh --workspace $WORKSPACE 1-project 3-gke first"
[[ -f "$GKE_OUTPUT" ]] || die "missing stage output snapshot: ${GKE_OUTPUT#$ROOT_DIR/}; run ./docs/bootstrap/export-stage-contracts.sh --workspace $WORKSPACE 1-project 3-gke first"

eval "$({
  python3 - "$PROJECT_OUTPUT" "$GKE_OUTPUT" <<'PY'
import json
import shlex
import sys

project = json.load(open(sys.argv[1]))
gke = json.load(open(sys.argv[2]))

project_contract = project.get("project_contract", {}).get("value", {})
cluster_contract = gke.get("cluster_contract", {}).get("value", {})
proof_contract = gke.get("proof_contract", {}).get("value", {})

values = {
    "PROJECT_ID": project_contract.get("project_id") or project.get("project_id", {}).get("value") or "",
    "REGION": project_contract.get("region") or gke.get("region", {}).get("value") or "europe-west3",
    "CLUSTER_NAME": cluster_contract.get("name") or "msdemo-public",
    "NAMESPACE": gke.get("namespace", {}).get("value") or "msdemo-public",
    "AR_REPO": gke.get("artifact_registry_repository_id", {}).get("value") or "services",
    "DEFAULT_REPO": gke.get("artifact_registry_repository_url", {}).get("value") or "",
    "HTTP_MODE": proof_contract.get("http_mode") or "public-load-balancer",
    "GRPC_MODE": proof_contract.get("grpc_mode") or "port-forward",
    "FRONTEND_SERVICE_NAME": proof_contract.get("frontend_service_name") or "frontend",
    "FRONTEND_PUBLIC_SERVICE_NAME": proof_contract.get("frontend_public_service_name") or "frontend-external",
    "FRONTEND_LOCAL_PORT": str(proof_contract.get("frontend_local_port") or 8081),
    "PRODUCTCATALOG_SERVICE_NAME": proof_contract.get("productcatalog_service_name") or "productcatalogservice",
    "PRODUCTCATALOG_LOCAL_PORT": str(proof_contract.get("productcatalog_local_port") or 3550),
    "RECOMMENDATION_SERVICE_NAME": proof_contract.get("recommendation_service_name") or "recommendationservice",
}

for key, value in values.items():
    print(f"{key}={shlex.quote(str(value))}")
PY
} )"

FRONTEND_BASE_URL="http://REPLACE_WITH_FRONTEND_EXTERNAL_IP"

if [[ "$HTTP_MODE" == "port-forward" ]]; then
  FRONTEND_BASE_URL="http://127.0.0.1:${FRONTEND_LOCAL_PORT}"
else
  if command -v kubectl >/dev/null 2>&1; then
    ip="$(kubectl get service "$FRONTEND_PUBLIC_SERVICE_NAME" -n "$NAMESPACE" -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || true)"
    if [[ -n "$ip" ]]; then
      FRONTEND_BASE_URL="http://${ip}"
    fi
  fi
fi

mkdir -p "$(dirname "$OUTPUT_PATH")"

cat > "$OUTPUT_PATH" <<EOF
# yaml-language-server: \$schema=./schema/proof-bind.schema.json

project_id: $PROJECT_ID
region: $REGION
cluster: $CLUSTER_NAME
namespace: $NAMESPACE
artifact_registry_repository: $AR_REPO

default_repo: $DEFAULT_REPO
frontend_public_base_url: $FRONTEND_BASE_URL

frontend:
  mode: catalog-only
  service_name: $FRONTEND_SERVICE_NAME
  public_service_name: $FRONTEND_PUBLIC_SERVICE_NAME

productcatalogservice:
  service_name: $PRODUCTCATALOG_SERVICE_NAME

recommendationservice:
  service_name: $RECOMMENDATION_SERVICE_NAME

proof:
  http_mode: $HTTP_MODE
  grpc_mode: $GRPC_MODE
  frontend_local_port: $FRONTEND_LOCAL_PORT
  productcatalog_local_port: $PRODUCTCATALOG_LOCAL_PORT
EOF

log "rendered ${OUTPUT_PATH#$ROOT_DIR/} from workspace ${WORKSPACE}"

cat <<EOF
Proof commands for workspace ${WORKSPACE}:

skaffold run --namespace "$NAMESPACE" --default-repo "$DEFAULT_REPO"

kubectl wait --for=condition=available deployment/frontend deployment/productcatalogservice deployment/$RECOMMENDATION_SERVICE_NAME --timeout=600s -n "$NAMESPACE"
EOF

if [[ "$HTTP_MODE" == "public-load-balancer" ]]; then
  cat <<EOF

kubectl get service $FRONTEND_PUBLIC_SERVICE_NAME -n "$NAMESPACE"

curl -fsS "$FRONTEND_BASE_URL/" | grep "Catalog-only mode is active"

curl -fsS "$FRONTEND_BASE_URL/product/OLJCESPC7Z" | grep "You May Also Like"
EOF
else
  cat <<EOF

kubectl port-forward svc/$FRONTEND_SERVICE_NAME $FRONTEND_LOCAL_PORT:80 -n "$NAMESPACE"

curl -fsS "$FRONTEND_BASE_URL/" | grep "Catalog-only mode is active"

curl -fsS "$FRONTEND_BASE_URL/product/OLJCESPC7Z" | grep "You May Also Like"
EOF
fi

cat <<EOF

kubectl port-forward svc/$PRODUCTCATALOG_SERVICE_NAME $PRODUCTCATALOG_LOCAL_PORT:3550 -n "$NAMESPACE"

grpcurl -plaintext -import-path protos -proto demo.proto -d '{}' localhost:$PRODUCTCATALOG_LOCAL_PORT hipstershop.ProductCatalogService/ListProducts
EOF
