# Infra Manual Walkthrough

This walkthrough is the staged-infra companion to the existing runtime proof
docs.

Use it when you want to shape or validate the repo's cloud-contract lane without
jumping straight into a live GKE proof run.

## Goal

Move the repo forward structurally by binding datasets and validating Terraform
stages in a consistent human-reviewed order.

This walkthrough is intentionally:

- contract-first
- human-run
- plan-first
- separate from app rollout proof

## Recommended order

1. review the architecture docs in this directory
2. bind `config/datasets/project/overrides.yaml`
3. bind `config/datasets/iam/overrides.yaml` only if identities are needed now
4. bind `config/datasets/gke/overrides.yaml`
5. bind `config/datasets/ci/overrides.yaml` only when CI-trigger provisioning work is in scope
6. validate and plan stages in order:
   - `infra/1-project`
   - `infra/2-iam`
   - `infra/3-gke`
   - `infra/4-ci`

## Choosing a starting profile

Start by choosing a whole-environment bundle in `config/profiles/`:

- `sandbox-public/`
- `org-hardened/`
- `adopt-existing-foundation/`

Then apply the referenced per-dataset examples.

For `config/datasets/gke/`, start from one of these examples:

- `overrides.sandbox-public.example.yaml`
  - simple public sandbox shape
  - default-style networking assumptions
  - no private nodes
- `overrides.org-hardened.example.yaml`
  - explicit network contract
  - private nodes enabled
  - control-plane CIDR bound
  - public load balancer path considered restricted by default
- `overrides.adopt-existing.example.yaml`
  - existing Artifact Registry repo and cluster
  - useful when foundation resources were created elsewhere already

These are profile examples, not automation. Copy one to `overrides.yaml` and then
edit it by hand.

For `config/datasets/iam/`, there are now two common shapes as well:

- `overrides.proof-bind.example.yaml` for repo-managed service-account creation
- `overrides.adopt-existing.example.yaml` for existing service accounts bound by email

For `config/datasets/ci/`, use one of these shapes:

- `overrides.placeholder-only.example.yaml` for the current branch-default contract-only mode
- `overrides.trigger-enabled.example.yaml` only when future trigger provisioning work is intentionally in scope

The current recommended staged-infra default remains the placeholder profile.
Keep the trigger-enabled profile for later CI provisioning work, not for the
current stable baseline.

Current workspace convention:

- `default` for the simple sandbox/public environment
- `org-constrained` for the stricter org-managed environment

## Runtime validation order

When you run the accepted real-GKE proof baseline, the recommended order is:

1. use the simpler non-constrained project first
   - this validates the baseline happy path with fewer environment variables
   - it is the fastest way to catch contract mistakes in the staged lane itself
2. then use the org-constrained project
   - this validates the hardened/real-world path
   - it confirms that private-node, network, and load-balancer restrictions are
     represented cleanly as config differences instead of ad hoc shell fixes

So both accounts matter, but not at the same time: use the simple account for the
first runtime regression, then the constrained account for the stronger follow-up check.

## Validation commands

```bash
./docs/bootstrap/staged-infra-cycle.sh --workspace default

# or target individual stages:
./docs/bootstrap/staged-infra-cycle.sh --workspace org-constrained 1-project 2-iam

# export stage contracts when readable state exists:
./docs/bootstrap/export-stage-contracts.sh --workspace default
```

The export helper is optional during the staged validation loop. It becomes useful after
real applies exist and you want local JSON snapshots of the emitted stage contracts.

If you want validation without export, use:

```bash
./docs/bootstrap/staged-infra-cycle.sh --workspace default --skip-export
```

Once stage exports exist, render the runtime proof bind with:

```bash
./docs/bootstrap/render-gke-proof-bind.sh --workspace default
```

For remote backend migration later, render local `*.tfbackend` files from the
tracked examples with:

```bash
./docs/bootstrap/render-tfbackends.sh --bucket YOUR_TF_STATE_BUCKET
```

That remains a later step, not a requirement for the current stable baseline.

## What to defer

Do not turn this walkthrough into repeated cluster reproving unless you are
actually validating a contract change.

Keep these as separate runtime-validation steps:

- live GKE apply
- namespace bootstrap and Skaffold rollout
- public frontend and gRPC proof calls
- debugging environment-specific org-policy runtime issues

## Why this exists

The repo now needs to support two truths at once:

- a simple app-first proof lane that already works
- a more explicit infra contract lane that can later absorb stricter org-managed
  requirements

This walkthrough keeps the repo-shape work moving without spending the current
iteration on cluster execution.
