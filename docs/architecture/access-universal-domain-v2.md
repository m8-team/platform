# M8 Universal Access — Domain Model v2

**Status: PROPOSED.** Implementation increment 1: catalog/domain/application ports with unit tests. No production traffic switched.

## Ownership

- Source modules own object creation, resource identity, business lifecycle, tenant/owner/hierarchy and current revision. Access stores protected resource references and authorization metadata only.
- Identity/Authentication owns principals and assurance. Access stores subject references.
- Access owns authorization manifests, definitions, reusable templates, composite profiles, assignments, grants, restrictions, publication revisions and policy decisions.
- SpiceDB is a derived graph; it is not the authoritative store for grants or role bindings.

## Module publication

Each registered M8 module may declare namespaced resource types (finance.invoice), actions (finance.invoice.view) and role templates (finance.accountant@1.0.0). Resource types carry the authorized scope kinds. Registration of finance.invoice does NOT create a particular invoice; an instance must be resolved through Finance/Resource Manager before a protected operation.

The app/catalog service checks a trusted authenticated service-account principal, matches the manifest publisher, and separately authorizes namespace ownership. The ManifestRepository port requires transactional compare-and-swap, idempotency and immutable release versions when implemented. Never allow a client to self-assert service identity or module ownership.

## Role composition

A composite Bundle is a version-pinned set of role-template references and nested bundles. For example, profiles.accountant@1.0.0 can include profiles.employee@1.0.0 and finance.accountant@2.0.0. Expansion fails on missing nested bundle versions and cycles. Before publication, the application must also prove that every resulting role template exists in the published catalog and supports the requested scope. Editing an existing published profile must be disallowed; instead publish a new version with explicit migration.

One Assignment references a single version of a role template or profile and one constrained Scope. An Assignment has its own UUID, tenant, revision and independent revocation. Multiple overlapping assignments or grants remain independently revocable.

## Scoped permissions

- INSTANCE: one particular resource, exact ID match.
- COLLECTION: bounded explicit list of instances of one resource type.
- ORGANIZATION: resources of one type belonging to one organization, as verified by the owner.
- SUBTREE: resource and descendants resolved from the authoritative owner, not inferred from SpiceDB.
- RESOURCE_KIND: all resources of one exact resource type in a tenant.
- TENANT: tenant-wide access to one explicit registered resource type. Never implicitly cover newly published module types.

These are authorization scopes, NOT Resource Manager OwnerScope values. Do not interpret Project as a supported ownership scope before its ADR and contract migration. All matching must be based on trusted owner-resolved identity, tenant, lifecycle and revision. Resource-owning services must revalidate revisions in the actual write CAS to avoid PDP/check-use races.

## Context policies

Policy conditions are typed values rather than client-supplied executable expressions: network IP/CIDR, verified MFA, an absolute deadline, and a verified resource attribute. EvaluationContext must be sourced from authenticated middleware and authoritative resource data; constructors alone do NOT authenticate external claims.

A policy has either REQUIRE (must match) or DENY_WHEN (must not match) effects. Neither grants an entitlement by itself.

Target decision formula:

Permit = verified principal AND live resource AND applicable entitlement for action + scope AND all mandatory REQUIRE policies AND no matching DENY_WHEN policy AND certified consistency barriers.

An IP-limited grant does NOT cancel another unconditional broad role. To require a corporate IP for an operation, publish an applicable mandatory policy covering all matching entitlements. Unknown IP, missing MFA proof or missing business attributes fail closed. Trust forwarded IP only from a configured and authenticated proxy boundary.

## Compatibility and projection

Existing production Access continues to use the v1 domain types; the new domain packages are not yet wired to the production PDP. The present schema file internal/access/schema/m8.zed has a fixed set of resource kinds. Publishing a namespaced kind does not automatically extend SpiceDB permissions.

Two candidate graph implementations require a separate reviewed ADR:

1. Generic capability objects with stable graph permission use and tenant/resource/action keyed IDs, retaining distinct binding/grant identity for independent revocations.
2. Generated specialized SpiceDB schema for proven high-volume modules after lifecycle, schema-evolution and load tests.

Keep distributed fencing, YDB checkpoint/ZedToken, tenant-wide policy journal barrier, owner revision barrier and reconciliation; do not bypass them for dynamic kinds or conditional policies. Benchmark object/action cardinality, lookup latency and 2-replica worker failover before production rollout.

## Remaining milestones (not implemented here)

1. Contract-first Protobuf for PublishModuleManifest, resource/permission discovery, publish bundles, assignments, policy management and context-aware AuthZEN evaluation; Buf lint/breaking review and generated implementations.
2. YDB authoritative catalog/template/bundle/assignment/policy repositories; transactional outbox, idempotent publication journal and policy revision fanout.
3. Trusted resource resolver, live module namespace bindings and policy discovery; no client-origin context.
4. SpiceDB generic graph schema/worker and reconciliation, or reviewed generated schema; complete authorization behavior tests.
5. Owner CAS integration, policy freshness and revocation tests; full migration from v1 domain consumers.
6. Resolve existing public GrantState vs domain APPROVED mismatch via reviewed contract change.

**Security gate:** Unit tests of the new domain types cannot be interpreted as production enforcement. Until all required trusted resolvers, policy composition and graph projections are wired, retain the v1 deny-by-default production behavior.
