# Dormant scaffold guide

> Branch context: `spike/dormant-scaffold-v2`

This branch keeps the original multi-service repository shape intact as a **reference scaffold**, while narrowing the **active loop** to `productcatalogservice` only.

The goal is to preserve the old service bodies, deployment layout, CI history, and release knowledge so services can be reactivated or replaced gradually later, without prematurely rewriting the whole repo around one service.

## Current operating model

### Active right now

Only `productcatalogservice` is active in the current loop.

That currently means:

- `skaffold.yaml`
  - only `productcatalogservice` is active
  - other service artifacts remain in place as commented history
- `kubernetes-manifests/kustomization.yaml`
  - only `productcatalogservice.yaml` is active
- `kustomize/base/kustomization.yaml`
  - only `productcatalogservice.yaml` is active
- `release/kubernetes-manifests.yaml`
  - narrowed so the upstream-style `kubectl apply -f ./release/kubernetes-manifests.yaml` path still works for the active service only
- `.github/workflows/ci-main.yaml`
  - only `productcatalogservice` is tested/deployed/waited on
  - older full-app flow is preserved as comments
- `.github/workflows/ci-pr.yaml`
  - same idea for PR flow

### Runtime policy

For the active service on this branch, runtime behavior is controlled at the **manifest layer**, not by rewriting vendor source semantics.

Current manifest policy:

- profiler is off via `DISABLE_PROFILER=1`
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

This branch supports both familiar upstream entry styles, narrowed to the
active service only:

- upstream-style release path: `kubectl apply -f ./release/kubernetes-manifests.yaml`
- upstream-style dev path: `skaffold run --default-repo=...`

### Goal

Prove all of the following with the smallest possible scope:

1. the repo branch deploys into a GCP project
2. only `productcatalogservice` is active
3. the service becomes reachable
4. one real gRPC request succeeds

This is intentionally **not** a full app deployment.

### What this proof includes

- one small GKE cluster
- one Artifact Registry repo
- one active service: `productcatalogservice`
- one validation call: `ListProducts`

### What this proof does not include

- no frontend
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
- `skaffold` **2.0.2+**
- Docker if you want the local-build path
- optional: `grpcurl` for the proof call

### Environment

```bash
cd /home/mpo/work/vendors/microservices-demo
git checkout spike/dormant-scaffold-v2

export PROJECT_ID="<your-gcp-project-id>"
export REGION="europe-west3"
export CLUSTER="msdemo-poc"
export AR_REPO="services"
export NAMESPACE="msdemo-poc"

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

### 4) Build and deploy the active service

You have two valid entry styles on this branch.

#### Option A — upstream-style release path

If you want to stay closest to the upstream README shape, first build and push
`productcatalogservice`, then apply the narrowed release manifest.

Build and push the image:

```bash
skaffold build \
  --default-repo "${REGION}-docker.pkg.dev/${PROJECT_ID}/${AR_REPO}"
```

Then deploy with the upstream-style command:

```bash
kubectl apply -f ./release/kubernetes-manifests.yaml -n "$NAMESPACE"
```

#### Option B — upstream-style dev path via Skaffold run

Preferred if you want Skaffold to handle build + deploy together.

```bash
skaffold run \
  --namespace "$NAMESPACE" \
  --default-repo "${REGION}-docker.pkg.dev/${PROJECT_ID}/${AR_REPO}"
```

Notes:
- first build/deploy can take a while
- if local Docker/build is slow or problematic, use the `gcb` fallback below

#### Fallback: Google Cloud Build

Use this if local Docker/build is problematic.

```bash
skaffold run -p gcb \
  --namespace "$NAMESPACE" \
  --default-repo "${REGION}-docker.pkg.dev/${PROJECT_ID}/${AR_REPO}"
```

Because this branch comments out the other services in the active loop, either
path should only build and deploy `productcatalogservice`.

### 5) Wait for readiness

```bash
kubectl get all -n "$NAMESPACE"
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
kubectl logs deployment/productcatalogservice -n "$NAMESPACE" --tail=100
```

### 6) Make one real proof call

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

### 7) If `grpcurl` is not installed

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
- `productcatalogservice` becomes `Available`
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
