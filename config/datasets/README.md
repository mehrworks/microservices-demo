# Dataset Inputs

This directory holds the canonical staged infra inputs for the thin GKE lane.

## Current datasets

- `project/` - project adoption and minimum API enablement inputs for `infra/1-project`
- `iam/` - deploy/runtime identity contracts for `infra/2-iam`
- `gke/` - Artifact Registry, cluster, network, and hardening placeholders for `infra/3-gke`
- `ci/` - future trigger/provisioning inputs for `infra/4-ci`

## Contract pattern

Each dataset follows the same pattern:

- `defaults.yaml` - reusable baseline values
- `overrides.yaml` - optional local environment bind, intentionally untracked
- `.config.yaml` - documents the defaults/overrides pairing
- `schema/*.json` - narrow bind schemas plus partial override schemas

The staged infra lane is inspired by `incubation/my-org-monorepo` and selective
Cloud Foundation Fabric module boundaries, but the active runtime proof bind for
the current app slice still lives in `config/gke-exposure/`.

For the staged validation order, use
`docs/architecture/infra-manual-walkthrough.md`.

For whole-environment bundle choices spanning multiple datasets, use
`config/profiles/README.md`.
