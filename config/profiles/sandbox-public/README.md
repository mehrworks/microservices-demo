# Sandbox Public Profile

Use this bundle for the simplest staged-infra shape:

- existing project
- small or repo-managed IAM lane
- public/sandbox-style GKE contract
- placeholder-only CI contract

## Bundle map

- `project` -> `config/datasets/project/overrides.adopt-existing.example.yaml`
- `iam` -> `config/datasets/iam/overrides.proof-bind.example.yaml`
- `gke` -> `config/datasets/gke/overrides.sandbox-public.example.yaml`
- `ci` -> `config/datasets/ci/overrides.placeholder-only.example.yaml`

## Best fit

- simple non-constrained project
- first staged runtime regression later
- fastest way to validate the baseline happy path
