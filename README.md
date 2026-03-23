This branch is a single-service scaffold derived from the upstream
`GoogleCloudPlatform/microservices-demo` repository.

The only active application in this branch is
[`productcatalogservice`](/src/productcatalogservice). The original multi-service
application has been removed from the active scaffold so this repository can
serve as a clean starting point for a single Go gRPC service on Kubernetes.

## What remains

- [`src/productcatalogservice`](/src/productcatalogservice): Go gRPC service that
  serves the product catalog.
- [`skaffold.yaml`](/skaffold.yaml): local build and deploy entrypoint for the
  single service.
- [`kubernetes-manifests`](/kubernetes-manifests): Skaffold-oriented manifests
  with a local image reference.
- [`kustomize`](/kustomize): minimal Kustomize base for the same service.
- [`protos`](/protos): the proto source retained for the checked-in
  `productcatalogservice` gRPC bindings.

## Default runtime mode

`productcatalogservice` defaults to file-backed local data. Unless AlloyDB
environment variables are explicitly set, the service loads catalog data from
[`src/productcatalogservice/products.json`](/src/productcatalogservice/products.json).

No frontend is included on this branch. Database-backed catalog loading remains
optional and disabled by default.

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

Regenerate the checked-in gRPC bindings after editing [`protos/demo.proto`](/protos/demo.proto):

```sh
cd src/productcatalogservice
./genproto.sh
```

## Upstream origin

This scaffold started from the upstream
[`GoogleCloudPlatform/microservices-demo`](https://github.com/GoogleCloudPlatform/microservices-demo)
repository and has been intentionally reduced to a single active service path
on branch `spike/single-service-scaffold`.
