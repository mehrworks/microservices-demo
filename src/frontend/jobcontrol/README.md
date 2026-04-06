# Job Control Package

This package is the first intended cloud-side implementation host for the hybrid
job-control model.

It stays inside the existing frontend module for now so the cloud-side control
plane can evolve without forcing a new deployed service too early.

## Scope

- in-memory reference store
- claim/lease behavior
- status and result-access behavior
- implementation-facing interfaces for a later real control-plane path
- handler-facing controller for operator submit/get/cancel and worker claim/status simulation

## Non-goals

- no database integration yet
- no new runtime deployment yet
- no user-facing UI integration yet

## Relationship to other artifacts

- payload/API/storage contracts: `config/contracts/job-control/`
- architecture docs: `docs/architecture/job-control-*.md`
- executable wrapper/reference module: `tools/jobcontrolref/`
