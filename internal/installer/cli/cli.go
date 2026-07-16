package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	installerv1alpha1 "github.com/m8platform/platform/api/installer/v1alpha1"
	"github.com/m8platform/platform/internal/installer/catalog"
	"github.com/m8platform/platform/internal/installer/config"
	installerhelm "github.com/m8platform/platform/internal/installer/helm"
	installerinstall "github.com/m8platform/platform/internal/installer/install"
	installerkubernetes "github.com/m8platform/platform/internal/installer/kubernetes"
	"github.com/m8platform/platform/internal/installer/output"
	"github.com/m8platform/platform/internal/installer/planner"
	"github.com/m8platform/platform/internal/installer/preflight"
	installeruninstall "github.com/m8platform/platform/internal/installer/uninstall"
	"sigs.k8s.io/yaml"
)

const Version = "1.0.0-dev"

const defaultArgoCDManifestURL = "https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml"

const (
	ExitOK             = 0
	ExitError          = 1
	ExitCheckFailed    = 2
	ExitUsage          = 3
	ExitNotImplemented = 4
)

type App struct {
	Stdout io.Writer
	Stderr io.Writer
	Now    func() time.Time
}

func New(stdout io.Writer, stderr io.Writer) App {
	return App{
		Stdout: stdout,
		Stderr: stderr,
		Now:    func() time.Time { return time.Now().UTC() },
	}
}

func (a App) Run(ctx context.Context, args []string) int {
	if len(args) == 0 {
		a.usage()
		return ExitUsage
	}

	switch args[0] {
	case "version":
		_, _ = fmt.Fprintf(a.Stdout, "m8ctl %s\n", Version)
		return ExitOK
	case "preflight":
		return a.runPreflight(ctx, args[1:])
	case "plan":
		return a.runPlan(ctx, args[1:])
	case "status":
		return a.runStatus(ctx, args[1:])
	case "install":
		return a.runInstall(ctx, args[1:])
	case "uninstall":
		return a.runUninstall(ctx, args[1:])
	case "bootstrap", "doctor", "upgrade", "rollback", "backup", "restore", "bundle":
		_, _ = fmt.Fprintf(a.Stderr, "m8ctl %s is defined in the CLI contract but not implemented in this MVP skeleton\n", args[0])
		return ExitNotImplemented
	case "-h", "--help", "help":
		a.usage()
		return ExitOK
	default:
		_, _ = fmt.Fprintf(a.Stderr, "unknown command %q\n", args[0])
		a.usage()
		return ExitUsage
	}
}

type installReport struct {
	APIVersion     string                     `json:"apiVersion" yaml:"apiVersion"`
	Kind           string                     `json:"kind" yaml:"kind"`
	Installation   planner.InstallationRef    `json:"installation" yaml:"installation"`
	Release        planner.ReleaseRef         `json:"release" yaml:"release"`
	Mode           string                     `json:"mode" yaml:"mode"`
	ApplySupported bool                       `json:"applySupported" yaml:"applySupported"`
	Message        string                     `json:"message" yaml:"message"`
	ConfigDigest   string                     `json:"configDigest" yaml:"configDigest"`
	ReleaseDigest  string                     `json:"releaseCatalogDigest" yaml:"releaseCatalogDigest"`
	Steps          []planner.InstallationStep `json:"steps" yaml:"steps"`
	Applied        []string                   `json:"applied,omitempty" yaml:"applied,omitempty"`
	Skipped        []string                   `json:"skipped,omitempty" yaml:"skipped,omitempty"`
	Operation      string                     `json:"operation,omitempty" yaml:"operation,omitempty"`
}

type installSource struct {
	Plan         planner.InstallationPlan
	Installation installerv1alpha1.PlatformInstallation
	Release      installerv1alpha1.PlatformRelease
	PreviewOnly  bool
}

