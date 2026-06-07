package health

import "time"

type Status string

const (
	StatusUnknown   Status = "UNKNOWN"
	StatusHealthy   Status = "HEALTHY"
	StatusDegraded  Status = "DEGRADED"
	StatusUnhealthy Status = "UNHEALTHY"
)

type Check struct {
	Spec    Config
	Checker Checker
}

type Config struct {
	Name        string
	Target      Target
	Kinds       []Kind
	Criticality Criticality
	Timeout     time.Duration
	Interval    time.Duration
}

type Kind string

const (
	KindLiveness  Kind = "LIVENESS"
	KindReadiness Kind = "READINESS"
	KindStartup   Kind = "STARTUP"
)

type Level string

const (
	LevelSummary Level = "SUMMARY"
	LevelVerbose Level = "VERBOSE"
)

type TargetKind string

const (
	TargetKindApplication TargetKind = "APPLICATION"
	TargetKindModule      TargetKind = "MODULE"
	TargetKindService     TargetKind = "SERVICE"
	TargetKindDependency  TargetKind = "DEPENDENCY"
)

type Target struct {
	Kind   TargetKind `json:"kind"`
	Name   string     `json:"name"`
	Module string     `json:"module,omitempty"`
}

type Criticality string

const (
	CriticalityRequired Criticality = "REQUIRED"
	CriticalityOptional Criticality = "OPTIONAL"
)

type Result struct {
	Name        string            `json:"name"`
	Status      Status            `json:"status"`
	Message     string            `json:"message,omitempty"`
	Error       string            `json:"error,omitempty"`
	Latency     time.Duration     `json:"latency"`
	CheckedAt   time.Time         `json:"checked_at"`
	Target      Target            `json:"target"`
	Criticality Criticality       `json:"criticality"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type Snapshot struct {
	Status    Status    `json:"status"`
	Kind      Kind      `json:"kind"`
	CheckedAt time.Time `json:"checked_at"`
	Results   []Result  `json:"results,omitempty"`
}
