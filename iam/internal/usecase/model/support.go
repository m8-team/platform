package model

import (
	"time"

	authzentity "github.com/m8platform/platform/iam/internal/entity/authz"
	tenantentity "github.com/m8platform/platform/iam/internal/entity/tenant"
)

type GrantSupportAccessCommand struct {
	RequestID   string
	TenantID    string
	Subject     authzentity.SubjectRef
	Resource    authzentity.ResourceRef
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
	Grant       tenantentity.SupportGrant
}

type SupportGrantRevokedEvent struct {
	EventID    string
	OccurredAt time.Time
	RevokedBy  string
	Reason     string
	Grant      tenantentity.SupportGrant
}

type SupportGrantExpiryWorkflow struct {
	SupportGrantID string
	TenantID       string
	RequestedBy    string
	Reason         string
	TTL            time.Duration
}

type SupportGrantResult struct {
	Grant    tenantentity.SupportGrant
	Warnings []error
}
