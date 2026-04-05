# Infra Operator Runbook

This runbook ties the local staged-infra helpers together into one human-run flow.

It is for the current repo-shape phase, where the goal is to evolve the contract
surface without spending the iteration on live cluster proof work.

## Default operator loop

1. choose the right environment bundle from `config/profiles/`
   - `sandbox-public/`
   - `org-hardened/`
   - `adopt-existing-foundation/`
2. materialize the bundle into local overrides:

   ```bash
   ./docs/bootstrap/apply-profile-bundle.sh --dry-run sandbox-public
   ./docs/bootstrap/apply-profile-bundle.sh sandbox-public
   ```

   Use `--dry-run` first if you want to inspect the copy plan.
3. run:

   ```bash
   ./docs/bootstrap/staged-infra-cycle.sh
   ```

4. if state exists, inspect the exported snapshots under `config/stages/*/outputs/`
5. only later, when the repo-shape pass is stable, move to the dedicated real-GKE runtime check

## Helper roles

- `validate-infra-contracts.sh`
  - checks Terraform formatting and validates/plans the requested stages
- `export-stage-contracts.sh`
  - captures emitted JSON outputs for readable local handoff snapshots
- `render-tfbackends.sh`
  - prepares local backend files only after a real state bucket exists
- `staged-infra-cycle.sh`
  - the default wrapper for the common validate-then-export loop

## When to use the runtime helpers instead

Use `bootstrap-gke-proof.sh` and the runtime proof docs only when you are in the
later dedicated GKE runtime pass.

For that later pass, keep the order:

1. simple non-constrained project
2. org-constrained project

That keeps baseline staged-lane mistakes separate from genuine hardened-environment differences.
