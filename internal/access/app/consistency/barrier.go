// Package consistency enforces a trusted Resource Manager to SpiceDB freshness
// barrier before permitting critical Access decisions.
package consistency

import (
	"context"
	"errors"

	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/domain"
)

var (
	ErrInvalidBarrier              = errors.New("invalid Access consistency barrier")
	ErrProjectionLag               = errors.New("authorization resource projection is behind Resource Manager")
	ErrResourceRevisionUnavailable = errors.New("authoritative Resource Manager revision unavailable")
)

// ResourceRevisionReader MUST query the authoritative Resource Manager owner and
// enforce the trusted tenant on the resource. A client-supplied revision is not
// acceptable for a security-critical barrier.
type ResourceRevisionReader interface {
	AuthoritativeRevision(context.Context, domain.TenantID, domain.Resource) (uint64, error)
}

type Watermark struct {
	Revision uint64
	ZedToken string
}

func (w Watermark) covers(revision uint64) bool {
	return w.Revision >= revision && w.ZedToken != ""
}

type WatermarkReader interface {
	LoadWatermark(context.Context, domain.TenantID, domain.Resource) (Watermark, error)
}

// Barrier never returns a positive decision until Resource Manager's latest
// authoritative revision is acknowledged by the graph projection. This does
// not replace Access grant/role revocation fencing, which is separate.
type Barrier struct {
	source   ResourceRevisionReader
	progress WatermarkReader
	pdp      ports.PermissionEngine
}

func New(source ResourceRevisionReader, progress WatermarkReader, pdp ports.PermissionEngine) (*Barrier, error) {
	if source == nil || progress == nil || pdp == nil {
		return nil, ErrInvalidBarrier
	}
	return &Barrier{source: source, progress: progress, pdp: pdp}, nil
}

func (b *Barrier) CheckPermission(ctx context.Context, req domain.CheckPermissionRequest) (domain.PermissionDecision, error) {
	decision, _, err := b.checkResource(ctx, req, false)
	return decision, err
}

func (b *Barrier) checkResource(ctx context.Context, req domain.CheckPermissionRequest, fullyConsistent bool) (domain.PermissionDecision, uint64, error) {
	if b == nil || b.source == nil || b.progress == nil || b.pdp == nil {
		return domain.PermissionDecision{}, 0, ErrInvalidBarrier
	}
	if err := req.Validate(); err != nil {
		return domain.PermissionDecision{}, 0, err
	}
	if err := ctx.Err(); err != nil {
		return domain.PermissionDecision{}, 0, err
	}
	latest, err := b.source.AuthoritativeRevision(ctx, req.TenantID(), req.Resource())
	if err != nil {
		return domain.PermissionDecision{}, 0, err
	}
	if latest == 0 {
		return domain.PermissionDecision{}, 0, ErrResourceRevisionUnavailable
	}
	committed, err := b.progress.LoadWatermark(ctx, req.TenantID(), req.Resource())
	if err != nil {
		return domain.PermissionDecision{}, 0, err
	}
	if !committed.covers(latest) {
		return domain.PermissionDecision{}, 0, ErrProjectionLag
	}
	// SpiceDB ZedTokens are opaque and not comparable. A second independent
	// caller-required token cannot be merged: use fully consistent instead.
	token := committed.ZedToken
	if fullyConsistent || (req.MinZedToken() != "" && req.MinZedToken() != token) {
		token = ""
	}
	// Internal barrier failures must always fail closed; never propagate a
	// caller's non-critical fail-open setting into a privileged check.
	verified, err := domain.NewCheckPermissionRequest(domain.CheckPermissionInput{
		TenantID: req.TenantID(), MinZedToken: token, Subject: req.Subject(),
		Permission: req.Permission(), Resource: req.Resource(), ModelRevision: req.ModelRevision(),
		FailMode: domain.FailModeDeny, Critical: true,
	})
	if err != nil {
		return domain.PermissionDecision{}, 0, err
	}
	decision, err := b.pdp.CheckPermission(ctx, verified)
	if err != nil {
		return domain.PermissionDecision{}, 0, err
	}
	// This check's graph snapshot is at least as fresh as the committed
	// watermark; authorized resource ownership is still checked by its owner.
	return decision, latest, nil
}
