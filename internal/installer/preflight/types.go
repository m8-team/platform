package preflight

import (
	"context"
	"time"

	installerv1alpha1 "github.com/m8-team/platform/api/installer/v1alpha1"
)

type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

type Status string

const (
	StatusPass Status = "pass"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
	StatusSkip Status = "skip"
)

type Result struct {
	ID               string            `json:"id" yaml:"id"`
	Category         string            `json:"category" yaml:"category"`
	Severity         Severity          `json:"severity" yaml:"severity"`
	Status           Status            `json:"status" yaml:"status"`
	Message          string            `json:"message" yaml:"message"`
	Details          map[string]string `json:"details,omitempty" yaml:"details,omitempty"`
	Remediation      string            `json:"remediation,omitempty" yaml:"remediation,omitempty"`
	DocumentationRef string            `json:"documentationRef,omitempty" yaml:"documentationRef,omitempty"`
	Duration         time.Duration     `json:"durationNanos" yaml:"durationNanos"`
}

type Report struct {
	Installation string    `json:"installation" yaml:"installation"`
	Profile      string    `json:"profile" yaml:"profile"`
	CheckedAt    time.Time `json:"checkedAt" yaml:"checkedAt"`
	Results      []Result  `json:"results" yaml:"results"`
	Summary      Summary   `json:"summary" yaml:"summary"`
}

type Summary struct {
	Passed   int `json:"passed" yaml:"passed"`
	Warnings int `json:"warnings" yaml:"warnings"`
	Failed   int `json:"failed" yaml:"failed"`
	Skipped  int `json:"skipped" yaml:"skipped"`
}

type Check interface {
	ID() string
	Run(ctx context.Context, installation installerv1alpha1.PlatformInstallation) Result
}

type ClusterReader interface {
	ServerVersion(ctx context.Context) (string, error)
	NodeSummary(ctx context.Context) (NodeSummary, error)
	HasAPIResource(ctx context.Context, groupVersion string, kind string) (bool, error)
	StorageClasses(ctx context.Context) ([]string, error)
}

type NodeSummary struct {
	Total         int
	Ready         int
	Architectures map[string]int
	Zones         map[string]int
}
