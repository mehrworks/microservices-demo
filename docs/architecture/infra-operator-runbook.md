# Infra Operator Runbook

This runbook ties the local staged-infra helpers together into one human-run flow.

It now describes the current stable operator baseline for the accepted GKE proof
environments.

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
   ./docs/bootstrap/staged-infra-cycle.sh --workspace default
   ```

4. if state exists, inspect the exported snapshots under `config/stages/*/outputs/`
5. render the runtime proof bind when stage exports are ready:

   ```bash
   ./docs/bootstrap/render-gke-proof-bind.sh --workspace default
   ```

6. run the environment-specific proof checks described by the rendered bind and
   compare outcomes against `docs/architecture/acceptance-environment-matrix.md`

## Helper roles

- `validate-infra-contracts.sh`
  - checks Terraform formatting and validates/plans the requested stages
- `export-stage-contracts.sh`
  - captures emitted JSON outputs for readable local handoff snapshots
- `render-tfbackends.sh`
  - prepares local backend files only after a real state bucket exists
- `staged-infra-cycle.sh`
  - the default wrapper for the common validate-then-export loop
- `render-gke-proof-bind.sh`
  - renders the local runtime proof bind from staged outputs and proof mode

## Runtime proof order

Use `bootstrap-gke-proof.sh`, `render-gke-proof-bind.sh`, and the runtime proof docs
when you are running the accepted GKE proof baseline.

For that later pass, keep the order:

1. simple non-constrained project
2. org-constrained project

That keeps baseline staged-lane mistakes separate from genuine hardened-environment differences.

Current recommended workspace names:

- `default` for the simple sandbox/public environment
- `org-constrained` for the stricter org-managed environment
