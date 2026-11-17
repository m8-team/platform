# ADR: Tenant-scoped ReferenceData and Pricing v1

## Status

Proposed for M8 Platform integration. Previously agreed Semantic Registry
business attributes are retained; tenant and HTTP API conventions are new.

## Context

Three Semantic Registry entities have been validated against the 1C
storage metadata:

| Entity | Context | 1C source |
| --- | --- | --- |
| Currency | ReferenceData | _Reference4 |
| PriceType | Pricing | _Reference50 |
| PriceRecord | Pricing | _InfoRg4139 |

Currency is shared reference data; PriceType and PriceRecord are
Pricing-owned. All APIs are tenant-scoped.

## Decision: ownership

- m8.platform.referencedata.v1 owns Currency and CurrencyService.
- m8.platform.pricing.v1 owns PriceType, PriceRecord and PricingService.
- Catalog owns Product / SKU; Resource Manager owns tenants; Access
  owns authorization. Cross-context references use UUIDs.

## Decision: tenant isolation

Every entity has a REQUIRED IMMUTABLE tenant_id UUID.
Every Get/List request carries tenant_id in its HTTP path.

The 1C sources do not contain tenant_id. The ingestion connector must
bind the source installation to a configured tenant. Persistence must
scope keys and lookups by (tenant_id, id), not by id alone. The reader
must derive caller authorization from authenticated identity, NOT
from user-supplied tenant_id. Foreign references and pagination tokens
must be scoped to the same tenant. Never leak whether a foreign-tenant
record exists. The current PR does not implement these runtime guards.

## Currency

Field order: id, tenant_id, state, code, name, label, description, version.

- id: _IDRRef
- state: _Marked, with additional INACTIVE business logic if available
- code: _Code (numeric ISO 4217 string, e.g. "643")
- name: _Fld304 (full name)
- label: _Description (short display form)
- description: _Fld305 (verbatim Russian spell-out parameters)
- version: _Version

CurrencyState is independent. _PredefinedID stays excluded.

## PriceType

Field order: id, tenant_id, state, code, name, description,
currency_id, department_id, version.

- id: _IDRRef; code: _Code; name: _Description;
  description: _Fld549; currency_id: _Fld542RRef;
  department_id: _Fld5445RRef; version: _Version.
- state: DELETED if _Marked, otherwise INACTIVE if _Fld5676,
  otherwise ACTIVE.
- Zero department references are modeled as an absent department_id.
- Other _Reference50 fields are excluded from the public contract.
- Organization association rows (_Reference50_VT6572) are relationships.

PriceTypeState is independent of CurrencyState.

## PriceRecord

Field order: id, tenant_id, state, product_id, sku_id, type_id,
currency_id, unit_id, price, discount, valid_from, document_id.

| Field | Source |
| --- | --- |
| id | UUID generated from namespaced register key |
| tenant_id | source connector binding |
| state | _Active + CDC deletes |
| product_id | _Fld4141RRef |
| sku_id | _Fld4142RRef, zero reference -> absent |
| type_id | _Fld4140RRef |
| currency_id | _Fld4143RRef |
| unit_id | _Fld4145RRef |
| price | _Fld4144 |
| discount | _Fld4146 |
| valid_from | _Period |
| document_id | _RecorderRRef |

The stable UUID is derived using a fixed versioned namespace from
tenant_id, source installation, _Period, _RecorderTRef, _RecorderRRef,
_LineNo. Source key bytes and the namespace MUST be consistent across
backfills and CDC ingestion. _RecorderTRef and _LineNo remain internal
Canonical CDC fields. The observed _RecorderTRef 0x00000099 maps to
_Document153. If other recorder types exist, they must be disambiguated.

ACTIVE refers to an active register movement and does not imply that
the price is currently applicable. PriceRecordState is independent.

For PriceCurrent and PriceHistory, the business grouping is:
(tenant_id, product_id, sku_id, type_id). Currency and unit are
attributes of each historical setting, not fetched from the current
PriceType or Product settings.

## Pricing formula

price is original RRP; discount is a percentage of RRP, e.g. 80.01.
selling_price = price * (1 - discount / 100).

With RRP 9999 and discount 80.01, the exact price is 1998.8001.
Use exact decimal arithmetic; rounding is deferred to an explicit
currency policy. The intended discount domain is [0,100].
Ingestion must quarantine values outside the supported range.

## Time

valid_from comes from the register business period _Period, not
Kafka event arrival time. The actual 1C source timezone must be
configured before conversion to protobuf Timestamp (UTC instant).
Retroactive changes and deletions require PriceHistory rebuilding.

## API and validation

Buf Protovalidate applies UUID formats, code formats, string length
limits, required Decimal/Timestamp messages and pagination boundaries.
google.api.field_behavior and google.api.http are used for annotations
and tenant-scoped read-only Get/List methods. Creation and modification
remain source-driven via CDC.

Contract-only changes do not make the HTTP or ConnectRPC endpoints
operational. Generated bindings, runtime authorization, ingestion,
storage adapters and deployment are follow-up implementation work.

## Consequences

Independent bounded contexts and tenant-safe design invariants.
More implementation work is necessary to enforce isolation in a
running service; Proto validation does not perform authorization.
