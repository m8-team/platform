package topic

import (
	"context"
	"time"

	authzentity "github.com/m8platform/platform/iam/internal/entity/authz"
	"github.com/m8platform/platform/iam/internal/usecase/model"
	tenantuc "github.com/m8platform/platform/iam/internal/usecase/tenant"
)

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

type SupportAccessConsumer struct {
	useCase *tenantuc.SupportAccessUseCase
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
		Subject: authzentity.SubjectRef{
			TenantID: message.TenantID,
			Type:     message.SubjectType,
			ID:       message.SubjectID,
		},
		Resource: authzentity.ResourceRef{
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
