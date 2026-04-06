# frontend

Run the following command to restore dependencies to `vendor/` directory:

    dep ensure --vendor-only

## Catalog-only mode

This branch adds a thin public-edge mode for the frontend:

- `FRONTEND_MODE=full` keeps the original multi-service behavior
- `FRONTEND_MODE=catalog-only` reduces the required backend surface to
  `PRODUCT_CATALOG_SERVICE_ADDR`, while allowing `RECOMMENDATION_SERVICE_ADDR`
  as an optional browsing enhancement

In `catalog-only` mode, the frontend keeps product browsing active while cart,
checkout, ads, currency switching, and assistant flows stay disabled.
When `RECOMMENDATION_SERVICE_ADDR` is configured, product-page recommendations
remain available in that reduced browsing path.

## Internal job-control host

This module now also carries the first cloud-side implementation host for the
planned hybrid control plane under `src/frontend/jobcontrol/`.

Current state:

- the package exists and is tested
- a bearer-token-gated internal route layer exists for submit/get/cancel plus worker simulation endpoints
- no user-facing UI wiring exists yet
- no persistence backend exists beyond the in-memory store

Enable the internal operator routes by setting:

- `JOB_CONTROL_OPERATOR_TOKEN`

Enable worker-simulation endpoints by also setting:

- `JOB_CONTROL_WORKER_TOKEN`

When set, the frontend registers:

- `POST /internal/jobs`
- `GET /internal/jobs/{job_id}`
- `GET /internal/jobs/{job_id}/result`
- `POST /internal/jobs/{job_id}:cancel`

When `JOB_CONTROL_WORKER_TOKEN` is also set, the frontend additionally registers:

- `POST /internal/worker/claim`
- `POST /internal/jobs/{job_id}/status`

This is still only a manual/operator simulation path. It does not introduce a
real local worker process yet.

For a manual end-to-end simulation against a running frontend, use:

```bash
JOB_CONTROL_OPERATOR_TOKEN=... JOB_CONTROL_WORKER_TOKEN=... \
  ./docs/bootstrap/job-control-smoke.sh
```
