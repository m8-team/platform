package temporalclient

import (
	"context"
	"fmt"

	tenantmodel "github.com/m8platform/platform/iam/internal/module/tenant/model"
)

type SupportGrantWorkflowStarter struct {
	starter *WorkflowStarter
}

func NewSupportGrantWorkflowStarter(starter *WorkflowStarter) *SupportGrantWorkflowStarter {
	return &SupportGrantWorkflowStarter{starter: starter}
}

func (s *SupportGrantWorkflowStarter) StartSupportGrantExpiry(ctx context.Context, workflow tenantmodel.SupportGrantExpiryWorkflow) error {
	if s == nil || s.starter == nil {
		return nil
	}
	_, err := s.starter.StartWorkflow(ctx, GrantSupportAccessWorkflowName, fmt.Sprintf("grant-support-%s", workflow.SupportGrantID), GrantTemporarySupportAccessInput{
		SupportGrantID: workflow.SupportGrantID,
		TenantID:       workflow.TenantID,
		RequestedBy:    workflow.RequestedBy,
		Reason:         workflow.Reason,
		TTL:            workflow.TTL,
	})
	return err
}
