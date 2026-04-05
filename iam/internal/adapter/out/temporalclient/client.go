package temporalclient

import (
	"context"

	"github.com/m8platform/platform/iam/internal/foundation/config"
	sdkclient "go.temporal.io/sdk/client"
)

type WorkflowStarter struct {
	client    sdkclient.Client
	taskQueue string
}

func NewWorkflowStarter(cfg config.TemporalConfig) (*WorkflowStarter, error) {
	if !cfg.Enabled {
		return &WorkflowStarter{taskQueue: cfg.TaskQueue}, nil
	}

	temporalClient, err := sdkclient.Dial(sdkclient.Options{
		HostPort:  cfg.Address,
		Namespace: cfg.Namespace,
	})
	if err != nil {
		return nil, err
	}

	return &WorkflowStarter{client: temporalClient, taskQueue: cfg.TaskQueue}, nil
}

func (w *WorkflowStarter) Close() error {
	if w == nil || w.client == nil {
		return nil
	}
	w.client.Close()
	return nil
}

func (w *WorkflowStarter) StartWorkflow(ctx context.Context, workflowName string, workflowID string, input any) (string, error) {
	if w == nil || w.client == nil {
		return workflowID, nil
	}

	run, err := w.client.ExecuteWorkflow(ctx, sdkclient.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: w.taskQueue,
	}, workflowName, input)
	if err != nil {
		return "", err
	}
	return run.GetID(), nil
}
