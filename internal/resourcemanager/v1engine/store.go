// Package v1engine implements the canonical v1 ownership rules with a
// development-only memory store. Production requires durable YDB adapters.
package v1engine

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalid  = errors.New("invalid resource")
	ErrNotFound = errors.New("resource not found")
	ErrConflict = errors.New("resource conflict")
)

type Kind string

const (
	Tenant            Kind = "tenant"
	OrganizationGroup Kind = "group"
	Organization      Kind = "organization"
	Workspace         Kind = "workspace"
	Project           Kind = "project"
	Environment       Kind = "environment"
	WorkloadService   Kind = "workload_service"
	Descriptor        Kind = "descriptor"
)

type Owner struct {
	Kind Kind   `json:"kind"`
	ID   string `json:"id"`
}
type Record struct {
	Kind           Kind              `json:"kind"`
	ID             string            `json:"id"`
	TenantID       string            `json:"tenantId,omitempty"`
	Owner          Owner             `json:"owner"`
	OrganizationID string            `json:"owningOrganizationId,omitempty"`
	Name           string            `json:"name"`
	Description    string            `json:"description,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	State          string            `json:"state"`
	Revision       int64             `json:"revision"`
	EnvironmentKey string            `json:"environmentKey,omitempty"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
	DeletedAt      *time.Time        `json:"deletedAt,omitempty"`
}
type Binding struct {
	ID            string `json:"id"`
	Resource      Owner  `json:"resource"`
	EnvironmentID string `json:"environmentId"`
	Version       string `json:"version"`
	ConfigRef     string `json:"configurationRef,omitempty"`
	Revision      int64  `json:"revision"`
}

type OrganizationGroupLink struct {
	GroupID        string
	OrganizationID string
	CreatedAt      time.Time
}

type Store struct {
	mu                   sync.RWMutex
	objects              map[string]Record
	requestIDs           map[string]string
	links                map[string]time.Time
	bindings             map[string]Binding
	relationships        map[string]OrganizationRelationship
	relationshipHistory  map[string][]OrganizationRelationship
	relationshipRequests map[string]relationshipReceipt
}

