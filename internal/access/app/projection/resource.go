package projection

import (
	"fmt"
	"strconv"

	"github.com/m8-team/platform/internal/access/domain"
)

// RM projections contain only a state/revision marker. They deliberately do
// not invent organization inheritance, ownership edges or effective permissions.
type ResourceState struct {
	TenantID  domain.TenantID
	Resource  domain.Resource
	Revision  uint64
	Lifecycle string
}

func IsResourceKind(kind Kind) bool {
	switch kind {
	case KindRMOrganization, KindRMWorkspace, KindRMEnvironment, KindRMWorkloadService:
		return true
	}
	return false
}
func ResourceTypeForKind(kind Kind) (string, error) {
	switch kind {
	case KindRMOrganization:
		return "organization", nil
	case KindRMWorkspace:
		return "workspace", nil
	case KindRMEnvironment:
		return "environment", nil
	case KindRMWorkloadService:
		return "workload_service", nil
	}
	return "", ErrInvalidEvent
}
func KindForResourceType(kind string) (Kind, error) {
	switch kind {
	case "organization":
		return KindRMOrganization, nil
	case "workspace":
		return KindRMWorkspace, nil
	case "environment":
		return KindRMEnvironment, nil
	case "workload_service":
		return KindRMWorkloadService, nil
	}
	return "", ErrInvalidEvent
}

func (s ResourceState) Validate(key Key, revision int64) error {
	kind, err := ResourceTypeForKind(key.Kind)
	if err != nil || s.TenantID != key.TenantID || s.Resource.Type() != kind || s.Resource.ID() != key.AggregateID ||
		s.Revision == 0 || s.Revision != uint64(revision) || s.Resource.Validate() != nil {
		return ErrInvalidEvent
	}
	switch s.Lifecycle {
	case "ACTIVE", "SUSPENDED", "ARCHIVED", "DELETED":
		return nil
	default:
		return ErrInvalidEvent
	}
}

// ResourceMarkerID separates equal UUIDs belonging to different RM kinds.
func ResourceMarkerID(kind Kind, id string) (string, error) {
	resourceType, err := ResourceTypeForKind(kind)
	if err != nil || !canonicalAggregateID.MatchString(id) {
		return "", ErrInvalidEvent
	}
	return resourceType + ":" + id, nil
}

func resourceRelationships(key Key, state *ResourceState) ([]domain.Relationship, error) {
	if state == nil {
		return nil, ErrInvalidEvent
	}
	if err := state.Validate(key, int64(state.Revision)); err != nil {
		return nil, err
	}
	markerID, err := ResourceMarkerID(key.Kind, key.AggregateID)
	if err != nil {
		return nil, err
	}
	marker, err := domain.NewResource("rm_projection", markerID)
	if err != nil {
		return nil, err
	}
	epoch, err := domain.NewSubject("projection_epoch", strconv.FormatUint(state.Revision, 10))
	if err != nil {
		return nil, err
	}
	relationship, err := domain.NewRelationship(domain.RelationshipInput{
		TenantID: key.TenantID, Resource: marker, Relation: "observed", Subject: epoch,
	})
	if err != nil {
		return nil, fmt.Errorf("resource marker: %w", err)
	}
	return []domain.Relationship{relationship}, nil
}
