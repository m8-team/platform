# M8 Pricing

## Responsibility

Tenant-scoped price-type directory, historical price settings and pricing
business rules.

## Owns

- PriceType, PriceRecord and their separate state enums.
- PricingService Get/List contracts.
- Exact discount-based selling-price calculation.
- Future PriceCurrent and PriceHistory projections.

## Does not own

- Currency (ReferenceData).
- Products and SKUs (Catalog).
- Tenant hierarchy (Resource Manager).
- The original 1C database schema.

## Source and semantics

- PriceType: _Reference50.
- PriceRecord: _InfoRg4139.
- Document: _Document153 for observed recorder type 0x00000099.

Selling price = price * (1 - discount / 100).
Price is original RRP; discount is a percentage, not a price multiplier.
Use exact math without implicit monetary rounding.

## Isolation requirements

Every stored row uses a (tenant_id, id) key. Every request is scoped to a
tenant_id authorized from caller identity. Cross-context UUID references
must resolve within the same tenant. Pagination tokens must be tenant
bound. Proto validation is not a substitute for authorization.

## Deployment status

The change introduces Proto contracts plus a small pure Go domain
calculation with unit tests. API handlers, CDC adapters, persistence,
generated client artifacts and production deployment are later steps.
