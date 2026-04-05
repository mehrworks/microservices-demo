# GKE Exposure Proof Contract

This directory records the smallest environment bind needed for the active thin
GKE exposure path:

- public HTTP edge: `frontend` running in `catalog-only` mode
- backing service: `productcatalogservice`
- no DB, service mesh, or private-network requirement in the baseline proof

## Branch defaults

These values are branch defaults, not rediscovery items:

- `region: europe-west3`
- `cluster: msdemo-public`
- `namespace: msdemo-public`
- `artifact_registry_repository: services`
- `frontend.mode: catalog-only`
- `frontend.service_name: frontend`
- `frontend.public_service_name: frontend-external`
- `productcatalogservice.service_name: productcatalogservice`

## Files

- `defaults.yaml` - reusable proof defaults for the branch
- `overrides.example.yaml` - copy to `overrides.yaml` for a real environment bind
- `schema/proof-bind.schema.json` - narrow schema for the proof-bind shape
- `.config.yaml` - documents the defaults/overrides pairing used in this folder

## Workflow

1. If this is a fresh GCP account or project, start with `docs/gcp-bootstrap.md`.
2. Copy `overrides.example.yaml` to `overrides.yaml`.
3. Bind the project-specific values: `project_id`, `default_repo`, and later `frontend_public_base_url`.
4. Use those values with `docs/bootstrap/bootstrap-gke-proof.sh` or the branch proof flow in `docs/dormant-scaffold.md`.

This contract is intentionally small and human-reviewed. It is not a full repo
configuration system.

The new staged infra lane lives under `infra/` with dataset inputs under
`config/datasets/`, but this folder remains the active proof-bind contract until
that infra path becomes the normal way to recreate the same two-service runtime proof.
