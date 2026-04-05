# Org Hardened Profile

Use this bundle when the target environment is governed by stronger platform or
org-policy constraints.

## Bundle map

- `project` -> `config/datasets/project/overrides.adopt-existing.example.yaml`
- `iam` -> `config/datasets/iam/overrides.adopt-existing.example.yaml`
- `gke` -> `config/datasets/gke/overrides.org-hardened.example.yaml`
- `ci` -> `config/datasets/ci/overrides.placeholder-only.example.yaml`

## Best fit

- private-node requirements
- explicit network/CIDR contract
- existing identity adoption is more likely than repo-managed creation
- later hardened-environment runtime regression after the baseline sandbox check

Recommended workspace:

- `org-constrained`
