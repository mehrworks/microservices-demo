# Job Control API Surface

This document turns the hybrid job-control planning contracts into a minimal
cloud-side interface shape without introducing a new runtime service yet.

## Why this exists

The repo now has:

- payload-level job-control schemas in `config/contracts/job-control/schema/`
- example payloads in `config/contracts/job-control/*.example.json`

The next missing layer was the API surface that connects those payloads into a
 cloud-side control-plane interface.

That interface now lives in:

- `config/contracts/job-control/job-control.openapi.json`

## Endpoints

The first minimal interface includes five endpoints:

- `POST /v1/jobs`
  - submit a new job
- `GET /v1/jobs/{job_id}`
  - read current job state
- `POST /v1/jobs/{job_id}:cancel`
  - request cancellation
- `POST /v1/worker/claim`
  - allow the local worker to claim work pull-style
- `POST /v1/jobs/{job_id}/status`
  - allow the local worker to report progress and terminal state

## Why this shape

- it matches the current planning rule of pull + polling first
- it avoids queue-specific semantics in the first cut
- it keeps the worker contract explicit without requiring a new runtime platform

## Auth and lease assumptions

- operator-facing endpoints use a generic bearer-style operator auth scheme
- worker-facing endpoints use a separate bearer-style worker auth scheme
- worker claim carries structured worker identity rather than only a bare string id
- claimed work now carries an explicit lease object rather than only an expiry timestamp

This keeps worker identity and claim ownership explicit before any implementation
exists.

## Current status

This is an interface artifact only.

- no new service implementation exists yet
- no frontend wiring is implied yet
- no queue, scheduler, or Cloud Run path is introduced by this document alone

## Companion artifacts

- `docs/architecture/hybrid-execution-model.md`
- `docs/architecture/runtime-responsibility-matrix.md`
- `docs/architecture/job-lifecycle-contract.md`
- `docs/architecture/local-worker-agent.md`
- `config/contracts/job-control/README.md`
