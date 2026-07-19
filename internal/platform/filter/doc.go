// Package filter provides transport-neutral parsing primitives for API list
// filters.
//
// A CELParser accepts a deliberately small, bounded CEL subset: scalar or map
// element equality, literal-list membership, and conjunction with &&. It
// returns an immutable Conjunction and never exposes cel-go AST or runtime
// values. Business modules configure the public fields and translate every
// predicate into their own typed repository query models. Repository adapters
// must not receive raw CEL expressions or generic field names.
package filter
