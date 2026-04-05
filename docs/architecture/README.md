# Architecture Docs

This directory collects the branch-local architecture documents for the thin GKE
infra lane and the current app-first delivery model.

## Read in this order

1. `thin-gke-infra-lane.md`
   - explains why the repo now separates app delivery from staged infra contracts
2. `app-to-infra-map.md`
   - shows how the active app slice sits inside the new staged infra lane
3. `infra-dependency-map.md`
   - shows stage ownership and dependency boundaries
4. `stage-contract-matrix.md`
   - shows what each stage consumes and emits today
5. `infra-manual-walkthrough.md`
   - shows the human-run order for binding datasets and validating stages
6. `infra-operator-runbook.md`
   - ties the local helper scripts into one operator workflow

The local command companion for these docs is
`docs/bootstrap/validate-infra-contracts.sh`.

The default wrapper for the local repo-shape loop is
`docs/bootstrap/staged-infra-cycle.sh`.

After real state exists, `docs/bootstrap/export-stage-contracts.sh` can snapshot
emitted contracts into `config/stages/*/outputs/`.

After stage exports exist, `docs/bootstrap/render-gke-proof-bind.sh` can render
the active local runtime proof bind from those staged outputs.

After a real state bucket exists, `docs/bootstrap/render-tfbackends.sh` can
materialize local `config/backends/*.tfbackend` files from the tracked examples.

## Scope

These docs describe the current staged-infra direction only. They do not replace
the existing runtime proof docs under `docs/gcp-bootstrap.md` and
`docs/dormant-scaffold.md`.

For the later real-GKE runtime pass, use the simple non-constrained project first,
then the org-constrained project.
