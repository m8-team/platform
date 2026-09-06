package workspaceapp

import (
	"errors"
	"fmt"
	"time"

	"github.com/m8-team/platform/internal/platform/telemetry"

	platformfilter "github.com/m8-team/platform/internal/platform/filter"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
)

var (
	ErrInvalidSoftDeleteRetention   = errors.New("soft-delete retention must be positive")
	ErrInvalidPageTokenKey          = errors.New("page token key must contain at least 32 bytes")
	ErrWorkspaceRepositoryRequired  = errors.New("workspace repository is required")
	ErrOrganizationLookupRequired   = errors.New("organization lookup is required")
	ErrWorkspaceAuthorizerRequired  = errors.New("workspace authorizer is required")
	ErrWorkspaceClockRequired       = errors.New("workspace clock is required")
	ErrWorkspaceIDGeneratorRequired = errors.New("workspace id generator is required")
	ErrGeneratedWorkspaceID         = errors.New("generated workspace id is invalid")
	ErrWorkspaceAuthorizationScope  = errors.New("workspace authorization scope key is required")
)

type WorkspaceServiceConfig struct {
	Observer            *telemetry.Observer
	SoftDeleteRetention time.Duration
	PageTokenKey        []byte
}

// WorkspaceService coordinates workspace use cases. The domain aggregate and
// persistence port are workspace-owned; the parent Organization is accessed
// only through the explicit lookup port.
type WorkspaceService struct {
	observer      *telemetry.Observer
	filterParser  *platformfilter.CELParser
	repository    ports.WorkspaceRepository
	organizations ports.OrganizationLookup
	authorizer    ports.Authorizer
	clock         ports.Clock
	idGenerator   ports.WorkspaceIDGenerator
	retention     time.Duration
	pageTokens    workspacePageTokenCodec
}

func NewWorkspaceService(
	repository ports.WorkspaceRepository,
	organizations ports.OrganizationLookup,
	authorizer ports.Authorizer,
	clock ports.Clock,
	idGenerator ports.WorkspaceIDGenerator,
	config WorkspaceServiceConfig,
) (*WorkspaceService, error) {
	if repository == nil {
		return nil, ErrWorkspaceRepositoryRequired
	}
	if organizations == nil {
		return nil, ErrOrganizationLookupRequired
	}
	if authorizer == nil {
		return nil, ErrWorkspaceAuthorizerRequired
	}
	if clock == nil {
		return nil, ErrWorkspaceClockRequired
	}
	if idGenerator == nil {
		return nil, ErrWorkspaceIDGeneratorRequired
	}
	if config.SoftDeleteRetention <= 0 {
		return nil, ErrInvalidSoftDeleteRetention
	}
	if len(config.PageTokenKey) < minimumPageTokenKeyLength {
		return nil, ErrInvalidPageTokenKey
	}
	parser, err := newFilterParser()
	if err != nil {
		return nil, fmt.Errorf("initialize workspace filter parser: %w", err)
	}

	return &WorkspaceService{
		observer:      config.Observer,
		filterParser:  parser,
		repository:    repository,
		organizations: organizations,
		authorizer:    authorizer,
		clock:         clock,
		idGenerator:   idGenerator,
		retention:     config.SoftDeleteRetention,
		pageTokens:    newWorkspacePageTokenCodec(config.PageTokenKey),
	}, nil
}
