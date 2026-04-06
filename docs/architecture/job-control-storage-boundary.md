# Job Control Storage Boundary

This document defines the first persistence-facing boundary for the cloud-side
job-control module.

It does not choose a database. It defines the shapes and responsibilities the
eventual implementation should preserve regardless of storage backend.

## Why this exists

The job-control API and module specs already define:

- external payloads
- worker identity and lease semantics
- the cloud-side module boundary

The next missing layer is how those concepts are expected to look when persisted.

## Storage-facing artifacts

The first concrete persistence artifacts now live under
`config/contracts/job-control/`:

- `schema/job-store-record.schema.json`
- `schema/job-store-update.schema.json`
- `schema/result-access-record.schema.json`
- `job-store-record.example.json`
- `job-store-update.example.json`
- `result-access-record.example.json`

## First storage assumptions

### Job record persistence

Persisted job state should keep:

- the canonical job record
- an optimistic concurrency marker such as `revision`
- an external cache/invalidation token such as `etag`
- optional result-access metadata
- optional retention expiry

### Lease persistence

Lease state should be stored explicitly enough to answer:

- who currently owns the job
- when the lease was issued
- when the lease expires
- whether the worker should renew before expiry

### Result retrieval persistence

The cloud side should not assume the result body must be stored inline.

The first model assumes:

- the job record can hold a `result_ref`
- persistence may also track a `result_access` record that says whether the result
  is currently available and how it should be retrieved
- retrieval can stay reference-only at first

## First implementation rule

Do not bury claim/lease concurrency inside generic storage update calls without a
visible compare-and-swap or revision expectation.

That is why `job-store-update` includes `expected_revision` explicitly.

## Current implementation checkpoint

The first concrete persistence implementation is still intentionally small:

- inside `src/frontend/jobcontrol/`
- file-backed only when `JOB_CONTROL_STATE_PATH` is set
- otherwise in-memory

This keeps the persistence step reversible while the control-plane boundary is
still being proven.

## Non-goals

- no database selection yet
- no migration format yet
- no ORM or storage framework decision yet
- no requirement that result payloads live in the same store as job metadata
