# Stage Contract Matrix

This document summarizes the current staged-infra contract boundaries, including
what each stage consumes and what it emits for later lanes.

## Current matrix

| Stage | Primary dataset | Current modes | Key emitted contract |
| --- | --- | --- | --- |
| `infra/1-project` | `config/datasets/project` | adopt existing project | `project_contract` |
| `infra/2-iam` | `config/datasets/iam` | create or adopt identities | `service_accounts`, `service_account_project_roles`, `iam_contract` |
| `infra/3-gke` | `config/datasets/gke` | create or adopt repo/cluster | `artifact_registry_contract`, `cluster_contract`, `network_contract` |
| `infra/4-ci` | `config/datasets/ci` | placeholder profile now, future trigger-enabled profile later | `ci_contract`, `ci_contract_summary` |

## Handoff expectations

- `infra/1-project`
  - emits the project id, project number, region, and enabled API list
  - gives downstream stages a stable project contract without doing project factory work
- `infra/2-iam`
  - emits identity members/emails whether the identities were created here or adopted
  - gives later lanes a stable service-account contract without forcing creation mode
- `infra/3-gke`
  - emits both cluster-side and Artifact Registry contracts
  - makes create-vs-adopt status explicit in outputs
- `infra/4-ci`
  - currently documents CI intent only
  - distinguishes the current GitHub Actions placeholder mode from a future Cloud Build trigger mode
  - validates the minimum enablement contract but does not yet provision real triggers

## Local output snapshots

When a stage has readable Terraform state, the local helper
`docs/bootstrap/export-stage-contracts.sh` can export its emitted contract to:

- `config/stages/1-project/outputs/current.json`
- `config/stages/2-iam/outputs/current.json`
- `config/stages/3-gke/outputs/current.json`
- `config/stages/4-ci/outputs/current.json`

These are local generated artifacts, not canonical inputs.

## Why this matters

The repo is moving toward explicit stage-to-stage handoffs instead of relying on
shell history or environment memory. This keeps the app repo compatible with both:

- simple sandbox projects
- stricter org-managed projects

without forcing the same runtime or provisioning mode everywhere.

Profile bundles under `config/profiles/` sit one layer above this matrix and group
these stage contracts into reviewable environment choices.
