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
- cancel response
- result reference

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
- `schema/cancel-job-response.schema.json`
- `schema/result-reference.schema.json`
- `*.example.json` companions for each shape

## Design intent

- cloud side owns submission and status visibility
- local worker owns heavy execution
- both sides share a stable lifecycle/result contract

For the higher-level rationale, use:

- `docs/architecture/hybrid-execution-model.md`
- `docs/architecture/runtime-responsibility-matrix.md`
- `docs/architecture/job-lifecycle-contract.md`
