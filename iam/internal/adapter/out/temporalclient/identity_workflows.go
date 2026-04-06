package temporalclient

import (
	"context"

	identitymodel "github.com/m8platform/platform/iam/internal/module/iam/model"
)

type IdentityWorkflowStarter struct {
	starter *WorkflowStarter
}

func NewIdentityWorkflowStarter(starter *WorkflowStarter) *IdentityWorkflowStarter {
	return &IdentityWorkflowStarter{starter: starter}
}

func (s *IdentityWorkflowStarter) StartCreateServiceAccount(ctx context.Context, workflow identitymodel.CreateServiceAccountWorkflow) error {
	if s == nil || s.starter == nil {
		return nil
	}
	_, err := s.starter.StartWorkflow(ctx, CreateServiceAccountWorkflowName, workflowIDOrDefault("create-service-account", workflow.OperationID), CreateServiceAccountInput{
		ServiceAccountID: workflow.ServiceAccountID,
		TenantID:         workflow.TenantID,
		DisplayName:      workflow.DisplayName,
		Description:      workflow.Description,
		RequestedBy:      workflow.RequestedBy,
	})
	return err
}

func (s *IdentityWorkflowStarter) StartRotateOAuthClientSecret(ctx context.Context, workflow identitymodel.RotateOAuthClientSecretWorkflow) error {
	if s == nil || s.starter == nil {
		return nil
	}
	_, err := s.starter.StartWorkflow(ctx, RotateClientSecretWorkflowName, workflow.OperationID, RotateClientSecretInput{
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
