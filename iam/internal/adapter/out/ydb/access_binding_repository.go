package ydb

import (
	"context"

	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	legacycore "github.com/m8platform/platform/iam/internal/core"
	authzentity "github.com/m8platform/platform/iam/internal/module/authz/entity"
	"github.com/m8platform/platform/iam/internal/shared/principal"
	"github.com/m8platform/platform/iam/internal/shared/resource"
)

type AccessBindingRepository struct {
	store legacycore.DocumentStore
}

func NewAccessBindingRepository(store legacycore.DocumentStore) *AccessBindingRepository {
	return &AccessBindingRepository{store: store}
}

func (r *AccessBindingRepository) ListByResource(ctx context.Context, ref resource.Ref) ([]authzentity.AccessBinding, error) {
	documents, _, err := r.store.ListDocuments(ctx, TableBindingOperations, ref.TenantID, 0, 1000)
	if err != nil {
		return nil, err
	}

	bindings := make([]authzentity.AccessBinding, 0, len(documents))
	for _, document := range documents {
		record := &authzv1.AccessBinding{}
		if err := legacycore.UnmarshalProto(document.Payload, record); err != nil {
			return nil, err
		}
		if record.GetResource().GetType().String() != ref.Type || record.GetResource().GetId() != ref.ID {
			continue
		}
		bindings = append(bindings, authzentity.AccessBinding{
			ID:     record.GetBindingId(),
			RoleID: record.GetRoleId(),
			Subject: principal.Principal{
				TenantID: record.GetSubject().GetTenantId(),
				Type:     record.GetSubject().GetType().String(),
				ID:       record.GetSubject().GetId(),
			},
			Resource: resource.Ref{
				TenantID: record.GetResource().GetTenantId(),
				Type:     record.GetResource().GetType().String(),
				ID:       record.GetResource().GetId(),
			},
		})
	}
	return bindings, nil
}
