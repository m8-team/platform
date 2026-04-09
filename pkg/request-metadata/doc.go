// Package requestmetadata defines reusable request-scoped operational metadata
// for M8 Platform services.
//
// The package models metadata such as actor identity, correlation identifiers,
// request identifiers, idempotency keys, and the request source. This metadata
// is operational context, not aggregate state, and must not be persisted as
// business data inside domain entities.
//
// The package is transport-agnostic and reusable across M8 Platform services.
// HTTP, gRPC, CLI, and worker adapters should be built around this package
// rather than embedded into it.
package requestmetadata
