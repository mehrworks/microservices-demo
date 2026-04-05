# Runtime Responsibility Matrix

This document assigns responsibilities between the cloud side, the local worker,
and the shared contract surface for the planned hybrid model.

## Default model

- cloud side stays on the current GKE baseline
- one trusted local worker agent runs on one machine
- shared contracts stay simple and explicit

## Matrix

| Concern | Cloud side | Local worker | Shared contract |
| --- | --- | --- | --- |
| public UI | own | none | none |
| job submission | own | none | request shape |
| job pickup | expose pull endpoint | poll / claim work | claim semantics |
| slow execution | none | own | state transitions |
| progress updates | receive and display | emit | progress schema |
| result materialization | serve references / retrieval | produce results | result schema |
| retry / cancellation | own control rules | cooperate during execution | status + retry/cancel rules |
| machine-local tools / dependencies | none | own | capability declaration only if needed later |

## Practical interpretation

### Cloud side owns

- user-facing frontend or thin control API
- creation of job records
- user-visible status and result retrieval
- operator-visible control actions such as cancel or retry

The first minimal API shape for that cloud-side control plane is described in
`docs/architecture/job-control-api-surface.md`.

### Local worker owns

- actual heavy or slow execution
- machine-local dependencies
- progress and terminal-state reporting
- graceful handling of retry/cancel instructions

The first worker identity and lease assumptions are described in
`docs/architecture/local-worker-agent.md`.

### Shared contract owns

- job identifiers
- lifecycle states
- progress schema
- result schema
- idempotency and retry expectations

The first concrete artifacts for those shared payloads now live in
`config/contracts/job-control/`.

## Current constraint

This is a planning contract. It does not yet require a queue, Pub/Sub, or any
specific execution framework.
