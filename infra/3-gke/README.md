# Stage 3 - GKE Contract

This stage defines the cluster-side infrastructure contract for the app-first GKE
lane.

Current scope:

- Artifact Registry repository contract
- optional repository IAM bindings
- optional custom VPC/subnetwork creation
- optional Autopilot cluster creation
- hardening placeholders for private nodes, control-plane CIDR, NAT, and public
  exposure policy

The stage supports both create and adopt modes for the Artifact Registry repo and
the GKE cluster contract:

- `create: true` creates the resource
- `create: false` expects an existing resource and reads it as the emitted contract

The stage also emits a dedicated Artifact Registry contract output so later lanes
can consume a consistent repo URL and create/adopt status without inspecting the
resource path directly.

Inputs come from `config/datasets/gke/`.

Profile examples live alongside that dataset:

- `overrides.sandbox-public.example.yaml`
- `overrides.org-hardened.example.yaml`

These examples make the recent simple-project versus org-managed-project split a
first-class repo contract instead of leaving it in ad hoc notes only.

Selective CFF alignment for this stage:

- Artifact Registry concerns line up with `artifact-registry`
- Autopilot cluster concerns line up with `gke-cluster-autopilot`
- later egress hardening concerns line up with `net-cloudnat`

Validation flow:

```bash
terraform -chdir=infra/3-gke init -backend=false -input=false
terraform -chdir=infra/3-gke validate
terraform -chdir=infra/3-gke plan -input=false -lock=false -no-color
```
