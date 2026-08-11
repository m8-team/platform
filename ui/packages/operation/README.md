# @m8/operation

Registered, Zod-validated command execution for M8 UI. The runtime composes
authorization, confirmation, audit and query-invalidation adapters and supports
AbortSignal. Declarative pages invoke stable operation IDs only; URLs, HTTP
methods and executable source are never operation spec fields.

Long-running definitions wait through `LongRunningOperationAdapter` and only
invalidate successful completion dependencies after `SUCCEEDED`. Audit and
cache-invalidation failures are reported as secondary-effect errors and never
turn an already successful mutation into a failed mutation.
