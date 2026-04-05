package model

import (
	"time"

	"github.com/m8platform/platform/iam/internal/module/tenant/entity"
	"github.com/m8platform/platform/iam/internal/shared/principal"
	"github.com/m8platform/platform/iam/internal/shared/resource"
)

type GrantSupportAccessCommand struct {
	RequestID   string
	TenantID    string
	Subject     principal.Principal
	Resource    resource.Ref
	RoleID      string
	TTL         time.Duration
	Reason      string
	RequestedBy string
}

type ApproveSupportAccessCommand struct {
	SupportGrantID string
	ApprovalTicket string
	Reason         string
	ApprovedBy     string
}

type RevokeSupportAccessCommand struct {
	SupportGrantID string
	Reason         string
	RevokedBy      string
}

type ListSupportGrantsQuery struct {
	TenantID  string
	PageSize  int
	PageToken string
}

type SupportGrantCreatedEvent struct {
	EventID     string
	OccurredAt  time.Time
	RequestedBy string
	Grant       entity.SupportGrant
}

type SupportGrantRevokedEvent struct {
	EventID    string
	OccurredAt time.Time
	RevokedBy  string
	Reason     string
	Grant      entity.SupportGrant
}

type SupportGrantExpiryWorkflow struct {
	SupportGrantID string
	TenantID       string
	RequestedBy    string
	Reason         string
	TTL            time.Duration
}

type SupportGrantResult struct {
	Grant    entity.SupportGrant
	Warnings []error
}
