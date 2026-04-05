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
5. `acceptance-environment-matrix.md`
   - freezes the current accepted environments and expected proof behavior
6. `infra-manual-walkthrough.md`
   - shows the human-run order for binding datasets and validating stages
7. `infra-operator-runbook.md`
   - ties the local helper scripts into one operator workflow
8. `hybrid-execution-model.md`
   - defines the intended next-step cloud/local boundary
9. `runtime-responsibility-matrix.md`
   - assigns runtime responsibilities between cloud, local worker, and shared contracts
10. `job-lifecycle-contract.md`
   - defines the first planned job/status/result contract for the hybrid phase
11. `job-control-api-surface.md`
   - defines the first minimal cloud-side API shape for the hybrid control plane
12. `local-worker-agent.md`
   - defines the first worker identity and lease assumptions for the hybrid phase
13. `job-control-module-spec.md`
   - defines the first implementation-facing module boundary for the cloud-side control plane
14. `job-control-storage-boundary.md`
   - defines the first persistence-facing boundary for the cloud-side control plane

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

The accepted runtime environments and their expected outcomes are frozen in
`docs/architecture/acceptance-environment-matrix.md`.

Concrete hybrid contract artifacts now live under `config/contracts/job-control/`.
