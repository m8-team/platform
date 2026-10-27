package ydb

import (
	"context"
	"database/sql"
	"errors"
	"os"
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
	// Resource hierarchy, ownership and environments must survive repository recreation.
	group, err := cat.Create(ctx, v1engine.Record{Kind: v1engine.Group, Owner: owner, Name: "Holding"}, seed+"-group")
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
	updated, err := cat.Update(ctx, v1engine.Project, project.ID, 1, "ERP New", "Owned by Alpha")
	if err != nil || updated.Revision != 2 {
		t.Fatalf("update CAS: %#v %v", updated, err)
	}
	if _, err = cat.Update(ctx, v1engine.Project, project.ID, 1, "stale", ""); !errors.Is(err, v1engine.ErrConflict) {
		t.Fatalf("stale update: %v", err)
	}
	if err = cat.SetGroupLink(ctx, group.ID, org.ID); err != nil {
		t.Fatal(err)
	}
	links, err := cat.ListGroupOrganizations(ctx, group.ID)
	if err != nil || len(links) != 1 || links[0] != org.ID {
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
	// Publication and bootstrap are in ONE transaction; no second initial-head event.
	var resourceEvents uint64
	err = db.QueryRowContext(ctx, graphStore.q("DECLARE $org AS Utf8;SELECT COUNT(*) FROM `m8-resource-manager/outbox` WHERE aggregate_id=$org;"), sql.Named("org", org.ID)).Scan(&resourceEvents)
	if err != nil || resourceEvents != 1 {
		t.Fatalf("Organization Outbox: count=%d err=%v", resourceEvents, err)
	}
}
