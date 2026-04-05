# App To Infra Map

This document shows the current app-first delivery path and where the new staged
infra lane fits around it.

## Active runtime path

The current active app slice remains:

`src/frontend` + `src/productcatalogservice` -> `skaffold.yaml` -> `kubernetes-manifests/` / `kustomize/` -> GKE cluster

The existing proof bind for that runtime path stays under `config/gke-exposure/`.

## New surrounding infra path

The new staged infra lane is:

`config/datasets/project` -> `infra/1-project`
`config/datasets/iam` -> `infra/2-iam`
`config/datasets/gke` -> `infra/3-gke`
`config/datasets/ci` -> `infra/4-ci`

Those stages define the cloud-side contract around the app delivery path, but do
not deploy the workloads themselves.

## Current map

```mermaid
flowchart LR
  classDef app fill:#e8f5e9,stroke:#2e7d32,color:#1b1b1b;
  classDef infra fill:#e3f2fd,stroke:#1565c0,color:#1b1b1b;
  classDef config fill:#fff8e1,stroke:#b26a00,color:#1b1b1b;
  classDef future fill:#f5f5f5,stroke:#9e9e9e,color:#424242,stroke-dasharray: 5 5;

  subgraph app_layer[App / Delivery]
    FE[src/frontend]
    PC[src/productcatalogservice]
    SK[skaffold.yaml]
    MAN[kubernetes-manifests / kustomize]
    BIND[config/gke-exposure]
  end

  subgraph infra_layer[Infra / Contracts]
    PCFG[config/datasets/project]
    ICFG[config/datasets/iam]
    GCFG[config/datasets/gke]
    CCFG[config/datasets/ci]
    P1[infra/1-project]
    I2[infra/2-iam]
    G3[infra/3-gke]
    C4[infra/4-ci]
  end

  subgraph runtime[Runtime]
    CLUSTER[GKE cluster]
    AR[Artifact Registry repo]
  end

  FE --> SK
  PC --> SK
  BIND --> SK
  SK --> MAN --> CLUSTER

  PCFG --> P1
  ICFG --> I2
  GCFG --> G3
  CCFG --> C4

  P1 --> I2 --> G3
  G3 --> CLUSTER
  G3 --> AR
  C4 -. later CI trigger plumbing .-> SK

  FE:::app
  PC:::app
  SK:::app
  MAN:::app
  BIND:::config
  PCFG:::config
  ICFG:::config
  GCFG:::config
  CCFG:::config
  P1:::infra
  I2:::infra
  G3:::infra
  C4:::future
  CLUSTER:::infra
  AR:::infra
```

## Selective CFF inspiration

The stage boundaries intentionally line up with small Cloud Foundation Fabric
module families without importing FAST/landing-zone behavior into the app repo:

- `infra/1-project` aligns with project/API enablement concerns
- `infra/2-iam` aligns with service-account and IAM contract concerns
- `infra/3-gke` aligns with Artifact Registry, Autopilot cluster, and later NAT/network placeholders

This keeps the repo compatible with stricter environments later while staying
small and app-delivery-first now.

For the human-run stage order that corresponds to this map, use
`docs/architecture/infra-manual-walkthrough.md`.
