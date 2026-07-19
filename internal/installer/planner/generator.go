package planner

import (
	"context"
	"fmt"
	"time"

	installerv1alpha1 "github.com/m8-team/platform/api/installer/v1alpha1"
	"github.com/m8-team/platform/internal/installer/config"
	"github.com/m8-team/platform/internal/installer/graph"
)

type GenerateInput struct {
	Installation         installerv1alpha1.PlatformInstallation
	Release              installerv1alpha1.PlatformRelease
	ConfigDigest         string
	ReleaseCatalogDigest string
	CreatedAt            time.Time
}

func Generate(ctx context.Context, input GenerateInput) (InstallationPlan, error) {
	if err := ctx.Err(); err != nil {
		return InstallationPlan{}, err
	}
	installation := input.Installation.Defaulted()
	if err := installation.Validate(); err != nil {
		return InstallationPlan{}, err
	}
	if err := input.Release.Validate(); err != nil {
		return InstallationPlan{}, err
	}
	if input.ConfigDigest == "" {
		digest, err := config.DigestObject(installation)
		if err != nil {
			return InstallationPlan{}, err
		}
		input.ConfigDigest = digest
	}
	if input.ReleaseCatalogDigest == "" {
		digest, err := config.DigestObject(input.Release)
		if err != nil {
			return InstallationPlan{}, err
		}
		input.ReleaseCatalogDigest = digest
	}
	if input.CreatedAt.IsZero() {
		input.CreatedAt = time.Now().UTC()
	}

	steps := buildSteps(installation, input.Release)
	nodes := make([]graph.Node, 0, len(steps))
	stepByID := make(map[string]InstallationStep, len(steps))
	for _, step := range steps {
		nodes = append(nodes, graph.Node{ID: step.ID, Wave: step.Wave, Dependencies: step.Dependencies})
		stepByID[step.ID] = step
	}

	dependencyGraph, err := graph.New(nodes)
	if err != nil {
		return InstallationPlan{}, err
	}
	orderedNodes, err := dependencyGraph.Topological()
	if err != nil {
		return InstallationPlan{}, err
	}

	orderedSteps := make([]InstallationStep, 0, len(orderedNodes))
	for _, node := range orderedNodes {
		orderedSteps = append(orderedSteps, stepByID[node.ID])
	}

	return InstallationPlan{
		APIVersion:           installerv1alpha1.GroupName + "/" + installerv1alpha1.Version,
		Kind:                 "InstallationPlan",
		Metadata:             PlanMetadata{Name: installation.Name + "-plan", CreatedAt: input.CreatedAt},
		Installation:         InstallationRef{Name: installation.Name, Namespace: installation.Namespace},
		Release:              ReleaseRef{Version: installation.Spec.PlatformVersion, Name: input.Release.Name},
		ConfigDigest:         input.ConfigDigest,
		ReleaseCatalogDigest: input.ReleaseCatalogDigest,
		Profile:              installation.Spec.Profile,
		Steps:                orderedSteps,
		Risks:                planRisks(installation),
		IrreversibleActions:  irreversibleActions(installation),
		ResourceEstimate:     resourceEstimate(input.Release),
	}, nil
}

