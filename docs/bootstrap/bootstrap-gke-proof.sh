#!/usr/bin/env bash
set -euo pipefail

MODE="${1:-bootstrap}"
PROJECT_ID="${PROJECT_ID:-}"
REGION="${REGION:-europe-west3}"
AR_REPO="${AR_REPO:-services}"
CLUSTER="${CLUSTER:-msdemo-public}"
NAMESPACE="${NAMESPACE:-msdemo-public}"
CREATE_CLUSTER="${CREATE_CLUSTER:-false}"

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

is_true() {
  case "${1,,}" in
    1|true|yes|y)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

need_cmd gcloud
need_cmd kubectl
need_cmd skaffold
need_cmd docker

ACTIVE_ACCOUNT="$(gcloud config get-value account 2>/dev/null || true)"
ACTIVE_PROJECT="$(gcloud config get-value project 2>/dev/null || true)"

[[ -n "$ACTIVE_ACCOUNT" ]] || die "no active gcloud account; run 'gcloud auth login' first"
[[ -n "$PROJECT_ID" ]] || die "PROJECT_ID is required"

DEFAULT_REPO="${DEFAULT_REPO:-${REGION}-docker.pkg.dev/${PROJECT_ID}/${AR_REPO}}"
REQUIRED_APIS=(
  container.googleapis.com
  artifactregistry.googleapis.com
  cloudbuild.googleapis.com
  compute.googleapis.com
)

show_state() {
  log "active gcloud account: $ACTIVE_ACCOUNT"
  log "gcloud config project: ${ACTIVE_PROJECT:-<unset>}"
  log "target project: $PROJECT_ID"
  log "region: $REGION"
  log "artifact registry repo: $AR_REPO"
  log "cluster: $CLUSTER"
  log "namespace: $NAMESPACE"
  log "default repo: $DEFAULT_REPO"
}

enable_apis() {
  log "enabling required APIs"
  gcloud services enable "${REQUIRED_APIS[@]}" --project "$PROJECT_ID"
}

ensure_artifact_registry() {
  if gcloud artifacts repositories describe "$AR_REPO" --project "$PROJECT_ID" --location "$REGION" >/dev/null 2>&1; then
    log "artifact registry repo already exists: $AR_REPO"
    return
  fi

  log "creating artifact registry repo: $AR_REPO"
  gcloud artifacts repositories create "$AR_REPO" \
    --project "$PROJECT_ID" \
    --repository-format=docker \
    --location "$REGION" \
    --description "Microservices demo proof images"
}

configure_docker() {
  log "configuring Docker credential helper for ${REGION}-docker.pkg.dev"
  gcloud auth configure-docker "${REGION}-docker.pkg.dev" -q
}

cluster_exists() {
  gcloud container clusters describe "$CLUSTER" --project "$PROJECT_ID" --region "$REGION" >/dev/null 2>&1
}

print_cluster_create_command() {
  cat <<EOF
Long wait step still required:

gcloud container clusters create-auto "$CLUSTER" --project "$PROJECT_ID" --region "$REGION"

gcloud container clusters get-credentials "$CLUSTER" --project "$PROJECT_ID" --region "$REGION"
EOF
}

ensure_cluster_for_proof() {
  if cluster_exists; then
    log "cluster already exists: $CLUSTER"
    return
  fi

  if is_true "$CREATE_CLUSTER"; then
    log "creating cluster because CREATE_CLUSTER=$CREATE_CLUSTER"
    gcloud container clusters create-auto "$CLUSTER" \
      --project "$PROJECT_ID" \
      --region "$REGION"
    return
  fi

  print_cluster_create_command
  die "cluster '$CLUSTER' does not exist; rerun after it is created or set CREATE_CLUSTER=true"
}

require_gke_plugin() {
  need_cmd gke-gcloud-auth-plugin
}

fetch_credentials() {
  log "fetching cluster credentials"
  gcloud container clusters get-credentials "$CLUSTER" \
    --project "$PROJECT_ID" \
    --region "$REGION"
}

ensure_namespace() {
  log "creating or reusing namespace: $NAMESPACE"
  kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
  kubectl config set-context --current --namespace="$NAMESPACE" >/dev/null
}

print_overrides_reminder() {
  cat <<EOF
Recommended config bind:

cp config/gke-exposure/overrides.example.yaml config/gke-exposure/overrides.yaml

Then set:
- project_id: $PROJECT_ID
- default_repo: $DEFAULT_REPO
- frontend_public_base_url: http://REPLACE_WITH_FRONTEND_EXTERNAL_IP
EOF
}

print_proof_commands() {
  cat <<EOF
Proof commands:

skaffold run --namespace "$NAMESPACE" --default-repo "$DEFAULT_REPO"

kubectl wait --for=condition=available deployment/frontend deployment/productcatalogservice deployment/recommendationservice --timeout=600s -n "$NAMESPACE"

kubectl get service frontend-external -n "$NAMESPACE"

curl -fsS "http://EXTERNAL_IP/" | grep "Catalog-only mode is active"

curl -fsS "http://EXTERNAL_IP/product/OLJCESPC7Z" | grep "You May Also Like"

kubectl port-forward svc/productcatalogservice 3550:3550 -n "$NAMESPACE"

grpcurl -plaintext -import-path protos -proto demo.proto -d '{}' localhost:3550 hipstershop.ProductCatalogService/ListProducts
EOF
}

gcloud config set project "$PROJECT_ID" >/dev/null
show_state

case "$MODE" in
  bootstrap)
    enable_apis
    ensure_artifact_registry
    configure_docker
    print_overrides_reminder
    print_cluster_create_command
    ;;
  proof-ready)
    require_gke_plugin
    enable_apis
    ensure_artifact_registry
    configure_docker
    ensure_cluster_for_proof
    fetch_credentials
    kubectl get nodes
    ensure_namespace
    print_overrides_reminder
    print_proof_commands
    ;;
  *)
    die "unsupported mode '$MODE'; expected 'bootstrap' or 'proof-ready'"
    ;;
esac
