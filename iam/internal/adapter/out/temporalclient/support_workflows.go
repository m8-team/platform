package temporalclient

import (
	"context"
	"fmt"

	"github.com/m8platform/platform/iam/internal/temporalx"
	"github.com/m8platform/platform/iam/internal/usecase/model"
)

type SupportGrantWorkflowStarter struct {
	starter *temporalx.WorkflowStarter
}

func NewSupportGrantWorkflowStarter(starter *temporalx.WorkflowStarter) *SupportGrantWorkflowStarter {
	return &SupportGrantWorkflowStarter{starter: starter}
}

func (s *SupportGrantWorkflowStarter) StartSupportGrantExpiry(ctx context.Context, workflow model.SupportGrantExpiryWorkflow) error {
	if s == nil || s.starter == nil {
		return nil
	}
	_, err := s.starter.StartWorkflow(ctx, temporalx.GrantSupportAccessWorkflowName, fmt.Sprintf("grant-support-%s", workflow.SupportGrantID), temporalx.GrantTemporarySupportAccessInput{
		SupportGrantID: workflow.SupportGrantID,
		TenantID:       workflow.TenantID,
		RequestedBy:    workflow.RequestedBy,
		Reason:         workflow.Reason,
		TTL:            workflow.TTL,
	})
	return err
}
