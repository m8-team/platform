package ydbcatalog

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/m8-team/platform/internal/access/app/catalog"
	"github.com/m8-team/platform/internal/access/app/consistency"
	"github.com/m8-team/platform/internal/access/domain"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

var (
	ErrTenantModuleUnavailable = errors.New("tenant module is not installed or not enabled")
	ErrInvalidTenantModule = errors.New("invalid tenant module installation state")
)

// TenantModuleInstallation is an authoritative tenant-specific module pin.
// Disabling an installation must be reflected here by the module lifecycle
// owner, independently of catalog publication or user-supplied claims.
type TenantModuleInstallation struct {
	ModuleID string
	Version string
	Revision uint64
}
func(i TenantModuleInstallation) Validate() error {
	if !shared.ValidModuleName(i.ModuleID) || !shared.ValidReleaseVersion(i.Version) || i.Revision == 0 {
		return ErrInvalidTenantModule
	}
	return nil
}

// TenantModuleResolver must query a trusted module installation authority.
// This interface is NOT filled from gateway headers or JWT custom claims.
type TenantModuleResolver interface {
	ResolveInstalledModule(context.Context, domain.TenantID, string) (TenantModuleInstallation, error)
}

type ReleaseReader interface {
	Get(context.Context, string, string) (catalog.PublishedRecord, error)
}

// UniversalPermissionReader verifies exact action/type mapping in a pinned
// immutable release, and never treats a globally published module as installed
// for every tenant.
type UniversalPermissionReader struct {
	installations TenantModuleResolver
	releases ReleaseReader
}
var _ consistency.UniversalCatalogReader = (*UniversalPermissionReader)(nil)

func NewUniversalPermissionReader(installations TenantModuleResolver, releases ReleaseReader) (*UniversalPermissionReader, error) {
	if installations == nil || releases == nil {
		return nil, ErrInvalidStore
	}
	return &UniversalPermissionReader{installations: installations, releases: releases}, nil
}
func (r *UniversalPermissionReader) HasCatalogPermission(ctx context.Context, tenant domain.TenantID, resourceType, action string) (bool, error) {
	if r == nil || r.installations == nil || r.releases == nil || tenant.Validate() != nil ||
		!shared.ValidQualifiedName(resourceType) || !shared.ValidQualifiedName(action) {
		return false, ErrInvalidTenantModule
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	moduleID, _, qualified := strings.Cut(resourceType, ".")
	if !qualified || !shared.ValidModuleName(moduleID) || !strings.HasPrefix(action, moduleID+".") {
		return false, nil
	}
	install, err := r.installations.ResolveInstalledModule(ctx, tenant, moduleID)
	if err != nil {
		return false, err
	}
	if err := install.Validate(); err != nil || install.ModuleID != moduleID {
		return false, ErrInvalidTenantModule
	}
	release, err := r.releases.Get(ctx, moduleID, install.Version)
	if err != nil {
		return false, err
	}
	if release.Validate() != nil || release.Manifest.ModuleID() != moduleID ||
		release.Manifest.Version() != install.Version {
		return false, catalog.ErrCatalogStoreInvalid
	}
	for _, d := range release.Manifest.Permissions() {
		if d.Permission().Name() == action && d.ResourceType() == resourceType &&
			d.GraphPermission() == "use" {
			return true, nil
		}
	}
	return false, nil
}

const installedModule = "DECLARE $tenant AS Utf8;DECLARE $module AS Utf8;\nSELECT release_version,enabled,revision FROM `m8-access/catalog_tenant_modules` WHERE tenant_id=$tenant AND module_id=$module;"

// ResolveInstalledModule is read-only: writes belong to a trusted module
// installation command with lifecycle authorization and revision control.
// A missing, disabled or malformed record fails closed.
func (s *Store) ResolveInstalledModule(ctx context.Context, tenant domain.TenantID, moduleID string) (TenantModuleInstallation, error) {
	if s == nil || s.db == nil || tenant.Validate() != nil || !shared.ValidModuleName(moduleID) {
		return TenantModuleInstallation{}, ErrInvalidTenantModule
	}
	var version string
	var enabled bool
	var revision uint64
	err := s.db.QueryRowContext(ctx, s.q(installedModule),
		sql.Named("tenant", string(tenant)), sql.Named("module", moduleID)).
		Scan(&version, &enabled, &revision)
	if errors.Is(err, sql.ErrNoRows) || err == nil && !enabled {
		return TenantModuleInstallation{}, ErrTenantModuleUnavailable
	}
	if err != nil {
		return TenantModuleInstallation{}, err
	}
	pin := TenantModuleInstallation{ModuleID: moduleID, Version: version, Revision: revision}
	if err := pin.Validate(); err != nil {
		return TenantModuleInstallation{}, err
	}
	return pin, nil
}
