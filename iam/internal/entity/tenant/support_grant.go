package tenant

import (
	"errors"
	"fmt"
	"strings"
	"time"

	authzentity "github.com/m8platform/platform/iam/internal/entity/authz"
)

var ErrSupportGrantTenantRequired = errors.New("support grant tenant id is required")
var ErrSupportGrantRoleRequired = errors.New("support grant role id is required")
var ErrSupportGrantReasonRequired = errors.New("support grant reason is required")
var ErrSupportGrantTTLRequired = errors.New("support grant ttl must be positive")
var ErrSupportGrantSubjectRequired = errors.New("support grant subject is required")
var ErrSupportGrantResourceRequired = errors.New("support grant resource is required")

type SupportGrantStatus string

const (
	SupportGrantStatusPendingApproval SupportGrantStatus = "pending_approval"
	SupportGrantStatusActive          SupportGrantStatus = "active"
	SupportGrantStatusRevoked         SupportGrantStatus = "revoked"
	SupportGrantStatusExpired         SupportGrantStatus = "expired"
	SupportGrantStatusDenied          SupportGrantStatus = "denied"
)

type SupportGrant struct {
	ID             string
	TenantID       string
	Subject        authzentity.SubjectRef
	Resource       authzentity.ResourceRef
	RoleID         string
	TTL            time.Duration
	Status         SupportGrantStatus
	Reason         string
	ApprovalTicket string
	ApprovedAt     time.Time
	ExpiresAt      time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type NewSupportGrantParams struct {
	ID       string
	TenantID string
	Subject  authzentity.SubjectRef
	Resource authzentity.ResourceRef
	RoleID   string
	TTL      time.Duration
	Reason   string
	Now      time.Time
}

func NewSupportGrant(params NewSupportGrantParams) (SupportGrant, error) {
	tenantID := strings.TrimSpace(params.TenantID)
	if tenantID == "" {
		return SupportGrant{}, ErrSupportGrantTenantRequired
	}
	if strings.TrimSpace(params.RoleID) == "" {
		return SupportGrant{}, ErrSupportGrantRoleRequired
	}
	if strings.TrimSpace(params.Reason) == "" {
		return SupportGrant{}, ErrSupportGrantReasonRequired
	}
	if params.TTL <= 0 {
		return SupportGrant{}, ErrSupportGrantTTLRequired
	}
	if strings.TrimSpace(params.Subject.ID) == "" || strings.TrimSpace(params.Subject.Type) == "" {
		return SupportGrant{}, ErrSupportGrantSubjectRequired
	}
	if strings.TrimSpace(params.Resource.ID) == "" || strings.TrimSpace(params.Resource.Type) == "" {
		return SupportGrant{}, ErrSupportGrantResourceRequired
	}

	now := params.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	id := strings.TrimSpace(params.ID)
	if id == "" {
		id = fmt.Sprintf("support-%d", now.UnixNano())
	}

	return SupportGrant{
		ID:       id,
		TenantID: tenantID,
		Subject: authzentity.SubjectRef{
			TenantID: tenantID,
			Type:     strings.TrimSpace(params.Subject.Type),
			ID:       strings.TrimSpace(params.Subject.ID),
		},
		Resource: authzentity.ResourceRef{
			TenantID: tenantID,
			Type:     strings.TrimSpace(params.Resource.Type),
			ID:       strings.TrimSpace(params.Resource.ID),
		},
		RoleID:    strings.TrimSpace(params.RoleID),
		TTL:       params.TTL,
		Status:    SupportGrantStatusPendingApproval,
		Reason:    strings.TrimSpace(params.Reason),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (g SupportGrant) Approve(approvalTicket string, now time.Time) SupportGrant {
	now = now.UTC()
	g.Status = SupportGrantStatusActive
	g.ApprovalTicket = strings.TrimSpace(approvalTicket)
	g.ApprovedAt = now
	g.ExpiresAt = now.Add(g.TTL)
	g.UpdatedAt = now
	return g
}

func (g SupportGrant) Revoke(now time.Time) SupportGrant {
	g.Status = SupportGrantStatusRevoked
	g.UpdatedAt = now.UTC()
	return g
}
