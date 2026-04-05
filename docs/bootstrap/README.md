# Bootstrap Helpers

This directory collects the branch-local helper scripts for the thin staged infra
lane and the current runtime proof lane.

## Helpers

- `bootstrap-gke-proof.sh`
  - current runtime proof helper for API enablement, registry setup, auth, and proof prep
- `validate-infra-contracts.sh`
  - validates the staged infra lane locally with `fmt`, `init`, `validate`, and `plan`
- `export-stage-contracts.sh`
  - exports `terraform output -json` into `config/stages/*/outputs/current.json` when state exists
- `render-tfbackends.sh`
  - renders local `*.tfbackend` files from tracked examples once a real state bucket exists
- `staged-infra-cycle.sh`
  - convenience wrapper for the common local repo-shape cycle: validate first, then export stage contracts if possible
- `apply-profile-bundle.sh`
  - safely copies a chosen profile bundle's dataset examples into local `overrides.yaml` files
- `validate-profile-bundles.sh`
  - checks that every profile bundle is structurally valid and points at real dataset examples

## Recommended use right now

For the current repo-shape pass:

1. apply a profile bundle or bind datasets by hand
2. run `staged-infra-cycle.sh`
3. inspect emitted contracts if state exists

The normal validation path already includes profile-bundle checks.

Keep real GKE proof work for the later dedicated runtime pass.
