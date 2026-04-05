# CI Dataset Contract

This directory holds the future trigger/provisioning inputs for `infra/4-ci`.

The first pass is intentionally placeholder-only. It describes the CI contract and
the path filters that matter for the thin app + infra lane, but it does not yet
turn CI trigger provisioning into an active Terraform mutation path.

## Use these files

- `defaults.yaml` - reusable CI defaults
- `overrides.yaml` - local environment-specific overrides
- `overrides.trigger-bind.example.yaml` - minimal example bind for a real repo
- `overrides.placeholder-only.example.yaml` - explicit placeholder-only profile
- `overrides.trigger-enabled.example.yaml` - explicit future trigger-enabled profile
- `schema/trigger-bind.schema.json` - narrow schema for the first bind
- `schema/placeholder-only.schema.json` - narrow schema for placeholder-only mode
- `schema/trigger-enabled.schema.json` - narrow schema for future trigger-enabled mode
- `schema/overrides.schema.json` - generic partial schema for local overrides

## Workflow

1. copy the example file to `overrides.yaml`
2. bind the real project and GitHub repository owner when trigger provisioning is
    intentionally adopted later
3. keep `triggers_enabled: false` until that work is explicitly in scope

## Profile examples

- `overrides.placeholder-only.example.yaml`
  - current branch-default mode
  - `provider: github-actions`
  - `trigger_mode: placeholder`
  - contract only, no trigger provisioning intent yet
- `overrides.trigger-enabled.example.yaml`
  - future mode for when trigger provisioning work is explicitly in scope
  - `provider: cloud-build`
  - `trigger_mode: cloud-build-triggers`
  - requires a real project and repo owner bind
