// Package catalog implements universal module authorization catalog commands and queries.
package catalog

import (
	"context"
	"errors"
	"strings"

	"github.com/m8-team/platform/internal/access/domain/module"
	"github.com/m8-team/platform/internal/access/domain/permission"
	"github.com/m8-team/platform/internal/access/domain/resource"
	"github.com/m8-team/platform/internal/access/domain/role"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

var (
	ErrCatalogUnauthenticated = errors.New("authenticated catalog principal required")
	ErrCatalogForbidden       = errors.New("catalog namespace access denied")
	ErrCatalogArgument        = errors.New("invalid catalog API argument")
	ErrCatalogNotFound        = errors.New("catalog release not found")
	ErrCatalogConflict        = errors.New("catalog release or idempotency conflict")
	ErrCatalogStoreInvalid    = errors.New("catalog repository returned an invalid or cross-namespace record")
)

// PublishedRecord is a committed immutable manifest and a catalog revision.
// A catalog revision is NOT a SpiceDB ZedToken.
type PublishedRecord struct {
	Manifest module.Manifest
	Revision int64
}

func (r PublishedRecord) Validate() error {
	if err := r.Manifest.Validate(); err != nil || r.Revision <= 0 {
		return ErrCatalogStoreInvalid
	}
	return nil
}

// PublishInput is a definition-only command. Publisher identity MUST be
// derived from authenticated service context, never a request field.
type PublishInput struct {
	ModuleID    string
	Version     string
	Resources   []resource.TypeDefinition
	Permissions []permission.Definition
	Roles       []role.Template
}

// IdentityResolver authenticates the caller via verified workload or user identity.
type IdentityResolver interface {
	AuthenticatedCatalogPrincipal(context.Context) (shared.Subject, error)
}

// CatalogAuthorizer supplies independent, trusted publisher/reader policy.
// No module can grant itself publishing rights by registering a manifest.
type CatalogAuthorizer interface {
	CanPublishModule(context.Context, shared.Subject, string) (bool, error)
	CanReadModule(context.Context, shared.Subject, string) (bool, error)
}

// Registry persists catalog changes with CAS, durable outbox and replay protection.
// Publish MUST return the original committed record on an identical idempotent
// retry; reuse of a key with different content MUST return ErrCatalogConflict.
// List cursors MUST be opaque, authenticated and bound to principal and module.
// Production requires an authoritative YDB adapter; no in-memory fallback.
type Registry interface {
	Publish(ctx context.Context, manifest module.Manifest, expectedRevision int64, idempotencyKey string) (PublishedRecord, error)
	Get(ctx context.Context, moduleID, version string) (PublishedRecord, error)
	List(ctx context.Context, principal shared.Subject, moduleID string, pageSize int, pageToken string) ([]PublishedRecord, string, error)
}

type API struct {
	identity   IdentityResolver
	authorizer CatalogAuthorizer
	registry   Registry
}

func NewAPI(identity IdentityResolver, authorizer CatalogAuthorizer, registry Registry) (*API, error) {
	if identity == nil || authorizer == nil || registry == nil {
		return nil, ErrCatalogUnavailable
	}
	return &API{identity: identity, authorizer: authorizer, registry: registry}, nil
}

func (a *API) principal(ctx context.Context, moduleID string, publish bool) (shared.Subject, error) {
	if a == nil || a.identity == nil || a.authorizer == nil || a.registry == nil {
		return shared.Subject{}, ErrCatalogUnavailable
	}
	if err := ctx.Err(); err != nil {
		return shared.Subject{}, err
	}
	if !shared.ValidModuleName(moduleID) {
		return shared.Subject{}, ErrCatalogArgument
	}
	subject, err := a.identity.AuthenticatedCatalogPrincipal(ctx)
	if err != nil || subject.Validate() != nil {
		return shared.Subject{}, ErrCatalogUnauthenticated
	}
	if publish {
		if subject.Type() != "service_account" {
			return shared.Subject{}, ErrCatalogForbidden
		}
		allowed, err := a.authorizer.CanPublishModule(ctx, subject, moduleID)
		if err != nil || !allowed {
			return shared.Subject{}, ErrCatalogForbidden
		}
	} else {
		allowed, err := a.authorizer.CanReadModule(ctx, subject, moduleID)
		if err != nil || !allowed {
			return shared.Subject{}, ErrCatalogForbidden
		}
	}
	return subject, nil
}

func (a *API) Publish(ctx context.Context, input PublishInput, expectedRevision int64, idempotencyKey string) (PublishedRecord, error) {
	if expectedRevision < 0 || strings.TrimSpace(idempotencyKey) != idempotencyKey ||
		idempotencyKey == "" || len(idempotencyKey) > 256 {
		return PublishedRecord{}, ErrCatalogArgument
	}
	subject, err := a.principal(ctx, input.ModuleID, true)
	if err != nil {
		return PublishedRecord{}, err
	}
	manifest, err := module.NewManifest(module.ManifestInput{
		ModuleID: input.ModuleID, Version: input.Version, Publisher: subject,
		Resources: input.Resources, Permissions: input.Permissions, Roles: input.Roles,
	})
	if err != nil {
		return PublishedRecord{}, ErrCatalogArgument
	}
	record, err := a.registry.Publish(ctx, manifest, expectedRevision, idempotencyKey)
	if err != nil {
		return PublishedRecord{}, err
	}
	if record.Validate() != nil || record.Manifest.ModuleID() != manifest.ModuleID() ||
		record.Manifest.Version() != manifest.Version() || record.Manifest.Publisher() != subject {
		return PublishedRecord{}, ErrCatalogStoreInvalid
	}
	return record, nil
}

func (a *API) Get(ctx context.Context, moduleID, version string) (PublishedRecord, error) {
	if !shared.ValidReleaseVersion(version) {
		return PublishedRecord{}, ErrCatalogArgument
	}
	if _, err := a.principal(ctx, moduleID, false); err != nil {
		return PublishedRecord{}, err
	}
	record, err := a.registry.Get(ctx, moduleID, version)
	if err != nil {
		return PublishedRecord{}, err
	}
	if record.Validate() != nil || record.Manifest.ModuleID() != moduleID || record.Manifest.Version() != version {
		return PublishedRecord{}, ErrCatalogStoreInvalid
	}
	return record, nil
}

func (a *API) List(ctx context.Context, moduleID string, pageSize int, pageToken string) ([]PublishedRecord, string, error) {
	if pageSize < 0 || pageSize > 100 || len(pageToken) > 2048 || strings.TrimSpace(pageToken) != pageToken {
		return nil, "", ErrCatalogArgument
	}
	principal, err := a.principal(ctx, moduleID, false)
	if err != nil {
		return nil, "", err
	}
	if pageSize == 0 {
		pageSize = 20
	}
	records, next, err := a.registry.List(ctx, principal, moduleID, pageSize, pageToken)
	if err != nil {
		return nil, "", err
	}
	if len(records) > pageSize || len(next) > 2048 {
		return nil, "", ErrCatalogStoreInvalid
	}
	for _, record := range records {
		if record.Validate() != nil || record.Manifest.ModuleID() != moduleID {
			return nil, "", ErrCatalogStoreInvalid
		}
	}
	return records, next, nil
}
