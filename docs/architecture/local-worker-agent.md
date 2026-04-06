# Local Worker Agent

This document defines the first intended shape of the trusted local worker in the
hybrid model.

## Default worker model

The first worker model is intentionally small:

- one trusted worker agent on one machine
- authenticated with a bearer-style identity
- claims work pull-style from the cloud side
- holds a short-lived lease while executing
- reports progress and terminal state back to the control plane

## Identity assumptions

The worker must present a stable identity with at least:

- `worker_id`
- `auth_mode`
- `auth_subject`

Optional but useful identity fields include:

- `auth_issuer`
- `machine_name`
- `worker_version`
- accepted job types and max lease preference

The first concrete identity artifact is:

- `config/contracts/job-control/schema/worker-identity.schema.json`

The first real worker process now lives in:

- `tools/jobcontrolworker/`

That worker now sends explicit worker identity in its claim request, rather than
relying on an implicit anonymous worker shape.

## Lease assumptions

Claimed work is not assumed to be owned forever.

The cloud side grants a lease with:

- `lease_id`
- `issued_at`
- `lease_expires_at`
- `lease_duration_seconds`
- optional `renew_after_seconds`

The worker is expected to treat the lease as authoritative.

The next operational behavior after simple claim/status support is:

- renew the lease before expiry when work is still running
- allow the cloud side to requeue work when the lease expires without renewal

## Behavioral rule of thumb

- do not start work without a valid lease
- do not keep assuming ownership after lease expiry
- do not hide worker identity behind anonymous machine behavior

These rules are meant to keep the first control plane understandable before any
queue, scheduler, or multi-worker complexity exists.

The matching cloud-side implementation boundary is described in
`docs/architecture/job-control-module-spec.md`.

The worker process currently talks to the internal frontend-hosted job-control API.

That API is expected to become stricter over time around lease renewal and stale
worker recovery before any long-running daemon behavior is added.
