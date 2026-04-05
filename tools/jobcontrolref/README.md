# Job Control Reference

This package is a tiny in-memory reference implementation for the hybrid
job-control contracts.

It is intentionally not a running service. Its purpose is to make the new
contract stack more concrete by modeling:

- job submission
- in-memory job persistence
- claim and lease behavior
- status updates
- result-access resolution

## Scope

- no HTTP server
- no database
- no queue
- no scheduler
- no framework choice

## Why it exists

The hybrid contract layer now includes:

- payload schemas and examples in `config/contracts/job-control/`
- an OpenAPI surface
- worker identity and lease semantics
- storage-boundary and module-boundary docs

This package is the first small executable reference that ties those ideas
together without forcing a real deployment decision.
