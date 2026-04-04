package testutil

import (
	"context"
	"sort"
	"sync"
	"time"

	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	identityv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/identity/v1"
	"github.com/m8platform/platform/iam/internal/core"
	"google.golang.org/protobuf/proto"
)

type FakeStore struct {
	mu     sync.Mutex
	tables map[string]map[string]core.StoredDocument
}

func NewFakeStore() *FakeStore {
	return &FakeStore{tables: make(map[string]map[string]core.StoredDocument)}
}

func (s *FakeStore) GetDocument(_ context.Context, table string, id string) (core.StoredDocument, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tables[table] == nil {
		return core.StoredDocument{}, core.ErrNotFound
	}
	document, ok := s.tables[table][id]
	if !ok {
		return core.StoredDocument{}, core.ErrNotFound
	}
	return document, nil
}

func (s *FakeStore) UpsertDocument(_ context.Context, table string, doc core.StoredDocument) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tables[table] == nil {
		s.tables[table] = make(map[string]core.StoredDocument)
	}
	s.tables[table][doc.ID] = doc
	return nil
}

func (s *FakeStore) DeleteDocument(_ context.Context, table string, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tables[table] == nil {
		return core.ErrNotFound
	}
	delete(s.tables[table], id)
	return nil
}

func (s *FakeStore) ListDocuments(_ context.Context, table string, tenantID string, offset int, limit int) ([]core.StoredDocument, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if limit <= 0 {
		limit = core.DefaultPageSize
	}
	records := make([]core.StoredDocument, 0, len(s.tables[table]))
	for _, document := range s.tables[table] {
		if tenantID == "" || document.TenantID == tenantID {
			records = append(records, document)
		}
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].ID < records[j].ID
	})
	if offset >= len(records) {
		return nil, "", nil
	}
	end := offset + limit
	if end > len(records) {
		end = len(records)
	}
	next := ""
	if end < len(records) {
		next = core.EncodePageToken(end)
	}
	return records[offset:end], next, nil
}

type FakePublisher struct {
	mu      sync.Mutex
	Topics  []string
	Payload []proto.Message
}

func (p *FakePublisher) PublishProto(_ context.Context, topic string, msg proto.Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Topics = append(p.Topics, topic)
	p.Payload = append(p.Payload, msg)
	return nil
}

type WorkflowCall struct {
	WorkflowName string
	WorkflowID   string
	Input        any
}

type FakeWorkflowStarter struct {
	mu    sync.Mutex
	Calls []WorkflowCall
}

func (w *FakeWorkflowStarter) StartWorkflow(_ context.Context, workflowName string, workflowID string, input any) (string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.Calls = append(w.Calls, WorkflowCall{
		WorkflowName: workflowName,
		WorkflowID:   workflowID,
		Input:        input,
	})
	return workflowID, nil
}

type FakeAuthorizationRuntime struct {
	Result                   *authzv1.AccessCheckResult
	Err                      error
	CheckErr                 error
	SyncErr                  error
	WriteGroupMembershipErr  error
	DeleteGroupMembershipErr error
}

func (r *FakeAuthorizationRuntime) Check(_ context.Context, _ *authzv1.CheckAccessRequest) (*authzv1.AccessCheckResult, error) {
	if r.CheckErr != nil {
		return r.Result, r.CheckErr
	}
	return r.Result, r.Err
}

func (r *FakeAuthorizationRuntime) SyncResource(_ context.Context, _ *authzv1.ResourceRef, _ []*authzv1.AccessBinding) error {
	if r.SyncErr != nil {
		return r.SyncErr
	}
	return r.Err
}

func (r *FakeAuthorizationRuntime) WriteGroupMembership(_ context.Context, _ string, _ *identityv1.GroupMember) error {
	if r.WriteGroupMembershipErr != nil {
		return r.WriteGroupMembershipErr
	}
	return r.Err
}

func (r *FakeAuthorizationRuntime) DeleteGroupMembership(_ context.Context, _ string, _ string, _ string, _ string) error {
	if r.DeleteGroupMembershipErr != nil {
		return r.DeleteGroupMembershipErr
	}
	return r.Err
}

type FakeCache struct {
	mu   sync.Mutex
	data map[string]string
}

func NewFakeCache() *FakeCache {
	return &FakeCache{data: make(map[string]string)}
}

func (c *FakeCache) Get(_ context.Context, key string) (string, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.data[key]
	return value, ok, nil
}

func (c *FakeCache) Set(_ context.Context, key string, value string, _ time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
	return nil
}

func (c *FakeCache) Delete(_ context.Context, keys ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, key := range keys {
		delete(c.data, key)
	}
	return nil
}
