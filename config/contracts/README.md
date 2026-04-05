# Contract Artifacts

This directory holds repo-wide, environment-independent contract artifacts.

Unlike `config/datasets/`, these files do not bind a specific GCP project,
cluster, or profile. They describe stable payload shapes and shared boundaries the
repo intends to implement.

## Current contracts

- `job-control/`
  - minimal cloud-side control-plane contract for the planned hybrid model
  - covers job submission, worker claim, status update, job record, result
    reference, error envelope, and a minimal API surface

## Relationship to the rest of `config/`

- `config/datasets/` = environment/stage inputs
- `config/profiles/` = grouped environment bundles
- `config/gke-exposure/` = active runtime proof bind
- `config/contracts/` = canonical shared payload contracts for future implementation

These artifacts are planning and interface assets first. They do not create a new
runtime by themselves.
