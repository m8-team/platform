package resourcemanager

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/m8-team/platform/internal/resourcemanager/adapter/authz"
	"github.com/m8-team/platform/internal/resourcemanager/adapter/memory"
	"github.com/m8-team/platform/internal/resourcemanager/adapter/system"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/app/usecase"
	"go.uber.org/fx"
)

var ErrEmptyServiceName = errors.New("resource manager service name is empty")

const generatedPageTokenKeyLength = 32

type Config struct {
	ServiceName          string
	Debug                bool
	AllowUnauthenticated bool
	SoftDeleteRetention  time.Duration
	PageTokenKey         []byte
}

func Module(cfg Config) fx.Option {
	return fx.Module(
		"resourcemanager",
		fx.Supply(cfg.normalized()),
		fx.Provide(newOrganizationRepository),
		fx.Provide(newWorkspaceRepository),
		fx.Provide(newOrganizationLookup),
		fx.Provide(newOrganizationAuthorizer),
		fx.Provide(newClock),
		fx.Provide(newSystemIDGenerator),
		fx.Provide(newIDGenerator),
		fx.Provide(newWorkspaceIDGenerator),
		fx.Provide(newWorkspaceChildren),
		fx.Provide(newOrganizationServiceConfig),
		fx.Provide(newWorkspaceServiceConfig),
		fx.Provide(usecase.NewOrganizationService),
		fx.Provide(usecase.NewWorkspaceService),
		fx.Invoke(configureModule),
	)
}

func configureModule(cfg Config) error {
	return cfg.Validate()
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.ServiceName) == "" {
		return ErrEmptyServiceName
	}
	if c.SoftDeleteRetention <= 0 {
		return usecase.ErrInvalidSoftDeleteRetention
	}
	if len(c.PageTokenKey) > 0 && len(c.PageTokenKey) < generatedPageTokenKeyLength {
		return usecase.ErrInvalidPageTokenKey
	}

	return nil
}

func (c Config) normalized() Config {
	c.ServiceName = strings.TrimSpace(c.ServiceName)
	if c.SoftDeleteRetention == 0 {
		c.SoftDeleteRetention = usecase.DefaultSoftDeleteRetention
	}
	c.PageTokenKey = append([]byte(nil), c.PageTokenKey...)
	return c
}

func newOrganizationRepository() ports.OrganizationRepository {
	return memory.NewOrganizationRepository()
}

func newWorkspaceRepository() ports.WorkspaceRepository { return memory.NewWorkspaceRepository() }
func newOrganizationLookup(repository ports.OrganizationRepository) ports.OrganizationLookup {
	return repository
}

func newOrganizationAuthorizer(cfg Config) ports.Authorizer {
	if cfg.AllowUnauthenticated {
		return authz.AllowAll()
	}
	return authz.DenyAll()
}

func newClock() ports.Clock {
	return system.NewClock()
}

func newSystemIDGenerator() *system.IDGenerator {
	return system.NewIDGenerator()
}

func newIDGenerator(generator *system.IDGenerator) ports.IDGenerator {
	return generator
}

func newWorkspaceIDGenerator(generator *system.IDGenerator) ports.WorkspaceIDGenerator {
	return generator
}

func newWorkspaceChildren(repository ports.WorkspaceRepository) ports.WorkspaceChildren {
	return repository
}

func newWorkspaceServiceConfig(config usecase.OrganizationServiceConfig) usecase.WorkspaceServiceConfig {
	return usecase.WorkspaceServiceConfig{
		SoftDeleteRetention: config.SoftDeleteRetention,
		PageTokenKey:        append([]byte(nil), config.PageTokenKey...),
	}
}

func newOrganizationServiceConfig(cfg Config) (usecase.OrganizationServiceConfig, error) {
	key := append([]byte(nil), cfg.PageTokenKey...)
	if len(key) == 0 {
		key = make([]byte, generatedPageTokenKeyLength)
		if _, err := rand.Read(key); err != nil {
			return usecase.OrganizationServiceConfig{}, fmt.Errorf("generate page token key: %w", err)
		}
	}

	return usecase.OrganizationServiceConfig{
		SoftDeleteRetention: cfg.SoftDeleteRetention,
		PageTokenKey:        key,
	}, nil
}
