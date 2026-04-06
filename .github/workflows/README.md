# GitHub Actions Workflows

> **Dormant scaffold note**
>
> This branch keeps the original multi-service repo shape as reference, but only
> `frontend` (catalog-only mode), `productcatalogservice`, and
> `recommendationservice` are active in the
> current CI/deploy loop. The older full-app workflow steps remain in the
> workflow YAML files as commented history. When you add or remove active
> services later, update the workflow files together with `skaffold.yaml`,
> `kubernetes-manifests/kustomization.yaml`, and `kustomize/base/kustomization.yaml`.

This page describes the CI/CD workflows for the Online Boutique app, which run in [Github Actions](https://github.com/GoogleCloudPlatform/microservices-demo/actions).

## Current branch delta

On `spike/dormant-scaffold-v2`, treat the rest of this document as **historical full-app workflow reference**. The current active branch behavior is narrower:

- `frontend`, `productcatalogservice`, and `recommendationservice` are active in CI/deploy waits
- CI smoke tests now verify the catalog-only frontend via port-forward
- full-app staging/comment behavior stays preserved in workflow YAML as commented history
- infra/config contract changes now have a separate non-GKE validation lane in `infra-contracts.yaml`
- the canonical branch guide is [`docs/dormant-scaffold.md`](/docs/dormant-scaffold.md)

## Infrastructure

The CI/CD pipelines for Online Boutique run in Github Actions, using a pool of two [self-hosted runners]((https://help.github.com/en/actions/automating-your-workflow-with-github-actions/about-self-hosted-runners)). These runners are GCE instances (virtual machines) that, for every open Pull Request in the repo, run the code test pipeline, deploy test pipeline, and (on main) deploy the latest version of the app to [cymbal-shops.retail.cymbal.dev](https://cymbal-shops.retail.cymbal.dev)

We also host a test GKE cluster, which is where the deploy tests run. Every PR has its own namespace in the cluster.

## Workflows

**Note**: In order for the current CI/CD setup to work on your pull request, you must branch directly off the repo (no forks). This is because the Github secrets necessary for these tests aren't copied over when you fork.

### Code Tests - [ci-pr.yaml](ci-pr.yaml)

These tests run on every commit for every open PR, as well as any commit to main / any release branch. Currently, this workflow runs only Go unit tests.

### Infra Contract Checks - [infra-contracts.yaml](infra-contracts.yaml)

This workflow runs only for staged infra/config changes and intentionally avoids
real GKE deployment work. It validates the repo-shape lane by running:

1. profile-bundle validation
2. hybrid contract artifact validation
3. `go test ./...` in `src/frontend/jobcontrol`
4. `go test ./...` in `tools/jobcontrolref`
5. `go test ./...` in `tools/jobcontrolworker`
6. `terraform fmt -check -recursive infra`
7. `terraform init -backend=false -input=false`
8. `terraform validate`
9. `terraform plan -input=false -lock=false -no-color`

for `infra/1-project`, `infra/2-iam`, `infra/3-gke`, and `infra/4-ci`.

The local equivalent is `docs/bootstrap/validate-infra-contracts.sh`, and the
default operator wrapper is `docs/bootstrap/staged-infra-cycle.sh`.


### Deploy Tests- [ci-pr.yaml](ci-pr.yaml)

These tests run on every commit for every open PR, as well as any commit to main / any release branch. This workflow:

1. Creates a dedicated GKE namespace for that PR, if it doesn't already exist, in the PR GKE cluster.
2. Uses `skaffold run` to build and push the images specific to that PR commit. Then skaffold deploys those images, via `kubernetes-manifests`, to the PR namespace in the test cluster.
3. Tests to make sure the active pods start up and become ready.
4. Port-forwards the internal `frontend` service, verifies the catalog-only home page loads, and checks that product-page recommendations render.

### Push and Deploy Latest - [push-deploy](push-deploy.yml)

This is the Continuous Deployment workflow, and it runs on every commit to the main branch. This workflow:

1. Builds the container images for every service, tagging as `latest`.
2. Pushes those images to Google Container Registry.

Note that this workflow does not update the image tags used in `release/kubernetes-manifests.yaml` - these release manifests are tied to a stable `v0.x.x` release.

### Cleanup - [cleanup.yaml](cleanup.yaml)

This workflow runs when a PR closes, regardless of whether it was merged into main. This workflow deletes the PR-specific GKE namespace in the test cluster.

## Appendix - Creating a new Actions runner

Should one of the two self-hosted Github Actions runners (GCE instances) fail, or you want to add more runner capacity, this is how to provision a new runner. Note that you need IAM access to the admin Online Boutique GCP project in order to do this.

1. Create a GCE instance.
    - VM should be at least n1-standard-4 with 50GB persistent disk
    - VM should use custom service account with permissions to: access a GKE cluster, create GCS storage buckets, and push to GCR.
2. SSH into new VM through the Google Cloud Console.
3. Install project-specific dependencies, including go, docker, skaffold, and kubectl:

```
wget -O - https://raw.githubusercontent.com/GoogleCloudPlatform/microservices-demo/main/.github/workflows/install-dependencies.sh | bash
```

The instance will restart when the script completes in order to finish the Docker install.

4. SSH back into the VM.

5. Follow the instructions to add a new runner on the [Actions Settings page](https://github.com/GoogleCloudPlatform/microservices-demo/settings/actions) to authenticate the new runner
6. Start GitHub Actions as a background service:
```
sudo ~/actions-runner/svc.sh install ; sudo ~/actions-runner/svc.sh start
```