func buildSteps(installation installerv1alpha1.PlatformInstallation, release installerv1alpha1.PlatformRelease) []InstallationStep {
	spec := installation.Spec
	steps := []InstallationStep{
		{
			ID:    "crds",
			Wave:  -100,
			Title: "Install required CRDs",
			Phase: "bootstrap",
			ChangeSet: ChangeSet{
				CRDs: []string{
					"gateway.networking.k8s.io",
					"cert-manager.io",
					"trust.cert-manager.io",
					"argoproj.io",
					"external-secrets.io",
					"installer.m8.io",
				},
			},
			Readiness: []string{"CRD Established conditions are true"},
			Rollback:  RollbackPolicy{Supported: false, Boundary: "CRD deletion is manual in M8 Installer 1.0"},
		},
		{
			ID:           "namespaces",
			Wave:         -90,
			Title:        "Create platform namespaces",
			Phase:        "bootstrap",
			Dependencies: []string{"crds"},
			ChangeSet: ChangeSet{
				Namespaces: []string{
					"m8-system",
					"m8-security",
					"m8-data",
					"m8-observability",
					"m8-gateway",
					spec.GitOps.ArgoCD.Namespace,
				},
			},
			Readiness: []string{"Namespaces exist with Pod Security labels"},
			Rollback:  RollbackPolicy{Supported: true},
		},
	}

	if spec.Certificates.CertManager.Enabled || spec.Trust.TrustManager.Enabled || spec.Trust.SPIRE.Enabled {
		steps = append(steps, InstallationStep{
			ID:           "pki-trust",
			Wave:         -80,
			Title:        "Install PKI and trust components",
			Phase:        "bootstrap",
			Dependencies: []string{"namespaces"},
			ChangeSet: ChangeSet{
				HelmReleases: releaseHelmChanges(release, "cert-manager", "trust-manager", "spire"),
				Certificates: []string{"internal-ca", "cluster-issuer", "trust-bundle"},
			},
			Readiness: []string{"cert-manager webhook ready", "trust-manager bundle synced", "SPIRE server ready when enabled"},
			Rollback:  RollbackPolicy{Supported: true},
		})
	}

	steps = append(steps, InstallationStep{
		ID:           "security-secrets",
		Wave:         -70,
		Title:        "Install security and secrets operators",
		Phase:        "bootstrap",
		Dependencies: dependenciesWhenPresent(steps, "pki-trust", "namespaces"),
		ChangeSet: ChangeSet{
			HelmReleases: releaseHelmChanges(release, "external-secrets-operator", "kyverno", "trivy-operator"),
			Secrets:      []string{"external secret stores", "initial Argo CD repository credential references"},
			Policies:     []string{"Pod Security Standards", "supply-chain policies", "trusted registry policies"},
		},
		Readiness: []string{"External Secrets webhook ready", "Kyverno admission ready when enabled"},
		Rollback:  RollbackPolicy{Supported: true},
	})

	if spec.Network.Cilium.Enabled {
		steps = append(steps, InstallationStep{
			ID:           "cilium",
			Wave:         -65,
			Title:        "Install or adopt Cilium networking",
			Phase:        "bootstrap",
			Dependencies: []string{"namespaces"},
			ChangeSet: ChangeSet{
				HelmReleases: releaseHelmChanges(release, "cilium"),
				Policies:     []string{"NetworkPolicy baseline", "DNS policies"},
			},
			Readiness: []string{"Cilium nodes ready", "Hubble relay ready when enabled"},
			Rollback:  RollbackPolicy{Supported: false, Boundary: "CNI rollback requires cluster-specific network recovery"},
		})
	}

	steps = append(steps, InstallationStep{
		ID:           "data-operators",
		Wave:         -60,
		Title:        "Install data operators",
		Phase:        "gitops",
		Dependencies: []string{"security-secrets"},
		ChangeSet: ChangeSet{
			ArgoApplications: []ArgoApplicationChange{{Name: "data-operators", Namespace: spec.GitOps.ArgoCD.Namespace, Path: "gitops/components/data-operators", Wave: -60}},
			HelmReleases:     releaseHelmChanges(release, "cloudnative-pg", "ydb-operator", "strimzi", "redis-operator"),
		},
		Readiness: []string{"Data operator deployments ready"},
		Rollback:  RollbackPolicy{Supported: true},
	})

	steps = append(steps, InstallationStep{
		ID:           "data-clusters",
		Wave:         -50,
		Title:        "Create data clusters and external data checks",
		Phase:        "gitops",
		Dependencies: []string{"data-operators"},
		ChangeSet: ChangeSet{
			ArgoApplications: []ArgoApplicationChange{{Name: "data-clusters", Namespace: spec.GitOps.ArgoCD.Namespace, Path: "gitops/components/data-clusters", Wave: -50}},
			ExternalChecks:   []string{"PostgreSQL readiness", "YDB readiness", "Kafka readiness", "Redis readiness"},
			Secrets:          []string{"database roles", "connection references"},
		},
		Readiness: []string{"Database endpoints reachable", "Kafka topics reconciled", "Redis ready"},
		Rollback:  RollbackPolicy{Supported: false, Boundary: "Stateful data creation requires backup-aware rollback"},
	})

	steps = append(steps, InstallationStep{
		ID:           "identity-authorization",
		Wave:         -40,
		Title:        "Install identity and authorization systems",
		Phase:        "gitops",
		Dependencies: []string{"data-clusters", "pki-trust"},
		ChangeSet: ChangeSet{
			ArgoApplications: []ArgoApplicationChange{{Name: "identity-authorization", Namespace: spec.GitOps.ArgoCD.Namespace, Path: "gitops/components/identity-authorization", Wave: -40}},
			Secrets:          []string{"Keycloak admin external secret", "SpiceDB preshared key external secret"},
			Migrations:       []string{"SpiceDB schema bootstrap"},
		},
		Readiness: []string{"Keycloak realm m8 imported", "SpiceDB schema version ready"},
		Rollback:  RollbackPolicy{Supported: true, Boundary: "SpiceDB schema downgrade may be manual"},
	})

	steps = append(steps, InstallationStep{
		ID:           "observability",
		Wave:         -30,
		Title:        "Install observability stack",
		Phase:        "gitops",
		Dependencies: []string{"security-secrets"},
		ChangeSet: ChangeSet{
			ArgoApplications: []ArgoApplicationChange{{Name: "observability", Namespace: spec.GitOps.ArgoCD.Namespace, Path: "gitops/components/observability", Wave: -30}},
			Policies:         []string{"ServiceMonitor", "PrometheusRule", "dashboards"},
		},
		Readiness: []string{"Prometheus ready", "OTel collector ready", "Grafana ready when enabled"},
		Rollback:  RollbackPolicy{Supported: true},
	})

	steps = append(steps, InstallationStep{
		ID:           "envoy-gateway",
		Wave:         -20,
		Title:        "Install Envoy Gateway and Gateway API resources",
		Phase:        "gitops",
		Dependencies: dependenciesWhenPresent(steps, "cilium", "pki-trust", "security-secrets"),
		ChangeSet: ChangeSet{
			ArgoApplications: []ArgoApplicationChange{{Name: "envoy-gateway", Namespace: spec.GitOps.ArgoCD.Namespace, Path: "gitops/components/envoy-gateway", Wave: -20}},
			Routes:           []string{"GatewayClass", "Gateway"},
			Policies:         []string{"SecurityPolicy", "BackendTrafficPolicy", "ClientTrafficPolicy"},
		},
		Readiness: []string{"GatewayClass accepted", "Gateway programmed"},
		Rollback:  RollbackPolicy{Supported: true},
	})

	steps = append(steps, InstallationStep{
		ID:           "m8-shared-services",
		Wave:         -10,
		Title:        "Install shared M8 services",
		Phase:        "gitops",
		Dependencies: []string{"identity-authorization", "observability"},
		ChangeSet: ChangeSet{
			ArgoApplications: []ArgoApplicationChange{{Name: "m8-shared-services", Namespace: spec.GitOps.ArgoCD.Namespace, Path: "gitops/components/m8-shared-services", Wave: -10}},
		},
		Readiness: []string{"M8 operations and idempotency APIs ready"},
		Rollback:  RollbackPolicy{Supported: true},
	})

	steps = append(steps, InstallationStep{
		ID:           "m8-applications",
		Wave:         0,
		Title:        "Install M8 application modules",
		Phase:        "gitops",
		Dependencies: []string{"m8-shared-services", "envoy-gateway"},
		ChangeSet: ChangeSet{
			ArgoApplications: []ArgoApplicationChange{{Name: "m8-applications", Namespace: spec.GitOps.ArgoCD.Namespace, Path: "gitops/components/m8-applications", Wave: 0}},
		},
		Readiness: []string{"Enabled M8 module deployments ready"},
		Rollback:  RollbackPolicy{Supported: true},
	})

	steps = append(steps, InstallationStep{
		ID:           "routes-policies",
		Wave:         10,
		Title:        "Publish external routes and API policies",
		Phase:        "gitops",
		Dependencies: []string{"m8-applications"},
		ChangeSet: ChangeSet{
			ArgoApplications: []ArgoApplicationChange{{Name: "routes-policies", Namespace: spec.GitOps.ArgoCD.Namespace, Path: "gitops/components/routes-policies", Wave: 10}},
			Routes:           []string{"HTTPRoute", "GRPCRoute", "TLSRoute", "BackendTLSPolicy"},
			Policies:         []string{"rate limits", "quotas", "auth policies"},
		},
		Readiness: []string{"Routes accepted", "Routes programmed"},
		Rollback:  RollbackPolicy{Supported: true},
	})

	steps = append(steps, InstallationStep{
		ID:           "bootstrap-data",
		Wave:         20,
		Title:        "Apply bootstrap data",
		Phase:        "gitops",
		Dependencies: []string{"routes-policies"},
		ChangeSet: ChangeSet{
			Migrations: []string{"Keycloak clients", "Temporal namespaces", "M8 default platform configuration"},
		},
		Readiness: []string{"Bootstrap jobs completed once with checkpoint"},
		Rollback:  RollbackPolicy{Supported: false, Boundary: "Bootstrap data may require restore or compensating migration"},
	})

	steps = append(steps, InstallationStep{
		ID:           "smoke-tests",
		Wave:         30,
		Title:        "Run smoke tests",
		Phase:        "verify",
		Dependencies: []string{"bootstrap-data"},
		ChangeSet: ChangeSet{
			SmokeTests: []string{"TLS issuance", "SPIFFE SVID", "OIDC token", "YDB write/read", "Kafka produce/consume", "Temporal workflow", "SpiceDB check", "Gateway route", "OTLP trace"},
		},
		Readiness: []string{"All required smoke tests pass"},
		Rollback:  RollbackPolicy{Supported: true},
	})

	return steps
}

