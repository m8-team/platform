package grpcadapter

import (
	"context"

	"github.com/m8-team/platform/internal/resourcemanager/adapter/ydb"
	"github.com/m8-team/platform/internal/resourcemanager/v1engine"
)

// ResourceBackend keeps the existing v1 transport independent of storage.
// Every method accepts the caller context so YDB transactions respect
// cancellations, deadlines and scope-scoped service workflows.
type ResourceBackend interface {
	Create(context.Context, v1engine.Record, string) (v1engine.Record, error)
	Get(context.Context, v1engine.Kind, string) (v1engine.Record, error)
	List(context.Context, v1engine.Kind, string) ([]v1engine.Record, error)
	ListTenantOrganizations(context.Context, string) ([]v1engine.Record, error)
	ListTenantGroups(context.Context, string) ([]v1engine.Record, error)
	CreateRelationship(context.Context, v1engine.OrganizationRelationship, string) (v1engine.OrganizationRelationship, error)
	GetRelationship(context.Context, string, string) (v1engine.OrganizationRelationship, error)
	ListRelationshipVersions(context.Context, string, string) ([]v1engine.OrganizationRelationship, error)
	ListRelationships(context.Context, string, string, bool) ([]v1engine.OrganizationRelationship, error)
	UpdateRelationship(context.Context, v1engine.RelationshipUpdate, string) (v1engine.OrganizationRelationship, error)
	CloseRelationship(context.Context, string, string, int64, int64, string) (v1engine.OrganizationRelationship, error)
	Update(context.Context, v1engine.Kind, string, int64, string, string) (v1engine.Record, error)
	ChangeState(context.Context, v1engine.Kind, string, int64, string) (v1engine.Record, error)
	MoveProject(context.Context, string, v1engine.Owner, int64) (v1engine.Record, error)
	Resolve(context.Context, v1engine.Kind, string) ([]v1engine.Owner, error)
	SetGroupLink(context.Context, string, string) error
	RemoveGroupLink(context.Context, string, string) error
	ListGroupOrganizations(context.Context, string) ([]string, error)
	SetBinding(context.Context, v1engine.Owner, string, string, string) (v1engine.Binding, error)
	RemoveBinding(context.Context, string, int64) error
	ListBindings(context.Context, v1engine.Owner, string) ([]v1engine.Binding, error)
}
type memoryResourceBackend struct{ store *v1engine.Store }

func (b memoryResourceBackend) Create(_ context.Context, r v1engine.Record, k string) (v1engine.Record, error) {
	return b.store.Create(r, k)
}
func (b memoryResourceBackend) Get(_ context.Context, k v1engine.Kind, id string) (v1engine.Record, error) {
	return b.store.Get(k, id)
}
func (b memoryResourceBackend) List(_ context.Context, k v1engine.Kind, org string) ([]v1engine.Record, error) {
	return b.store.List(k, org)
}
func (b memoryResourceBackend) ListTenantOrganizations(_ context.Context, tenantID string) ([]v1engine.Record, error) {
	return b.store.ListTenantOrganizations(tenantID)
}
func (b memoryResourceBackend) ListTenantGroups(_ context.Context, tenantID string) ([]v1engine.Record, error) {
	return b.store.ListTenantGroups(tenantID)
}
func (b memoryResourceBackend) Update(_ context.Context, k v1engine.Kind, id string, rev int64, name, description string) (v1engine.Record, error) {
	return b.store.Update(k, id, rev, name, description)
}
func (b memoryResourceBackend) ChangeState(_ context.Context, k v1engine.Kind, id string, rev int64, state string) (v1engine.Record, error) {
	return b.store.ChangeState(k, id, rev, state)
}
func (b memoryResourceBackend) MoveProject(_ context.Context, id string, o v1engine.Owner, rev int64) (v1engine.Record, error) {
	return b.store.MoveProject(id, o, rev)
}
func (b memoryResourceBackend) Resolve(_ context.Context, k v1engine.Kind, id string) ([]v1engine.Owner, error) {
	return b.store.Resolve(k, id)
}
func (b memoryResourceBackend) SetGroupLink(_ context.Context, g, o string) error {
	return b.store.SetGroupLink(g, o)
}
func (b memoryResourceBackend) RemoveGroupLink(_ context.Context, g, o string) error {
	return b.store.RemoveGroupLink(g, o)
}
func (b memoryResourceBackend) ListGroupOrganizations(_ context.Context, g string) ([]string, error) {
	return b.store.ListGroupOrganizations(g)
}
func (b memoryResourceBackend) SetBinding(_ context.Context, r v1engine.Owner, e, v, c string) (v1engine.Binding, error) {
	return b.store.SetBinding(r, e, v, c)
}
func (b memoryResourceBackend) RemoveBinding(_ context.Context, id string, rev int64) error {
	return b.store.RemoveBinding(id, rev)
}
func (b memoryResourceBackend) ListBindings(_ context.Context, r v1engine.Owner, e string) ([]v1engine.Binding, error) {
	return b.store.ListBindings(r, e)
}

func (b memoryResourceBackend) CreateRelationship(_ context.Context, r v1engine.OrganizationRelationship, k string) (v1engine.OrganizationRelationship,error) {
	return b.store.CreateRelationship(r,k)
}
func (b memoryResourceBackend) GetRelationship(_ context.Context, tenant,id string) (v1engine.OrganizationRelationship,error) {
	return b.store.GetRelationship(tenant,id)
}
func (b memoryResourceBackend) ListRelationshipVersions(_ context.Context, tenant,id string) ([]v1engine.OrganizationRelationship,error) {
	return b.store.ListRelationshipVersions(tenant,id)
}
func (b memoryResourceBackend) ListRelationships(_ context.Context, tenant,org string, closed bool) ([]v1engine.OrganizationRelationship,error) {
	return b.store.ListRelationships(tenant,org,closed)
}
func (b memoryResourceBackend) UpdateRelationship(_ context.Context, u v1engine.RelationshipUpdate, k string) (v1engine.OrganizationRelationship,error) {
	return b.store.UpdateRelationship(u,k)
}
func (b memoryResourceBackend) CloseRelationship(_ context.Context, tenant,id string, rev,at int64, k string) (v1engine.OrganizationRelationship,error) {
	return b.store.CloseRelationship(tenant,id,rev,at,k)
}

var _ ResourceBackend = (*ydb.Catalog)(nil)
var _ ResourceBackend = memoryResourceBackend{}
