# Acceptance Environment Matrix

This document freezes the current accepted environments and expected proof shape
for the active branch baseline.

## Active baseline

The current active slice is:

- `frontend` in `catalog-only` mode
- `productcatalogservice`
- `recommendationservice`

No additional boutique services are part of the accepted branch baseline.

## Acceptance matrix

| Environment | Terraform workspace | Profile bundle | Project | HTTP proof mode | gRPC proof mode | Expected outcome |
| --- | --- | --- | --- | --- | --- | --- |
| simple/public | `default` | `sandbox-public` | `project-6a8a62e6-d3c9-43a0-82f` | public load balancer | port-forward | home page, product recommendations, and `ListProducts` all pass publicly/internally as expected |
| org-constrained | `org-constrained` | `org-hardened` | `mwk27x7-factory-test-drive` | frontend port-forward | port-forward | internal HTTP and `ListProducts` pass; public `LoadBalancer` remains blocked by org policy |

## Environment-specific notes

### `default`

- cluster can be created and used with the simpler public baseline
- `frontend-external` should receive a public IP
- public proof should check:
  - home page contains `Catalog-only mode is active`
  - product page contains `You May Also Like`

### `org-constrained`

- cluster remains private-node constrained
- `frontend-external` is expected to stay `EXTERNAL-IP <pending>`
- the supported proof path is:
  - frontend port-forward for HTTP
  - productcatalogservice port-forward for gRPC

Known blocker:

- org policy `constraints/compute.restrictLoadBalancerCreationForTypes`
  blocks the public external forwarding rule type needed by the current Service `LoadBalancer` path

## How to use this matrix

- choose the environment first
- use its recommended profile bundle and workspace
- run the operator flow in `docs/architecture/infra-operator-runbook.md`
- compare the observed result with the expected outcome here before assuming a regression
