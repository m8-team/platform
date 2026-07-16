package install

import (
	"context"
	"fmt"
	"os/user"
	"strings"
	"time"

	installerv1alpha1 "github.com/m8platform/platform/api/installer/v1alpha1"
	installerhelm "github.com/m8platform/platform/internal/installer/helm"
	installerkubernetes "github.com/m8platform/platform/internal/installer/kubernetes"
	"github.com/m8platform/platform/internal/installer/operations"
	"github.com/m8platform/platform/internal/installer/planner"
)

type Executor struct {
	Kubernetes *installerkubernetes.Client
	Helm       installerhelm.Client
	Now        func() time.Time
}

type Request struct {
	Plan                 planner.InstallationPlan
	Installation         installerv1alpha1.PlatformInstallation
	Release              installerv1alpha1.PlatformRelease
	RequestedBy          string
	ArgoCDManifest       []byte
	RootGitOpsManifests  [][]byte
	SkipCilium           bool
	SkipCertManager      bool
	SkipArgoCD           bool
	SkipRootGitOps       bool
	WaitGitOpsHandoff    bool
	GitOpsHandoffTimeout time.Duration
}

type Result struct {
	Operation installerv1alpha1.InstallationOperation
	Applied   []string
	Skipped   []string
}

func (e Executor) Apply(ctx context.Context, request Request) (Result, error) {
	if e.Kubernetes == nil {
		return Result{}, fmt.Errorf("kubernetes client is required")
	}
	if request.Installation.Name == "" {
		return Result{}, fmt.Errorf("platform installation is required")
	}
	if request.Release.Name == "" {
		return Result{}, fmt.Errorf("platform release is required")
	}
	now := time.Now().UTC
	if e.Now != nil {
		now = e.Now
	}
	requestedBy := request.RequestedBy
	if requestedBy == "" {
		requestedBy = currentUsername()
	}
	if request.Installation.Namespace == "" {
		request.Installation.Namespace = "m8-system"
	}

	operation := operations.NewInstallationOperation(
		request.Installation.Name,
		installerv1alpha1.OperationTypeInstall,
		request.Installation.Spec.PlatformVersion,
		request.Plan.ConfigDigest,
		requestedBy,
		now(),
	)
	operation.Namespace = request.Installation.Namespace

	result := Result{Operation: operation}

	if err := e.Kubernetes.ApplyInstallerCRDs(ctx); err != nil {
		return result, err
	}
	result.Applied = append(result.Applied, "installer-crds")

	if err := e.Kubernetes.WaitForInstallerAPI(ctx, 45*time.Second); err != nil {
		return result, err
	}
	result.Applied = append(result.Applied, "installer-api-ready")

	for _, namespace := range namespacesFromPlan(request.Plan) {
		if err := e.Kubernetes.ApplyNamespace(ctx, namespace); err != nil {
			return result, err
		}
		result.Applied = append(result.Applied, "namespace/"+namespace)
	}

	if !request.SkipCilium {
		if e.Helm == nil {
			result.Skipped = append(result.Skipped, "cilium helm client not configured")
		} else if request.Installation.Spec.Network.Cilium.Enabled {
			if err := e.installCilium(ctx, request); err != nil {
				return result, err
			}
			result.Applied = append(result.Applied, "helm/kube-system/cilium")
		}
	}

	if !request.SkipCertManager {
		if e.Helm == nil {
			result.Skipped = append(result.Skipped, "cert-manager helm client not configured")
		} else if request.Installation.Spec.Certificates.CertManager.Enabled {
			if err := e.installCertManager(ctx, request); err != nil {
				return result, err
			}
			result.Applied = append(result.Applied, "helm/m8-security/cert-manager")
		}
	}

	if !request.SkipArgoCD {
		if len(request.ArgoCDManifest) == 0 {
			result.Skipped = append(result.Skipped, "argocd install manifest not provided")
		} else {
			applied, err := e.Kubernetes.ApplyYAMLDocuments(ctx, request.ArgoCDManifest, request.Installation.Spec.GitOps.ArgoCD.Namespace)
			if err != nil {
				return result, err
			}
			result.Applied = append(result.Applied, fmt.Sprintf("argocd/%d resources", len(applied)))
			for _, kind := range []string{"Application", "AppProject", "ApplicationSet"} {
				if err := e.Kubernetes.WaitForAPIResource(ctx, "argoproj.io/v1alpha1", kind, 90*time.Second); err != nil {
					return result, err
				}
				result.Applied = append(result.Applied, "argocd-"+strings.ToLower(kind)+"-api-ready")
			}
		}
	}

	if err := e.Kubernetes.ApplyPlatformRelease(ctx, request.Release); err != nil {
		return result, err
	}
	result.Applied = append(result.Applied, "platformrelease/"+request.Release.Name)

	if err := e.Kubernetes.ApplyPlatformInstallation(ctx, request.Installation); err != nil {
		return result, err
	}
	result.Applied = append(result.Applied, "platforminstallation/"+request.Installation.Namespace+"/"+request.Installation.Name)

	if err := e.Kubernetes.ApplyInstallationOperation(ctx, operation); err != nil {
		return result, err
	}
	result.Applied = append(result.Applied, "installationoperation/"+operation.Namespace+"/"+operation.Name)

	if !request.SkipRootGitOps {
		for index, manifest := range request.RootGitOpsManifests {
			if len(manifest) == 0 {
				continue
			}
			applied, err := e.Kubernetes.ApplyYAMLDocuments(ctx, manifest, request.Installation.Spec.GitOps.ArgoCD.Namespace)
			if err != nil {
				return result, err
			}
			result.Applied = append(result.Applied, fmt.Sprintf("root-gitops-%d/%d resources", index+1, len(applied)))
		}
		applications := argoApplicationNames(request.Plan)
		if len(applications) > 0 {
			if request.WaitGitOpsHandoff {
				if err := e.Kubernetes.WaitForArgoApplications(ctx, request.Installation.Spec.GitOps.ArgoCD.Namespace, applications, request.GitOpsHandoffTimeout); err != nil {
					return result, err
				}
				result.Applied = append(result.Applied, fmt.Sprintf("gitops-reconciliation-handoff/%d applications", len(applications)))
			} else {
				result.Applied = append(result.Applied, fmt.Sprintf("gitops-reconciliation-handoff/submitted-%d applications", len(applications)))
			}
		}
	}

	result.Skipped = append(result.Skipped,
		"trust-manager helm install",
		"external-secrets helm install",
	)

	return result, nil
}

