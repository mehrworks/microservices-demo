# Fresh GCP Bootstrap

Use this path when you are starting the dormant proof flow from a brand-new GCP
project or account and do not want to rediscover the required APIs, registry
shape, cluster defaults, auth expectations, or proof commands.

This document still describes the supported runtime proof path for the branch.
The staged `infra/` lane is being introduced separately as a thin cloud-contract
layer and does not replace the Skaffold proof flow yet.

This branch's active proof scope is intentionally small:

- public HTTP edge: `frontend` in `catalog-only` mode
- backing browsing services: `productcatalogservice` and `recommendationservice`
- no DB, mesh, or private-network requirement in the baseline proof

## Branch defaults

- `REGION=europe-west3`
- `AR_REPO=services`
- `CLUSTER=msdemo-public`
- `NAMESPACE=msdemo-public`
- `DEFAULT_REPO=${REGION}-docker.pkg.dev/${PROJECT_ID}/${AR_REPO}`

## Local prerequisites

Required:

- `gcloud`
- `kubectl`
- `skaffold` 2.x
- Docker

Also required for `kubectl` access to GKE:

- `gke-gcloud-auth-plugin`

Common ways to install the auth plugin:

```bash
gcloud components install gke-gcloud-auth-plugin
```

If your `gcloud` installation uses Debian/Ubuntu packages instead of the
component manager:

```bash
sudo apt-get update
sudo apt-get install -y google-cloud-sdk-gke-gcloud-auth-plugin
```

Optional:

- `grpcurl` for the gRPC proof call
- if `grpcurl` is missing, use the Docker fallback in `docs/dormant-scaffold.md`

## One-time fresh-project bootstrap

1. Set the proof env vars.

```bash
export PROJECT_ID="<your-project-id>"
export REGION="europe-west3"
export AR_REPO="services"
export CLUSTER="msdemo-public"
export NAMESPACE="msdemo-public"
export DEFAULT_REPO="${REGION}-docker.pkg.dev/${PROJECT_ID}/${AR_REPO}"
```

2. Run the helper script in bootstrap mode.

```bash
PROJECT_ID="$PROJECT_ID" \
REGION="$REGION" \
AR_REPO="$AR_REPO" \
CLUSTER="$CLUSTER" \
NAMESPACE="$NAMESPACE" \
DEFAULT_REPO="$DEFAULT_REPO" \
./docs/bootstrap/bootstrap-gke-proof.sh bootstrap
```

This mode:

- validates local tooling
- shows the active gcloud account and project
- enables the required APIs
- creates or reuses the `services` Artifact Registry repo
- configures Docker auth for `${REGION}-docker.pkg.dev`
- prints the next manual cluster step

## Long wait: create the cluster

Cluster creation is intentionally kept explicit.

```bash
gcloud container clusters describe "$CLUSTER" \
  --project "$PROJECT_ID" \
  --region "$REGION" >/dev/null 2>&1 || \
gcloud container clusters create-auto "$CLUSTER" \
  --project "$PROJECT_ID" \
  --region "$REGION"
```

Then fetch credentials:

```bash
gcloud container clusters get-credentials "$CLUSTER" \
  --project "$PROJECT_ID" \
  --region "$REGION"
```

## Bind the proof config

Copy the example file and fill the project-specific values:

```bash
cp config/gke-exposure/overrides.example.yaml \
  config/gke-exposure/overrides.yaml
```

Set:

- `project_id`
- `default_repo`
- `frontend_public_base_url` after the load balancer IP exists

The branch defaults for region, repo, cluster, namespace, and service names are
already encoded in the example file.

If staged outputs already exist, you can render this local proof bind from them
instead:

```bash
./docs/bootstrap/render-gke-proof-bind.sh --workspace default
```

## Prepare the cluster for the proof

Once the cluster exists and credentials work, run:

```bash
PROJECT_ID="$PROJECT_ID" \
REGION="$REGION" \
AR_REPO="$AR_REPO" \
CLUSTER="$CLUSTER" \
NAMESPACE="$NAMESPACE" \
DEFAULT_REPO="$DEFAULT_REPO" \
./docs/bootstrap/bootstrap-gke-proof.sh proof-ready
```

This mode:

- verifies the cluster exists
- fetches credentials
- creates or reuses the namespace
- sets the current kubectl namespace
- prints the exact deploy and proof commands

If the cluster is missing, the helper fails by default and prints the exact
`gcloud container clusters create-auto` command instead of creating it
automatically.

## Deploy the active proof surface

```bash
skaffold run \
  --namespace "$NAMESPACE" \
  --default-repo "$DEFAULT_REPO"
```

## Verify the proof

The exact HTTP verification path depends on `proof.http_mode` in the rendered
`config/gke-exposure/overrides.yaml` bind.

Wait for the active deployments:

```bash
kubectl wait \
  --for=condition=available \
  deployment/frontend \
  deployment/productcatalogservice \
  deployment/recommendationservice \
  --timeout=600s \
  -n "$NAMESPACE"
```

For `proof.http_mode: public-load-balancer`, wait for the public IP:

```bash
kubectl get service frontend-external -n "$NAMESPACE"
```

Then run the public HTTP proof:

```bash
curl -fsS "http://EXTERNAL_IP/" | grep "Catalog-only mode is active"
```

Then confirm the recommendation surface on a product page:

```bash
curl -fsS "http://EXTERNAL_IP/product/OLJCESPC7Z" | grep "You May Also Like"
```

For `proof.http_mode: port-forward`, use the internal HTTP proof instead:

```bash
kubectl port-forward svc/frontend 8081:80 -n "$NAMESPACE"
curl -fsS "http://127.0.0.1:8081/" | grep "Catalog-only mode is active"
curl -fsS "http://127.0.0.1:8081/product/OLJCESPC7Z" | grep "You May Also Like"
```

gRPC proof:

```bash
kubectl port-forward svc/productcatalogservice 3550:3550 -n "$NAMESPACE"
grpcurl \
  -plaintext \
  -import-path protos \
  -proto demo.proto \
  -d '{}' \
  localhost:3550 \
  hipstershop.ProductCatalogService/ListProducts
```

## Long waits and resume points

- API enablement can take a short time to propagate
- `gcloud container clusters create-auto` is the main human-wait step
- `skaffold run` is the build/push/deploy wait
- `frontend-external` IP allocation may take extra time after the pods are ready

If you stop at any of those boundaries, resume from the next section rather than
redoing the entire flow.
