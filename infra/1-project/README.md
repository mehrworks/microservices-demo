# Stage 1 - Project Contract

This stage adopts an existing project and manages the minimum API enablement
needed for the thin GKE lane.

It is intentionally small:

- no billing attachment
- no folder placement
- no project factory behavior

This is intentionally an adopt-existing stage in the current pass. It aligns with
the project/API-enable concern split seen in Cloud Foundation Fabric, but does not
try to become a full project-factory lane here.

Inputs come from `config/datasets/project/`.

Validation flow:

```bash
terraform -chdir=infra/1-project init -backend=false -input=false
terraform -chdir=infra/1-project validate
terraform -chdir=infra/1-project plan -input=false -lock=false -no-color
```