type uninstallReport struct {
	APIVersion   string                  `json:"apiVersion" yaml:"apiVersion"`
	Kind         string                  `json:"kind" yaml:"kind"`
	Installation planner.InstallationRef `json:"installation" yaml:"installation"`
	Mode         string                  `json:"mode" yaml:"mode"`
	Message      string                  `json:"message" yaml:"message"`
	Deleted      []string                `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Preserved    []string                `json:"preserved,omitempty" yaml:"preserved,omitempty"`
	Skipped      []string                `json:"skipped,omitempty" yaml:"skipped,omitempty"`
}

func (a App) runInstall(ctx context.Context, args []string) int {
	flags := flag.NewFlagSet("install", flag.ContinueOnError)
	flags.SetOutput(a.Stderr)
	planPath := flags.String("plan", "", "InstallationPlan YAML file")
	file := flags.String("f", "", "PlatformInstallation YAML file")
	outputValue := flags.String("output", "table", "Output format: table, json, yaml")
	catalogDir := flags.String("catalog", "catalog/releases", "PlatformRelease catalog directory")
	allowUnsigned := flags.Bool("allow-unsigned-release", false, "Allow unsigned release catalogs for local development")
	dryRun := flags.Bool("dry-run", false, "Render the installation plan without applying changes")
	kubeconfig := flags.String("kubeconfig", "", "Path to kubeconfig")
	kubeContext := flags.String("context", "", "Kubernetes context")
	helmTimeout := flags.Duration("helm-timeout", 10*time.Minute, "Timeout for Helm SDK bootstrap releases")
	skipCilium := flags.Bool("skip-cilium", false, "Skip Cilium Helm installation")
	skipCertManager := flags.Bool("skip-cert-manager", false, "Skip cert-manager Helm installation")
	argoCDManifestPath := flags.String("argocd-manifest", "", "Local Argo CD install manifest")
	argoCDManifestURL := flags.String("argocd-manifest-url", defaultArgoCDManifestURL, "Argo CD install manifest URL")
	skipArgoCD := flags.Bool("skip-argocd", false, "Skip Argo CD manifest installation")
	skipRootGitOps := flags.Bool("skip-root-gitops", false, "Skip root AppProject/ApplicationSet installation")
	waitGitOpsHandoff := flags.Bool("wait-gitops-handoff", false, "Wait until ApplicationSet creates planned Argo CD Applications")
	gitOpsHandoffTimeout := flags.Duration("gitops-handoff-timeout", 5*time.Minute, "Timeout for --wait-gitops-handoff")
	if err := flags.Parse(args); err != nil {
		return ExitUsage
	}
	if flags.NArg() > 0 {
		_, _ = fmt.Fprintf(a.Stderr, "unexpected install argument %q\n", flags.Arg(0))
		return ExitUsage
	}

	var source installSource
	var err error
	switch {
	case *planPath != "":
		plan, readErr := readPlanFile(*planPath)
		err = readErr
		source = installSource{Plan: plan, PreviewOnly: true}
	case *file != "":
		source, err = a.loadInstallSource(ctx, *file, resolveCatalogDir(*catalogDir), *allowUnsigned)
	default:
		defaultFile, ok := findDefaultInstallationFile()
		if !ok {
			_, _ = fmt.Fprintln(a.Stderr, "install requires --plan plan.yaml or -f installation.yaml")
			return ExitUsage
		}
		source, err = a.loadInstallSource(ctx, defaultFile, resolveCatalogDir(*catalogDir), *allowUnsigned)
	}
	if err != nil {
		_, _ = fmt.Fprintf(a.Stderr, "prepare install plan: %v\n", err)
		return ExitError
	}

	report := installReport{
		APIVersion:     installerv1alpha1.GroupName + "/" + installerv1alpha1.Version,
		Kind:           "InstallReport",
		Installation:   source.Plan.Installation,
		Release:        source.Plan.Release,
		Mode:           "preview",
		ApplySupported: true,
		Message:        "Dry run only; no cluster changes were made.",
		ConfigDigest:   source.Plan.ConfigDigest,
		ReleaseDigest:  source.Plan.ReleaseCatalogDigest,
		Steps:          source.Plan.Steps,
	}

	if *dryRun || source.PreviewOnly {
		if source.PreviewOnly {
			report.ApplySupported = false
			report.Message = "Plan-only preview; pass -f installation.yaml to apply installer metadata resources."
		}
		if err := writeInstallReport(a.Stdout, output.ParseFormat(*outputValue), report); err != nil {
			_, _ = fmt.Fprintf(a.Stderr, "write output: %v\n", err)
			return ExitError
		}
		return ExitOK
	}

	client, err := installerkubernetes.NewClient(installerkubernetes.ClientOptions{Kubeconfig: *kubeconfig, Context: *kubeContext})
	if err != nil {
		_, _ = fmt.Fprintf(a.Stderr, "create Kubernetes client: %v\n", err)
		return ExitError
	}
	var argoCDManifest []byte
	if !*skipArgoCD {
		argoCDManifest, err = loadArgoCDManifest(ctx, *argoCDManifestPath, *argoCDManifestURL)
		if err != nil {
			_, _ = fmt.Fprintf(a.Stderr, "load Argo CD manifest: %v\n", err)
			return ExitError
		}
	}
	var rootGitOpsManifests [][]byte
	if !*skipRootGitOps {
		rootGitOpsManifests, err = loadRootGitOpsManifests()
		if err != nil {
			_, _ = fmt.Fprintf(a.Stderr, "load root GitOps manifests: %v\n", err)
			return ExitError
		}
	}
	helmClient := installerhelm.SDKClient{
		Kubeconfig: *kubeconfig,
		Context:    *kubeContext,
		Timeout:    *helmTimeout,
	}
	result, err := installerinstall.Executor{Kubernetes: client, Helm: helmClient, Now: a.Now}.Apply(ctx, installerinstall.Request{
		Plan:                 source.Plan,
		Installation:         source.Installation,
		Release:              source.Release,
		ArgoCDManifest:       argoCDManifest,
		RootGitOpsManifests:  rootGitOpsManifests,
		SkipCilium:           *skipCilium,
		SkipCertManager:      *skipCertManager,
		SkipArgoCD:           *skipArgoCD,
		SkipRootGitOps:       *skipRootGitOps,
		WaitGitOpsHandoff:    *waitGitOpsHandoff,
		GitOpsHandoffTimeout: *gitOpsHandoffTimeout,
	})
	if err != nil {
		_, _ = fmt.Fprintf(a.Stderr, "install failed: %v\n", err)
		return ExitError
	}
	report.Mode = "applied"
	report.Message = "Applied installer API resources, bootstrap namespaces, Cilium/cert-manager Helm releases when enabled, Argo CD bootstrap manifests, PlatformRelease, PlatformInstallation and InstallationOperation."
	report.Applied = result.Applied
	report.Skipped = result.Skipped
	report.Operation = result.Operation.Namespace + "/" + result.Operation.Name

	if err := writeInstallReport(a.Stdout, output.ParseFormat(*outputValue), report); err != nil {
		_, _ = fmt.Fprintf(a.Stderr, "write output: %v\n", err)
		return ExitError
	}
	return ExitOK
}

func (a App) runUninstall(ctx context.Context, args []string) int {
	flags := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	flags.SetOutput(a.Stderr)
	file := flags.String("f", "", "PlatformInstallation YAML file")
	name := flags.String("name", "", "PlatformInstallation name")
	namespace := ""
	flags.StringVar(&namespace, "namespace", "", "PlatformInstallation namespace")
	flags.StringVar(&namespace, "n", "", "PlatformInstallation namespace")
	outputValue := flags.String("output", "table", "Output format: table, json, yaml")
	dryRun := flags.Bool("dry-run", false, "Show resources that would be removed without changing the cluster")
	kubeconfig := flags.String("kubeconfig", "", "Path to kubeconfig")
	kubeContext := flags.String("context", "", "Kubernetes context")
	helmTimeout := flags.Duration("helm-timeout", 10*time.Minute, "Timeout for Helm SDK uninstall actions")
	confirmation := flags.String("confirmation", "", "Required installation name for destructive uninstall options")
	deleteAll := flags.Bool("all", false, "Delete all M8 bootstrap resources, including Cilium, installer CRDs and M8 namespaces")
	deleteNetwork := flags.Bool("delete-network", false, "Also uninstall Cilium Helm release")
	deleteInstallerCRDs := flags.Bool("delete-installer-crds", false, "Also delete M8 installer CRDs")
	deleteNamespaces := flags.Bool("delete-namespaces", false, "Also delete M8, Argo CD and security namespaces")
	deleteData := flags.Bool("delete-data", false, "Allow namespace deletion to remove namespaced persistent data")
	skipGitOpsApplications := flags.Bool("skip-gitops-applications", false, "Skip generated Argo CD Application deletion")
	skipRootGitOps := flags.Bool("skip-root-gitops", false, "Skip root AppProject/ApplicationSet deletion")
	skipArgoCD := flags.Bool("skip-argocd", false, "Skip Argo CD manifest deletion")
	skipCertManager := flags.Bool("skip-cert-manager", false, "Skip cert-manager Helm uninstall")
	skipInstallerMetadata := flags.Bool("skip-installer-metadata", false, "Skip PlatformInstallation and PlatformRelease deletion")
	argoCDManifestPath := flags.String("argocd-manifest", "", "Local Argo CD install manifest")
	argoCDManifestURL := flags.String("argocd-manifest-url", defaultArgoCDManifestURL, "Argo CD install manifest URL")
	if err := flags.Parse(args); err != nil {
		return ExitUsage
	}
	if flags.NArg() > 0 {
		_, _ = fmt.Fprintf(a.Stderr, "unexpected uninstall argument %q\n", flags.Arg(0))
		return ExitUsage
	}

	source, err := a.loadUninstallSource(*file)
	if err != nil {
		_, _ = fmt.Fprintf(a.Stderr, "prepare uninstall: %v\n", err)
		return ExitError
	}
	if *name != "" {
		source.Installation.Name = *name
	}
	if namespace != "" {
		source.Installation.Namespace = namespace
	}
	if source.Installation.Name == "" {
		source.Installation.Name = "m8-production"
	}
	if source.Installation.Namespace == "" {
		source.Installation.Namespace = "m8-system"
	}
	if source.Installation.Spec.GitOps.ArgoCD.Namespace == "" {
		source.Installation.Spec.GitOps.ArgoCD.Namespace = "argocd"
	}

	effectiveDeleteNetwork := *deleteNetwork || *deleteAll
	effectiveDeleteInstallerCRDs := *deleteInstallerCRDs || *deleteAll
	effectiveDeleteNamespaces := *deleteNamespaces || *deleteAll || *deleteData
	if effectiveDeleteNamespaces && !*deleteData && !*deleteAll {
		_, _ = fmt.Fprintln(a.Stderr, "namespace deletion can remove namespaced persistent data; pass --delete-data or --all with --confirmation to continue")
		return ExitUsage
	}
	if (*deleteAll || effectiveDeleteNetwork || effectiveDeleteInstallerCRDs || effectiveDeleteNamespaces) && *confirmation != source.Installation.Name {
		_, _ = fmt.Fprintf(a.Stderr, "destructive uninstall option requires --confirmation %s\n", source.Installation.Name)
		return ExitUsage
	}

	report := uninstallReport{
		APIVersion: installerv1alpha1.GroupName + "/" + installerv1alpha1.Version,
		Kind:       "UninstallReport",
		Installation: planner.InstallationRef{
			Name:      source.Installation.Name,
			Namespace: source.Installation.Namespace,
		},
		Mode:      "preview",
		Message:   "Dry run only; no cluster changes were made.",
		Deleted:   previewUninstallDeletes(*skipGitOpsApplications, *skipRootGitOps, *skipArgoCD, *skipCertManager, *skipInstallerMetadata, effectiveDeleteNetwork, effectiveDeleteInstallerCRDs, effectiveDeleteNamespaces),
		Preserved: previewUninstallPreserved(effectiveDeleteNetwork, effectiveDeleteInstallerCRDs, effectiveDeleteNamespaces),
	}
	if *deleteAll {
		report.Message = "Dry run only; full uninstall selected, no cluster changes were made."
	}

	if *dryRun {
		if err := writeUninstallReport(a.Stdout, output.ParseFormat(*outputValue), report); err != nil {
			_, _ = fmt.Fprintf(a.Stderr, "write output: %v\n", err)
			return ExitError
		}
		return ExitOK
	}

	client, err := installerkubernetes.NewClient(installerkubernetes.ClientOptions{Kubeconfig: *kubeconfig, Context: *kubeContext})
	if err != nil {
		_, _ = fmt.Fprintf(a.Stderr, "create Kubernetes client: %v\n", err)
		return ExitError
	}
	var argoCDManifest []byte
	if !*skipArgoCD {
		argoCDManifest, err = loadArgoCDManifest(ctx, *argoCDManifestPath, *argoCDManifestURL)
		if err != nil {
			_, _ = fmt.Fprintf(a.Stderr, "load Argo CD manifest: %v\n", err)
			return ExitError
		}
	}
	var rootGitOpsManifests [][]byte
	if !*skipRootGitOps {
		rootGitOpsManifests, err = loadRootGitOpsManifests()
		if err != nil {
			_, _ = fmt.Fprintf(a.Stderr, "load root GitOps manifests: %v\n", err)
			return ExitError
		}
	}
	helmClient := installerhelm.SDKClient{
		Kubeconfig: *kubeconfig,
		Context:    *kubeContext,
		Timeout:    *helmTimeout,
	}
	result, err := installeruninstall.Executor{Kubernetes: client, Helm: helmClient}.Apply(ctx, installeruninstall.Request{
		Plan:                     source.Plan,
		InstallationName:         source.Installation.Name,
		InstallationNamespace:    source.Installation.Namespace,
		PlatformReleaseName:      source.Installation.Spec.PlatformVersion,
		ArgoCDNamespace:          source.Installation.Spec.GitOps.ArgoCD.Namespace,
		ArgoCDManifest:           argoCDManifest,
		RootGitOpsManifests:      rootGitOpsManifests,
		DeleteNetwork:            effectiveDeleteNetwork,
		DeleteInstallerCRDs:      effectiveDeleteInstallerCRDs,
		DeleteNamespaces:         effectiveDeleteNamespaces,
		SkipRootGitOps:           *skipRootGitOps,
		SkipArgoCD:               *skipArgoCD,
		SkipCertManager:          *skipCertManager,
		SkipInstallerMetadata:    *skipInstallerMetadata,
		SkipGitOpsApplications:   *skipGitOpsApplications,
		PreserveOperationHistory: !effectiveDeleteInstallerCRDs,
	})
	if err != nil {
		_, _ = fmt.Fprintf(a.Stderr, "uninstall failed: %v\n", err)
		return ExitError
	}
	report.Mode = "applied"
	report.Message = "Removed M8 bootstrap resources. Destructive resources were removed only when explicitly selected."
	report.Deleted = result.Deleted
	report.Preserved = result.Preserved
	report.Skipped = result.Skipped

	if err := writeUninstallReport(a.Stdout, output.ParseFormat(*outputValue), report); err != nil {
		_, _ = fmt.Fprintf(a.Stderr, "write output: %v\n", err)
		return ExitError
	}
	return ExitOK
}

func previewUninstallPreserved(deleteNetwork bool, deleteInstallerCRDs bool, deleteNamespaces bool) []string {
	preserved := []string{}
	if !deleteNetwork {
		preserved = append(preserved, "helm/kube-system/cilium")
	}
	if !deleteInstallerCRDs {
		preserved = append(preserved,
			"installer-crds",
			"installationoperations",
		)
	}
	if !deleteNamespaces {
		preserved = append(preserved,
			"argocd",
			"m8-data",
			"m8-gateway",
			"m8-observability",
			"m8-security",
			"m8-system",
		)
		if !deleteNetwork {
			preserved = append(preserved, "cilium-secrets")
		}
	}
	preserved = append(preserved,
		"persistentvolumeclaims",
		"persistentvolumes",
		"external-database-data",
		"kafka-topics",
		"backups",
		"external-secrets-backends",
		"audit-archives",
	)
	return dedupeStrings(preserved)
}

func (a App) runStatus(ctx context.Context, args []string) int {
	flags := flag.NewFlagSet("status", flag.ContinueOnError)
	flags.SetOutput(a.Stderr)
	outputValue := flags.String("output", "table", "Output format: table, json, yaml")
	kubeconfig := flags.String("kubeconfig", "", "Path to kubeconfig")
	kubeContext := flags.String("context", "", "Kubernetes context")
	namespace := ""
	flags.StringVar(&namespace, "namespace", "", "PlatformInstallation namespace")
	flags.StringVar(&namespace, "n", "", "PlatformInstallation namespace")
	allNamespaces := flags.Bool("all-namespaces", false, "List PlatformInstallation resources across all namespaces")
	if err := flags.Parse(args); err != nil {
		return ExitUsage
	}

	name := ""
	if flags.NArg() > 0 {
		name = flags.Arg(0)
	}
	if flags.NArg() > 1 {
		_, _ = fmt.Fprintln(a.Stderr, "status accepts at most one PlatformInstallation name")
		return ExitUsage
	}

	client, err := installerkubernetes.NewClient(installerkubernetes.ClientOptions{Kubeconfig: *kubeconfig, Context: *kubeContext})
	if err != nil {
		_, _ = fmt.Fprintf(a.Stderr, "create Kubernetes client: %v\n", err)
		return ExitError
	}

	var installations []installerv1alpha1.PlatformInstallation
	if name != "" {
		if namespace == "" {
			namespace = "m8-system"
		}
		installation, err := client.GetPlatformInstallation(ctx, namespace, name)
		if err != nil {
			_, _ = fmt.Fprintf(a.Stderr, "get status: %v\n", err)
			return ExitError
		}
		installations = []installerv1alpha1.PlatformInstallation{installation}
	} else {
		listNamespace := namespace
		if *allNamespaces || namespace == "" {
			listNamespace = ""
		}
		installations, err = client.ListPlatformInstallations(ctx, listNamespace)
		if err != nil {
			_, _ = fmt.Fprintf(a.Stderr, "list status: %v\n", err)
			return ExitError
		}
	}

	sort.SliceStable(installations, func(i, j int) bool {
		if installations[i].Namespace != installations[j].Namespace {
			return installations[i].Namespace < installations[j].Namespace
		}
		return installations[i].Name < installations[j].Name
	})

	if err := output.WriteInstallationsStatus(a.Stdout, output.ParseFormat(*outputValue), installations); err != nil {
		_, _ = fmt.Fprintf(a.Stderr, "write output: %v\n", err)
		return ExitError
	}
	return ExitOK
}

func (a App) runPreflight(ctx context.Context, args []string) int {
	flags := flag.NewFlagSet("preflight", flag.ContinueOnError)
	flags.SetOutput(a.Stderr)
	file := flags.String("f", "", "PlatformInstallation YAML file")
	outputValue := flags.String("output", "table", "Output format: table, json, yaml")
	kubeconfig := flags.String("kubeconfig", "", "Path to kubeconfig")
	kubeContext := flags.String("context", "", "Kubernetes context")
	skipCluster := flags.Bool("skip-cluster", false, "Skip live Kubernetes checks")
	if err := flags.Parse(args); err != nil {
		return ExitUsage
	}
	if *file == "" {
		_, _ = fmt.Fprintln(a.Stderr, "preflight requires -f installation.yaml")
		return ExitUsage
	}

	installationFile, err := config.LoadInstallationFile(*file)
	if err != nil {
		_, _ = fmt.Fprintf(a.Stderr, "load installation: %v\n", err)
		return ExitError
	}

	var cluster preflight.ClusterReader
	if !*skipCluster {
		client, err := installerkubernetes.NewClient(installerkubernetes.ClientOptions{Kubeconfig: *kubeconfig, Context: *kubeContext})
		if err != nil {
			_, _ = fmt.Fprintf(a.Stderr, "create Kubernetes client: %v\n", err)
			return ExitError
		}
		cluster = client
	}

	report := preflight.DefaultRunner(cluster).Run(ctx, installationFile.Installation)
	if err := output.WritePreflight(a.Stdout, output.ParseFormat(*outputValue), report); err != nil {
		_, _ = fmt.Fprintf(a.Stderr, "write output: %v\n", err)
		return ExitError
	}
	if report.HasFailures() {
		return ExitCheckFailed
	}
	return ExitOK
}

func (a App) runPlan(ctx context.Context, args []string) int {
	flags := flag.NewFlagSet("plan", flag.ContinueOnError)
	flags.SetOutput(a.Stderr)
	file := flags.String("f", "", "PlatformInstallation YAML file")
	outputValue := flags.String("output", "table", "Output format table/json/yaml, or a .yaml/.json path to save the plan")
	catalogDir := flags.String("catalog", "catalog/releases", "PlatformRelease catalog directory")
	allowUnsigned := flags.Bool("allow-unsigned-release", false, "Allow unsigned release catalogs for local development")
	if err := flags.Parse(args); err != nil {
		return ExitUsage
	}
	if *file == "" {
		_, _ = fmt.Fprintln(a.Stderr, "plan requires -f installation.yaml")
		return ExitUsage
	}

	plan, err := a.generatePlanFromFile(ctx, *file, resolveCatalogDir(*catalogDir), *allowUnsigned)
	if err != nil {
		_, _ = fmt.Fprintf(a.Stderr, "generate plan: %v\n", err)
		return ExitError
	}

	if isPathOutput(*outputValue) {
		if err := writePlanFile(*outputValue, plan); err != nil {
			_, _ = fmt.Fprintf(a.Stderr, "write plan: %v\n", err)
			return ExitError
		}
		_, _ = fmt.Fprintf(a.Stdout, "Plan written to %s\n", *outputValue)
		return ExitOK
	}

	if err := output.WritePlan(a.Stdout, output.ParseFormat(*outputValue), plan); err != nil {
		_, _ = fmt.Fprintf(a.Stderr, "write output: %v\n", err)
		return ExitError
	}
	return ExitOK
}

func (a App) generatePlanFromFile(ctx context.Context, path string, catalogDir string, allowUnsigned bool) (planner.InstallationPlan, error) {
	source, err := a.loadInstallSource(ctx, path, catalogDir, allowUnsigned)
	if err != nil {
		return planner.InstallationPlan{}, err
	}
	return source.Plan, nil
}

func (a App) loadInstallSource(ctx context.Context, path string, catalogDir string, allowUnsigned bool) (installSource, error) {
	installationFile, err := config.LoadInstallationFile(path)
	if err != nil {
		return installSource{}, fmt.Errorf("load installation: %w", err)
	}
	installation := installationFile.Installation
	if installation.Namespace == "" {
		installation.Namespace = "m8-system"
	}

	releaseCatalog := catalog.NewFileCatalog(catalogDir)
	releaseCatalog.AllowUnsigned = allowUnsigned
	release, err := releaseCatalog.Resolve(ctx, installation.Spec.PlatformVersion)
	if err != nil {
		return installSource{}, fmt.Errorf("resolve release catalog: %w", err)
	}
	if err := releaseCatalog.Verify(ctx, release); err != nil {
		if errors.Is(err, catalog.ErrUnsignedRelease) {
			return installSource{}, fmt.Errorf("release catalog is unsigned; use --allow-unsigned-release only for local development")
		}
		return installSource{}, fmt.Errorf("verify release catalog: %w", err)
	}
	releaseDigest, err := releaseCatalog.Digest(ctx, release)
	if err != nil {
		return installSource{}, fmt.Errorf("digest release catalog: %w", err)
	}

	plan, err := planner.Generate(ctx, planner.GenerateInput{
		Installation:         installation,
		Release:              release,
		ConfigDigest:         installationFile.Digest,
		ReleaseCatalogDigest: releaseDigest,
		CreatedAt:            a.Now(),
	})
	if err != nil {
		return installSource{}, err
	}
	return installSource{Plan: plan, Installation: installation, Release: release}, nil
}

func (a App) loadUninstallSource(path string) (installSource, error) {
	if path == "" {
		defaultFile, ok := findDefaultInstallationFile()
		if ok {
			path = defaultFile
		}
	}
	if path == "" {
		return installSource{
			Installation: installerv1alpha1.PlatformInstallation{
				Spec: installerv1alpha1.PlatformInstallationSpec{
					GitOps: installerv1alpha1.GitOpsSpec{
						ArgoCD: installerv1alpha1.ArgoCDSpec{Namespace: "argocd"},
					},
				},
			},
		}, nil
	}
	installationFile, err := config.LoadInstallationFile(path)
	if err != nil {
		return installSource{}, fmt.Errorf("load installation: %w", err)
	}
	installation := installationFile.Installation
	if installation.Namespace == "" {
		installation.Namespace = "m8-system"
	}
	plan := planner.InstallationPlan{
		Installation: planner.InstallationRef{
			Name:      installation.Name,
			Namespace: installation.Namespace,
		},
		Release: planner.ReleaseRef{
			Name:    installation.Spec.PlatformVersion,
			Version: installation.Spec.PlatformVersion,
		},
		Profile: installation.Spec.Profile,
	}
	return installSource{Plan: plan, Installation: installation}, nil
}

func isPathOutput(value string) bool {
	lower := strings.ToLower(value)
	return strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml") || strings.HasSuffix(lower, ".json")
}

func readPlanFile(path string) (planner.InstallationPlan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return planner.InstallationPlan{}, fmt.Errorf("read plan file %q: %w", path, err)
	}
	var plan planner.InstallationPlan
	if err := yaml.Unmarshal(data, &plan); err != nil {
		return planner.InstallationPlan{}, fmt.Errorf("parse plan file %q: %w", path, err)
	}
	if plan.Kind != "InstallationPlan" {
		return planner.InstallationPlan{}, fmt.Errorf("plan file %q has kind %q, want InstallationPlan", path, plan.Kind)
	}
	if plan.ConfigDigest == "" || plan.ReleaseCatalogDigest == "" {
		return planner.InstallationPlan{}, fmt.Errorf("plan file %q is missing required digests", path)
	}
	if len(plan.Steps) == 0 {
		return planner.InstallationPlan{}, fmt.Errorf("plan file %q has no steps", path)
	}
	return plan, nil
}

func writePlanFile(path string, plan planner.InstallationPlan) error {
	data, err := yaml.Marshal(plan)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func writeUninstallReport(w io.Writer, format output.Format, report uninstallReport) error {
	switch format {
	case output.FormatTable:
		if _, err := fmt.Fprintf(w, "M8 Installer 1.0\n\nInstallation: %s\nNamespace: %s\nMode: %s\n\n%s\n\n", report.Installation.Name, report.Installation.Namespace, report.Mode, report.Message); err != nil {
			return err
		}
		if len(report.Deleted) > 0 {
			if _, err := fmt.Fprintln(w, "Deleted:"); err != nil {
				return err
			}
			for _, item := range report.Deleted {
				if _, err := fmt.Fprintf(w, "- %s\n", item); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if len(report.Preserved) > 0 {
			if _, err := fmt.Fprintln(w, "Preserved:"); err != nil {
				return err
			}
			for _, item := range report.Preserved {
				if _, err := fmt.Fprintf(w, "- %s\n", item); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if len(report.Skipped) > 0 {
			if _, err := fmt.Fprintln(w, "Skipped:"); err != nil {
				return err
			}
			for _, item := range report.Skipped {
				if _, err := fmt.Fprintf(w, "- %s\n", item); err != nil {
					return err
				}
			}
		}
		return nil
	case output.FormatJSON, output.FormatYAML:
		return output.Write(w, format, report)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

func previewUninstallDeletes(skipGitOpsApplications bool, skipRootGitOps bool, skipArgoCD bool, skipCertManager bool, skipInstallerMetadata bool, deleteNetwork bool, deleteInstallerCRDs bool, deleteNamespaces bool) []string {
	var deleted []string
	if !skipGitOpsApplications {
		deleted = append(deleted, "argocd-applications")
	}
	if !skipRootGitOps {
		deleted = append(deleted, "root-gitops")
	}
	if !skipCertManager {
		deleted = append(deleted, "helm/m8-security/cert-manager")
	}
	if deleteNetwork {
		deleted = append(deleted, "helm/kube-system/cilium")
	}
	if !skipArgoCD {
		deleted = append(deleted, "argocd-bootstrap-manifest")
	}
	if !skipInstallerMetadata {
		deleted = append(deleted, "installer-metadata")
	}
	if deleteInstallerCRDs {
		deleted = append(deleted, "installer-crds")
	}
	if deleteNamespaces {
		deleted = append(deleted, "namespaces")
	}
	return dedupeStrings(deleted)
}

func dedupeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	deduped := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		deduped = append(deduped, value)
	}
	return deduped
}

func writeInstallReport(w io.Writer, format output.Format, report installReport) error {
	switch format {
	case output.FormatTable:
		if _, err := fmt.Fprintf(w, "M8 Installer 1.0\n\nInstallation: %s\nPlatform version: %s\nMode: %s\nApply supported: %t\n\n%s\n\n", report.Installation.Name, report.Release.Version, report.Mode, report.ApplySupported, report.Message); err != nil {
			return err
		}
		if len(report.Applied) > 0 {
			if _, err := fmt.Fprintln(w, "Applied:"); err != nil {
				return err
			}
			for _, item := range report.Applied {
				if _, err := fmt.Fprintf(w, "- %s\n", item); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if len(report.Skipped) > 0 {
			if _, err := fmt.Fprintln(w, "Not applied in this MVP:"); err != nil {
				return err
			}
			for _, item := range report.Skipped {
				if _, err := fmt.Fprintf(w, "- %s\n", item); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		for _, step := range report.Steps {
			if _, err := fmt.Fprintf(w, "%4d  %-28s %s\n", step.Wave, step.ID, step.Title); err != nil {
				return err
			}
		}
		return nil
	case output.FormatJSON, output.FormatYAML:
		return output.Write(w, format, report)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

func findDefaultInstallationFile() (string, bool) {
	candidates := []string{
		"installation.yaml",
		"platform-installation.yaml",
		filepath.Join("gitops", "environments", "production", "platform-installation.yaml"),
		filepath.Join("..", "gitops", "environments", "production", "platform-installation.yaml"),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, true
		}
	}
	return "", false
}

func resolveCatalogDir(path string) string {
	candidates := []string{path}
	if path == "catalog/releases" {
		candidates = append(candidates, filepath.Join("..", "catalog", "releases"))
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return path
}

func loadArgoCDManifest(ctx context.Context, path string, url string) ([]byte, error) {
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read Argo CD manifest %q: %w", path, err)
		}
		return data, nil
	}
	if url == "" {
		return nil, fmt.Errorf("Argo CD manifest URL is empty")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create Argo CD manifest request: %w", err)
	}
	client := http.Client{Timeout: 60 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("download Argo CD manifest from %s: %w", url, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return nil, fmt.Errorf("download Argo CD manifest from %s: HTTP %d", url, response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 25*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("read Argo CD manifest response: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("Argo CD manifest from %s is empty", url)
	}
	return data, nil
}

func loadRootGitOpsManifests() ([][]byte, error) {
	dir, ok := findRootGitOpsDir()
	if !ok {
		return nil, nil
	}
	files := []string{"appproject.yaml", "applicationset.yaml"}
	manifests := make([][]byte, 0, len(files))
	for _, file := range files {
		path := filepath.Join(dir, file)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read root GitOps manifest %q: %w", path, err)
		}
		manifests = append(manifests, data)
	}
	return manifests, nil
}

func findRootGitOpsDir() (string, bool) {
	candidates := []string{
		filepath.Join("gitops", "root"),
		filepath.Join("..", "gitops", "root"),
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, true
		}
	}
	return "", false
}

func (a App) usage() {
	_, _ = fmt.Fprint(a.Stderr, `M8 Installer 1.0

Usage:
  m8ctl preflight -f installation.yaml [--output table|json|yaml]
  m8ctl plan -f installation.yaml [--output table|json|yaml|plan.yaml]
  m8ctl install [--plan plan.yaml|-f installation.yaml] [--dry-run]
  m8ctl status [name] [-n namespace] [--output table|json|yaml]
  m8ctl uninstall [-f installation.yaml] [--dry-run]
  m8ctl uninstall --all --confirmation NAME
  m8ctl version

Defined commands:
  preflight, plan, bootstrap, install, status, doctor, upgrade, rollback,
  backup, restore, bundle, uninstall
`)
}
