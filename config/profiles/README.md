# Profile Bundles

This directory groups the per-dataset example binds into reviewable environment
bundles.

Each profile points at one recommended example file per staged dataset:

- `project`
- `iam`
- `gke`
- `ci`

These bundles are not auto-applied. They exist so an operator can choose a whole
environment shape first, then copy the matching dataset examples into local
`overrides.yaml` files with fewer ad hoc decisions.

## Available bundles

- `sandbox-public/`
  - baseline simple-project shape
- `org-hardened/`
  - stricter org-managed shape
- `adopt-existing-foundation/`
  - use when core project/identity/cluster resources already exist

## Typical use

1. pick a profile directory
2. read its `README.md`
3. copy the referenced dataset example files into local `config/datasets/*/overrides.yaml`, or run:

   ```bash
   ./docs/bootstrap/apply-profile-bundle.sh sandbox-public
   ```

   Use `--dry-run` to inspect the copy plan first.
4. run `./docs/bootstrap/staged-infra-cycle.sh --workspace <workspace>` using the profile's recommended workspace

Bundle validation helper:

```bash
./docs/bootstrap/validate-profile-bundles.sh
```

The current later runtime-check order remains:

1. `sandbox-public`
2. `org-hardened`
