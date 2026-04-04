package temporalclient

import (
	"context"

	"github.com/m8platform/platform/iam/internal/temporalx"
	"github.com/m8platform/platform/iam/internal/usecase/model"
)

type IdentityWorkflowStarter struct {
	starter *temporalx.WorkflowStarter
}

func NewIdentityWorkflowStarter(starter *temporalx.WorkflowStarter) *IdentityWorkflowStarter {
	return &IdentityWorkflowStarter{starter: starter}
}

func (s *IdentityWorkflowStarter) StartCreateServiceAccount(ctx context.Context, workflow model.CreateServiceAccountWorkflow) error {
	if s == nil || s.starter == nil {
		return nil
	}
	_, err := s.starter.StartWorkflow(ctx, temporalx.CreateServiceAccountWorkflowName, workflowIDOrDefault("create-service-account", workflow.OperationID), temporalx.CreateServiceAccountInput{
		ServiceAccountID: workflow.ServiceAccountID,
		TenantID:         workflow.TenantID,
		DisplayName:      workflow.DisplayName,
		Description:      workflow.Description,
		RequestedBy:      workflow.RequestedBy,
	})
	return err
}

func (s *IdentityWorkflowStarter) StartRotateOAuthClientSecret(ctx context.Context, workflow model.RotateOAuthClientSecretWorkflow) error {
	if s == nil || s.starter == nil {
		return nil
	}
	_, err := s.starter.StartWorkflow(ctx, temporalx.RotateClientSecretWorkflowName, workflow.OperationID, temporalx.RotateClientSecretInput{
		OAuthClientID: workflow.OAuthClientID,
		RequestedBy:   workflow.RequestedBy,
		Reason:        workflow.Reason,
	})
	return err
}

func workflowIDOrDefault(prefix string, entityID string) string {
	if entityID == "" {
		return prefix
	}
	return prefix + "-" + entityID
}
