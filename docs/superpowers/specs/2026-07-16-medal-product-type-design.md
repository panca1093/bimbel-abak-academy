# Medal Product Type Design

## Context

PR #36 adds `merchandise` as a generic physical product type. Client feedback requires medals to be a distinct product type, while retaining the same physical-product lifecycle. Migration `0027` has not run in shared or staging environments, so it can be amended in place.

## Decision

Add `medal` as a first-class `product.type`. It is not a merchandise subtype and is visible independently in product management and the catalog.

`book`, `merchandise`, and `medal` are physical types. Each has stock, optional weight and image, checkout stock decrement, paid-order processing, required shipping before completion, and the same admin-store permissions.

UI labels are `Medali` in Indonesian and `Medal` in English.

## Design

### Persistence and backend

- Amend migration `0027` so the product CHECK allows `book`, `course`, `exam`, `merchandise`, and `medal`.
- Its down migration converts both `merchandise` and `medal` rows to `book` before restoring the original CHECK.
- Extend the existing physical-type checks in service, repository, and worker paths to include `medal`.
- Add the `Medal` Midtrans item category and permit `admin_store` to manage medals.

### Frontend

- Add `medal` to `ProductType`, product input shapes, admin type dropdown, filters, badges, gradients, catalog metadata, and image rendering mappings.
- Treat medal as physical in the product form, product detail, and admin orders UI.
- Add `product_type_medal` translations: `Medali` (ID) and `Medal` (EN).

## Error handling and compatibility

- No new API endpoint or schema table is introduced.
- Existing books, merchandise, courses, and exams are unchanged.
- Since migration `0027` has not been deployed, no data conversion or compatibility migration is needed on the up path.
- Rollback is safe because both new physical types map to the pre-existing physical `book` type.

## Validation

- Migration tests cover accepting `medal`, rejecting unknown types, and safe rollback.
- Backend tests cover medal RBAC, zero-stock add-to-cart rejection, checkout decrement and insufficient-stock behavior, paid-order ship-pending behavior, and ship-before-complete enforcement.
- Frontend tests cover medal selection/physical fields, labels, catalog rendering, and the admin shipping lifecycle.
- Run the existing backend build, vet, race/shuffle suite; frontend Vitest; and production build before pushing the amended branch.

## Out of scope

- Medal-specific variants, engraving fields, pricing rules, fulfillment workflow, or a category/subtype model.
- Changes to existing product behavior beyond recognizing medal as a physical product.
