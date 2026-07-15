package helm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	helmcli "helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
)

type SDKClient struct {
	Kubeconfig string
	Context    string
	Timeout    time.Duration
	DebugLog   func(format string, v ...any)
}

func (c SDKClient) Plan(ctx context.Context, release Release) (ChangeSet, error) {
	if err := ctx.Err(); err != nil {
		return ChangeSet{}, err
	}
	return ChangeSet{Update: []string{release.Namespace + "/" + release.Name}}, nil
}

func (c SDKClient) Apply(ctx context.Context, release Release) error {
	if release.Name == "" {
		return fmt.Errorf("helm release name is required")
	}
	if release.Namespace == "" {
		return fmt.Errorf("helm release namespace is required")
	}
	if release.Chart == "" {
		return fmt.Errorf("helm chart is required for release %s", release.Name)
	}

	settings := helmcli.New()
	settings.KubeConfig = c.Kubeconfig
	settings.KubeContext = c.Context
	settings.SetNamespace(release.Namespace)

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), release.Namespace, "secret", c.logf); err != nil {
		return fmt.Errorf("initialize Helm action configuration for %s/%s: %w", release.Namespace, release.Name, err)
	}

	chartPathOptions := action.ChartPathOptions{
		RepoURL: release.Repository,
		Version: release.Version,
	}
	chartPath, err := chartPathOptions.LocateChart(release.Chart, settings)
	if err != nil {
		return fmt.Errorf("locate Helm chart %s version %s: %w", release.Chart, release.Version, err)
	}
	chart, err := loader.Load(chartPath)
	if err != nil {
		return fmt.Errorf("load Helm chart %s: %w", chartPath, err)
	}

	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}

	upgrade := action.NewUpgrade(actionConfig)
	upgrade.Namespace = release.Namespace
	upgrade.Timeout = timeout
	upgrade.Wait = true
	upgrade.WaitForJobs = true
	upgrade.ChartPathOptions = chartPathOptions
	upgrade.TakeOwnership = true

	if _, err := upgrade.RunWithContext(ctx, release.Name, chart, release.Values); err != nil {
		if !isReleaseNotFound(err) {
			return fmt.Errorf("upgrade Helm release %s/%s: %w", release.Namespace, release.Name, err)
		}

		install := action.NewInstall(actionConfig)
		install.ReleaseName = release.Name
		install.Namespace = release.Namespace
		install.CreateNamespace = true
		install.Timeout = timeout
		install.Wait = true
		install.WaitForJobs = true
		install.ChartPathOptions = chartPathOptions
		install.TakeOwnership = true

		if _, err := install.RunWithContext(ctx, chart, release.Values); err != nil {
			return fmt.Errorf("install Helm release %s/%s: %w", release.Namespace, release.Name, err)
		}
	}

	return nil
}

func (c SDKClient) Status(ctx context.Context, namespace string, name string) (Status, error) {
	if err := ctx.Err(); err != nil {
		return Status{}, err
	}
	settings := helmcli.New()
	settings.KubeConfig = c.Kubeconfig
	settings.KubeContext = c.Context
	settings.SetNamespace(namespace)

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, "secret", c.logf); err != nil {
		return Status{}, fmt.Errorf("initialize Helm action configuration for %s/%s: %w", namespace, name, err)
	}
	statusAction := action.NewStatus(actionConfig)
	release, err := statusAction.Run(name)
	if err != nil {
		return Status{}, fmt.Errorf("get Helm release status %s/%s: %w", namespace, name, err)
	}
	return Status{
		Name:      release.Name,
		Namespace: release.Namespace,
		Revision:  release.Version,
		Phase:     release.Info.Status.String(),
	}, nil
}

func (c SDKClient) Rollback(ctx context.Context, namespace string, name string, revision int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	settings := helmcli.New()
	settings.KubeConfig = c.Kubeconfig
	settings.KubeContext = c.Context
	settings.SetNamespace(namespace)

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, "secret", c.logf); err != nil {
		return fmt.Errorf("initialize Helm action configuration for %s/%s: %w", namespace, name, err)
	}
	rollback := action.NewRollback(actionConfig)
	rollback.Version = revision
	rollback.Wait = true
	rollback.Timeout = c.Timeout
	if rollback.Timeout <= 0 {
		rollback.Timeout = 10 * time.Minute
	}
	if err := rollback.Run(name); err != nil {
		return fmt.Errorf("rollback Helm release %s/%s to revision %d: %w", namespace, name, revision, err)
	}
	return nil
}

func (c SDKClient) logf(format string, v ...any) {
	if c.DebugLog != nil {
		c.DebugLog(format, v...)
	}
}

func isReleaseNotFound(err error) bool {
	return errors.Is(err, driver.ErrReleaseNotFound) ||
		errors.Is(err, driver.ErrNoDeployedReleases) ||
		strings.Contains(strings.ToLower(err.Error()), "release: not found") ||
		strings.Contains(strings.ToLower(err.Error()), "has no deployed releases")
}
