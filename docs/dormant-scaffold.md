# Dormant scaffold guide

> Branch context: `spike/dormant-scaffold-v2`

This branch keeps the original multi-service repository shape intact as a **reference scaffold**, while narrowing the **active loop** to `productcatalogservice` only.

The goal is to preserve the old service bodies, deployment layout, CI history, and release knowledge so services can be reactivated or replaced gradually later, without prematurely rewriting the whole repo around one service.

## Current operating model

### Active right now

Only `productcatalogservice` is active in the current loop.

That currently means:

- `skaffold.yaml`
  - only `productcatalogservice` is active
  - other service artifacts remain in place as commented history
- `kubernetes-manifests/kustomization.yaml`
  - only `productcatalogservice.yaml` is active
- `kustomize/base/kustomization.yaml`
  - only `productcatalogservice.yaml` is active
- `release/kubernetes-manifests.yaml`
  - narrowed so the upstream-style `kubectl apply -f ./release/kubernetes-manifests.yaml` path still works for the active service only
- `.github/workflows/ci-main.yaml`
  - only `productcatalogservice` is tested/deployed/waited on
  - older full-app flow is preserved as comments
- `.github/workflows/ci-pr.yaml`
  - same idea for PR flow

### Runtime policy

For the active service on this branch, runtime behavior is controlled at the **manifest layer**, not by rewriting vendor source semantics.

Current manifest policy:

- profiler is off via `DISABLE_PROFILER=1`
- tracing is off because `ENABLE_TRACING` is omitted
- AlloyDB remains dormant unless its environment is explicitly wired in later

## Dormant but intentionally retained

These surfaces are still present and still mostly describe the full original app:

- service directories under `src/`
- `helm-chart/`
- `release/`
- `cloudbuild.yaml`
- most of `docs/`
- `.github/release-cluster/`
- Kustomize components such as:
  - network policies
  - AlloyDB
  - service mesh / Istio
  - shopping assistant
  - without-loadgenerator

These are not mistakes. They are preserved on purpose as reference material and future replacement scaffolding.

## Reactivation rule

When reactivating an existing service or adding a new active one, update the active surfaces **together** in one coherent pass.

Minimum checklist:

1. `skaffold.yaml`
2. `kubernetes-manifests/kustomization.yaml`
3. `kustomize/base/kustomization.yaml`
4. `.github/workflows/ci-main.yaml`
5. `.github/workflows/ci-pr.yaml`
6. any service-specific waits / smoke tests / frontend assumptions
7. any branch notes that describe the active loop

Prefer **commenting services in or out** rather than deleting their surrounding history.

## What not to do

- Do not silently reactivate a service in only one file.
- Do not change vendor `src/` semantics just to enforce branch policy if the same behavior can live in manifests or CI.
- Do not delete the old service bodies unless there is a deliberate fork away from the vendor scaffold model.
- Do not assume full-app release docs describe the current active path.

## Recommended next step

The next meaningful step should be app-directed, not more repo-wide cleanup:

- gradually adapt `productcatalogservice`, or
- add a new service beside it when there is a real use case

Try to keep future changes localized and intentional rather than repeating another large scaffold rewrite.
