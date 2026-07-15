package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

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