func releaseHelmChanges(release installerv1alpha1.PlatformRelease, componentNames ...string) []HelmReleaseChange {
	changes := make([]HelmReleaseChange, 0, len(componentNames))
	for _, name := range componentNames {
		component, ok := release.Spec.Components[name]
		if !ok || component.Chart.Repository == "" {
			continue
		}
		changes = append(changes, HelmReleaseChange{
			Name:      name,
			Namespace: namespaceForComponent(name),
			Chart:     component.Chart.Repository,
			Version:   component.Chart.Version,
			Digest:    component.Chart.Digest,
		})
	}
	return changes
}

func namespaceForComponent(name string) string {
	switch name {
	case "cert-manager", "trust-manager", "spire", "external-secrets-operator", "kyverno", "trivy-operator":
		return "m8-security"
	case "cilium":
		return "kube-system"
	case "cloudnative-pg", "ydb-operator", "strimzi", "redis-operator":
		return "m8-data"
	default:
		return "m8-system"
	}
}

func dependenciesWhenPresent(steps []InstallationStep, candidates ...string) []string {
	known := make(map[string]struct{}, len(steps))
	for _, step := range steps {
		known[step.ID] = struct{}{}
	}
	dependencies := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if _, ok := known[candidate]; ok {
			dependencies = append(dependencies, candidate)
		}
	}
	return dependencies
}

