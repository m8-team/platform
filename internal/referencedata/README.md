# M8 ReferenceData

## Responsibility

Tenant-scoped reference directories shared across business contexts.

## Owns

- Currency and CurrencyState.
- CurrencyService GetCurrency / ListCurrencies contracts.

## Does not own

- Price types and historical price records (Pricing).
- Currency exchange-rate calculation (future separate capability).
- Tenant lifecycle (Resource Manager) or authorization (Access).

## Source

The initial directory projects the 1C _Reference4 catalog. Tenant identity
is added by the source-to-tenant binding during ingestion, never read from
1C itself.

## Deployment status

This change defines the bounded-context contracts only. An authenticated
read adapter, persistence model, CDC integration and API service executable
are needed before the endpoints can receive traffic. Request tenant_id alone
is never evidence of authorization.
