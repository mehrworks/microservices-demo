# `productcatalogservice-only`

This directory provides an additive single-service deployment path for
`productcatalogservice`.

Use it with [`/skaffold.productcatalogservice.yaml`](/skaffold.productcatalogservice.yaml)
to build and deploy only `productcatalogservice` while keeping the original
multi-service manifests in [`/kubernetes-manifests`](/kubernetes-manifests)
unchanged.
