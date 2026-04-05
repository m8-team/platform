package main

import (
	"github.com/m8platform/platform/iam/internal/foundation/config"
	"github.com/m8platform/platform/iam/internal/observability"
	"github.com/m8platform/platform/iam/internal/temporalx"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
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
	w.RegisterWorkflowWithOptions(temporalx.CreateServiceAccountWorkflow, workflow.RegisterOptions{
		Name: temporalx.CreateServiceAccountWorkflowName,
	})
	w.RegisterWorkflowWithOptions(temporalx.RotateClientSecretWorkflow, workflow.RegisterOptions{
		Name: temporalx.RotateClientSecretWorkflowName,
	})
	w.RegisterWorkflowWithOptions(temporalx.GrantTemporarySupportAccessWorkflow, workflow.RegisterOptions{
		Name: temporalx.GrantSupportAccessWorkflowName,
	})
	w.RegisterWorkflowWithOptions(temporalx.SyncRelationshipsToSpiceDBWorkflow, workflow.RegisterOptions{
		Name: temporalx.SyncRelationshipsWorkflowName,
	})
	w.RegisterWorkflowWithOptions(temporalx.RebuildAccessReadModelsWorkflow, workflow.RegisterOptions{
		Name: temporalx.RebuildAccessReadModelsWorkflowName,
	})
	w.RegisterWorkflowWithOptions(temporalx.ImportFederatedUserWorkflow, workflow.RegisterOptions{
		Name: temporalx.ImportFederatedUserWorkflowName,
	})
	w.RegisterActivity(activities)

	if err := w.Run(worker.InterruptCh()); err != nil {
		panic(err)
	}
}
