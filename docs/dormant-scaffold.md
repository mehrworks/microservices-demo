# Dormant scaffold guide

> Branch context: `spike/dormant-scaffold-v2`

This branch keeps the original multi-service repository shape intact as a **reference scaffold**, while narrowing the **active loop** to a thin public path: `frontend` in `catalog-only` mode plus `productcatalogservice`.

The goal is to preserve the old service bodies, deployment layout, CI history, and release knowledge so services can be reactivated or replaced gradually later, without prematurely rewriting the whole repo around one service.

## Current operating model

### Active right now

`frontend` and `productcatalogservice` are active in the current loop.

That currently means:

- `skaffold.yaml`
  - only `frontend` and `productcatalogservice` are active
  - other service artifacts remain in place as commented history
- `kubernetes-manifests/kustomization.yaml`
  - only `frontend.yaml` and `productcatalogservice.yaml` are active
- `kustomize/base/kustomization.yaml`
  - only `frontend.yaml` and `productcatalogservice.yaml` are active
- `release/kubernetes-manifests.yaml`
  - narrowed so the active public path stays synchronized in one release reference file, even though direct `kubectl apply -f` still needs image rewriting first
- `.github/workflows/ci-main.yaml`
  - only `frontend` and `productcatalogservice` are tested/deployed/waited on
  - older full-app flow is preserved as comments
- `.github/workflows/ci-pr.yaml`
  - same idea for PR flow

### Runtime policy

For the active services on this branch, runtime behavior is controlled at the **manifest layer**, not by rewriting vendor source semantics.

Current manifest policy:

- `frontend` runs with `FRONTEND_MODE=catalog-only`
- `frontend` keeps only `PRODUCT_CATALOG_SERVICE_ADDR` in the active env surface
- profiler is off via `DISABLE_PROFILER=1` for `productcatalogservice` and `ENABLE_PROFILER=0` for `frontend`
- tracing is off because `ENABLE_TRACING` is omitted
- AlloyDB remains dormant unless its environment is explicitly wired in later

## Dormant but intentionally retained

These surfaces are still present and still mostly describe the full original app:

- service directories under `src/`
- `helm-chart/`
- `release/`
- `cloudbuild.yaml`
- most of `docs/`
- `.github/release-cluster/`
- Kustomize components such as:
  - network policies
  - AlloyDB
  - service mesh / Istio
  - shopping assistant
  - without-loadgenerator

These are not mistakes. They are preserved on purpose as reference material and future replacement scaffolding.

## Reactivation rule

When reactivating an existing service or adding a new active one, update the active surfaces **together** in one coherent pass.

Minimum checklist:

1. `skaffold.yaml`
2. `kubernetes-manifests/kustomization.yaml`
3. `kustomize/base/kustomization.yaml`
4. `release/kubernetes-manifests.yaml`
5. `.github/workflows/ci-main.yaml`
6. `.github/workflows/ci-pr.yaml`
7. any service-specific waits / smoke tests / frontend assumptions
8. any branch notes that describe the active loop

Prefer **commenting services in or out** rather than deleting their surrounding history.

## What not to do

- Do not silently reactivate a service in only one file.
- Do not change vendor `src/` semantics just to enforce branch policy if the same behavior can live in manifests or CI.
- Do not delete the old service bodies unless there is a deliberate fork away from the vendor scaffold model.
- Do not assume full-app release docs describe the current active path.

## Bootstrap proof (GKE)

This branch keeps the original release manifest in sync with the active surface,
but the supported bootstrap proof path is the Skaffold-driven one:

- supported proof path: `skaffold run --default-repo=...`
- retained release reference: `./release/kubernetes-manifests.yaml` tracks the same resources, but its bare image names must be rewritten before direct `kubectl apply`
- fresh-account bootstrap entrypoint: `docs/gcp-bootstrap.md`
- staged infra contract lane: `infra/` with inputs under `config/datasets/`, kept separate from the current runtime proof flow

### Goal

Prove all of the following with the smallest possible scope:

1. the repo branch deploys into a GCP project
2. only `frontend` and `productcatalogservice` are active
3. the public frontend becomes reachable
4. one real gRPC request succeeds

This is intentionally **not** a full app deployment.

### What this proof includes

- one small GKE cluster
- one Artifact Registry repo
- two active services: `frontend` and `productcatalogservice`
- one public HTTP validation call to the catalog-only frontend
- one gRPC validation call: `ListProducts`

### What this proof does not include

- no DB / Cloud SQL
- no service mesh
- no multi-service restore
- no domain adaptation yet

### Assumptions

- local repo path: `/home/mpo/work/vendors/microservices-demo`
- branch: `spike/dormant-scaffold-v2`
- experimental GCP project only
- region example: `europe-west3`

### Prerequisites

- `gcloud`
- `kubectl`
- `gke-gcloud-auth-plugin` for GKE `kubectl` access
- `skaffold` **2.0.2+**
- Docker if you want the local-build path
- optional: `grpcurl` for the proof call

### Environment

If you are starting from a fresh GCP account or project, use `docs/gcp-bootstrap.md`
first. The steps below mirror that bootstrap path but stay here as the branch-local
execution reference.

Optional but recommended first step:

1. copy `config/gke-exposure/overrides.example.yaml` to `config/gke-exposure/overrides.yaml`
2. bind the real project/cluster/namespace/image values there
3. mirror those values into the shell exports below

The repo does not auto-load that YAML yet. It is the human-reviewed proof bind contract for this branch.

```bash
cd /home/mpo/work/vendors/microservices-demo
git checkout spike/dormant-scaffold-v2

export PROJECT_ID="<your-gcp-project-id>"
export REGION="europe-west3"
export CLUSTER="msdemo-public"
export AR_REPO="services"
export NAMESPACE="msdemo-public"

gcloud config set project "$PROJECT_ID"
```

