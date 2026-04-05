# IAM Dataset Contract

This directory holds the identity-layer dataset inputs for `infra/2-iam`.

The first pass keeps IAM intentionally small:

- optional service accounts for deploy and runtime lanes
- optional adoption of existing service accounts by email
- optional project-role bindings for those service accounts

Artifact Registry repository creation remains in `infra/3-gke`, but the identity
contract created here is intended to feed that stage later.

## Use these files

- `defaults.yaml` - reusable IAM defaults
- `overrides.yaml` - local environment-specific overrides
- `overrides.proof-bind.example.yaml` - minimal example for the first bind
- `overrides.adopt-existing.example.yaml` - adopt-existing example for service accounts already present in the project
- `schema/proof-bind.schema.json` - narrow proof-bind schema
- `schema/adopt-existing.schema.json` - narrow schema for adopting existing identities
- `schema/overrides.schema.json` - generic partial schema for local overrides

## When editing `overrides.yaml`

Confirm first:

- `project_id`

Then only opt into what you really need:

- `service_accounts.<name>.create`
- `service_accounts.<name>.email`
- `service_account_project_roles.<name>`

Choose one mode per identity:

- `create: true` when this stage should create the service account
- `create: false` plus `email: ...` when this stage should adopt an existing service account

## Workflow

1. copy the example file to `overrides.yaml`
2. choose create-or-adopt mode for only the identities needed for the next step
3. review the resulting role bindings carefully
4. validate and plan `infra/2-iam`
