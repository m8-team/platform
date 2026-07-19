package preflight

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	installerv1alpha1 "github.com/m8-team/platform/api/installer/v1alpha1"
)

type InstallationValidationCheck struct{}

func (InstallationValidationCheck) ID() string { return "M8-PREFLIGHT-INSTALLATION-001" }

func (c InstallationValidationCheck) Run(ctx context.Context, installation installerv1alpha1.PlatformInstallation) Result {
	if err := ctx.Err(); err != nil {
		return failed(c.ID(), "configuration", "Context cancelled before validation completed.", err.Error(), "")
	}
	if err := installation.Validate(); err != nil {
		return Result{
			ID:               c.ID(),
			Category:         "configuration",
			Severity:         SeverityError,
			Status:           StatusFail,
			Message:          "PlatformInstallation schema validation failed.",
			Details:          map[string]string{"error": err.Error()},
			Remediation:      "Fix the installation YAML before running plan or install.",
			DocumentationRef: "docs/engineering-artifacts/installer/api.md",
		}
	}
	return passed(c.ID(), "configuration", "PlatformInstallation is valid.")
}

type ModuleDependencyCheck struct{}

func (ModuleDependencyCheck) ID() string { return "M8-PREFLIGHT-MODULES-001" }

func (c ModuleDependencyCheck) Run(ctx context.Context, installation installerv1alpha1.PlatformInstallation) Result {
	if err := ctx.Err(); err != nil {
		return failed(c.ID(), "configuration", "Context cancelled before module dependency validation completed.", err.Error(), "")
	}
	if err := installation.Validate(); err != nil {
		return Result{
			ID:               c.ID(),
			Category:         "configuration",
			Severity:         SeverityError,
			Status:           StatusFail,
			Message:          "Module dependency graph is invalid.",
			Details:          map[string]string{"error": err.Error()},
			Remediation:      "Enable required dependencies or disable dependent M8 modules.",
			DocumentationRef: "docs/engineering-artifacts/installer/architecture.md#dependency-graph",
		}
	}
	return passed(c.ID(), "configuration", "Module dependency graph is valid.")
}

type SkippedClusterCheck struct{}

func (SkippedClusterCheck) ID() string { return "M8-PREFLIGHT-CLUSTER-000" }

func (c SkippedClusterCheck) Run(context.Context, installerv1alpha1.PlatformInstallation) Result {
	return Result{
		ID:               c.ID(),
		Category:         "kubernetes",
		Severity:         SeverityWarning,
		Status:           StatusSkip,
		Message:          "Kubernetes checks were skipped.",
		Remediation:      "Run without --skip-cluster before bootstrap or install.",
		DocumentationRef: "docs/engineering-artifacts/installer/cli.md#preflight",
	}
}

type KubernetesAPICheck struct {
	Cluster ClusterReader
}

func (KubernetesAPICheck) ID() string { return "M8-PREFLIGHT-K8S-001" }

func (c KubernetesAPICheck) Run(ctx context.Context, installation installerv1alpha1.PlatformInstallation) Result {
	version, err := c.Cluster.ServerVersion(ctx)
	if err != nil {
		return failed(c.ID(), "kubernetes", "Kubernetes API is not reachable.", err.Error(), "Check kubeconfig, context, TLS, and network access to the API server.")
	}
	result := passed(c.ID(), "kubernetes", "Kubernetes API is reachable.")
	result.Details = map[string]string{"serverVersion": version}

	minVersion := "1.29.0"
	if installation.Spec.Profile == installerv1alpha1.ProfileProduction {
		minVersion = "1.30.0"
	}
	if compareMajorMinor(version, minVersion) < 0 {
		result.Status = StatusFail
		result.Severity = SeverityError
		result.Message = "Kubernetes version is below the supported minimum."
		result.Remediation = "Upgrade Kubernetes before installing M8 Platform."
		result.Details["minimumVersion"] = minVersion
	}

	return result
}

type NodeCapacityCheck struct {
	Cluster ClusterReader
}

func (NodeCapacityCheck) ID() string { return "M8-PREFLIGHT-K8S-002" }

func (c NodeCapacityCheck) Run(ctx context.Context, installation installerv1alpha1.PlatformInstallation) Result {
	summary, err := c.Cluster.NodeSummary(ctx)
	if err != nil {
		return failed(c.ID(), "kubernetes", "Unable to inspect Kubernetes nodes.", err.Error(), "Grant node list permissions to the installer identity.")
	}
	details := map[string]string{
		"nodes":         strconv.Itoa(summary.Total),
		"ready":         strconv.Itoa(summary.Ready),
		"architectures": joinCountMap(summary.Architectures),
		"zones":         joinCountMap(summary.Zones),
	}
	minimumNodes := int(installation.Spec.Cluster.MinimumNodes)
	if summary.Ready < minimumNodes {
		return Result{
			ID:               c.ID(),
			Category:         "kubernetes",
			Severity:         SeverityError,
			Status:           StatusFail,
			Message:          "Cluster does not have enough ready nodes for the selected profile.",
			Details:          details,
			Remediation:      fmt.Sprintf("Provide at least %d ready nodes or choose a smaller profile.", minimumNodes),
			DocumentationRef: "docs/engineering-artifacts/installer/architecture.md#profiles",
		}
	}
	if installation.Spec.Profile == installerv1alpha1.ProfileProduction && len(summary.Zones) < 3 {
		return Result{
			ID:               c.ID(),
			Category:         "kubernetes",
			Severity:         SeverityWarning,
			Status:           StatusWarn,
			Message:          "Production profile should span at least three topology zones.",
			Details:          details,
			Remediation:      "Add topology.kubernetes.io/zone labels and distribute worker nodes across zones.",
			DocumentationRef: "docs/engineering-artifacts/installer/architecture.md#profiles",
		}
	}

	result := passed(c.ID(), "kubernetes", "Node capacity satisfies the selected profile.")
	result.Details = details
	return result
}

