package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	installerv1alpha1 "github.com/m8platform/platform/api/installer/v1alpha1"
	"github.com/m8platform/platform/internal/installer/planner"
	"github.com/m8platform/platform/internal/installer/preflight"
	"sigs.k8s.io/yaml"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

func ParseFormat(value string) Format {
	switch strings.ToLower(value) {
	case "", "table":
		return FormatTable
	case "json":
		return FormatJSON
	case "yaml", "yml":
		return FormatYAML
	default:
		return Format(value)
	}
}

func Write(w io.Writer, format Format, value any) error {
	switch format {
	case FormatJSON:
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(value)
	case FormatYAML:
		data, err := yaml.Marshal(value)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

func WritePreflight(w io.Writer, format Format, report preflight.Report) error {
	switch format {
	case FormatTable:
		_, err := fmt.Fprintf(w, "M8 Installer 1.0\n\nInstallation: %s\nProfile: %s\n\n", report.Installation, report.Profile)
		if err != nil {
			return err
		}
		for _, result := range report.Results {
			marker := markerForStatus(result.Status)
			if _, err := fmt.Fprintf(w, "%s %-34s %s\n", marker, result.ID, result.Message); err != nil {
				return err
			}
			if result.Remediation != "" && result.Status != preflight.StatusPass {
				if _, err := fmt.Fprintf(w, "  remediation: %s\n", result.Remediation); err != nil {
					return err
				}
			}
		}
		_, err = fmt.Fprintf(w, "\nSummary: %d passed, %d warnings, %d failed, %d skipped\n", report.Summary.Passed, report.Summary.Warnings, report.Summary.Failed, report.Summary.Skipped)
		return err
	case FormatJSON, FormatYAML:
		return Write(w, format, report)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

func WritePlan(w io.Writer, format Format, plan planner.InstallationPlan) error {
	switch format {
	case FormatTable:
		if _, err := fmt.Fprintf(w, "M8 Installer 1.0\n\nInstallation: %s\nPlatform version: %s\nProfile: %s\nConfig digest: %s\nRelease digest: %s\n\n", plan.Installation.Name, plan.Release.Version, plan.Profile, plan.ConfigDigest, plan.ReleaseCatalogDigest); err != nil {
			return err
		}
		for _, step := range plan.Steps {
			if _, err := fmt.Fprintf(w, "%4d  %-28s %s\n", step.Wave, step.ID, step.Title); err != nil {
				return err
			}
		}
		if len(plan.Risks) > 0 {
			if _, err := fmt.Fprintln(w, "\nRisks:"); err != nil {
				return err
			}
			for _, risk := range plan.Risks {
				if _, err := fmt.Fprintf(w, "- %s: %s\n", risk.Severity, risk.Message); err != nil {
					return err
				}
			}
		}
		return nil
	case FormatJSON, FormatYAML:
		return Write(w, format, plan)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

func WriteInstallationsStatus(w io.Writer, format Format, installations []installerv1alpha1.PlatformInstallation) error {
	switch format {
	case FormatTable:
		if _, err := fmt.Fprint(w, "M8 Installer 1.0\n\n"); err != nil {
			return err
		}
		if len(installations) == 0 {
			_, err := fmt.Fprintln(w, "No PlatformInstallation resources found.")
			return err
		}
		if _, err := fmt.Fprintf(w, "%-24s %-16s %-14s %-13s %-18s %-10s\n", "NAME", "NAMESPACE", "PHASE", "VERSION", "COMPONENTS", "ENDPOINTS"); err != nil {
			return err
		}
		for _, installation := range installations {
			phase := installation.Status.Phase
			if phase == "" {
				phase = installerv1alpha1.PhasePending
			}
			if _, err := fmt.Fprintf(
				w,
				"%-24s %-16s %-14s %-13s %-18s %-10s\n",
				installation.Name,
				installation.Namespace,
				phase,
				firstNonEmpty(installation.Status.PlatformVersion, installation.Spec.PlatformVersion),
				componentSummary(installation.Status.Components),
				endpointSummary(installation.Status.Endpoints),
			); err != nil {
				return err
			}
		}
		return nil
	case FormatJSON, FormatYAML:
		return Write(w, format, installations)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

func componentSummary(components []installerv1alpha1.ComponentStatus) string {
	if len(components) == 0 {
		return "0/0 ready"
	}
	ready := 0
	for _, component := range components {
		if component.Ready {
			ready++
		}
	}
	return fmt.Sprintf("%d/%d ready", ready, len(components))
}

func endpointSummary(endpoints []installerv1alpha1.EndpointStatus) string {
	if len(endpoints) == 0 {
		return "0/0 ready"
	}
	ready := 0
	for _, endpoint := range endpoints {
		if endpoint.Ready {
			ready++
		}
	}
	return fmt.Sprintf("%d/%d ready", ready, len(endpoints))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return "-"
}

func markerForStatus(status preflight.Status) string {
	switch status {
	case preflight.StatusPass:
		return "OK"
	case preflight.StatusWarn:
		return "!!"
	case preflight.StatusFail:
		return "XX"
	case preflight.StatusSkip:
		return "--"
	default:
		return "??"
	}
}
