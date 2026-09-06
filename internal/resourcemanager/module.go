// Package resourcemanager is the composition boundary of M8 Resource Manager.
// A host may mount these use cases in a modular monolith or a dedicated service.
package resourcemanager

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/m8-team/platform/internal/platform/telemetry"

	organizationapp "github.com/m8-team/platform/internal/resourcemanager/app/organization"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	workspaceapp "github.com/m8-team/platform/internal/resourcemanager/app/workspace"
)

var ErrEmptyServiceName = errors.New("resource manager service name is empty")

type Config struct {
	ServiceName         string
	Debug               bool
	SoftDeleteRetention time.Duration
	PageTokenKey        []byte
}

// Dependencies are selected by the host, never implicitly by the module.
// Organizations and Workspaces must share the same hierarchy consistency boundary.
type Dependencies struct {
	Observer        *telemetry.Observer
	Organizations   ports.OrganizationRepository
	Workspaces      ports.WorkspaceRepository
	Authorizer      ports.Authorizer
	Clock           ports.Clock
	OrganizationIDs ports.IDGenerator
	WorkspaceIDs    ports.WorkspaceIDGenerator
}

type Services struct {
	Organizations *organizationapp.OrganizationService
	Workspaces    *workspaceapp.WorkspaceService
}

func New(cfg Config, deps Dependencies) (*Services, error) {
	cfg = cfg.normalized()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	key := cfg.PageTokenKey
	if len(key) == 0 {
		key = make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, fmt.Errorf("generate page token key: %w", err)
		}
	}
	organizations, err := organizationapp.NewOrganizationService(
		deps.Organizations, deps.Authorizer, deps.Clock, deps.OrganizationIDs, deps.Workspaces,
		organizationapp.OrganizationServiceConfig{SoftDeleteRetention: cfg.SoftDeleteRetention, PageTokenKey: key, Observer: deps.Observer})
	if err != nil {
		return nil, fmt.Errorf("compose organizations: %w", err)
	}
	workspaces, err := workspaceapp.NewWorkspaceService(
		deps.Workspaces, deps.Organizations, deps.Authorizer, deps.Clock, deps.WorkspaceIDs,
		workspaceapp.WorkspaceServiceConfig{SoftDeleteRetention: cfg.SoftDeleteRetention, PageTokenKey: key, Observer: deps.Observer})
	if err != nil {
		return nil, fmt.Errorf("compose workspaces: %w", err)
	}
	return &Services{Organizations: organizations, Workspaces: workspaces}, nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.ServiceName) == "" {
		return ErrEmptyServiceName
	}
	if c.SoftDeleteRetention <= 0 {
		return organizationapp.ErrInvalidSoftDeleteRetention
	}
	if len(c.PageTokenKey) > 0 && len(c.PageTokenKey) < 32 {
		return organizationapp.ErrInvalidPageTokenKey
	}
	return nil
}

func (c Config) normalized() Config {
	c.ServiceName = strings.TrimSpace(c.ServiceName)
	if c.SoftDeleteRetention == 0 {
		c.SoftDeleteRetention = organizationapp.DefaultSoftDeleteRetention
	}
	c.PageTokenKey = append([]byte(nil), c.PageTokenKey...)
	return c
}
