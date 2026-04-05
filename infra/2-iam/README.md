# Stage 2 - IAM Contract

This stage defines the small identity layer for the thin GKE lane.

Current scope:

- optional deploy service account
- optional runtime service account
- optional adoption of existing service accounts by explicit email
- optional project-role bindings for those service accounts

This stage now supports both modes cleanly:

- create mode for new identities managed by this repo
- adopt mode for existing identities already provisioned elsewhere

Inputs come from `config/datasets/iam/`.

Validation flow:

```bash
terraform -chdir=infra/2-iam init -backend=false -input=false
terraform -chdir=infra/2-iam validate
terraform -chdir=infra/2-iam plan -input=false -lock=false -no-color
```
