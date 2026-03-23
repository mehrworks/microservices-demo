# Product Catalog Kustomize Base

This directory now contains a minimal Kustomize path for the
`productcatalogservice`-only scaffold on `spike/single-service-scaffold`.

## Render or apply

From the repository root:

```bash
kubectl kustomize kustomize
kubectl apply -k kustomize
```

The rendered output contains a single Deployment, Service, and ServiceAccount
for `productcatalogservice`.
