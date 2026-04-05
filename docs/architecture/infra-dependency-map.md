# Infra Dependency Map

This document describes the practical staged-infra topology that is active in the
repository now.

## Active stage chain

The first-pass chain is:

`infra/1-project` -> `infra/2-iam` -> `infra/3-gke`

with `infra/4-ci` reserved as a future contract lane.

## What each stage owns

- `infra/1-project`
  - project adoption
  - minimum API enablement
- `infra/2-iam`
  - optional deploy/runtime identities
  - optional project-role bindings
- `infra/3-gke`
  - Artifact Registry repo contract
  - optional repo IAM bindings
  - optional custom network/subnet contract
  - optional Autopilot cluster creation
  - hardening placeholders for private nodes, control-plane CIDR, NAT, and public LB policy
- `infra/4-ci`
  - placeholder-only CI trigger contract for later work

## Current dependency graph

```mermaid
flowchart TB
  classDef active fill:#e8f5e9,stroke:#2e7d32,color:#1b1b1b;
  classDef config fill:#fff8e1,stroke:#b26a00,color:#1b1b1b;
  classDef future fill:#f5f5f5,stroke:#9e9e9e,color:#424242,stroke-dasharray: 5 5;

  PCFG[config/datasets/project]
  ICFG[config/datasets/iam]
  GCFG[config/datasets/gke]
  CCFG[config/datasets/ci]

  P1[infra/1-project]
  I2[infra/2-iam]
  G3[infra/3-gke]
  C4[infra/4-ci]

  PCFG --> P1
  ICFG --> I2
  GCFG --> G3
  CCFG --> C4

  P1 --> I2 --> G3
  I2 -. identity outputs later .-> G3
  G3 -. cluster / repo contract later .-> C4

  PCFG:::config
  ICFG:::config
  GCFG:::config
  CCFG:::config
  P1:::active
  I2:::active
  G3:::active
  C4:::future
```

## Placeholder hardening path

The repo now has a defined place for stricter environment requirements, but those
settings remain off by default:

- `network.*`
- `hardening.private_nodes_enabled`
- `hardening.master_ipv4_cidr`
- `hardening.cloud_nat_enabled`
- `hardening.public_load_balancer_enabled`

That means the repo can describe org-managed environments without forcing those
concerns into the default sandbox path.

## CFF alignment

These stage boundaries intentionally stay close to the kinds of concerns split by
Cloud Foundation Fabric modules, especially:

- project
- iam-service-account
- artifact-registry
- gke-cluster-autopilot
- net-cloudnat

The repo uses that as structure inspiration first, while keeping direct module
adoption as a later optimization choice.
