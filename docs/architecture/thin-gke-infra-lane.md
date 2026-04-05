# Thin GKE Infra Lane

This branch now has two related but separate layers:

1. app delivery
2. cloud infrastructure contracts

## App delivery stays primary

The active app slice is still:

- `frontend` in `catalog-only` mode
- `productcatalogservice`

That slice is still deployed through:

- `skaffold.yaml`
- `kubernetes-manifests/`
- `kustomize/`
- the proof flow in `docs/gcp-bootstrap.md` and `docs/dormant-scaffold.md`

This remains the authoritative runtime proof path for the branch.

## Infra now has a staged lane

The new `infra/` directory introduces a thin staged contract inspired by the
`my-org-monorepo` repo shape:

- `infra/1-project` - project adoption and API enablement
- `infra/2-iam` - optional deploy/runtime identities
- `infra/3-gke` - Artifact Registry, cluster contract, and optional hardening placeholders
- `infra/4-ci` - future CI trigger contract, placeholder only for now

Its input layer lives under `config/datasets/`.

The stage boundaries are also intentionally aligned with selective Cloud
Foundation Fabric concerns such as project setup, IAM identities, Artifact
Registry, Autopilot GKE, and later optional NAT/network hardening.

## Why this split exists

Recent proof runs showed that the app path itself is viable, but infrastructure
requirements can vary by environment:

- simple sandbox projects may work with default networking and public load balancers
- org-managed projects may require private nodes, explicit control-plane CIDR,
  custom networking, or may block some public load balancer types

Those differences belong in an infra contract layer, not in ad hoc shell history.

## What is intentionally not happening yet

- Terraform does not run `kubectl apply`
- Terraform does not wait on Pods or Services
- Terraform does not own the current workload rollout path
- this repo does not manage org policies or a full landing zone

## Transitional rule

`config/gke-exposure/` stays in place as the active proof-bind contract until the
staged infra lane becomes the normal way to recreate the same two-service proof.

Until then:

- use `config/gke-exposure/` for the live proof bind
- use `config/datasets/` for infra-stage inputs and validation
- treat the legacy `terraform/` directory as reference material, not the target
  architecture

For the staged infra lane, use `docs/architecture/infra-manual-walkthrough.md` as
the human-run order of operations.
