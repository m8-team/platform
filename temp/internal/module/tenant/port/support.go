package port

import (
	"context"

	"github.com/m8platform/platform/iam/internal/module/tenant/entity"
	"github.com/m8platform/platform/iam/internal/module/tenant/model"
)

type SupportGrantRepository interface {
	Save(ctx context.Context, grant entity.SupportGrant) error
	GetByID(ctx context.Context, supportGrantID string) (entity.SupportGrant, error)
	ListByTenant(ctx context.Context, tenantID string, pageSize int, pageToken string) ([]entity.SupportGrant, string, error)
}

type SupportGrantEventPublisher interface {
	PublishSupportGrantCreated(ctx context.Context, event model.SupportGrantCreatedEvent) error
	PublishSupportGrantRevoked(ctx context.Context, event model.SupportGrantRevokedEvent) error
}

type SupportGrantWorkflowStarter interface {
	StartSupportGrantExpiry(ctx context.Context, workflow model.SupportGrantExpiryWorkflow) error
}
