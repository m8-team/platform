package model

import (
	"time"

	identityentity "github.com/m8platform/platform/iam/internal/entity/identity"
)

type CreateServiceAccountCommand struct {
	ServiceAccountID string
	TenantID         string
	DisplayName      string
	Description      string
	PerformedBy      string
}

type CreateServiceAccountWorkflow struct {
	OperationID      string
	ServiceAccountID string
	TenantID         string
	DisplayName      string
	Description      string
	RequestedBy      string
}

type ServiceAccountCreatedEvent struct {
	OperationID string
	OccurredAt  time.Time
	PerformedBy string
	Account     identityentity.ServiceAccount
}

type CreateServiceAccountResult struct {
	Account  identityentity.ServiceAccount
	Warnings []error
}

type RotateOAuthClientSecretCommand struct {
	OAuthClientID string
	PerformedBy   string
	Reason        string
}

type RotateOAuthClientSecretWorkflow struct {
	OperationID   string
	OAuthClientID string
	RequestedBy   string
	Reason        string
}

type RotateOAuthClientSecretResult struct {
	OperationID string
	SecretRef   string
	Warnings    []error
}
