//go:build integration

package ydbroles

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/domain/role"
	"github.com/m8-team/platform/internal/access/domain/shared"
	ydbsdk "github.com/ydb-platform/ydb-go-sdk/v3"
)

func liveStore(t *testing.T) *Store {
	t.Helper()
	if os.Getenv("M8_ACCESS_LIVE_YDB") != "1" {
		t.Skip("requires disposable YDB")
	}
	dsn := os.Getenv("M8_ACCESS_TEST_YDB_DSN")
	base := os.Getenv("M8_ACCESS_TEST_YDB_PREFIX")
	if !strings.HasPrefix(base, "/local/m8_access_ci") || !strings.Contains(dsn, "localhost:2136/local") {
		t.Fatal("refuse to test outside disposable local YDB")
	}
	suffix := strings.NewReplacer("/", "_", " ", "_").Replace(strings.ToLower(t.Name()))
	prefix := base + "_roles_" + suffix
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	t.Cleanup(cancel)
	db, err := sql.Open("ydb", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	for attempt := 0; attempt < 40; attempt++ {
		probe, done := context.WithTimeout(ctx, 3*time.Second)
		err = db.PingContext(probe)
		done()
		if err == nil {
			break
		}
		if ctx.Err() != nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		t.Fatal(err)
	}
	sdk, err := ydbsdk.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer sdk.Close(ctx)
	for _, path := range []string{base, prefix, prefix + "/m8-access"} {
		if err := sdk.Scheme().MakeDirectory(ctx, path); err != nil {
			if _, exists := sdk.Scheme().ListDirectory(ctx, path); exists != nil {
				t.Fatalf("directory %s: %v", path, err)
			}
		}
	}
	source, err := os.ReadFile(filepath.Join("schema.sql"))
	if err != nil {
		t.Fatal(err)
	}
	var clean []string
	for _, line := range strings.Split(string(source), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "--") {
			clean = append(clean, line)
		}
	}
	for _, fragment := range strings.Split(strings.Join(clean, "\n"), ";") {
		stmt := strings.TrimSpace(fragment)
		if stmt == "" {
			continue
		}
		_, err = db.ExecContext(ctx, fmt.Sprintf("PRAGMA TablePathPrefix(%q);\n%s;", prefix, stmt))
		if err != nil {
			t.Fatalf("schema migration: %v, statement: %s", err, stmt)
		}
	}
	store, err := New(db, prefix)
	if err != nil {
		t.Fatal(err)
	}
	return store
}
func TestLiveYDBRoleBundleAndAssignmentCQRS(t *testing.T) {
	s := liveStore(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	employee, _ := role.NewBundle(role.BundleRef{Name: "profiles.employee", Version: "1.0.0"},
		[]role.TemplateRef{{Name: "identity.employee", Version: "1.0.0"}}, nil)
	r, err := s.PublishBundle(ctx, employee, 0, "employee-1")
	if err != nil || r.Revision != 1 {
		t.Fatalf("create employee bundle: %v", err)
	}
	root := testBundle(t, "1.0.0")
	record, err := s.PublishBundle(ctx, root, 0, "accountant-1")
	if err != nil || record.Revision != 1 {
		t.Fatalf("bundle create: %v", err)
	}
	retry, err := s.PublishBundle(ctx, root, 0, "accountant-1")
	if err != nil || retry.Revision != 1 {
		t.Fatalf("idempotent bundle retry: %v", err)
	}
	if _, err := s.PublishBundle(ctx, testBundle(t, "1.0.1"), 0, "accountant-2"); !errors.Is(err, shared.ErrRevisionConflict) {
		t.Fatalf("stale bundle CAS: %v", err)
	}
	if _, err := s.PublishBundle(ctx, testBundle(t, "1.0.1"), 1, "accountant-1"); !errors.Is(err, ErrRoleConflict) {
		t.Fatalf("reused bundle request key: %v", err)
	}
	second, err := s.PublishBundle(ctx, testBundle(t, "1.0.1"), 1, "accountant-2")
	if err != nil || second.Revision != 2 {
		t.Fatalf("second version: %v", err)
	}
	got, err := s.GetBundle(ctx, root.Ref())
	if err != nil || got.Revision != 1 {
		t.Fatalf("immutable profile version: %v", err)
	}
	versions, err := s.ListBundleVersions(ctx, root.Ref().Name, 0, 10)
	if err != nil || len(versions) != 2 {
		t.Fatalf("bundle CQRS index: %+v %v", versions, err)
	}
	a := testAssignment(t, "33333333-3333-4333-8333-333333333333")
	created, err := s.CreateAssignment(ctx, a, "create-1")
	if err != nil || created.Status() != role.ProfileAssignmentPending {
		t.Fatalf("stage assignment: %v", err)
	}
	again, err := s.CreateAssignment(ctx, a, "create-1")
	if err != nil || again.Revision() != 1 {
		t.Fatalf("replay assignment creation: %v", err)
	}
	if _, err := s.CreateAssignment(ctx, a, "other-create"); !errors.Is(err, ErrAssignmentConflict) {
		t.Fatalf("duplicate assignment ID accepted: %v", err)
	}
	list, err := s.ListBySubject(ctx, a.TenantID(), a.Subject(), "", 10)
	if err != nil || len(list) != 1 || list[0].Status() != role.ProfileAssignmentPending {
		t.Fatalf("subject index: %+v %v", list, err)
	}
	if _, err := s.RevokeAssignment(ctx, a.TenantID(), a.ID(), 2, "revoke-2"); !errors.Is(err, shared.ErrRevisionConflict) {
		t.Fatalf("stale revocation: %v", err)
	}
	revoked, err := s.RevokeAssignment(ctx, a.TenantID(), a.ID(), 1, "revoke-1")
	if err != nil || revoked.Revision() != 2 || revoked.Status() != role.ProfileAssignmentRevoked {
		t.Fatalf("revoke: %+v %v", revoked, err)
	}
	replayed, err := s.RevokeAssignment(ctx, a.TenantID(), a.ID(), 1, "revoke-1")
	if err != nil || replayed.Revision() != 2 {
		t.Fatalf("revoke replay: %v", err)
	}
	original, err := s.CreateAssignment(ctx, a, "create-1")
	if err != nil || original.Status() != role.ProfileAssignmentPending || original.Revision() != 1 {
		t.Fatalf("creation replay returned mutable state: %+v %v", original, err)
	}
	latest, err := s.GetAssignment(ctx, a.TenantID(), a.ID())
	if err != nil || latest.Revision() != 2 || latest.Status() != role.ProfileAssignmentRevoked {
		t.Fatalf("latest state: %v", err)
	}
	events, err := s.Pending(ctx, "assignment", a.TenantID().String(), a.ID(), 0, 10)
	if err != nil || len(events) != 2 || events[0].Kind != "ProfileAssignmentStaged" ||
		events[1].Kind != "ProfileAssignmentRevoked" {
		t.Fatalf("assignment outbox: %+v %v", events, err)
	}
	if err := s.MarkPublished(ctx, events[0]); err != nil {
		t.Fatal(err)
	}
	if err := s.MarkPublished(ctx, events[0]); err != nil {
		t.Fatal(err)
	}
	remaining, err := s.Pending(ctx, "assignment", a.TenantID().String(), a.ID(), 0, 10)
	if err != nil || len(remaining) != 1 || remaining[0].Revision != 2 {
		t.Fatalf("durable broker ack: %+v %v", remaining, err)
	}
}
func TestLiveYDBRoleBundleConcurrentCAS(t *testing.T) {
	s := liveStore(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if _, err := s.PublishBundle(ctx, testBundle(t, "1.0.0"), 0, "seed"); err != nil {
		t.Fatal(err)
	}
	candidates := []role.Bundle{testBundle(t, "1.0.1"), testBundle(t, "1.0.2")}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, bundle := range candidates {
		bundle := bundle
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.PublishBundle(ctx, bundle, 1, "concurrent-"+bundle.Ref().Version)
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	winners := 0
	for err := range results {
		if err == nil {
			winners++
		}
	}
	if winners != 1 {
		t.Fatalf("expected one winning CAS publisher, got %d", winners)
	}
	versions, err := s.ListBundleVersions(ctx, "profiles.accountant", 0, 10)
	if err != nil || len(versions) != 2 {
		t.Fatalf("atomic CQRS: %+v %v", versions, err)
	}
}
func TestLiveYDBRoleAssignmentOutboxAtomicRollback(t *testing.T) {
	s := liveStore(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	_, err := s.db.ExecContext(ctx, s.q("DROP TABLE `m8-access/role_management_outbox`;"))
	if err != nil {
		t.Fatal(err)
	}
	a := testAssignment(t, "33333333-3333-4333-8333-333333333333")
	if _, err := s.CreateAssignment(ctx, a, "create"); err == nil {
		t.Fatal("write succeeded without outbox")
	}
	if _, err := s.GetAssignment(ctx, a.TenantID(), a.ID()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("orphan assignment: %v", err)
	}
	visible, err := s.ListBySubject(ctx, a.TenantID(), a.Subject(), "", 10)
	if err != nil || len(visible) != 0 {
		t.Fatalf("orphan CQRS index: %+v %v", visible, err)
	}
}