func planRisks(installation installerv1alpha1.PlatformInstallation) []PlanRisk {
	var risks []PlanRisk
	if installation.Spec.Network.Cilium.Enabled {
		risks = append(risks, PlanRisk{Severity: "high", Message: "CNI installation changes cluster networking and requires a tested recovery path."})
	}
	if installation.Spec.Profile == installerv1alpha1.ProfileDemo || installation.Spec.Profile == installerv1alpha1.ProfileDevelopment {
		risks = append(risks, PlanRisk{Severity: "medium", Message: fmt.Sprintf("%s profile is not production-grade.", installation.Spec.Profile)})
	}
	if !installation.Spec.Backup.Enabled {
		risks = append(risks, PlanRisk{Severity: "high", Message: "Backup is disabled; upgrades and destructive operations must be blocked."})
	}
	return risks
}

func irreversibleActions(installation installerv1alpha1.PlatformInstallation) []string {
	actions := []string{"CRD version migrations", "database schema migrations", "bootstrap identity data"}
	if installation.Spec.Network.Cilium.Enabled {
		actions = append(actions, "CNI installation")
	}
	return actions
}

func resourceEstimate(release installerv1alpha1.PlatformRelease) ResourceEstimate {
	var storage string
	if component, ok := release.Spec.Components["platform"]; ok {
		storage = component.Resources.Storage
	}
	return ResourceEstimate{
		CPU:     "profile-dependent",
		Memory:  "profile-dependent",
		Storage: storage,
	}
}
