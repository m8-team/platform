package ydb

import (
	"context"
	"encoding/json"

	legacycore "github.com/m8platform/platform/iam/internal/core"
	identityentity "github.com/m8platform/platform/iam/internal/module/iam/entity"
)

type serviceAccountDocument struct {
	ServiceAccountID string `json:"service_account_id"`
	TenantID         string `json:"tenant_id"`
	DisplayName      string `json:"display_name"`
	Description      string `json:"description,omitempty"`
	Disabled         bool   `json:"disabled"`
	OperationID      string `json:"operation_id,omitempty"`
	KeycloakClientID string `json:"keycloak_client_id,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
	UpdatedAt        string `json:"updated_at,omitempty"`
}

type ServiceAccountRepository struct {
	store legacycore.DocumentStore
}

func NewServiceAccountRepository(store legacycore.DocumentStore) *ServiceAccountRepository {
	return &ServiceAccountRepository{store: store}
}

func (r *ServiceAccountRepository) Save(ctx context.Context, account identityentity.ServiceAccount) error {
	payload, err := json.Marshal(serviceAccountDocument{
		ServiceAccountID: account.ID,
		TenantID:         account.TenantID,
		DisplayName:      account.DisplayName,
		Description:      account.Description,
		Disabled:         account.Disabled,
		OperationID:      account.OperationID,
		KeycloakClientID: account.KeycloakClientID,
		CreatedAt:        account.CreatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
		UpdatedAt:        account.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
	})
	if err != nil {
		return err
	}

	return r.store.UpsertDocument(ctx, TableServiceAccounts, legacycore.StoredDocument{
		ID:        account.ID,
		TenantID:  account.TenantID,
		Payload:   payload,
		CreatedAt: account.CreatedAt.UTC(),
		UpdatedAt: account.UpdatedAt.UTC(),
	})
}
