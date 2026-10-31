package ydb

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/resourcemanager/v1engine"
	_ "github.com/ydb-platform/ydb-go-sdk/v3"
)

func TestIntegrationUnifiedCatalogAndAtomicOrganization(t *testing.T) {
	dsn := os.Getenv("M8_RM_TEST_YDB_DSN")
	if dsn == "" {
		t.Skip("requires disposable local YDB")
	}
	if os.Getenv("M8_RM_TEST_YDB_BOOTSTRAP") != "1" {
		t.Skip("only run with explicit disposable test schema")
	}
	prefix := os.Getenv("M8_RM_TEST_YDB_PREFIX")
	if prefix == "" {
		prefix = "/local/m8"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	db, err := sql.Open("ydb", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for attempt := 0; attempt < 30; attempt++ {
		if err = db.PingContext(ctx); err == nil {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatalf("YDB start: %v", err)
		case <-time.After(time.Second):
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	if err = prepareIntegrationSchema(ctx, db, dsn, prefix); err != nil {
		t.Fatal(err)
	}
	if err = prepareResourceCatalogSchema(ctx, db, prefix); err != nil {
		t.Fatal(err)
	}
	cat, err := NewCatalog(db, prefix)
	if err != nil {
		t.Fatal(err)
	}
	seed, err := newID()
	if err != nil {
		t.Fatal(err)
	}
	// A syntactically valid Tenant UUID does not provision a Tenant.
	// This is the source of the observed gRPC code 5 / HTTP 404 response.
	missingTenant, err := newID()
	if err != nil {
		t.Fatal(err)
	}
	_, err = cat.Create(ctx, v1engine.Record{
		Kind:  v1engine.Organization,
		Owner: v1engine.Owner{Kind: v1engine.Tenant, ID: missingTenant},
		Name:  "Must not be created",
	}, seed+"-missing-parent")
	if !errors.Is(err, v1engine.ErrNotFound) {
		t.Fatalf("missing Tenant must return NotFound, got: %v", err)
	}
	tenant, err := cat.Create(ctx, v1engine.Record{Kind: v1engine.Tenant, Name: "Parent tenant"}, seed+"-tenant")
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	owner := v1engine.Owner{Kind: v1engine.Tenant, ID: tenant.ID}
	orgKey := seed + "-organization"
	org, err := cat.Create(ctx, v1engine.Record{Kind: v1engine.Organization, Owner: owner, Name: "Alpha Retail"}, orgKey)
	if err != nil {
		t.Fatalf("atomic organization creation: %v", err)
	}
	// Tenant discovery never mixes organizations across tenants.
	otherTenant, err := cat.Create(ctx, v1engine.Record{Kind: v1engine.Tenant, Name: "Second holding"}, seed+"-second-tenant")
	if err != nil {
		t.Fatal(err)
	}
	otherOrg, err := cat.Create(ctx, v1engine.Record{
		Kind: v1engine.Organization, Name: "Beta Retail",
		Owner: v1engine.Owner{Kind: v1engine.Tenant, ID: otherTenant.ID},
	}, seed+"-second-organization")
	if err != nil {
		t.Fatal(err)
	}
	firstList, err := cat.ListTenantOrganizations(ctx, tenant.ID)
	if err != nil || len(firstList) != 1 || firstList[0].ID != org.ID {
		t.Fatalf("first Tenant organizations: %#v err=%v", firstList, err)
	}
	secondList, err := cat.ListTenantOrganizations(ctx, otherTenant.ID)
	if err != nil || len(secondList) != 1 || secondList[0].ID != otherOrg.ID {
		t.Fatalf("second Tenant organizations: %#v err=%v", secondList, err)
	}
	if _, err = cat.ListTenantOrganizations(ctx, "00000000-0000-4000-8000-000000000001"); !errors.Is(err, v1engine.ErrNotFound) {
		t.Fatalf("missing tenant must return NotFound, got %v", err)
	}
	// Check canonical Resource and active OrgUnit head after reopening the repositories.
	another, err := NewCatalog(db, prefix)
	if err != nil {
		t.Fatal(err)
	}
	durable, err := another.Get(ctx, v1engine.Organization, org.ID)
	if err != nil || durable.TenantID != tenant.ID || durable.ID != org.ID {
		t.Fatalf("durability: %#v %v", durable, err)
	}
	graphStore, err := NewStore(db, prefix)
	if err != nil {
		t.Fatal(err)
	}
	graph, err := graphStore.Load(ctx, tenant.ID, org.ID)
	if err != nil || graph.Revision != 1 || len(graph.Units) != 0 || len(graph.Closure) != 0 {
		t.Fatalf("organization's initial graph: %#v %v", graph, err)
	}
	replay, err := cat.Create(ctx, v1engine.Record{Kind: v1engine.Organization, Owner: owner, Name: "Alpha Retail"}, orgKey)
	if err != nil || replay.ID != org.ID {
		t.Fatalf("idempotent replay: %#v %v", replay, err)
	}
	if _, err = cat.Create(ctx, v1engine.Record{Kind: v1engine.Organization, Owner: owner, Name: "Changed payload"}, orgKey); !errors.Is(err, v1engine.ErrConflict) {
		t.Fatalf("idempotency conflict: %v", err)
	}
	// Concurrent identical commands must converge on one Organization identity.
	concurrentKey := seed + "-concurrent-org"
	concurrentRequest := v1engine.Record{Kind: v1engine.Organization, Owner: owner, Name: "Concurrent"}
	type result struct {
		value v1engine.Record
		err   error
	}
	results := make(chan result, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := cat.Create(ctx, concurrentRequest, concurrentKey)
			results <- result{value, err}
		}()
	}
	wg.Wait()
	close(results)
	var winner string
	for outcome := range results {
		if outcome.err == nil {
			if winner != "" && winner != outcome.value.ID {
				t.Fatal("concurrent create produced two identities")
			}
			winner = outcome.value.ID
		}
	}
	if winner == "" {
		t.Fatal("no concurrent create succeeded")
	}
	final, err := cat.Create(ctx, concurrentRequest, concurrentKey)
	if err != nil || final.ID != winner {
		t.Fatalf("concurrent retry: %#v %v", final, err)
	}
	secondGraph, err := graphStore.Load(ctx, tenant.ID, winner)
	if err != nil || secondGraph.Revision != 1 {
		t.Fatalf("missing concurrent initial graph: %#v %v", secondGraph, err)
	}
	// Corporate relations are separate from Group affiliations and durable.
	relInput := v1engine.OrganizationRelationship{
		TenantID: tenant.ID, FromOrganizationID: org.ID, ToOrganizationID: winner,
		Type: v1engine.RelationshipParentSubsidiary, OwnershipBasisPoints: 5100,
		ValidFrom: time.Now().Add(-time.Hour).Unix(), EvidenceRef: "contract/123",
	}
	rel, err := cat.CreateRelationship(ctx, relInput, seed+"-corporate-relation")
	if err != nil || rel.Revision != 1 || rel.ID == "" {
		t.Fatalf("create durable relation: %#v %v", rel, err)
	}
	replayedRel, err := cat.CreateRelationship(ctx, relInput, seed+"-corporate-relation")
	if err != nil || replayedRel.ID != rel.ID {
		t.Fatalf("corporate idempotency: %#v %v", replayedRel, err)
	}
	if _, err = cat.CreateRelationship(ctx, v1engine.OrganizationRelationship{
		TenantID: tenant.ID, FromOrganizationID: org.ID, ToOrganizationID: otherOrg.ID,
		Type: v1engine.RelationshipPartner, ValidFrom: relInput.ValidFrom,
	}, seed+"-cross-tenant"); !errors.Is(err, v1engine.ErrInvalid) {
		t.Fatalf("cross-tenant corporate relation accepted: %v", err)
	}
	if _, err = cat.CreateRelationship(ctx, v1engine.OrganizationRelationship{
		TenantID: tenant.ID, FromOrganizationID: winner, ToOrganizationID: org.ID,
		Type: v1engine.RelationshipParentSubsidiary, ValidFrom: relInput.ValidFrom,
	}, seed+"-cycle"); !errors.Is(err, v1engine.ErrRelationshipCycle) {
		t.Fatalf("corporate ownership cycle accepted: %v", err)
	}
	if rows, err := cat.ListRelationships(ctx, otherTenant.ID, "", true); err != nil || len(rows) != 0 {
		t.Fatalf("corporate tenant isolation: %#v %v", rows, err)
	}
	if _, err = another.GetRelationship(ctx, otherTenant.ID, rel.ID); !errors.Is(err, v1engine.ErrNotFound) {
		t.Fatalf("cross-tenant relationship read: %v", err)
	}
	relUpdated, err := cat.UpdateRelationship(ctx, v1engine.RelationshipUpdate{
		TenantID: tenant.ID, ID: rel.ID, ExpectedRevision: 1, OwnershipBasisPoints: 5200,
		EvidenceRef: "contract/124",
	}, seed+"-update-relation")
	if err != nil || relUpdated.Revision != 2 {
		t.Fatalf("relationship optimistic update: %#v %v", relUpdated, err)
	}
	relClosed, err := cat.CloseRelationship(ctx, tenant.ID, rel.ID, 2, time.Now().Unix(), seed+"-close-relation")
	if err != nil || relClosed.State != "CLOSED" || relClosed.Revision != 3 {
		t.Fatalf("relationship close: %#v %v", relClosed, err)
	}
	versions, err := another.ListRelationshipVersions(ctx, tenant.ID, rel.ID)
	if err != nil || len(versions) != 3 || versions[0].Revision != 1 || versions[2].Revision != 3 {
		t.Fatalf("corporate audit revisions: %#v %v", versions, err)
	}

	// Resource hierarchy, ownership and environments must survive repository recreation.
	group, err := cat.Create(ctx, v1engine.Record{Kind: v1engine.OrganizationGroup, Owner: owner, Name: "Holding"}, seed+"-group")
	if err != nil {
		t.Fatal(err)
	}
	ws, err := cat.Create(ctx, v1engine.Record{Kind: v1engine.Workspace, Owner: v1engine.Owner{Kind: v1engine.Organization, ID: org.ID}, Name: "Engineering"}, seed+"-ws")
	if err != nil {
		t.Fatal(err)
	}
	project, err := cat.Create(ctx, v1engine.Record{Kind: v1engine.Project, Owner: v1engine.Owner{Kind: v1engine.Workspace, ID: ws.ID}, Name: "ERP"}, seed+"-project")
	if err != nil {
		t.Fatal(err)
	}
	env, err := cat.Create(ctx, v1engine.Record{Kind: v1engine.Environment, Owner: v1engine.Owner{Kind: v1engine.Organization, ID: org.ID}, Name: "Production", EnvironmentKey: "prod"}, seed+"-prod")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = cat.Create(ctx, v1engine.Record{Kind: v1engine.Environment, Owner: v1engine.Owner{Kind: v1engine.Organization, ID: org.ID}, Name: "Production Again", EnvironmentKey: "prod"}, seed+"-prod-duplicate"); !errors.Is(err, v1engine.ErrConflict) {
		t.Fatalf("duplicate environment key: %v", err)
	}
	service, err := cat.Create(ctx, v1engine.Record{Kind: v1engine.WorkloadService, Owner: v1engine.Owner{Kind: v1engine.Project, ID: project.ID}, Name: "ERP API"}, seed+"-service")
	if err != nil {
		t.Fatal(err)
	}
	got, err := another.List(ctx, "", org.ID)
	if err != nil || len(got) != 5 {
		t.Fatalf("scope listing: len=%d %v", len(got), err)
	}
	parents, err := another.Resolve(ctx, v1engine.WorkloadService, service.ID)
	if err != nil || len(parents) != 5 {
		t.Fatalf("ancestry: %#v %v", parents, err)
	}
	updated, err := cat.Update(ctx, v1engine.Project, project.ID, 1, "ERP New", "Owned by Alpha", nil)
	if err != nil || updated.Revision != 2 {
		t.Fatalf("update CAS: %#v %v", updated, err)
	}
	if _, err = cat.Update(ctx, v1engine.Project, project.ID, 1, "stale", "", nil); !errors.Is(err, v1engine.ErrConflict) {
		t.Fatalf("stale update: %v", err)
	}
	if err = cat.SetOrganizationGroupLink(ctx, group.ID, org.ID); err != nil {
		t.Fatal(err)
	}
	links, err := cat.ListOrganizationGroupLinks(ctx, group.ID)
	if err != nil || len(links) != 1 || links[0].GroupID != group.ID || links[0].OrganizationID != org.ID || links[0].CreatedAt.IsZero() {
		t.Fatalf("group links: %#v %v", links, err)
	}
	binding, err := cat.SetBinding(ctx, v1engine.Owner{Kind: v1engine.WorkloadService, ID: service.ID}, env.ID, "v1", "ref://config")
	if err != nil {
		t.Fatalf("binding: %v", err)
	}
	bindings, err := cat.ListBindings(ctx, service.Owner, env.ID)
	_ = bindings
	// The resource filter is explicitly the service, not the service's parent.
	bindings, err = cat.ListBindings(ctx, v1engine.Owner{Kind: v1engine.WorkloadService, ID: service.ID}, env.ID)
	if err != nil || len(bindings) != 1 || bindings[0].ID != binding.ID {
		t.Fatalf("bindings: %#v %v", bindings, err)
	}
	if _, err = cat.ChangeState(ctx, v1engine.Organization, org.ID, 1, "DELETED"); !errors.Is(err, v1engine.ErrConflict) {
		t.Fatalf("delete with descendants must fail: %v", err)
	}
	secondWorkspace, err := cat.Create(ctx, v1engine.Record{Kind: v1engine.Workspace, Owner: v1engine.Owner{Kind: v1engine.Organization, ID: org.ID}, Name: "Other Workspace"}, seed+"-ws2")
	if err != nil {
		t.Fatal(err)
	}
	moved, err := cat.MoveProject(ctx, project.ID, v1engine.Owner{Kind: v1engine.Workspace, ID: secondWorkspace.ID}, 2)
	if err != nil || moved.Revision != 3 || moved.Owner.ID != secondWorkspace.ID {
		t.Fatalf("move: %#v %v", moved, err)
	}
	if _, err = cat.MoveProject(ctx, project.ID, v1engine.Owner{Kind: v1engine.Workspace, ID: ws.ID}, 2); !errors.Is(err, v1engine.ErrConflict) {
		t.Fatalf("stale move: %v", err)
	}
	if err = cat.RemoveBinding(ctx, binding.ID, 1); err != nil {
		t.Fatalf("delete binding: %v", err)
	}
	remaining, err := cat.ListBindings(ctx, v1engine.Owner{Kind: v1engine.WorkloadService, ID: service.ID}, env.ID)
	if err != nil || len(remaining) != 0 {
		t.Fatalf("remaining bindings: %#v %v", remaining, err)
	}
	if err = cat.RemoveOrganizationGroupLink(ctx, group.ID, org.ID); err != nil {
		t.Fatal(err)
	}
	links, err = cat.ListOrganizationGroupLinks(ctx, group.ID)
	if err != nil || len(links) != 0 {
		t.Fatalf("remove group link: %#v %v", links, err)
	}
	if _, err = cat.ChangeState(ctx, v1engine.WorkloadService, service.ID, 1, "SUSPENDED"); err != nil {
		t.Fatalf("state change: %v", err)
	}
	// Publication and bootstrap are in ONE transaction; no second initial-head event.
	var resourceEvents uint64
	err = db.QueryRowContext(ctx, graphStore.q("DECLARE $org AS Utf8;SELECT COUNT(*) FROM `m8-resource-manager/outbox` WHERE aggregate_id=$org;"), sql.Named("org", org.ID)).Scan(&resourceEvents)
	if err != nil || resourceEvents != 1 {
		t.Fatalf("Organization Outbox: count=%d err=%v", resourceEvents, err)
	}
}
