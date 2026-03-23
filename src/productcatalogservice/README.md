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

## Data sources

By default, the service loads catalog data from `products.json`.

If `ALLOYDB_CLUSTER_NAME` is set, the service switches to AlloyDB-backed
loading and expects the related database environment variables to be present.
That integration remains optional and is not needed for scaffold use.

## Optional behavior flags

`EXTRA_LATENCY` injects a fixed delay into each request. The value must parse
as a Go `time.Duration`, for example `EXTRA_LATENCY=250ms`.

The service also keeps the upstream signal-driven catalog reload toggle:

- `SIGUSR1` enables reload-on-request behavior
- `SIGUSR2` disables it again

That mechanism is retained as part of the seed service behavior, but it is not
required for normal scaffold use.

## Proto bindings

The checked-in gRPC bindings in `genproto/` are generated from
`../../protos/demo.proto`. Regenerate them after editing the proto source:

```sh
./genproto.sh
```
