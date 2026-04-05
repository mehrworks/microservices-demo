# Job Control Contract

This directory holds the first minimal cloud-side contract for the planned hybrid
control-plane model.

It is intentionally small and does not imply a queue or scheduler implementation.

## Scope

Current contract artifacts cover:

- submission request
- submission response
- worker identity
- lease
- worker claim request
- worker claim response
- worker status update
- job record
- job store record
- job store update
- result access record
- cancel response
- result reference
- error envelope

- `job-control.openapi.json`
  - minimal cloud-side API surface that stitches the payloads together

## Files

- `schema/submit-job-request.schema.json`
- `schema/submit-job-response.schema.json`
- `schema/worker-identity.schema.json`
- `schema/lease.schema.json`
- `schema/claim-next-request.schema.json`
- `schema/claim-next-response.schema.json`
- `schema/status-update.schema.json`
- `schema/job-record.schema.json`
- `schema/job-store-record.schema.json`
- `schema/job-store-update.schema.json`
- `schema/result-access-record.schema.json`
- `schema/cancel-job-response.schema.json`
- `schema/result-reference.schema.json`
- `schema/error-envelope.schema.json`
- `*.example.json` companions for each shape

## Design intent

- cloud side owns submission and status visibility
- local worker owns heavy execution
- both sides share a stable lifecycle/result contract

For the higher-level rationale, use:

- `docs/architecture/hybrid-execution-model.md`
- `docs/architecture/runtime-responsibility-matrix.md`
- `docs/architecture/job-lifecycle-contract.md`
- `docs/architecture/job-control-storage-boundary.md`

For the first executable reference implementation, use:

- `tools/jobcontrolref/`
