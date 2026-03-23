# productcatalogservice

Run the following command to restore dependencies to `vendor/` directory:

    go mod vendor

## Runtime defaults

By default, this service starts in the lowest-friction local mode:

- Catalog data is loaded from `products.json`.
- Tracing is off.
- Profiler is off.
- AlloyDB hooks stay dormant unless the AlloyDB environment is configured.

Opt in to the runtime hooks with environment variables:

- `ENABLE_TRACING=1` enables OpenTelemetry tracing and the OTLP gRPC exporter. `COLLECTOR_SERVICE_ADDR` must also be set in that mode.
- `ENABLE_PROFILER=1` enables the Cloud Profiler startup path.
- `ALLOYDB_CLUSTER_NAME` switches catalog loading from `products.json` to AlloyDB. In that mode, the service also expects `PROJECT_ID`, `REGION`, `ALLOYDB_INSTANCE_NAME`, `ALLOYDB_DATABASE_NAME`, `ALLOYDB_TABLE_NAME`, and `ALLOYDB_SECRET_NAME`.
- `EXTRA_LATENCY=<time.Duration>` injects latency on every request.

If none of those opt-in variables are set, the service remains file-backed and runs without tracing or profiling.

## Dynamic catalog reloading / artificial delay

This service has a "dynamic catalog reloading" feature that is purposefully
not well implemented. The goal of this feature is to allow you to modify the
`products.json` file and have the changes be picked up without having to
restart the service.

However, this feature is bugged: the catalog is actually reloaded on each
request, introducing a noticeable delay in the frontend. This delay will also
show up in profiling tools: the `parseCatalog` function will take more than 80%
of the CPU time.

You can trigger this feature (and the delay) by sending a `USR1` signal and
remove it (if needed) by sending a `USR2` signal:

```
# Trigger bug
kubectl exec \
    $(kubectl get pods -l app=productcatalogservice -o jsonpath='{.items[0].metadata.name}') \
    -c server -- kill -USR1 1
# Remove bug
kubectl exec \
    $(kubectl get pods -l app=productcatalogservice -o jsonpath='{.items[0].metadata.name}') \
    -c server -- kill -USR2 1
```

## Latency injection

This service has an `EXTRA_LATENCY` environment variable. This will inject a sleep for the specified [time.Duration](https://golang.org/pkg/time/#ParseDuration) on every call to
to the server.

For example, use `EXTRA_LATENCY="5.5s"` to sleep for 5.5 seconds on every request.
