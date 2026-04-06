package model

import (
	"time"

	"github.com/m8platform/platform/iam/internal/module/iam/entity"
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
	Account     entity.ServiceAccount
}

type CreateServiceAccountResult struct {
	Account  entity.ServiceAccount
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
