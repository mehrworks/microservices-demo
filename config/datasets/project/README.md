# Project Dataset Contract

This directory holds the dataset inputs for `infra/1-project`.

The first pass adopts an existing project and manages only the minimum API
enablement contract for the thin GKE lane. It does not create billing links,
folders, or full project-factory behavior.

## Use these files

- `defaults.yaml` - reusable baseline values
- `overrides.yaml` - local environment-specific overrides
- `overrides.onboarding-bind.example.yaml` - minimal example bind for a real project
- `overrides.adopt-existing.example.yaml` - explicit adopt-existing example for a project that already exists
- `schema/onboarding-bind.schema.json` - narrow schema for the first project bind
- `schema/adopt-existing.schema.json` - narrow schema for explicit project adoption
- `schema/overrides.schema.json` - generic partial schema for local overrides

## When editing `overrides.yaml`

Minimum expected values:

- `project_id`

Optional values only when they differ from defaults:

- `region`
- `project_number`
- `enabled_apis`

## Workflow

1. copy the example file to `overrides.yaml`
2. bind the real project identifier
3. keep the override minimal when defaults already fit
4. validate and plan `infra/1-project`

This stage is adopt-existing only in the current pass. It does not create a new
project; it declares the contract for an existing project and manages the minimum
API surface around it.
