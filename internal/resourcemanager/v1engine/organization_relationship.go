package v1engine

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"sort"
	"time"
)

var ErrRelationshipCycle = errors.New("corporate parent relationship cycle")
var ErrRelationshipOverlap = errors.New("overlapping corporate relationship")

type RelationshipType string

const (
	RelationshipParentSubsidiary RelationshipType = "PARENT_SUBSIDIARY"
	RelationshipAffiliate        RelationshipType = "AFFILIATE"
	RelationshipPartner          RelationshipType = "PARTNER"
	RelationshipServiceProvider  RelationshipType = "SERVICE_PROVIDER"
)

type OrganizationRelationship struct {
	ID                   string           `json:"id"`
	TenantID             string           `json:"tenantId"`
	FromOrganizationID   string           `json:"fromOrganizationId"`
	ToOrganizationID     string           `json:"toOrganizationId"`
	Type                 RelationshipType `json:"type"`
	OwnershipBasisPoints int32            `json:"ownershipBasisPoints"`
	ValidFrom            int64            `json:"validFromUnix"`
	ValidTo              int64            `json:"validToUnix"` // zero = unbounded, exclusive
	EvidenceRef          string           `json:"evidenceRef"`
	State                string           `json:"state"`
	Revision             int64            `json:"revision"`
	CreatedAt            time.Time        `json:"createdAt"`
	UpdatedAt            time.Time        `json:"updatedAt"`
}

func (r OrganizationRelationship) Validate() error {
	if r.TenantID == "" || r.FromOrganizationID == "" || r.ToOrganizationID == "" ||
		r.FromOrganizationID == r.ToOrganizationID || r.ValidFrom <= 0 ||
		(r.ValidTo != 0 && r.ValidTo <= r.ValidFrom) || len(r.EvidenceRef) > 1024 {
		return ErrInvalid
	}
	switch r.Type {
	case RelationshipParentSubsidiary:
		if r.OwnershipBasisPoints < 0 || r.OwnershipBasisPoints > 10000 {
			return ErrInvalid
		}
	case RelationshipAffiliate, RelationshipPartner, RelationshipServiceProvider:
		if r.OwnershipBasisPoints != 0 {
			return ErrInvalid
		}
	default:
		return ErrInvalid
	}
	return nil
}

func overlap(a, b OrganizationRelationship) bool {
	return (a.ValidTo == 0 || b.ValidFrom < a.ValidTo) &&
		(b.ValidTo == 0 || a.ValidFrom < b.ValidTo)
}

func rangeEnd(a, b int64) int64 {
	if a == 0 {
		return b
	}
	if b == 0 {
		return a
	}
	if a < b {
		return a
	}
	return b
}

// ValidateRelationshipGraph ensures no overlapping duplicate tuples and
// no temporal cycle of PARENT_SUBSIDIARY relationships. All historic closed
// records participate in temporal checks using their bounded validity.
func ValidateRelationshipGraph(existing []OrganizationRelationship, candidate OrganizationRelationship) error {
	if err := candidate.Validate(); err != nil {
		return err
	}
	for _, current := range existing {
		if current.ID == candidate.ID && candidate.ID != "" {
			continue
		}
		if current.TenantID != candidate.TenantID || !overlap(current, candidate) {
			continue
		}
		if current.FromOrganizationID == candidate.FromOrganizationID &&
			current.ToOrganizationID == candidate.ToOrganizationID &&
			current.Type == candidate.Type {
			return ErrRelationshipOverlap
		}
	}
	if candidate.Type != RelationshipParentSubsidiary {
		return nil
	}
	adjacency := make(map[string][]OrganizationRelationship)
	for _, existingEdge := range existing {
		if existingEdge.ID == candidate.ID && candidate.ID != "" {
			continue
		}
		if existingEdge.TenantID == candidate.TenantID && existingEdge.Type == RelationshipParentSubsidiary {
			adjacency[existingEdge.FromOrganizationID] = append(adjacency[existingEdge.FromOrganizationID], existingEdge)
		}
	}
	type next struct {
		id         string
		start, end int64
	}
	stack := []next{{candidate.ToOrganizationID, candidate.ValidFrom, candidate.ValidTo}}
	seen := make(map[next]bool)
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if n.id == candidate.FromOrganizationID {
			return ErrRelationshipCycle
		}
		if seen[n] {
			continue
		}
		seen[n] = true
		for _, edge := range adjacency[n.id] {
			start := n.start
			if edge.ValidFrom > start {
				start = edge.ValidFrom
			}
			end := rangeEnd(n.end, edge.ValidTo)
			if end != 0 && start >= end {
				continue
			}
			stack = append(stack, next{edge.ToOrganizationID, start, end})
		}
	}
	return nil
}

func FilterRelationships(items []OrganizationRelationship, tenantID, organizationID string, includeClosed bool) []OrganizationRelationship {
	result := make([]OrganizationRelationship, 0)
	for _, rel := range items {
		if rel.TenantID != tenantID || (!includeClosed && rel.State != "ACTIVE") {
			continue
		}
		if organizationID != "" && rel.FromOrganizationID != organizationID && rel.ToOrganizationID != organizationID {
			continue
		}
		result = append(result, rel)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].ValidFrom != result[j].ValidFrom {
			return result[i].ValidFrom < result[j].ValidFrom
		}
		return result[i].ID < result[j].ID
	})
	return result
}

func RelationshipHash(operation string, relationship OrganizationRelationship) ([32]byte, error) {
	payload, err := json.Marshal(struct {
		Operation    string
		Relationship OrganizationRelationship
	}{operation, OrganizationRelationship{
		ID: relationship.ID, TenantID: relationship.TenantID,
		FromOrganizationID: relationship.FromOrganizationID,
		ToOrganizationID:   relationship.ToOrganizationID, Type: relationship.Type,
		OwnershipBasisPoints: relationship.OwnershipBasisPoints,
		ValidFrom:            relationship.ValidFrom, ValidTo: relationship.ValidTo,
		EvidenceRef: relationship.EvidenceRef,
		Revision:    relationship.Revision,
	}})
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(payload), nil
}
