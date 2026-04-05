# frontend

Run the following command to restore dependencies to `vendor/` directory:

    dep ensure --vendor-only

## Catalog-only mode

This branch adds a thin public-edge mode for the frontend:

- `FRONTEND_MODE=full` keeps the original multi-service behavior
- `FRONTEND_MODE=catalog-only` reduces the required backend surface to
  `PRODUCT_CATALOG_SERVICE_ADDR`, while allowing `RECOMMENDATION_SERVICE_ADDR`
  as an optional browsing enhancement

In `catalog-only` mode, the frontend keeps product browsing active while cart,
checkout, ads, currency switching, and assistant flows stay disabled.
When `RECOMMENDATION_SERVICE_ADDR` is configured, product-page recommendations
remain available in that reduced browsing path.