func NewStore() *Store {
	return &Store{objects: make(map[string]Record), requestIDs: make(map[string]string), links: make(map[string]time.Time), bindings: make(map[string]Binding), relationships: make(map[string]OrganizationRelationship), relationshipHistory: make(map[string][]OrganizationRelationship), relationshipRequests: make(map[string]relationshipReceipt)}
}
func newID() (string, error) {
	var b [16]byte
	if _, e := rand.Read(b[:]); e != nil {
		return "", e
	}
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	v := hex.EncodeToString(b[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s", v[:8], v[8:12], v[12:16], v[16:20], v[20:]), nil
}
func key(k Kind, v string) string { return string(k) + ":" + v }
func AllowedParent(child, parent Kind) bool {
	switch child {
	case Tenant:
		return parent == ""
	case OrganizationGroup, Organization:
		return parent == Tenant
	case Workspace, Environment:
		return parent == Organization
	case Project:
		return parent == Organization || parent == Workspace
	case WorkloadService:
		return parent == Workspace || parent == Project
	case Descriptor:
		return parent == Organization || parent == Workspace || parent == Project
	}
	return false
}
func (s *Store) Create(r Record, idem string) (Record, error) {
	if strings.TrimSpace(idem) == "" || strings.TrimSpace(r.Name) == "" {
		return Record{}, fmt.Errorf("%w: idempotencyKey and name required", ErrInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if rid, exists := s.requestIDs["create:"+idem]; exists {
		return s.objects[rid], nil
	}
	if !AllowedParent(r.Kind, r.Owner.Kind) {
		return Record{}, ErrInvalid
	}
	if r.Kind != Tenant {
		if r.Owner.ID == "" {
			return Record{}, ErrInvalid
		}
		parent, exists := s.objects[key(r.Owner.Kind, r.Owner.ID)]
		if !exists || parent.State != "ACTIVE" {
			return Record{}, ErrNotFound
		}
		switch r.Kind {
		case OrganizationGroup, Organization:
			r.TenantID = parent.ID
		default:
			r.TenantID = parent.TenantID
			if parent.Kind == Organization {
				r.OrganizationID = parent.ID
			} else {
				r.OrganizationID = parent.OrganizationID
			}
		}
	} else if r.Owner.ID != "" || r.Owner.Kind != "" {
		return Record{}, ErrInvalid
	}
	if r.Kind == Environment {
		if r.EnvironmentKey == "" {
			return Record{}, fmt.Errorf("%w: environmentKey required", ErrInvalid)
		}
		for _, obj := range s.objects {
			if obj.Kind == Environment && obj.OrganizationID == r.OrganizationID && obj.State != "DELETED" && obj.EnvironmentKey == r.EnvironmentKey {
				return Record{}, ErrConflict
			}
		}
	}
	id, err := newID()
	if err != nil {
		return Record{}, err
	}
	r.ID = id
	r.Revision = 1
	r.State = "ACTIVE"
	r.CreatedAt = time.Now().UTC()
	r.UpdatedAt = r.CreatedAt
	s.objects[key(r.Kind, id)] = r
	s.requestIDs["create:"+idem] = key(r.Kind, id)
	return r, nil
}
func (s *Store) Get(kind Kind, id string) (Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.objects[key(kind, id)]
	if !ok {
		return Record{}, ErrNotFound
	}
	return r, nil
}
func (s *Store) List(kind Kind, orgID string) ([]Record, error) {
	if orgID == "" {
		return nil, fmt.Errorf("%w: organization filter required", ErrInvalid)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Record{}
	for _, r := range s.objects {
		if (kind == "" || r.Kind == kind) && (r.OrganizationID == orgID || (r.Kind == Organization && r.ID == orgID)) {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// Tenant-scoped discovery never scans or returns records from other Tenants.
func (s *Store) ListTenantOrganizations(tenantID string) ([]Record, error) {
	return s.listTenantKind(tenantID, Organization)
}
func (s *Store) ListTenantGroups(tenantID string) ([]Record, error) {
	return s.listTenantKind(tenantID, OrganizationGroup)
}
func (s *Store) listTenantKind(tenantID string, kind Kind) ([]Record, error) {
	if tenantID == "" {
		return nil, ErrInvalid
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.objects[key(Tenant, tenantID)]; !ok {
		return nil, ErrNotFound
	}
	out := make([]Record, 0)
	for _, record := range s.objects {
		if record.Kind == kind && record.TenantID == tenantID && record.State != "DELETED" {
			out = append(out, record)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func (s *Store) Update(kind Kind, id string, rev int64, name, description string, labels map[string]string) (Record, error) {
	if rev <= 0 || strings.TrimSpace(name) == "" {
		return Record{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.objects[key(kind, id)]
	if !ok {
		return Record{}, ErrNotFound
	}
	if r.Revision != rev || r.State != "ACTIVE" {
		return Record{}, ErrConflict
	}
	r.Name = name
	r.Description = description
	r.Labels = maps.Clone(labels)
	r.Revision++
	r.UpdatedAt = time.Now().UTC()
	s.objects[key(kind, id)] = r
	return r, nil
}
func (s *Store) ChangeState(kind Kind, id string, rev int64, state string) (Record, error) {
	if rev <= 0 || (state != "SUSPENDED" && state != "ACTIVE" && state != "ARCHIVED" && state != "DELETED") {
		return Record{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.objects[key(kind, id)]
	if !ok {
		return Record{}, ErrNotFound
	}
	if r.Revision != rev || r.State == "DELETED" {
		return Record{}, ErrConflict
	}
	if state == "DELETED" || state == "ARCHIVED" {
		for _, x := range s.objects {
			if x.Owner.Kind == kind && x.Owner.ID == id && x.State != "DELETED" {
				return Record{}, fmt.Errorf("%w: active children", ErrConflict)
			}
		}
	}
	r.State = state
	if state == "DELETED" {
		deletedAt := time.Now().UTC()
		r.DeletedAt = &deletedAt
	}
	r.Revision++
	r.UpdatedAt = time.Now().UTC()
	s.objects[key(kind, id)] = r
	return r, nil
}
func (s *Store) MoveProject(id string, parent Owner, rev int64) (Record, error) {
	if !AllowedParent(Project, parent.Kind) || rev <= 0 {
		return Record{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.objects[key(Project, id)]
	if !ok {
		return Record{}, ErrNotFound
	}
	if r.Revision != rev || r.State != "ACTIVE" {
		return Record{}, ErrConflict
	}
	p, exists := s.objects[key(parent.Kind, parent.ID)]
	if !exists || p.State != "ACTIVE" {
		return Record{}, ErrNotFound
	}
	newOrg := p.OrganizationID
	if p.Kind == Organization {
		newOrg = p.ID
	}
	if newOrg != r.OrganizationID {
		return Record{}, fmt.Errorf("%w: cross-organization move prohibited", ErrInvalid)
	}
	r.Owner = parent
	r.Revision++
	r.UpdatedAt = time.Now().UTC()
	s.objects[key(Project, id)] = r
	return r, nil
}
func (s *Store) Resolve(kind Kind, id string) ([]Owner, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Owner{}
	visited := map[string]bool{}
	for i := 0; i < 10; i++ {
		r, ok := s.objects[key(kind, id)]
		if !ok {
			return nil, ErrNotFound
		}
		if visited[key(kind, id)] {
			return nil, ErrInvalid
		}
		visited[key(kind, id)] = true
		out = append([]Owner{{Kind: kind, ID: id}}, out...)
		if r.Owner.Kind == "" {
			return out, nil
		}
		kind, id = r.Owner.Kind, r.Owner.ID
	}
	return nil, ErrInvalid
}
func (s *Store) SetOrganizationGroupLink(groupID, organizationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, gok := s.objects[key(OrganizationGroup, groupID)]
	o, ook := s.objects[key(Organization, organizationID)]
	if !gok || !ook || g.TenantID != o.TenantID || g.State != "ACTIVE" || o.State != "ACTIVE" {
		return ErrInvalid
	}
	linkKey := groupID + ":" + organizationID
	if _, exists := s.links[linkKey]; !exists {
		s.links[linkKey] = time.Now().UTC()
	}
	return nil
}
func (s *Store) SetBinding(resource Owner, environmentID, version, config string) (Binding, error) {
	if version == "" {
		return Binding{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.objects[key(resource.Kind, resource.ID)]
	e, ok2 := s.objects[key(Environment, environmentID)]
	if !ok || !ok2 || r.State != "ACTIVE" || e.State != "ACTIVE" || r.OrganizationID != e.OrganizationID || r.OrganizationID == "" {
		return Binding{}, ErrInvalid
	}
	for _, b := range s.bindings {
		if b.Resource == resource && b.EnvironmentID == environmentID {
			b.Version = version
			b.ConfigRef = config
			b.Revision++
			s.bindings[b.ID] = b
			return b, nil
		}
	}
	bid, err := newID()
	if err != nil {
		return Binding{}, err
	}
	b := Binding{ID: bid, Resource: resource, EnvironmentID: environmentID, Version: version, ConfigRef: config, Revision: 1}
	s.bindings[bid] = b
	return b, nil
}

func (s *Store) RemoveOrganizationGroupLink(groupID, organizationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := groupID + ":" + organizationID
	if _, exists := s.links[k]; !exists {
		return ErrNotFound
	}
	delete(s.links, k)
	return nil
}
func (s *Store) ListOrganizationGroupLinks(groupID string) ([]OrganizationGroupLink, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.objects[key(OrganizationGroup, groupID)]; !ok {
		return nil, ErrNotFound
	}
	out := make([]OrganizationGroupLink, 0)
	prefix := groupID + ":"
	for k, createdAt := range s.links {
		if strings.HasPrefix(k, prefix) {
			out = append(out, OrganizationGroupLink{
				GroupID:        groupID,
				OrganizationID: strings.TrimPrefix(k, prefix),
				CreatedAt:      createdAt,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].OrganizationID < out[j].OrganizationID })
	return out, nil
}
func (s *Store) RemoveBinding(id string, revision int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.bindings[id]
	if !ok {
		return ErrNotFound
	}
	if revision <= 0 {
		return ErrInvalid
	}
	if b.Revision != revision {
		return ErrConflict
	}
	delete(s.bindings, id)
	return nil
}
func (s *Store) ListBindings(resource Owner, environmentID string) ([]Binding, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if resource.ID == "" && environmentID == "" {
		return nil, fmt.Errorf("%w: require resource or environment scope", ErrInvalid)
	}
	out := []Binding{}
	for _, b := range s.bindings {
		if resource.ID != "" && b.Resource != resource {
			continue
		}
		if environmentID != "" && b.EnvironmentID != environmentID {
			continue
		}
		out = append(out, b)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
