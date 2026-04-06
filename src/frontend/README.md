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
- optional file-backed persistence now exists via `JOB_CONTROL_STATE_PATH`

Enable the internal operator routes by setting:

- `JOB_CONTROL_OPERATOR_TOKEN`

Enable worker-simulation endpoints by also setting:

- `JOB_CONTROL_WORKER_TOKEN`

Optional worker identity settings for the internal API host:

- `JOB_CONTROL_WORKER_ID`
- `JOB_CONTROL_WORKER_SUBJECT`
- `JOB_CONTROL_WORKER_ISSUER`
- `JOB_CONTROL_WORKER_MACHINE_NAME`
- `JOB_CONTROL_WORKER_VERSION`
- `JOB_CONTROL_WORKER_MAX_LEASE_SECONDS`

Enable file-backed state persistence by also setting:

- `JOB_CONTROL_STATE_PATH`

When set, the frontend registers:

- `POST /internal/jobs`
- `GET /internal/jobs/{job_id}`
- `GET /internal/jobs/{job_id}/result`
- `POST /internal/jobs/{job_id}:cancel`

When `JOB_CONTROL_WORKER_TOKEN` is also set, the frontend additionally registers:

- `POST /internal/worker/claim`
- `POST /internal/jobs/{job_id}/lease:renew`
- `POST /internal/jobs/{job_id}/status`

This route layer still supports a manual/operator simulation path. A separate
real worker process now exists in `tools/jobcontrolworker/`, but these routes can
still be exercised manually for debugging.

For a manual end-to-end simulation against a running frontend, use:

```bash
JOB_CONTROL_OPERATOR_TOKEN=... JOB_CONTROL_WORKER_TOKEN=... \
  ./docs/bootstrap/job-control-smoke.sh
```
