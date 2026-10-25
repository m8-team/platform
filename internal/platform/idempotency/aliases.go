package idempotency

import "github.com/m8-team/platform/internal/platform/idempotency/domain"

// Keep the historical package-level Store interface compiling by sharing the
// canonical domain data types. This is a type alias, not a parallel data model.
type Record = domain.Record
type Result = domain.Result
