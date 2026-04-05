# GKE Dataset Contract

This directory holds the staged-cluster dataset inputs for `infra/3-gke`.

It defines the app repo's cloud-facing baseline:

- Artifact Registry repo contract
- Autopilot cluster contract
- namespace default
- placeholder hardening knobs for org-managed environments
- runtime proof-mode contract for public vs internal verification

The active human-reviewed proof bind for the currently deployed app slice still
lives under `config/gke-exposure/`. This dataset focuses on infrastructure, not
the live frontend URL bind.

## Use these files

- `defaults.yaml` - reusable baseline values
- `overrides.yaml` - local environment-specific overrides
- `overrides.proof-bind.example.yaml` - first concrete infrastructure bind
- `overrides.sandbox-public.example.yaml` - simple public sandbox profile example
- `overrides.org-hardened.example.yaml` - stricter org-managed profile example
- `overrides.adopt-existing.example.yaml` - adopt-existing profile for an already provisioned repo + cluster
- `schema/proof-bind.schema.json` - narrow schema for the first bind shape
- `schema/sandbox-public.schema.json` - narrow schema for the sandbox profile example
- `schema/org-hardened.schema.json` - narrow schema for the org-hardened profile example
- `schema/adopt-existing.schema.json` - narrow schema for the adopt-existing profile example
- `schema/overrides.schema.json` - generic partial schema for local overrides

## Baseline vs placeholders

Baseline-safe fields:

- `project_id`
- `region`
- `namespace`
- `artifact_registry.repository_id`
- `cluster.name`

Placeholder-only hardening fields:

- `network.*`
- `hardening.private_nodes_enabled`
- `hardening.master_ipv4_cidr`
- `hardening.cloud_nat_enabled`
- `hardening.public_load_balancer_enabled`

Keep those placeholder fields off until a target environment actually requires
them.

Runtime-proof fields:

- `proof.http_mode`
- `proof.grpc_mode`
- `proof.frontend_*`
- `proof.productcatalog_*`

Those values bridge the staged infra lane to the active runtime proof bind and
make the simple public path distinct from the org-constrained internal path.

## Workflow

1. copy the example file to `overrides.yaml`
2. bind the real project and decide whether this stage should create or adopt
   the repo and cluster
3. only opt into custom network or private-node settings when the environment
   forces them
4. validate and plan `infra/3-gke`

For repo and cluster resources, the two main modes are:

- `create: true` when this stage should create them
- `create: false` when this stage should adopt existing resources and emit the
  contract around them

## Profile examples

Use the two profile examples as reviewable starting points:

- `overrides.sandbox-public.example.yaml`
  - simplest public path
  - no private nodes
  - no NAT
  - public load balancer allowed
  - HTTP proof via public load balancer
- `overrides.org-hardened.example.yaml`
  - explicit network contract
  - private nodes enabled
  - control-plane CIDR required
  - NAT assumed on
  - public load balancer treated as restricted by default
  - HTTP proof via frontend port-forward
- `overrides.adopt-existing.example.yaml`
  - existing Artifact Registry repo
  - existing cluster contract
  - useful when foundation resources were provisioned elsewhere already

These examples reflect the kinds of differences already seen between a simple
sandbox project and a stricter org-managed project.