### 1) Enable minimum APIs

```bash
gcloud services enable \
  container.googleapis.com \
  artifactregistry.googleapis.com \
  cloudbuild.googleapis.com \
  compute.googleapis.com
```

### 2) Create or reuse Artifact Registry

```bash
gcloud artifacts repositories describe "$AR_REPO" \
  --location "$REGION" >/dev/null 2>&1 || \
gcloud artifacts repositories create "$AR_REPO" \
  --repository-format=docker \
  --location="$REGION" \
  --description="Microservices demo POC images"
```

If building locally with Docker:

```bash
gcloud auth configure-docker "${REGION}-docker.pkg.dev" -q
```

### 3) Create a small GKE cluster

Autopilot is the simplest POC path.

```bash
gcloud container clusters describe "$CLUSTER" \
  --region "$REGION" >/dev/null 2>&1 || \
gcloud container clusters create-auto "$CLUSTER" \
  --region "$REGION"
```

Fetch credentials:

```bash
gcloud container clusters get-credentials "$CLUSTER" \
  --region "$REGION" \
  --project "$PROJECT_ID"
```

Optional sanity check:

```bash
kubectl get nodes
```

Create namespace:

```bash
kubectl create namespace "$NAMESPACE" \
  --dry-run=client -o yaml | kubectl apply -f -
```

Optional:

```bash
kubectl config set-context --current --namespace="$NAMESPACE"
```

### 4) Build and deploy the active surface

Use the Skaffold path for the working branch proof. It is the smallest accurate
flow for this branch because Skaffold handles image tagging and manifest image
rewrites together.

```bash
skaffold run \
  --namespace "$NAMESPACE" \
  --default-repo "${REGION}-docker.pkg.dev/${PROJECT_ID}/${AR_REPO}"
```

Notes:
- first build/deploy can take a while
- if local Docker/build is slow or problematic, use the `gcb` fallback below
- `release/kubernetes-manifests.yaml` remains the synchronized reference for the
  active surface, but direct `kubectl apply -f` needs manual image rewriting
  before it is usable on a clean GKE cluster

#### Fallback: Google Cloud Build

Use this if local Docker/build is problematic.

```bash
skaffold run -p gcb \
  --namespace "$NAMESPACE" \
  --default-repo "${REGION}-docker.pkg.dev/${PROJECT_ID}/${AR_REPO}"
```

Because this branch comments out the other services in the active loop, either
path should only build and deploy `frontend` plus `productcatalogservice`.

### 5) Wait for readiness

```bash
kubectl get all -n "$NAMESPACE"
```

```bash
kubectl wait \
  --for=condition=available \
  deployment/frontend \
  --timeout=600s \
  -n "$NAMESPACE"
```

```bash
kubectl wait \
  --for=condition=available \
  deployment/productcatalogservice \
  --timeout=600s \
  -n "$NAMESPACE"
```

Check logs:

```bash
kubectl logs deployment/frontend -n "$NAMESPACE" --tail=100
```

```bash
kubectl logs deployment/productcatalogservice -n "$NAMESPACE" --tail=100
```

### 6) Make one public HTTP proof call

Wait for the public `LoadBalancer` IP and fetch the catalog-only frontend:

```bash
kubectl get service frontend-external -n "$NAMESPACE"
```

Once an external IP appears, open `http://EXTERNAL_IP` in a browser, or use:

```bash
curl -fsS "http://EXTERNAL_IP/" | grep "Catalog-only mode is active"
```

### 7) Make one real gRPC proof call

This service is gRPC, so the simplest proof call uses `grpcurl`.

#### Terminal A: port-forward

```bash
kubectl port-forward svc/productcatalogservice 3550:3550 -n "$NAMESPACE"
```

#### Terminal B: call `ListProducts`

```bash
cd /home/mpo/work/vendors/microservices-demo
grpcurl \
  -plaintext \
  -import-path protos \
  -proto demo.proto \
  -d '{}' \
  localhost:3550 \
  hipstershop.ProductCatalogService/ListProducts
```

Optional second check:

```bash
grpcurl \
  -plaintext \
  -import-path protos \
  -proto demo.proto \
  -d '{"query":"socks"}' \
  localhost:3550 \
  hipstershop.ProductCatalogService/SearchProducts
```

### 8) If `grpcurl` is not installed

#### Option A: install it

```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

#### Option B: run it with Docker

```bash
docker run --rm --network host \
  -v "$PWD/protos:/protos" \
  fullstorydev/grpcurl \
  -plaintext \
  -import-path /protos \
  -proto demo.proto \
  -d '{}' \
  localhost:3550 \
  hipstershop.ProductCatalogService/ListProducts
```

### Success criteria

This proof is successful if:

- the cluster exists
- `frontend` becomes `Available`
- `productcatalogservice` becomes `Available`
- the public frontend is reachable
- logs look healthy
- `ListProducts` returns data

That is enough to prove the repo can bootstrap into GCP and kick-start a live app path.

### Cheap cleanup after the proof

If you deployed with `skaffold run`, delete the deployed resources first:

```bash
skaffold delete --namespace "$NAMESPACE"
```

Then delete the cluster if this was just a bootstrap proof:

```bash
gcloud container clusters delete "$CLUSTER" \
  --region "$REGION" \
  --project "$PROJECT_ID"
```

You can keep the Artifact Registry repo if you want to reuse it.

## Recommended next step

The next meaningful step should be app-directed, not more repo-wide cleanup:

- gradually adapt `productcatalogservice`, or
- add a new service beside it when there is a real use case

Try to keep future changes localized and intentional rather than repeating another large scaffold rewrite.
