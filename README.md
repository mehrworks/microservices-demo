This branch is a single-service scaffold derived from the upstream
Google microservices-demo repository and now carries the
`github.com/mehrworks/microservices-demo` identity for the active scaffolded
service module and generated protobuf bindings.

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

Observability hooks are also local-first by default:

- tracing is off unless `ENABLE_TRACING=1`
- profiler is off unless `ENABLE_PROFILER=1`
- `COLLECTOR_SERVICE_ADDR` is only required when tracing is enabled

If those env vars are unset, the service starts without attempting cloud trace
export, Cloud Profiler startup, or extra gRPC OTel instrumentation.

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

## Runtime modes

Local file-backed mode is the default and requires no cloud-specific env vars:

```sh
cd src/productcatalogservice
go run .
```

Optional cloud hooks can be enabled explicitly:

```sh
cd src/productcatalogservice
ENABLE_TRACING=1 COLLECTOR_SERVICE_ADDR=otel-collector:4317 go run .
ENABLE_PROFILER=1 go run .
```

Optional AlloyDB-backed catalog loading is selected only when
`ALLOYDB_CLUSTER_NAME` is set. In that mode the service expects:

- `PROJECT_ID`
- `REGION`
- `ALLOYDB_CLUSTER_NAME`
- `ALLOYDB_INSTANCE_NAME`
- `ALLOYDB_DATABASE_NAME`
- `ALLOYDB_TABLE_NAME`
- `ALLOYDB_SECRET_NAME`

Without `ALLOYDB_CLUSTER_NAME`, none of the AlloyDB or Secret Manager setup is used.

## Upstream origin

This scaffold started from the upstream Google microservices-demo repository
and has been intentionally reduced to a single active service path
on branch `spike/single-service-scaffold`. The active Go module path for that
service is `github.com/mehrworks/microservices-demo/src/productcatalogservice`.
