# GKE bootstrap POC runbook

> Branch context: `spike/dormant-scaffold-v2`

This runbook proves the current dormant-scaffold branch can bootstrap into a GCP project and bring up a single active service.

## Goal

Prove all of the following with the smallest possible scope:

1. the repo branch deploys into a GCP project
2. only `productcatalogservice` is active
3. the service becomes reachable
4. one real gRPC request succeeds

This is intentionally **not** a full app deployment.

## What this POC includes

- one small GKE cluster
- one Artifact Registry repo
- one active service: `productcatalogservice`
- one validation call: `ListProducts`

## What this POC does not include

- no frontend
- no DB / Cloud SQL
- no service mesh
- no multi-service restore
- no domain adaptation yet

## Assumptions

- local repo path: `/home/mpo/work/vendors/microservices-demo`
- branch: `spike/dormant-scaffold-v2`
- experimental GCP project only
- region example: `europe-west3`

## Environment

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

## 1) Enable minimum APIs

```bash
gcloud services enable \
  container.googleapis.com \
  artifactregistry.googleapis.com \
  cloudbuild.googleapis.com \
  compute.googleapis.com
```

## 2) Create or reuse Artifact Registry

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

## 3) Create a small GKE cluster

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

Create namespace:

```bash
kubectl create namespace "$NAMESPACE" \
  --dry-run=client -o yaml | kubectl apply -f -
```

Optional:

```bash
kubectl config set-context --current --namespace="$NAMESPACE"
```

## 4) Deploy only the active service via Skaffold

### Preferred: local build

```bash
skaffold run \
  --namespace "$NAMESPACE" \
  --default-repo "${REGION}-docker.pkg.dev/${PROJECT_ID}/${AR_REPO}"
```

### Fallback: Google Cloud Build

Use this if local Docker/build is problematic.

```bash
skaffold run -p gcb \
  --namespace "$NAMESPACE" \
  --default-repo "${REGION}-docker.pkg.dev/${PROJECT_ID}/${AR_REPO}"
```

Because this branch comments out the other services in the active loop, this should only build and deploy `productcatalogservice`.

## 5) Wait for readiness

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

## 6) Make one real proof call

This service is gRPC, so the simplest proof call uses `grpcurl`.

### Terminal A: port-forward

```bash
kubectl port-forward svc/productcatalogservice 3550:3550 -n "$NAMESPACE"
```

### Terminal B: call `ListProducts`

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

## 7) If `grpcurl` is not installed

### Option A: install it

```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

### Option B: run it with Docker

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

## Success criteria

This POC is successful if:

- the cluster exists
- `productcatalogservice` becomes `Available`
- logs look healthy
- `ListProducts` returns data

That is enough to prove the repo can bootstrap into GCP and kick-start a live app path.

## Cheap cleanup after the proof

Delete the cluster if this is just a bootstrap proof:

```bash
gcloud container clusters delete "$CLUSTER" \
  --region "$REGION" \
  --project "$PROJECT_ID"
```

You can keep the Artifact Registry repo if you want to reuse it.

## Notes

- This runbook intentionally avoids frontend, DB, and mesh.
- If this works, the next step is **not** more scaffold hardening; it is either:
  - adapting `productcatalogservice`, or
  - adding your first adjacent service with a real use case.
