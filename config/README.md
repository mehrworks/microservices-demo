# Configuration Layout

This directory now uses a split between dataset inputs, backend examples, stage
output placeholders, and the existing proof-bind contract for the active app
slice.

## Current layout

```text
config/
  README.md
  gke-exposure/
  datasets/
    project/
    iam/
    gke/
    ci/
  profiles/
  backends/
  stages/
```

Guidance:

- `config/gke-exposure/` remains the active human-reviewed proof bind for the
  current `frontend` + `productcatalogservice` + `recommendationservice` slice.
- `config/datasets/` holds the new staged infra inputs for `infra/1-project`,
  `infra/2-iam`, `infra/3-gke`, and `infra/4-ci`.
- `config/profiles/` groups those dataset examples into reviewable environment bundles.
- `config/backends/` holds placeholder `*.tfbackend.example` files only.
- `config/stages/*/outputs/` is reserved for generated artifacts later; it is
  not hand-edited input config.

See also `config/datasets/README.md` for the stage-to-dataset mapping.
See also `config/profiles/README.md` for whole-environment bundle choices.

## Contract rules

- committed `defaults.yaml` files carry reusable baseline values
- local `overrides.yaml` files are optional and intentionally untracked under
  `config/datasets/`
- example overrides and narrow schemas describe the first safe bind shape for a
  stage
- override schemas keep top-level keys optional so partial local binds stay easy
  to review

## App delivery versus infra stages

The repo stays app-first:

- workload deployment still uses Skaffold, manifests, and the existing proof
  docs
- the new `infra/` lane only defines cloud prerequisites and contracts
- Terraform does not deploy Kubernetes workloads directly in this first pass

For the active runtime proof path, start with `docs/gcp-bootstrap.md` and
`config/gke-exposure/README.md`.

For staged infra validation, use `docs/architecture/infra-manual-walkthrough.md`
and `docs/bootstrap/staged-infra-cycle.sh`.

For local exported stage snapshots after state exists, use
`docs/bootstrap/export-stage-contracts.sh`.

For rendering the local runtime proof bind from staged outputs, use
`docs/bootstrap/render-gke-proof-bind.sh`.

For local backend file rendering once a real state bucket exists, use
`docs/bootstrap/render-tfbackends.sh`.
