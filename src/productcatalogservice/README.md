# productcatalogservice

`productcatalogservice` is the active seed service in this repository. It is a
small Go gRPC server that exposes a product catalog API, loads catalog data
from `products.json` by default, and can optionally read from AlloyDB when the
database-related environment variables are configured.

This scaffold is intentionally narrow:

- no frontend is included
- no database integration is required for local use
- no broader domain rename has been applied yet

## Local development

Run the tests:

```sh
go test ./...
```

Start the server:

```sh
go run .
```

The service listens on port `3550` by default. Set `PORT` to override it.
With no extra env vars, startup stays local-first: local file-backed catalog,
no trace exporter setup, and no Cloud Profiler startup attempt.

## Data sources

By default, the service loads catalog data from `products.json`.

If `ALLOYDB_CLUSTER_NAME` is set, the service switches to AlloyDB-backed
loading and expects these env vars to be present:

- `PROJECT_ID`
- `REGION`
- `ALLOYDB_CLUSTER_NAME`
- `ALLOYDB_INSTANCE_NAME`
- `ALLOYDB_DATABASE_NAME`
- `ALLOYDB_TABLE_NAME`
- `ALLOYDB_SECRET_NAME`

That integration remains optional and is not needed for scaffold use. If
`ALLOYDB_CLUSTER_NAME` is unset, the service does not attempt AlloyDB or Secret
Manager access.

## Optional behavior flags

`EXTRA_LATENCY` injects a fixed delay into each request. The value must parse
as a Go `time.Duration`, for example `EXTRA_LATENCY=250ms`.

Tracing is opt-in. Set `ENABLE_TRACING=1` to enable OTLP trace export, and set
`COLLECTOR_SERVICE_ADDR` at the same time so the exporter has a collector to
connect to.

Profiling is also opt-in. Set `ENABLE_PROFILER=1` to attempt Cloud Profiler
startup. If unset, profiler startup is skipped entirely.

The service also keeps the upstream signal-driven catalog reload toggle:

- `SIGUSR1` enables reload-on-request behavior
- `SIGUSR2` disables it again

That mechanism is retained as part of the seed service behavior, but it is not
required for normal scaffold use.

## Proto bindings

The checked-in gRPC bindings in `genproto/` are generated from
`../../protos/demo.proto` and use the
`github.com/mehrworks/microservices-demo/src/productcatalogservice/genproto`
package path. Regenerate them after editing the proto source:

```sh
./genproto.sh
```
