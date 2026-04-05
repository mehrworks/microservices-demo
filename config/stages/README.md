# Stage Output Placeholders

This directory is reserved for generated stage outputs only.

- `config/datasets/` holds the canonical inputs.
- `config/backends/` holds backend-init examples only.
- `config/stages/*/outputs/` is where generated artifacts or exported output
  snapshots can land later if the repo intentionally adopts that pattern.
- `config/stages/.gitignore` keeps exported local JSON snapshots untracked.

Do not hand-edit stage outputs as if they were input config.

Current local helper:

```bash
./docs/bootstrap/export-stage-contracts.sh
```

That helper exports `terraform output -json` into:

- `config/stages/<stage>/outputs/<workspace>.json`
- `config/stages/<stage>/outputs/current.json` as a convenience alias for the latest export

Recommended workspace contract for the current acceptance environments:

- `default` - simple sandbox/public project
- `org-constrained` - stricter org-managed project
