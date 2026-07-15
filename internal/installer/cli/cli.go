package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	installerv1alpha1 "github.com/m8platform/platform/api/installer/v1alpha1"
	"github.com/m8platform/platform/internal/installer/catalog"
	"github.com/m8platform/platform/internal/installer/config"
	installerkubernetes "github.com/m8platform/platform/internal/installer/kubernetes"
	"github.com/m8platform/platform/internal/installer/output"
	"github.com/m8platform/platform/internal/installer/planner"
	"github.com/m8platform/platform/internal/installer/preflight"
	"sigs.k8s.io/yaml"
)

const Version = "1.0.0-dev"

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
	case "bootstrap", "install", "doctor", "upgrade", "rollback", "backup", "restore", "bundle", "uninstall":
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

	installationFile, err := config.LoadInstallationFile(*file)
	if err != nil {
		_, _ = fmt.Fprintf(a.Stderr, "load installation: %v\n", err)
		return ExitError
	}

	releaseCatalog := catalog.NewFileCatalog(*catalogDir)
	releaseCatalog.AllowUnsigned = *allowUnsigned
	release, err := releaseCatalog.Resolve(ctx, installationFile.Installation.Spec.PlatformVersion)
	if err != nil {
		_, _ = fmt.Fprintf(a.Stderr, "resolve release catalog: %v\n", err)
		return ExitError
	}
	if err := releaseCatalog.Verify(ctx, release); err != nil {
		if errors.Is(err, catalog.ErrUnsignedRelease) {
			_, _ = fmt.Fprintln(a.Stderr, "release catalog is unsigned; use --allow-unsigned-release only for local development")
		} else {
			_, _ = fmt.Fprintf(a.Stderr, "verify release catalog: %v\n", err)
		}
		return ExitError
	}
	releaseDigest, err := releaseCatalog.Digest(ctx, release)
	if err != nil {
		_, _ = fmt.Fprintf(a.Stderr, "digest release catalog: %v\n", err)
		return ExitError
	}

	plan, err := planner.Generate(ctx, planner.GenerateInput{
		Installation:         installationFile.Installation,
		Release:              release,
		ConfigDigest:         installationFile.Digest,
		ReleaseCatalogDigest: releaseDigest,
		CreatedAt:            a.Now(),
	})
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

func isPathOutput(value string) bool {
	lower := strings.ToLower(value)
	return strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml") || strings.HasSuffix(lower, ".json")
}

func writePlanFile(path string, plan planner.InstallationPlan) error {
	data, err := yaml.Marshal(plan)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func (a App) usage() {
	_, _ = fmt.Fprint(a.Stderr, `M8 Installer 1.0

Usage:
  m8ctl preflight -f installation.yaml [--output table|json|yaml]
  m8ctl plan -f installation.yaml [--output table|json|yaml|plan.yaml]
  m8ctl version

Defined commands:
  preflight, plan, bootstrap, install, status, doctor, upgrade, rollback,
  backup, restore, bundle, uninstall
`)
}
