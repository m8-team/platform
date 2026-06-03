package health

import "time"

// Status is the normalized outcome of a health check or aggregated snapshot.
type Status string

const (
	// StatusUnknown means the check did not produce a known health state.
	StatusUnknown Status = "UNKNOWN"
	// StatusHealthy means the target is operating normally.
	StatusHealthy Status = "HEALTHY"
	// StatusDegraded means the target has a non-critical issue.
	StatusDegraded Status = "DEGRADED"
	// StatusUnhealthy means the target is not healthy enough for its criticality.
	StatusUnhealthy Status = "UNHEALTHY"
)

// Check binds immutable health metadata to the runtime checker implementation.
type Check struct {
	// Spec describes what is checked and how it participates in aggregation.
	Spec CheckSpec
	// Checker executes the runtime health check.
	Checker Checker
}

// CheckSpec describes a health check without its runtime executor.
type CheckSpec struct {
	// Name is the unique registry key for the check.
	Name string
	// Target identifies the application, module, service, or dependency being checked.
	Target Target
	// Kinds lists the probe categories this check participates in.
	Kinds []Kind
	// Criticality controls how failures affect aggregate status.
	Criticality Criticality
	// Timeout is the per-execution deadline for this check.
	Timeout time.Duration
	// Interval is metadata for future scheduled execution or caching.
	Interval time.Duration
}

// Kind identifies a probe category.
type Kind string

const (
	// KindLiveness is for process-local liveness checks.
	KindLiveness Kind = "LIVENESS"
	// KindReadiness is for checks that decide whether the service can receive traffic.
	KindReadiness Kind = "READINESS"
	// KindStartup is for checks that gate startup completion.
	KindStartup Kind = "STARTUP"
)

// Level controls how much health detail should be exposed by an adapter or API.
type Level string

const (
	// LevelSummary is intended for compact health output.
	LevelSummary Level = "SUMMARY"
	// LevelVerbose is intended for diagnostic health output.
	LevelVerbose Level = "VERBOSE"
)

// TargetKind classifies the thing being checked.
type TargetKind string

const (
	// TargetKindApplication identifies the whole application.
	TargetKindApplication TargetKind = "APPLICATION"
	// TargetKindModule identifies a platform or domain module.
	TargetKindModule TargetKind = "MODULE"
	// TargetKindService identifies a deployable service.
	TargetKindService TargetKind = "SERVICE"
	// TargetKindDependency identifies an external or infrastructure dependency.
	TargetKindDependency TargetKind = "DEPENDENCY"
)

// Target identifies the subject of a health check.
type Target struct {
	// Kind classifies the target.
	Kind TargetKind `json:"kind"`
	// Name is the target name within its kind.
	Name string `json:"name"`
	// Module optionally names the owning M8 module.
	Module string `json:"module,omitempty"`
}

// Criticality controls how an individual result contributes to aggregation.
type Criticality string

const (
	// CriticalityRequired failures make the aggregate status unhealthy.
	CriticalityRequired Criticality = "REQUIRED"
	// CriticalityOptional failures make the aggregate status degraded.
	CriticalityOptional Criticality = "OPTIONAL"
)

// Result is the normalized output of one check execution.
type Result struct {
	// Name is copied from the check spec.
	Name string `json:"name"`
	// Status is the normalized health outcome.
	Status Status `json:"status"`
	// Message is a human-readable diagnostic summary.
	Message string `json:"message,omitempty"`
	// Error is a safe error string for diagnostics.
	Error string `json:"error,omitempty"`
	// Latency is stored as milliseconds in registry-produced results.
	Latency time.Duration `json:"latency"`
	// CheckedAt is the UTC time when the check completed.
	CheckedAt time.Time `json:"checked_at"`
	// Target is copied from the check spec.
	Target Target `json:"target"`
	// Criticality is copied from the check spec.
	Criticality Criticality `json:"criticality"`
	// Metadata contains adapter-specific safe diagnostic values.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Snapshot is an aggregate health view for a single probe kind.
type Snapshot struct {
	// Status is the aggregate status of all included results.
	Status Status `json:"status"`
	// Kind is the probe category used to select checks.
	Kind Kind `json:"kind"`
	// CheckedAt is the UTC time when the snapshot was produced.
	CheckedAt time.Time `json:"checked_at"`
	// Results are sorted deterministically by check name.
	Results []Result `json:"results,omitempty"`
}
