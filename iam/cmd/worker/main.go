package main

import (
	"github.com/m8platform/platform/iam/internal/config"
	"github.com/m8platform/platform/iam/internal/observability"
	"github.com/m8platform/platform/iam/internal/temporalx"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	cfg := config.Load()
	logger, err := observability.NewLogger(cfg.Development)
	if err != nil {
		panic(err)
	}

	starter, err := temporalx.NewWorkflowStarter(cfg.Temporal)
	if err != nil {
		panic(err)
	}
	defer starter.Close()

	if !cfg.Temporal.Enabled {
		logger.Info("temporal disabled; worker exits without starting")
		return
	}

	temporalClient, err := client.Dial(client.Options{
		HostPort:  cfg.Temporal.Address,
		Namespace: cfg.Temporal.Namespace,
	})
	if err != nil {
		panic(err)
	}
	defer temporalClient.Close()

	w := worker.New(temporalClient, cfg.Temporal.TaskQueue, worker.Options{})
	activities := &temporalx.Activities{Logger: logger}
	w.RegisterWorkflow(temporalx.CreateServiceAccountWorkflow)
	w.RegisterWorkflow(temporalx.RotateClientSecretWorkflow)
	w.RegisterWorkflow(temporalx.GrantTemporarySupportAccessWorkflow)
	w.RegisterWorkflow(temporalx.SyncRelationshipsToSpiceDBWorkflow)
	w.RegisterWorkflow(temporalx.RebuildAccessReadModelsWorkflow)
	w.RegisterWorkflow(temporalx.ImportFederatedUserWorkflow)
	w.RegisterActivity(activities)

	if err := w.Run(worker.InterruptCh()); err != nil {
		panic(err)
	}
}