func (e Executor) installCilium(ctx context.Context, request Request) error {
	release, err := ciliumHelmRelease(request.Installation, request.Release)
	if err != nil {
		return err
	}
	return e.Helm.Apply(ctx, release)
}

func ciliumHelmRelease(installation installerv1alpha1.PlatformInstallation, release installerv1alpha1.PlatformRelease) (installerhelm.Release, error) {
	component, ok := release.Spec.Components["cilium"]
	if !ok {
		return installerhelm.Release{}, fmt.Errorf("release catalog does not contain cilium component")
	}
	version := component.Chart.Version
	if version == "" {
		version = component.Version
	}
	values, err := ciliumValues(installation)
	if err != nil {
		return installerhelm.Release{}, err
	}
	return installerhelm.Release{
		Name:       "cilium",
		Namespace:  "kube-system",
		Chart:      "cilium",
		Repository: "https://helm.cilium.io",
		Version:    version,
		Values:     values,
	}, nil
}

func ciliumValues(installation installerv1alpha1.PlatformInstallation) (map[string]any, error) {
	kubeProxyReplacement, err := ciliumKubeProxyReplacement(installation.Spec.Network.KubeProxyReplacement)
	if err != nil {
		return nil, err
	}
	values := map[string]any{
		"kubeProxyReplacement": kubeProxyReplacement,
		"hubble": map[string]any{
			"enabled": installation.Spec.Network.Cilium.HubbleRelay || installation.Spec.Network.Cilium.HubbleUI,
			"relay":   map[string]any{"enabled": installation.Spec.Network.Cilium.HubbleRelay},
			"ui":      map[string]any{"enabled": installation.Spec.Network.Cilium.HubbleUI},
		},
	}
	if installation.Spec.Network.WireGuardEncryption {
		values["encryption"] = map[string]any{
			"enabled": true,
			"type":    "wireguard",
		}
	}
	return values, nil
}

func ciliumKubeProxyReplacement(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "false", "disabled":
		return false, nil
	case "true", "enabled", "strict":
		return true, nil
	default:
		return false, fmt.Errorf("spec.network.kubeProxyReplacement must be true or false for Cilium Helm chart, got %q", value)
	}
}

func (e Executor) installCertManager(ctx context.Context, request Request) error {
	release, err := certManagerHelmRelease(request.Release)
	if err != nil {
		return err
	}
	return e.Helm.Apply(ctx, release)
}

func certManagerHelmRelease(release installerv1alpha1.PlatformRelease) (installerhelm.Release, error) {
	component, ok := release.Spec.Components["cert-manager"]
	if !ok {
		return installerhelm.Release{}, fmt.Errorf("release catalog does not contain cert-manager component")
	}
	version := component.Chart.Version
	if version == "" {
		version = component.Version
	}
	return installerhelm.Release{
		Name:       "cert-manager",
		Namespace:  "m8-security",
		Chart:      "cert-manager",
		Repository: "https://charts.jetstack.io",
		Version:    version,
		Values: map[string]any{
			"crds": map[string]any{"enabled": true},
		},
	}, nil
}

func namespacesFromPlan(plan planner.InstallationPlan) []string {
	seen := map[string]struct{}{}
	var namespaces []string
	for _, step := range plan.Steps {
		for _, namespace := range step.ChangeSet.Namespaces {
			if _, ok := seen[namespace]; ok {
				continue
			}
			seen[namespace] = struct{}{}
			namespaces = append(namespaces, namespace)
		}
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
			if !strings.HasPrefix(name, "m8-") {
				name = "m8-" + name
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			names = append(names, name)
		}
	}
	return names
}

func currentUsername() string {
	current, err := user.Current()
	if err != nil || current.Username == "" {
		return "unknown"
	}
	return current.Username
}
