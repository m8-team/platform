package preflight

import (
	"context"
	"time"

	installerv1alpha1 "github.com/m8-team/platform/api/installer/v1alpha1"
)

type Runner struct {
	checks []Check
}

func NewRunner(checks ...Check) Runner {
	return Runner{checks: append([]Check(nil), checks...)}
}

func DefaultRunner(cluster ClusterReader) Runner {
	checks := []Check{
		InstallationValidationCheck{},
		ModuleDependencyCheck{},
	}
	if cluster != nil {
		checks = append(checks,
			KubernetesAPICheck{Cluster: cluster},
			NodeCapacityCheck{Cluster: cluster},
			StorageClassCheck{Cluster: cluster},
			GatewayAPICheck{Cluster: cluster},
		)
	} else {
		checks = append(checks, SkippedClusterCheck{})
	}
	return NewRunner(checks...)
}

func (r Runner) Run(ctx context.Context, installation installerv1alpha1.PlatformInstallation) Report {
	report := Report{
		Installation: installation.Name,
		Profile:      string(installation.Spec.Profile),
		CheckedAt:    time.Now().UTC(),
		Results:      make([]Result, 0, len(r.checks)),
	}

	for _, check := range r.checks {
		startedAt := time.Now()
		result := check.Run(ctx, installation)
		result.Duration = time.Since(startedAt)
		report.Results = append(report.Results, result)
		switch result.Status {
		case StatusPass:
			report.Summary.Passed++
		case StatusWarn:
			report.Summary.Warnings++
		case StatusFail:
			report.Summary.Failed++
		case StatusSkip:
			report.Summary.Skipped++
		}
	}

	return report
}

func (r Report) HasFailures() bool {
	return r.Summary.Failed > 0
}
