package ydb

import (
	"context"

	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	legacycore "github.com/m8platform/platform/iam/internal/core"
	authzentity "github.com/m8platform/platform/iam/internal/entity/authz"
	legacystorage "github.com/m8platform/platform/iam/internal/storage/ydb"
)

type AccessBindingRepository struct {
	store legacycore.DocumentStore
}

func NewAccessBindingRepository(store legacycore.DocumentStore) *AccessBindingRepository {
	return &AccessBindingRepository{store: store}
}

func (r *AccessBindingRepository) ListByResource(ctx context.Context, resource authzentity.ResourceRef) ([]authzentity.AccessBinding, error) {
	documents, _, err := r.store.ListDocuments(ctx, legacystorage.TableBindingOperations, resource.TenantID, 0, 1000)
	if err != nil {
		return nil, err
	}

	bindings := make([]authzentity.AccessBinding, 0, len(documents))
	for _, document := range documents {
		record := &authzv1.AccessBinding{}
		if err := legacycore.UnmarshalProto(document.Payload, record); err != nil {
			return nil, err
		}
		if record.GetResource().GetType().String() != resource.Type || record.GetResource().GetId() != resource.ID {
			continue
		}
		bindings = append(bindings, authzentity.AccessBinding{
			ID:     record.GetBindingId(),
			RoleID: record.GetRoleId(),
			Subject: authzentity.SubjectRef{
				TenantID: record.GetSubject().GetTenantId(),
				Type:     record.GetSubject().GetType().String(),
				ID:       record.GetSubject().GetId(),
			},
			Resource: authzentity.ResourceRef{
				TenantID: record.GetResource().GetTenantId(),
				Type:     record.GetResource().GetType().String(),
				ID:       record.GetResource().GetId(),
			},
		})
	}
	return bindings, nil
}
