package install

import (
	"context"
	"fmt"
	"os/user"
	"time"

	installerv1alpha1 "github.com/m8platform/platform/api/installer/v1alpha1"
	installerkubernetes "github.com/m8platform/platform/internal/installer/kubernetes"
	"github.com/m8platform/platform/internal/installer/operations"
	"github.com/m8platform/platform/internal/installer/planner"
)

type Executor struct {
	Kubernetes *installerkubernetes.Client
	Now        func() time.Time
}

type Request struct {
	Plan                planner.InstallationPlan
	Installation        installerv1alpha1.PlatformInstallation
	Release             installerv1alpha1.PlatformRelease
	RequestedBy         string
	ArgoCDManifest      []byte
	RootGitOpsManifests [][]byte
	SkipArgoCD          bool
	SkipRootGitOps      bool
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

	if !request.SkipArgoCD {
		if len(request.ArgoCDManifest) == 0 {
			result.Skipped = append(result.Skipped, "argocd install manifest not provided")
		} else {
			applied, err := e.Kubernetes.ApplyYAMLDocuments(ctx, request.ArgoCDManifest, request.Installation.Spec.GitOps.ArgoCD.Namespace)
			if err != nil {
				return result, err
			}
			result.Applied = append(result.Applied, fmt.Sprintf("argocd/%d resources", len(applied)))
			if err := e.Kubernetes.WaitForAPIResource(ctx, "argoproj.io/v1alpha1", "Application", 90*time.Second); err != nil {
				return result, err
			}
			result.Applied = append(result.Applied, "argocd-api-ready")
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
	}

	result.Skipped = append(result.Skipped,
		"cilium helm install",
		"cert-manager helm install",
		"trust-manager helm install",
		"external-secrets helm install",
		"GitOps component reconciliation",
	)

	return result, nil
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

func currentUsername() string {
	current, err := user.Current()
	if err != nil || current.Username == "" {
		return "unknown"
	}
	return current.Username
}