type StorageClassCheck struct {
	Cluster ClusterReader
}

func (StorageClassCheck) ID() string { return "M8-PREFLIGHT-STORAGE-001" }

func (c StorageClassCheck) Run(ctx context.Context, installation installerv1alpha1.PlatformInstallation) Result {
	classes, err := c.Cluster.StorageClasses(ctx)
	if err != nil {
		return failed(c.ID(), "storage", "Unable to list StorageClass resources.", err.Error(), "Grant storage.k8s.io StorageClass list permissions.")
	}
	sort.Strings(classes)
	if len(classes) == 0 && installation.Spec.Profile != installerv1alpha1.ProfileDemo {
		return Result{
			ID:               c.ID(),
			Category:         "storage",
			Severity:         SeverityError,
			Status:           StatusFail,
			Message:          "No StorageClass resources found.",
			Remediation:      "Install a default StorageClass before installing stateful M8 components.",
			DocumentationRef: "docs/engineering-artifacts/installer/architecture.md#stateful-components",
		}
	}
	result := passed(c.ID(), "storage", "StorageClass resources are available.")
	result.Details = map[string]string{"storageClasses": strings.Join(classes, ",")}
	return result
}

type GatewayAPICheck struct {
	Cluster ClusterReader
}

func (GatewayAPICheck) ID() string { return "M8-PREFLIGHT-GATEWAY-001" }

func (c GatewayAPICheck) Run(ctx context.Context, installation installerv1alpha1.PlatformInstallation) Result {
	if !installation.Spec.Gateway.EnvoyGateway.Enabled {
		return Result{
			ID:       c.ID(),
			Category: "gateway",
			Severity: SeverityInfo,
			Status:   StatusSkip,
			Message:  "Envoy Gateway is disabled.",
		}
	}
	found, err := c.Cluster.HasAPIResource(ctx, "gateway.networking.k8s.io/v1", "GatewayClass")
	if err != nil {
		return failed(c.ID(), "gateway", "Unable to inspect Gateway API resources.", err.Error(), "Grant discovery permissions to the installer identity.")
	}
	if !found {
		return Result{
			ID:               c.ID(),
			Category:         "gateway",
			Severity:         SeverityWarning,
			Status:           StatusWarn,
			Message:          "Gateway API v1 CRDs are not installed yet.",
			Remediation:      "This is acceptable before bootstrap; bootstrap will install Gateway API CRDs.",
			DocumentationRef: "docs/engineering-artifacts/installer/gitops.md#sync-waves",
		}
	}
	return passed(c.ID(), "gateway", "Gateway API v1 resources are discoverable.")
}

func passed(id, category, message string) Result {
	return Result{
		ID:       id,
		Category: category,
		Severity: SeverityInfo,
		Status:   StatusPass,
		Message:  message,
	}
}

func failed(id, category, message, detail, remediation string) Result {
	return Result{
		ID:          id,
		Category:    category,
		Severity:    SeverityError,
		Status:      StatusFail,
		Message:     message,
		Details:     map[string]string{"error": detail},
		Remediation: remediation,
	}
}

func joinCountMap(values map[string]int) string {
	if len(values) == 0 {
		return ""
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", key, values[key]))
	}
	return strings.Join(parts, ",")
}

func compareMajorMinor(version string, minimum string) int {
	currentMajor, currentMinor := parseMajorMinor(version)
	minMajor, minMinor := parseMajorMinor(minimum)
	if currentMajor != minMajor {
		if currentMajor < minMajor {
			return -1
		}
		return 1
	}
	if currentMinor < minMinor {
		return -1
	}
	if currentMinor > minMinor {
		return 1
	}
	return 0
}

func parseMajorMinor(version string) (int, int) {
	clean := strings.TrimPrefix(version, "v")
	clean = strings.Split(clean, "-")[0]
	parts := strings.Split(clean, ".")
	if len(parts) < 2 {
		return 0, 0
	}
	major, _ := strconv.Atoi(parts[0])
	minorPart := strings.TrimRightFunc(parts[1], func(r rune) bool {
		return r < '0' || r > '9'
	})
	minor, _ := strconv.Atoi(minorPart)
	return major, minor
}
