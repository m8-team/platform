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
	installerinstall "github.com/m8platform/platform/internal/installer/install"
	installerkubernetes "github.com/m8platform/platform/internal/installer/kubernetes"
	"github.com/m8platform/platform/internal/installer/output"
	"github.com/m8platform/platform/internal/installer/planner"
	"github.com/m8platform/platform/internal/installer/preflight"
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
	case "bootstrap", "doctor", "upgrade", "rollback", "backup", "restore", "bundle", "uninstall":
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
	argoCDManifestPath := flags.String("argocd-manifest", "", "Local Argo CD install manifest")
	argoCDManifestURL := flags.String("argocd-manifest-url", defaultArgoCDManifestURL, "Argo CD install manifest URL")
	skipArgoCD := flags.Bool("skip-argocd", false, "Skip Argo CD manifest installation")
	skipRootGitOps := flags.Bool("skip-root-gitops", false, "Skip root AppProject/ApplicationSet installation")
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
	result, err := installerinstall.Executor{Kubernetes: client, Now: a.Now}.Apply(ctx, installerinstall.Request{
		Plan:                source.Plan,
		Installation:        source.Installation,
		Release:             source.Release,
		ArgoCDManifest:      argoCDManifest,
		RootGitOpsManifests: rootGitOpsManifests,
		SkipArgoCD:          *skipArgoCD,
		SkipRootGitOps:      *skipRootGitOps,
	})
	if err != nil {
		_, _ = fmt.Fprintf(a.Stderr, "install failed: %v\n", err)
		return ExitError
	}
	report.Mode = "applied"
	report.Message = "Applied installer API resources, namespaces, PlatformRelease, PlatformInstallation and InstallationOperation. Remaining Helm/GitOps component reconciliation is still planned but not executed by this MVP."
	report.Applied = result.Applied
	report.Skipped = result.Skipped
	report.Operation = result.Operation.Namespace + "/" + result.Operation.Name

	if err := writeInstallReport(a.Stdout, output.ParseFormat(*outputValue), report); err != nil {
		_, _ = fmt.Fprintf(a.Stderr, "write output: %v\n", err)
		return ExitError
	}
	return ExitOK
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
  m8ctl version

Defined commands:
  preflight, plan, bootstrap, install, status, doctor, upgrade, rollback,
  backup, restore, bundle, uninstall
`)
}
