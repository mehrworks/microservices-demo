![Continuous Integration](https://github.com/GoogleCloudPlatform/microservices-demo/workflows/Continuous%20Integration%20-%20Main/Release/badge.svg)

This branch is a single-service GKE scaffold derived from the upstream
GoogleCloudPlatform `microservices-demo` repository.

The only active application in this branch is
[`productcatalogservice`](/src/productcatalogservice). The original multi-service
boutique app, frontend, load generator, and database-backed deployment options
have been removed from the active scaffold so this repository can serve as a
clean starting point for a single gRPC service on GKE.

## What remains

- [`src/productcatalogservice`](/src/productcatalogservice): Go gRPC service that
  serves the product catalog.
- [`skaffold.yaml`](/skaffold.yaml): local build and deploy entrypoint for the
  single service.
- [`kubernetes-manifests`](/kubernetes-manifests): Skaffold-oriented manifests
  with a local image reference.
- [`kustomize`](/kustomize): minimal Kustomize base for the same service.
- [`protos`](/protos): upstream protocol definitions retained for reference.

## Default runtime mode

`productcatalogservice` still defaults to file-backed local data. Unless
AlloyDB-related environment variables are explicitly set, the service loads
catalog data from [`src/productcatalogservice/products.json`](/src/productcatalogservice/products.json).

No frontend is included on this branch. No database integration is enabled by
default.

## Quickstart

Build and deploy with Skaffold:

```sh
skaffold run
```

Render or apply the Kustomize base directly:

```sh
kubectl kustomize kustomize
kubectl apply -k kustomize
```

Run the service tests locally:

```sh
go test ./...
```

From the service directory:

```sh
cd src/productcatalogservice
go test ./...
```

## Upstream origin

This scaffold started from the upstream
[`GoogleCloudPlatform/microservices-demo`](https://github.com/GoogleCloudPlatform/microservices-demo)
vendor repository and has been intentionally reduced to a single active service
path on branch `spike/single-service-scaffold`.
