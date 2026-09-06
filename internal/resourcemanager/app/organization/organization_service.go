package organizationapp

import (
	"errors"
	"fmt"
	"time"

	"github.com/m8-team/platform/internal/platform/telemetry"

	platformfilter "github.com/m8-team/platform/internal/platform/filter"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
)

const DefaultSoftDeleteRetention = 30 * 24 * time.Hour

var (
	ErrOrganizationRepositoryRequired  = errors.New("organization repository is required")
	ErrOrganizationAuthorizerRequired  = errors.New("organization authorizer is required")
	ErrOrganizationClockRequired       = errors.New("organization clock is required")
	ErrOrganizationIDGeneratorRequired = errors.New("organization id generator is required")
	ErrWorkspaceChildrenRequired       = errors.New("workspace children reader is required")
	ErrInvalidSoftDeleteRetention      = errors.New("soft-delete retention must be positive")
	ErrInvalidPageTokenKey             = errors.New("page token key must contain at least 32 bytes")
	ErrOrganizationHasWorkspaces       = errors.New("organization has non-deleted workspaces")
	ErrGeneratedOrganizationID         = errors.New("generated organization id is invalid")
	ErrAuthorizationScopeRequired      = errors.New("authorization scope key is required")
)

type OrganizationServiceConfig struct {
	Observer            *telemetry.Observer
	SoftDeleteRetention time.Duration
	PageTokenKey        []byte
}

// OrganizationService coordinates organization use cases. Persistence,
// authorization, time, identity generation, and hierarchy checks are explicit
// ports so business behavior remains independent from gRPC and storage.
type OrganizationService struct {
	observer          *telemetry.Observer
	filterParser      *platformfilter.CELParser
	repository        ports.OrganizationRepository
	authorizer        ports.Authorizer
	clock             ports.Clock
	idGenerator       ports.IDGenerator
	workspaceChildren ports.WorkspaceChildren
	retention         time.Duration
	pageTokens        pageTokenCodec
}

func NewOrganizationService(
	repository ports.OrganizationRepository,
	authorizer ports.Authorizer,
	clock ports.Clock,
	idGenerator ports.IDGenerator,
	workspaceChildren ports.WorkspaceChildren,
	config OrganizationServiceConfig,
) (*OrganizationService, error) {
	if repository == nil {
		return nil, ErrOrganizationRepositoryRequired
	}
	if authorizer == nil {
		return nil, ErrOrganizationAuthorizerRequired
	}
	if clock == nil {
		return nil, ErrOrganizationClockRequired
	}
	if idGenerator == nil {
		return nil, ErrOrganizationIDGeneratorRequired
	}
	if workspaceChildren == nil {
		return nil, ErrWorkspaceChildrenRequired
	}
	if config.SoftDeleteRetention <= 0 {
		return nil, ErrInvalidSoftDeleteRetention
	}
	if len(config.PageTokenKey) < minimumPageTokenKeyLength {
		return nil, ErrInvalidPageTokenKey
	}
	parser, err := newFilterParser()
	if err != nil {
		return nil, fmt.Errorf("initialize organization filter parser: %w", err)
	}

	return &OrganizationService{
		observer:          config.Observer,
		filterParser:      parser,
		repository:        repository,
		authorizer:        authorizer,
		clock:             clock,
		idGenerator:       idGenerator,
		workspaceChildren: workspaceChildren,
		retention:         config.SoftDeleteRetention,
		pageTokens:        newPageTokenCodec(config.PageTokenKey),
	}, nil
}
