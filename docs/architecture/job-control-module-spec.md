# Job Control Module Spec

This document defines the first implementation-facing cloud-side module boundary
for the hybrid job-control surface.

It still does not introduce a new running service. It describes the internal
module responsibilities the eventual cloud-side implementation should keep.

The first concrete implementation host for this module boundary now lives in:

- `src/frontend/jobcontrol/`

The first handler-facing adapter in that host is:

- `src/frontend/jobcontrol/controller.go`

The first dormant internal route host that uses that adapter lives in:

- `src/frontend/jobcontrol_http.go`

## Goal

Make the future job-control implementation small, explicit, and modular before
any runtime is chosen.

## Module boundary

The first implementation should be decomposed into these responsibilities:

1. HTTP/API layer
   - translate HTTP requests and responses
   - bind to the payload contracts in `config/contracts/job-control/`
   - avoid embedding claim/lease or persistence rules directly in handlers

In the current first cut, `Controller` is the handler-facing adapter that verifies
operator identity and delegates submit/get/cancel behavior into the underlying service.

2. Auth layer
   - validate operator bearer auth for user-facing endpoints
   - validate worker bearer auth for worker-facing endpoints
   - expose a small trusted identity shape to the rest of the module

3. Job store
   - persist job records
   - support read/update/list-by-state behavior needed by the minimal control API
   - keep storage implementation abstract

4. Claim and lease service
   - decide which queued job can be claimed next
   - issue and validate leases
   - own lease expiry and renew-after semantics

5. Status/result updater
   - accept worker status updates
   - merge terminal states and progress into the job record
   - attach `result_ref` when work succeeds

## Suggested internal interfaces

These are conceptual interfaces, not a required language binding yet.

### `AuthVerifier`

- `VerifyOperator(request) -> operator identity`
- `VerifyWorker(request) -> worker identity`

### `JobStore`

- `CreateJob(submit_request) -> job_record`
- `GetJob(job_id) -> job_record`
- `UpdateJob(job_id, patch) -> job_record`
- `ListQueuedJobs(filters) -> []job_record`

The persistence-facing contract for those operations is defined in
`docs/architecture/job-control-storage-boundary.md`.

### `LeaseManager`

- `IssueLease(job_id, worker_identity, requested_seconds) -> lease`
- `ValidateLease(job_id, worker_identity, lease_id) -> ok/error`
- `ExpireLease(job_id, lease_id) -> updated job record`

### `ClaimService`

- `ClaimNext(worker_identity) -> claim_next_response`

### `StatusService`

- `ApplyStatusUpdate(job_id, worker_identity, status_update) -> job_record`
- `RequestCancel(job_id, operator_identity) -> cancel response`

## Design rules

- handlers should depend on interfaces, not storage details
- auth should not be mixed with lease logic
- lease issuance should not be hidden inside generic store update calls
- the first implementation should support one trusted worker well before any
  multi-worker sophistication is added

## Non-goals

- no framework choice yet
- no database choice yet
- no queue selection yet
- no background scheduler selection yet
- no requirement to split this into a separate deployment immediately

## Relationship to existing artifacts

- API contract: `config/contracts/job-control/job-control.openapi.json`
- payload schemas: `config/contracts/job-control/schema/*.json`
- lifecycle rules: `docs/architecture/job-lifecycle-contract.md`
- worker/lease assumptions: `docs/architecture/local-worker-agent.md`
- storage boundary: `docs/architecture/job-control-storage-boundary.md`
- cloud-side package host: `src/frontend/jobcontrol/`
- executable reference package: `tools/jobcontrolref/`
