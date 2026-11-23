package cataloggrpc

import (
	"context"
	"testing"

	"github.com/m8-team/platform/internal/access/app/catalog"
	"github.com/m8-team/platform/internal/access/domain/module"
	"github.com/m8-team/platform/internal/access/domain/shared"
	accesspb "github.com/m8-team/platform/internal/proto/m8/platform/access/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeIdentity struct{ principal shared.Subject }

func (r fakeIdentity) AuthenticatedCatalogPrincipal(context.Context) (shared.Subject, error) {
	return r.principal, nil
}

type fakeAuthorizer struct {
	publish bool
	read    bool
}

func (a fakeAuthorizer) CanPublishModule(context.Context, shared.Subject, string) (bool, error) {
	return a.publish, nil
}
func (a fakeAuthorizer) CanReadModule(context.Context, shared.Subject, string) (bool, error) {
	return a.read, nil
}

type fakeRegistry struct{ record catalog.PublishedRecord }

func (r *fakeRegistry) Publish(_ context.Context, m module.Manifest, _ int64, _ string) (catalog.PublishedRecord, error) {
	r.record = catalog.PublishedRecord{Manifest: m, Revision: 1}
	return r.record, nil
}
func (r *fakeRegistry) Get(_ context.Context, moduleID, version string) (catalog.PublishedRecord, error) {
	if r.record.Revision == 0 || r.record.Manifest.ModuleID() != moduleID || r.record.Manifest.Version() != version {
		return catalog.PublishedRecord{}, catalog.ErrCatalogNotFound
	}
	return r.record, nil
}
func (r *fakeRegistry) List(_ context.Context, _ shared.Subject, _ string, _ int, _ string) ([]catalog.PublishedRecord, string, error) {
	if r.record.Revision == 0 {
		return nil, "", nil
	}
	return []catalog.PublishedRecord{r.record}, "", nil
}

func exampleManifest() *accesspb.CatalogModuleManifest {
	return &accesspb.CatalogModuleManifest{
		ModuleId: "finance", Version: "1.0.0",
		Resources: []*accesspb.CatalogResourceType{
			{Name: "finance.invoice", AllowedScopes: []accesspb.CatalogScope{
				accesspb.CatalogScope_CATALOG_SCOPE_INSTANCE,
				accesspb.CatalogScope_CATALOG_SCOPE_ORGANIZATION,
			}},
		},
		Permissions: []*accesspb.CatalogPermissionDefinition{
			{Name: "finance.invoice.view", ResourceType: "finance.invoice", GraphPermission: "use"},
		},
		Roles: []*accesspb.CatalogRoleTemplate{
			{Name: "finance.accountant", Version: "1.0.0", Permissions: []string{"finance.invoice.view"},
				AllowedScopes: []accesspb.CatalogScope{accesspb.CatalogScope_CATALOG_SCOPE_ORGANIZATION}},
		},
	}
}
func TestCatalogGRPC_PublicationAndDiscovery(t *testing.T) {
	actor, _ := shared.NewSubject("service_account", "finance-service")
	registry := &fakeRegistry{}
	api, err := catalog.NewAPI(fakeIdentity{principal: actor}, fakeAuthorizer{publish: true, read: true}, registry)
	if err != nil {
		t.Fatal(err)
	}
	s, err := New(api)
	if err != nil {
		t.Fatal(err)
	}
	published, err := s.PublishModuleManifest(context.Background(), &accesspb.PublishModuleManifestRequest{
		Manifest: exampleManifest(), ExpectedRevision: 0, IdempotencyKey: "request-1",
	})
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if published.Revision != 1 || published.PublisherSubjectId != "finance-service" {
		t.Fatalf("publisher must derive from verified principal: %+v", published)
	}
	got, err := s.GetModuleManifest(context.Background(), &accesspb.GetModuleManifestRequest{ModuleId: "finance", Version: "1.0.0"})
	if err != nil || got.Manifest.GetModuleId() != "finance" {
		t.Fatal(err)
	}
	p, err := s.ListCatalogPermissions(context.Background(), &accesspb.ListCatalogPermissionsRequest{ModuleId: "finance", Version: "1.0.0"})
	if err != nil || len(p.Permissions) != 1 || p.Permissions[0].Name != "finance.invoice.view" {
		t.Fatal(err)
	}
	rs, err := s.ListCatalogResourceTypes(context.Background(), &accesspb.ListCatalogResourceTypesRequest{ModuleId: "finance", Version: "1.0.0"})
	if err != nil || len(rs.ResourceTypes) != 1 {
		t.Fatal(err)
	}
	roles, err := s.ListCatalogRoleTemplates(context.Background(), &accesspb.ListCatalogRoleTemplatesRequest{ModuleId: "finance", Version: "1.0.0"})
	if err != nil || len(roles.Roles) != 1 {
		t.Fatal(err)
	}
	list, err := s.ListModuleManifests(context.Background(), &accesspb.ListModuleManifestsRequest{ModuleId: "finance", PageSize: 20})
	if err != nil || len(list.Manifests) != 1 {
		t.Fatal(err)
	}
}
func TestCatalogGRPC_FailsClosedForUntrustedPublication(t *testing.T) {
	user, _ := shared.NewSubject("user", "alice")
	registry := &fakeRegistry{}
	api, _ := catalog.NewAPI(fakeIdentity{principal: user}, fakeAuthorizer{publish: true, read: true}, registry)
	s, _ := New(api)
	_, err := s.PublishModuleManifest(context.Background(), &accesspb.PublishModuleManifestRequest{Manifest: exampleManifest(), IdempotencyKey: "request-1"})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("user should not register global module: %v", err)
	}
	m := exampleManifest()
	m.Permissions[0].GraphPermission = "manage"
	_, err = s.PublishModuleManifest(context.Background(), &accesspb.PublishModuleManifestRequest{Manifest: m, IdempotencyKey: "request-2"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("graph permission escalation accepted: %v", err)
	}
	m = exampleManifest()
	m.Resources[0].AllowedScopes = []accesspb.CatalogScope{accesspb.CatalogScope_CATALOG_SCOPE_UNSPECIFIED}
	_, err = s.PublishModuleManifest(context.Background(), &accesspb.PublishModuleManifestRequest{Manifest: m, IdempotencyKey: "request-3"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("invalid scope accepted: %v", err)
	}
	if registry.record.Revision != 0 {
		t.Fatal("untrusted input mutated registry")
	}
	_, err = s.GetModuleManifest(context.Background(), &accesspb.GetModuleManifestRequest{ModuleId: "finance", Version: "1.0.0"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("unexpected lookup status: %v", err)
	}
}
