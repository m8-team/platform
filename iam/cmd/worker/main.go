package main

import (
	temporaladapter "github.com/m8platform/platform/iam/internal/adapter/out/temporalclient"
	"github.com/m8platform/platform/iam/internal/foundation/config"
	foundationlogging "github.com/m8platform/platform/iam/internal/foundation/logging"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func main() {
	cfg := config.Load()
	logger, err := foundationlogging.New(cfg.Development)
	if err != nil {
		panic(err)
	}

	starter, err := temporaladapter.NewWorkflowStarter(cfg.Temporal)
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
	activities := &temporaladapter.Activities{Logger: logger}
	w.RegisterWorkflowWithOptions(temporaladapter.CreateServiceAccountWorkflow, workflow.RegisterOptions{
		Name: temporaladapter.CreateServiceAccountWorkflowName,
	})
	w.RegisterWorkflowWithOptions(temporaladapter.RotateClientSecretWorkflow, workflow.RegisterOptions{
		Name: temporaladapter.RotateClientSecretWorkflowName,
	})
	w.RegisterWorkflowWithOptions(temporaladapter.GrantTemporarySupportAccessWorkflow, workflow.RegisterOptions{
		Name: temporaladapter.GrantSupportAccessWorkflowName,
	})
	w.RegisterWorkflowWithOptions(temporaladapter.SyncRelationshipsToSpiceDBWorkflow, workflow.RegisterOptions{
		Name: temporaladapter.SyncRelationshipsWorkflowName,
	})
	w.RegisterWorkflowWithOptions(temporaladapter.RebuildAccessReadModelsWorkflow, workflow.RegisterOptions{
		Name: temporaladapter.RebuildAccessReadModelsWorkflowName,
	})
	w.RegisterWorkflowWithOptions(temporaladapter.ImportFederatedUserWorkflow, workflow.RegisterOptions{
		Name: temporaladapter.ImportFederatedUserWorkflowName,
	})
	w.RegisterActivity(activities)

	if err := w.Run(worker.InterruptCh()); err != nil {
		panic(err)
	}
}
