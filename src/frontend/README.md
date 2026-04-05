# frontend

Run the following command to restore dependencies to `vendor/` directory:

    dep ensure --vendor-only

## Catalog-only mode

This branch adds a thin public-edge mode for the frontend:

- `FRONTEND_MODE=full` keeps the original multi-service behavior
- `FRONTEND_MODE=catalog-only` reduces the required backend surface to
  `PRODUCT_CATALOG_SERVICE_ADDR` only

In `catalog-only` mode, the frontend keeps product browsing active while cart,
checkout, recommendations, ads, currency switching, and assistant flows stay
disabled.
