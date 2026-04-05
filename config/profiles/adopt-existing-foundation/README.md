# Adopt Existing Foundation Profile

Use this bundle when most of the cloud-side foundation already exists and this
repo mainly needs to declare the contracts around it.

## Bundle map

- `project` -> `config/datasets/project/overrides.adopt-existing.example.yaml`
- `iam` -> `config/datasets/iam/overrides.adopt-existing.example.yaml`
- `gke` -> `config/datasets/gke/overrides.adopt-existing.example.yaml`
- `ci` -> `config/datasets/ci/overrides.placeholder-only.example.yaml`

## Best fit

- project already exists
- identities already exist
- Artifact Registry repo and cluster already exist
- repo-shape validation without provisioning ownership drift

Recommended workspace:

- choose an explicit environment-specific workspace rather than reusing another bundle's state by accident
