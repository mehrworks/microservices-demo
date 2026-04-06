# Bootstrap Helpers

This directory collects the branch-local helper scripts for the thin staged infra
lane and the current runtime proof lane.

## Helpers

- `bootstrap-gke-proof.sh`
  - current runtime proof helper for API enablement, registry setup, auth, and proof prep
- `validate-infra-contracts.sh`
  - validates the staged infra lane locally with `fmt`, `init`, `validate`, and `plan`
- `export-stage-contracts.sh`
  - exports `terraform output -json` into workspace-aware stage snapshots and refreshes `current.json` as a convenience alias
- `render-tfbackends.sh`
  - renders local `*.tfbackend` files from tracked examples once a real state bucket exists
- `staged-infra-cycle.sh`
  - convenience wrapper for the common local repo-shape cycle: validate first, then export stage contracts if possible
- `apply-profile-bundle.sh`
  - safely copies a chosen profile bundle's dataset examples into local `overrides.yaml` files
- `validate-profile-bundles.sh`
  - checks that every profile bundle is structurally valid and points at real dataset examples
- `render-gke-proof-bind.sh`
  - renders the local runtime proof bind from exported stage contracts and proof mode
- `job-control-smoke.sh`
  - exercises the dormant internal operator/worker simulation path against a running frontend

## Recommended use right now

For the current stable operator baseline:

1. apply a profile bundle or bind datasets by hand
2. run `staged-infra-cycle.sh` with the intended workspace
3. inspect emitted contracts if state exists
4. render `config/gke-exposure/overrides.yaml` from stage outputs when the runtime bind is needed

The normal validation path already includes profile-bundle and hybrid contract
artifact checks.

For manual hybrid control-plane smoke testing against a running frontend with
`JOB_CONTROL_OPERATOR_TOKEN` and `JOB_CONTROL_WORKER_TOKEN` set, use:

```bash
JOB_CONTROL_OPERATOR_TOKEN=... JOB_CONTROL_WORKER_TOKEN=... \
  ./docs/bootstrap/job-control-smoke.sh
```

Use `docs/architecture/infra-operator-runbook.md` as the single operator flow,
and `docs/architecture/acceptance-environment-matrix.md` for expected results by environment.
