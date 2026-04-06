# Job Lifecycle Contract

This document defines the first intended contract for slow or heavy work once the
repo pivots from boutique-service expansion to the hybrid control-plane model.

## Design intent

Keep the first lifecycle model small enough to implement without committing to a
message queue or scheduler too early.

## Core states

- `queued`
- `running`
- `succeeded`
- `failed`
- `cancelled`

Optional later extension states such as `retrying` or `stale` can be added later
if the simpler model proves too weak.

## Minimal record shape

The first contract should be able to represent at least:

- `job_id`
- `job_type`
- `submitted_at`
- `updated_at`
- `state`
- `progress_percent` or equivalent coarse progress field
- `status_message`
- `result_ref` or `null`
- `error_code` or `null`

The current concrete companion artifacts for this lifecycle live in
`config/contracts/job-control/`, especially:

- `job-record.example.json`
- `status-update.example.json`
- `schema/job-record.schema.json`
- `schema/status-update.schema.json`

## Transition rules

- `queued -> running`
  - when the local worker claims the job
- `running -> succeeded`
  - when the worker completes and publishes a result reference
- `running -> failed`
  - when execution ends unsuccessfully
- `queued -> cancelled`
  - when cancelled before claim
- `running -> cancelled`
  - when cancellation is accepted during execution

Disallowed by default:

- `succeeded -> running`
- `failed -> running`
- `cancelled -> running`

Any retry should create a new attempt or explicit retry action rather than
silently reusing the same state transition semantics.

## First protocol assumption

Use pull + polling first:

- cloud side exposes claimable queued work
- local worker polls for jobs
- local worker reports progress and final state back
- frontend or operator view polls for updated status

This keeps the initial contract understandable before introducing queue-specific
ack/nack behavior.

The first concrete API shape for this lifecycle is described in
`docs/architecture/job-control-api-surface.md`.

## Failure assumptions to define before implementation

- what marks a running job as abandoned
- whether claim ownership expires after a timeout
- whether retry creates a new job id or a new attempt on the same job id
- whether partial results are allowed
- how cancellation is represented when the worker cannot stop immediately

Current implementation direction for stale-worker recovery:

- lease expiry returns the running job to `queued`
- the next claim increments `attempt`
- lease renewal is expected before long-running work crosses the renewal threshold

The first explicit worker/lease companion doc is:

- `docs/architecture/local-worker-agent.md`

The first persistence-facing companion doc is:

- `docs/architecture/job-control-storage-boundary.md`
