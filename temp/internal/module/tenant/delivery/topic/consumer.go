package topic

import (
	"context"
	"time"

	"github.com/m8platform/platform/iam/internal/module/tenant/model"
	tenantuc "github.com/m8platform/platform/iam/internal/module/tenant/usecase"
	"github.com/m8platform/platform/iam/internal/shared/principal"
	"github.com/m8platform/platform/iam/internal/shared/resource"
)

type SupportAccessConsumer struct {
	useCase *tenantuc.SupportAccessUseCase
}

type GrantSupportAccessMessage struct {
	RequestID    string
	TenantID     string
	SubjectType  string
	SubjectID    string
	ResourceType string
	ResourceID   string
	RoleID       string
	TTL          time.Duration
	Reason       string
	RequestedBy  string
}

func NewSupportAccessConsumer(useCase *tenantuc.SupportAccessUseCase) *SupportAccessConsumer {
	return &SupportAccessConsumer{useCase: useCase}
}

func (c *SupportAccessConsumer) HandleGrantTemporaryAccess(ctx context.Context, message GrantSupportAccessMessage) error {
	if c == nil || c.useCase == nil {
		return nil
	}
	_, err := c.useCase.Grant(ctx, model.GrantSupportAccessCommand{
		RequestID: message.RequestID,
		TenantID:  message.TenantID,
		Subject: principal.Principal{
			TenantID: message.TenantID,
			Type:     message.SubjectType,
			ID:       message.SubjectID,
		},
		Resource: resource.Ref{
			TenantID: message.TenantID,
			Type:     message.ResourceType,
			ID:       message.ResourceID,
		},
		RoleID:      message.RoleID,
		TTL:         message.TTL,
		Reason:      message.Reason,
		RequestedBy: message.RequestedBy,
	})
	return err
}
