# Thin Infra Lane

This directory holds the staged infrastructure lane for the app-first GKE proof
shape.

Current intent:

- `1-project` - adopt an existing project and manage minimum API enablement
- `2-iam` - optional deploy/runtime identities and small project-role contracts
- `3-gke` - Artifact Registry, Autopilot cluster, namespace defaults, and future
  hardening placeholders
- `4-ci` - future trigger/provisioning contract, placeholder only in this pass

Selective inspiration comes from Cloud Foundation Fabric module boundaries such as
`project`, `iam-service-account`, `artifact-registry`, `gke-cluster-autopilot`,
and later `net-cloudnat`, without importing FAST/landing-zone weight into this
repo.

Non-goals for this first pass:

- no Terraform-managed workload deployment
- no `kubectl apply`
- no rollout waits or app mutation from Terraform
- no landing-zone or org-policy management
