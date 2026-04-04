package identity

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrServiceAccountIDRequired = errors.New("service account id is required")
var ErrTenantIDRequired = errors.New("tenant id is required")
var ErrServiceAccountDisplayNameRequired = errors.New("service account display name is required")

type ServiceAccount struct {
	ID               string
	TenantID         string
	DisplayName      string
	Description      string
	Disabled         bool
	OperationID      string
	KeycloakClientID string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type NewServiceAccountParams struct {
	ID          string
	TenantID    string
	DisplayName string
	Description string
	OperationID string
	Now         time.Time
}

func NewServiceAccount(params NewServiceAccountParams) (ServiceAccount, error) {
	id := strings.TrimSpace(params.ID)
	if id == "" {
		return ServiceAccount{}, ErrServiceAccountIDRequired
	}
	tenantID := strings.TrimSpace(params.TenantID)
	if tenantID == "" {
		return ServiceAccount{}, ErrTenantIDRequired
	}
	displayName := strings.TrimSpace(params.DisplayName)
	if displayName == "" {
		return ServiceAccount{}, ErrServiceAccountDisplayNameRequired
	}
	now := params.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}

	operationID := strings.TrimSpace(params.OperationID)
	if operationID == "" {
		operationID = fmt.Sprintf("op-sa-%d", now.UnixNano())
	}

	return ServiceAccount{
		ID:          id,
		TenantID:    tenantID,
		DisplayName: displayName,
		Description: strings.TrimSpace(params.Description),
		Disabled:    false,
		OperationID: operationID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (s ServiceAccount) WithKeycloakClientID(keycloakClientID string) ServiceAccount {
	s.KeycloakClientID = strings.TrimSpace(keycloakClientID)
	return s
}
