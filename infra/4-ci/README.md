# Stage 4 - CI Contract

This stage reserves the future CI trigger/provisioning lane.

In this first pass it is intentionally placeholder-only:

- dataset loading and outputs exist
- no cloud mutation resources are created yet

The stage now validates the minimum contract for any future trigger enablement and
emits a compact CI handoff summary, but it still does not provision live trigger
resources.

Current contract intent:

- `provider: github-actions` plus `trigger_mode: placeholder` matches the current branch reality
- `provider: cloud-build` plus `trigger_mode: cloud-build-triggers` is the future explicit provisioning direction if this lane grows later

Inputs come from `config/datasets/ci/`.

Two profile examples are available there now:

- `overrides.placeholder-only.example.yaml`
- `overrides.trigger-enabled.example.yaml`

Validation flow:

```bash
terraform -chdir=infra/4-ci init -backend=false -input=false
terraform -chdir=infra/4-ci validate
terraform -chdir=infra/4-ci plan -input=false -lock=false -no-color
```
