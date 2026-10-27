package v1engine

import (
	"crypto/sha256"
	"time"
)

type relationshipReceipt struct {
	hash [32]byte
	id string
}

func (s *Store) relationshipReplay(scope, key string, hash [32]byte) (OrganizationRelationship, bool, error) {
	receipt, ok := s.relationshipRequests[scope+":"+key]
	if !ok { return OrganizationRelationship{}, false, nil }
	if receipt.hash != hash { return OrganizationRelationship{}, true, ErrConflict }
	return s.relationships[receipt.id], true, nil
}
func (s *Store) validateRelationOrgs(r OrganizationRelationship) error {
	from, yesFrom := s.objects[key(Organization, r.FromOrganizationID)]
	to, yesTo := s.objects[key(Organization, r.ToOrganizationID)]
	if !yesFrom || !yesTo { return ErrNotFound }
	if from.TenantID != r.TenantID || to.TenantID != r.TenantID { return ErrInvalid }
	if from.State != "ACTIVE" || to.State != "ACTIVE" { return ErrConflict }
	return nil
}

func (s *Store) CreateRelationship(r OrganizationRelationship, idempotencyKey string) (OrganizationRelationship, error) {
	if idempotencyKey == "" || r.Validate() != nil { return OrganizationRelationship{}, ErrInvalid }
	hash, err := RelationshipHash("create", r)
	if err != nil { return OrganizationRelationship{}, err }
	s.mu.Lock()
	defer s.mu.Unlock()
	scope := r.TenantID
	if replay, found, err := s.relationshipReplay(scope, idempotencyKey, hash); found { return replay, err }
	if err := s.validateRelationOrgs(r); err != nil { return OrganizationRelationship{}, err }
	entries := make([]OrganizationRelationship, 0, len(s.relationships))
	for _, entry := range s.relationships { entries = append(entries, entry) }
	if err := ValidateRelationshipGraph(entries, r); err != nil { return OrganizationRelationship{}, err }
	id, err := newID()
	if err != nil { return OrganizationRelationship{}, err }
	now := time.Now().UTC()
	r.ID, r.State, r.Revision, r.CreatedAt, r.UpdatedAt = id, "ACTIVE", 1, now, now
	s.relationships[id] = r
	s.relationshipRequests[scope+":"+idempotencyKey] = relationshipReceipt{hash: hash, id: id}
	return r, nil
}

func (s *Store) GetRelationship(tenantID, id string) (OrganizationRelationship, error) {
	if tenantID == "" || id == "" { return OrganizationRelationship{}, ErrInvalid }
	s.mu.RLock()
	defer s.mu.RUnlock()
	rel, exists := s.relationships[id]
	if !exists || rel.TenantID != tenantID { return OrganizationRelationship{}, ErrNotFound }
	return rel, nil
}
func (s *Store) ListRelationships(tenantID, organizationID string, includeClosed bool) ([]OrganizationRelationship, error) {
	if tenantID == "" { return nil, ErrInvalid }
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.objects[key(Tenant,tenantID)]; !ok { return nil, ErrNotFound }
	out := make([]OrganizationRelationship,0,len(s.relationships))
	for _, rel := range s.relationships { out = append(out,rel) }
	return FilterRelationships(out, tenantID, organizationID, includeClosed), nil
}

// RelationshipUpdate can adjust stake, effective end and evidence. Identity,
// type, tenant and valid-from are immutable. ExpectedRevision is required.
type RelationshipUpdate struct {
	TenantID string
	ID string
	ExpectedRevision int64
	OwnershipBasisPoints int32
	ValidTo int64 // 0 = unchanged (not clear)
	EvidenceRef string
}

func updateHash(operation string, update RelationshipUpdate) [32]byte {
	sum, _ := RelationshipHash(operation, OrganizationRelationship{
		ID:update.ID,TenantID:update.TenantID, Revision:update.ExpectedRevision,
		OwnershipBasisPoints:update.OwnershipBasisPoints,ValidTo:update.ValidTo,EvidenceRef:update.EvidenceRef,
	})
	return sum
}
func (s *Store) UpdateRelationship(u RelationshipUpdate, idempotencyKey string) (OrganizationRelationship,error) {
	if idempotencyKey == "" || u.ID == "" || u.TenantID == "" || u.ExpectedRevision <= 0 { return OrganizationRelationship{},ErrInvalid }
	hash := updateHash("update",u)
	s.mu.Lock()
	defer s.mu.Unlock()
	if replay, found, err := s.relationshipReplay(u.TenantID,idempotencyKey,hash); found { return replay,err }
	r, exists := s.relationships[u.ID]
	if !exists || r.TenantID != u.TenantID { return OrganizationRelationship{}, ErrNotFound }
	if r.Revision != u.ExpectedRevision || r.State != "ACTIVE" { return OrganizationRelationship{},ErrConflict }
	r.OwnershipBasisPoints = u.OwnershipBasisPoints
	if u.ValidTo != 0 { r.ValidTo = u.ValidTo }
	r.EvidenceRef = u.EvidenceRef
	if err:=r.Validate(); err != nil {return OrganizationRelationship{},err}
	r.Revision++
	r.UpdatedAt = time.Now().UTC()
	s.relationships[r.ID] = r
	s.relationshipRequests[u.TenantID+":"+idempotencyKey] = relationshipReceipt{hash:hash,id:r.ID}
	return r,nil
}

func (s *Store) CloseRelationship(tenantID, id string, expectedRevision, closedAt int64, idempotencyKey string) (OrganizationRelationship,error) {
	if tenantID == "" || id == "" || expectedRevision <= 0 || closedAt <= 0 || idempotencyKey == "" { return OrganizationRelationship{},ErrInvalid }
	cmd := RelationshipUpdate{TenantID:tenantID,ID:id,ExpectedRevision:expectedRevision,ValidTo:closedAt}
	hash:=updateHash("close",cmd)
	s.mu.Lock()
	defer s.mu.Unlock()
	if replay, found, err := s.relationshipReplay(tenantID,idempotencyKey,hash); found {return replay,err}
	r,ok:=s.relationships[id]
	if !ok || r.TenantID != tenantID {return OrganizationRelationship{},ErrNotFound}
	if r.Revision != expectedRevision || r.State != "ACTIVE" {return OrganizationRelationship{},ErrConflict}
	if closedAt <= r.ValidFrom || (r.ValidTo != 0 && closedAt > r.ValidTo) ||
		closedAt > time.Now().Unix()+60 {return OrganizationRelationship{},ErrInvalid}
	r.ValidTo,r.State,r.Revision,r.UpdatedAt=closedAt,"CLOSED",r.Revision+1,time.Now().UTC()
	s.relationships[id]=r
	s.relationshipRequests[tenantID+":"+idempotencyKey]=relationshipReceipt{hash:hash,id:id}
	return r,nil
}

// Prevent accidental import removal by the Go formatter; SHA is part of
// the idempotency receipt, not a security token.
var _ [32]byte = sha256.Sum256(nil)
