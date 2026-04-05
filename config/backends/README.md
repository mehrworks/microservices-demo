# Backend Init Examples

This directory is reserved for future Terraform backend-init inputs.

- The new `infra/` stages still use local backends in this first pass.
- The `*.tfbackend.example` files here are placeholders only.
- Do not rename them to active `.tfbackend` files or wire them into automation
  until the target GCS state bucket and stable prefixes are real.
- Generated `*.tfbackend` files are kept out of git by `config/backends/.gitignore`.

Local helper once a real state bucket exists:

```bash
./docs/bootstrap/render-tfbackends.sh \
  --bucket YOUR_TF_STATE_BUCKET \
  --prefix-root microservices-demo
```

That writes local `config/backends/*.tfbackend` files from the tracked examples.

Planned usage after an intentional migration:

```bash
terraform -chdir=infra/3-gke init \
  -backend-config=../../config/backends/3-gke.tfbackend
```
