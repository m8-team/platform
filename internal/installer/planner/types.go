package planner

import (
	"time"

	installerv1alpha1 "github.com/m8platform/platform/api/installer/v1alpha1"
)

type InstallationPlan struct {
	APIVersion           string                    `json:"apiVersion" yaml:"apiVersion"`
	Kind                 string                    `json:"kind" yaml:"kind"`
	Metadata             PlanMetadata              `json:"metadata" yaml:"metadata"`
	Installation         InstallationRef           `json:"installation" yaml:"installation"`
	Release              ReleaseRef                `json:"release" yaml:"release"`
	ConfigDigest         string                    `json:"configDigest" yaml:"configDigest"`
	ReleaseCatalogDigest string                    `json:"releaseCatalogDigest" yaml:"releaseCatalogDigest"`
	Profile              installerv1alpha1.Profile `json:"profile" yaml:"profile"`
	Steps                []InstallationStep        `json:"steps" yaml:"steps"`
	Risks                []PlanRisk                `json:"risks,omitempty" yaml:"risks,omitempty"`
	IrreversibleActions  []string                  `json:"irreversibleActions,omitempty" yaml:"irreversibleActions,omitempty"`
	ResourceEstimate     ResourceEstimate          `json:"resourceEstimate,omitempty" yaml:"resourceEstimate,omitempty"`
}

type PlanMetadata struct {
	Name      string    `json:"name" yaml:"name"`
	CreatedAt time.Time `json:"createdAt" yaml:"createdAt"`
}

type InstallationRef struct {
	Name      string `json:"name" yaml:"name"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

type ReleaseRef struct {
	Version string `json:"version" yaml:"version"`
	Name    string `json:"name" yaml:"name"`
}

type InstallationStep struct {
	ID           string         `json:"id" yaml:"id"`
	Wave         int            `json:"wave" yaml:"wave"`
	Title        string         `json:"title" yaml:"title"`
	Phase        string         `json:"phase" yaml:"phase"`
	Dependencies []string       `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	ChangeSet    ChangeSet      `json:"changeSet" yaml:"changeSet"`
	Readiness    []string       `json:"readiness,omitempty" yaml:"readiness,omitempty"`
	Rollback     RollbackPolicy `json:"rollback" yaml:"rollback"`
}

type ChangeSet struct {
	Namespaces       []string                `json:"namespaces,omitempty" yaml:"namespaces,omitempty"`
	CRDs             []string                `json:"crds,omitempty" yaml:"crds,omitempty"`
	HelmReleases     []HelmReleaseChange     `json:"helmReleases,omitempty" yaml:"helmReleases,omitempty"`
	ArgoApplications []ArgoApplicationChange `json:"argoApplications,omitempty" yaml:"argoApplications,omitempty"`
	ExternalChecks   []string                `json:"externalChecks,omitempty" yaml:"externalChecks,omitempty"`
	Secrets          []string                `json:"secrets,omitempty" yaml:"secrets,omitempty"`
	Certificates     []string                `json:"certificates,omitempty" yaml:"certificates,omitempty"`
	Routes           []string                `json:"routes,omitempty" yaml:"routes,omitempty"`
	Policies         []string                `json:"policies,omitempty" yaml:"policies,omitempty"`
	Migrations       []string                `json:"migrations,omitempty" yaml:"migrations,omitempty"`
	SmokeTests       []string                `json:"smokeTests,omitempty" yaml:"smokeTests,omitempty"`
}

type HelmReleaseChange struct {
	Name      string `json:"name" yaml:"name"`
	Namespace string `json:"namespace" yaml:"namespace"`
	Chart     string `json:"chart" yaml:"chart"`
	Version   string `json:"version" yaml:"version"`
	Digest    string `json:"digest" yaml:"digest"`
}

type ArgoApplicationChange struct {
	Name      string `json:"name" yaml:"name"`
	Namespace string `json:"namespace" yaml:"namespace"`
	Path      string `json:"path" yaml:"path"`
	Wave      int    `json:"wave" yaml:"wave"`
}

type RollbackPolicy struct {
	Supported bool   `json:"supported" yaml:"supported"`
	Boundary  string `json:"boundary,omitempty" yaml:"boundary,omitempty"`
}

type PlanRisk struct {
	Severity string `json:"severity" yaml:"severity"`
	Message  string `json:"message" yaml:"message"`
}

type ResourceEstimate struct {
	CPU     string `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	Memory  string `json:"memory,omitempty" yaml:"memory,omitempty"`
	Storage string `json:"storage,omitempty" yaml:"storage,omitempty"`
}
