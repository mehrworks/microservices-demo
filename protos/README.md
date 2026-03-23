## `protos/`

This directory is kept because `src/productcatalogservice/genproto/` is checked
in and still generated from `demo.proto`.

Only the proto definitions required by the current scaffold remain:

- `ProductCatalogService`
- the catalog request and response messages it serves
- shared `Money` and `Empty` messages used by that API

The upstream `hipstershop` package name is intentionally unchanged in this pass
to avoid turning cleanup into a wire-level API rename. If this scaffold is
later renamed into a different domain, the proto package and generated bindings
should be updated together in the same change.
