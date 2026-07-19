package uninstall

import (
	"context"
	"fmt"

	installerv1alpha1 "github.com/m8-team/platform/api/installer/v1alpha1"
	installerhelm "github.com/m8-team/platform/internal/installer/helm"
	installerkubernetes "github.com/m8-team/platform/internal/installer/kubernetes"
	"github.com/m8-team/platform/internal/installer/planner"
)

type Executor struct {
	Kubernetes *installerkubernetes.Client
	Helm       installerhelm.Client
}

type Request struct {
	Plan                     planner.InstallationPlan
	InstallationName         string
	InstallationNamespace    string
	PlatformReleaseName      string
	ArgoCDNamespace          string
	ArgoCDManifest           []byte
	RootGitOpsManifests      [][]byte
	DeleteNetwork            bool
	DeleteInstallerCRDs      bool
	DeleteNamespaces         bool
	SkipRootGitOps           bool
	SkipArgoCD               bool
	SkipCertManager          bool
	SkipInstallerMetadata    bool
	SkipGitOpsApplications   bool
	PreserveOperationHistory bool
}

type Result struct {
	Deleted   []string `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Preserved []string `json:"preserved,omitempty" yaml:"preserved,omitempty"`
	Skipped   []string `json:"skipped,omitempty" yaml:"skipped,omitempty"`
}

func (e Executor) Apply(ctx context.Context, request Request) (Result, error) {
	if e.Kubernetes == nil {
		return Result{}, fmt.Errorf("kubernetes client is required")
	}
	result := Result{}
	namespace := firstNonEmpty(request.InstallationNamespace, request.Plan.Installation.Namespace, "m8-system")
	name := firstNonEmpty(request.InstallationName, request.Plan.Installation.Name)
	argoNamespace := firstNonEmpty(request.ArgoCDNamespace, "argocd")

	if !request.SkipGitOpsApplications {
		applications := argoApplicationNames(request.Plan)
		if len(applications) == 0 {
			applications = defaultArgoApplicationNames()
		}
		deleted, err := e.Kubernetes.DeleteArgoApplications(ctx, argoNamespace, applications)
		if err != nil {
			return result, err
		}
		result.Deleted = append(result.Deleted, deleted...)
	}

	if !request.SkipRootGitOps {
		for index, manifest := range request.RootGitOpsManifests {
			if len(manifest) == 0 {
				continue
			}
			deleted, err := e.Kubernetes.DeleteYAMLDocuments(ctx, manifest, argoNamespace)
			if err != nil {
				return result, err
			}
			if len(deleted) == 0 {
				result.Skipped = append(result.Skipped, fmt.Sprintf("root-gitops-%d/no resources found", index+1))
				continue
			}
			result.Deleted = append(result.Deleted, deleted...)
		}
	}

	if !request.SkipCertManager {
		if e.Helm == nil {
			result.Skipped = append(result.Skipped, "helm/m8-security/cert-manager helm client not configured")
		} else if err := e.Helm.Uninstall(ctx, "m8-security", "cert-manager"); err != nil {
			return result, err
		} else {
			result.Deleted = append(result.Deleted, "helm/m8-security/cert-manager")
		}
		deleted, err := e.Kubernetes.DeleteCertManagerArtifacts(ctx)
		if err != nil {
			return result, err
		}
		result.Deleted = append(result.Deleted, deleted...)
	}

	if request.DeleteNetwork {
		if e.Helm == nil {
			result.Skipped = append(result.Skipped, "helm/kube-system/cilium helm client not configured")
		} else if err := e.Helm.Uninstall(ctx, "kube-system", "cilium"); err != nil {
			return result, err
		} else {
			result.Deleted = append(result.Deleted, "helm/kube-system/cilium")
		}
		deleted, err := e.Kubernetes.DeleteCiliumArtifacts(ctx)
		if err != nil {
			return result, err
		}
		result.Deleted = append(result.Deleted, deleted...)
	} else {
		result.Preserved = append(result.Preserved, "helm/kube-system/cilium")
	}

	if !request.SkipArgoCD {
		if len(request.ArgoCDManifest) == 0 {
			result.Skipped = append(result.Skipped, "argocd install manifest not provided")
		} else {
			deleted, err := e.Kubernetes.DeleteYAMLDocuments(ctx, request.ArgoCDManifest, argoNamespace)
			if err != nil {
				return result, err
			}
			result.Deleted = append(result.Deleted, deleted...)
		}
		deleted, err := e.Kubernetes.DeleteArgoCDArtifacts(ctx)
		if err != nil {
			return result, err
		}
		result.Deleted = append(result.Deleted, deleted...)
	}

	if !request.SkipInstallerMetadata {
		deleted, err := e.Kubernetes.DeleteInstallerMetadata(ctx, namespace, name, request.PlatformReleaseName)
		if err != nil {
			return result, err
		}
		result.Deleted = append(result.Deleted, deleted...)
	}

	if request.DeleteInstallerCRDs {
		if err := e.Kubernetes.DeleteInstallerCRDs(ctx); err != nil {
			return result, err
		}
		result.Deleted = append(result.Deleted, "installer-crds")
	} else {
		result.Preserved = append(result.Preserved, "installer-crds")
	}

	if request.PreserveOperationHistory {
		result.Preserved = append(result.Preserved, "installationoperations")
	}
	if request.DeleteNamespaces {
		deleted, err := e.Kubernetes.DeleteNamespaces(ctx, uninstallNamespaces(request.DeleteNetwork))
		if err != nil {
			return result, err
		}
		result.Deleted = append(result.Deleted, deleted...)
	} else {
		result.Preserved = append(result.Preserved, uninstallNamespaces(request.DeleteNetwork)...)
	}
	result.Preserved = append(result.Preserved,
		"persistentvolumeclaims",
		"persistentvolumes",
		"database-data",
		"kafka-topics",
		"backups",
		"external-secrets",
		"audit-archives",
	)

	return result, nil
}

func uninstallNamespaces(includeNetwork bool) []string {
	namespaces := []string{
		"argocd",
		"m8-data",
		"m8-gateway",
		"m8-observability",
		"m8-security",
		"m8-system",
	}
	if includeNetwork {
		namespaces = append(namespaces, "cilium-secrets")
	}
	return namespaces
}

func argoApplicationNames(plan planner.InstallationPlan) []string {
	seen := map[string]struct{}{}
	var names []string
	for _, step := range plan.Steps {
		for _, application := range step.ChangeSet.ArgoApplications {
			name := application.Name
			if name == "" {
				continue
			}
			name = "m8-" + name
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			names = append(names, name)
		}
	}
	return names
}

func defaultArgoApplicationNames() []string {
	return []string{
		"m8-data-operators",
		"m8-data-clusters",
		"m8-identity-authorization",
		"m8-observability",
		"m8-envoy-gateway",
		"m8-m8-shared-services",
		"m8-m8-applications",
		"m8-routes-policies",
		"m8-bootstrap-data",
		"m8-smoke-tests",
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func InstallationRef(installation installerv1alpha1.PlatformInstallation) planner.InstallationRef {
	return planner.InstallationRef{Name: installation.Name, Namespace: installation.Namespace}
}
