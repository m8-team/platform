package port

import (
	"context"

	tenantentity "github.com/m8platform/platform/iam/internal/entity/tenant"
	"github.com/m8platform/platform/iam/internal/usecase/model"
)

type SupportGrantRepository interface {
	Save(ctx context.Context, grant tenantentity.SupportGrant) error
	GetByID(ctx context.Context, supportGrantID string) (tenantentity.SupportGrant, error)
	ListByTenant(ctx context.Context, tenantID string, pageSize int, pageToken string) ([]tenantentity.SupportGrant, string, error)
}

type SupportGrantEventPublisher interface {
	PublishSupportGrantCreated(ctx context.Context, event model.SupportGrantCreatedEvent) error
	PublishSupportGrantRevoked(ctx context.Context, event model.SupportGrantRevokedEvent) error
}

type SupportGrantWorkflowStarter interface {
	StartSupportGrantExpiry(ctx context.Context, workflow model.SupportGrantExpiryWorkflow) error
}
