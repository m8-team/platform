package event

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
)

func TestIntegrationEventCredentialSeparation(t *testing.T) {
	ctx, admin, path := liveDatabase(t)
	createTestTopic(t, ctx, admin, path, 1)
	owners := createTestTable(t, ctx, admin, path+"_owner", "replica Utf8 NOT NULL, pod_uid Utf8 NOT NULL, generation Uint64 NOT NULL, PRIMARY KEY (replica)")
	guards := createTestTable(t, ctx, admin, path+"_guard", "replica Utf8 NOT NULL, revision Uint64 NOT NULL, PRIMARY KEY (replica)")
	principal := func(grants ...string) *ydb.Driver {
		t.Helper()
		user := "evt" + strings.ReplaceAll(uuid.NewString(), "-", "")[:16]
		password := "EventAa1" + strings.ReplaceAll(uuid.NewString(), "-", "")
		if err := admin.Query().Exec(ctx, "CREATE USER "+user+" PASSWORD '"+password+"';"); err != nil {
			ydb.IterateByIssues(err, func(message string, _ Ydb.StatusIds_StatusCode, _ uint32) {
				t.Log(strings.ReplaceAll(message, password, "[redacted]"))
			})
			t.Fatal("could not create disposable credentials")
		}
		t.Cleanup(func() {
			cleanup, stop := context.WithTimeout(context.Background(), 10*time.Second)
			defer stop()
			_ = admin.Query().Exec(cleanup, "DROP USER "+user+";")
		})
		for _, grant := range append(grants, "CONNECT ON `"+admin.Name()+"`") {
			if err := admin.Query().Exec(ctx, "GRANT "+grant+" TO "+user+";"); err != nil {
				t.Fatalf("grant disposable event privileges: %v", err)
			}
		}
		_, db, _ := liveDatabase(t, ydb.WithStaticCredentials(user, password))
		return db
	}
	writer := principal("SELECT, UPDATE ROW ON " + owners)
	reader := principal("SELECT ON "+owners, "SELECT, UPDATE ROW ON "+guards+", `"+path+"`")

	assertDenied := func(db *ydb.Driver, sql string) {
		t.Helper()
		var denied error
		if err := db.Query().Do(ctx, func(ctx context.Context, session query.Session) error {
			denied = session.Exec(ctx, sql)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		ok := ydb.IsOperationError(denied, Ydb.StatusIds_UNAUTHORIZED)
		ydb.IterateByIssues(denied, func(message string, _ Ydb.StatusIds_StatusCode, _ uint32) {
			ok = ok || strings.Contains(message, "status: AccessDenied") ||
				ydb.IsOperationErrorSchemeError(denied) && strings.Contains(message, "you do not have access permissions")
		})
		if !ok {
			t.Fatalf("expected ACL rejection for statement %q, got %v", sql, denied)
		}
	}

	slot := ReplicaSlot{"test", uuid.NewString(), "publisher-0", path + "_owner"}
	uid := uuid.NewString()
	reconciler, _ := NewTransactor(writer, 5*time.Second)
	owner, _ := NewTransactor(reader, 5*time.Second)
	if err := slot.Reconcile(ctx, reconciler, func(context.Context) (string, bool, error) { return uid, false, nil }); err != nil {
		t.Fatal(err)
	}
	claim, err := slot.Bind(ctx, owner, uid)
	if err != nil {
		t.Fatal(err)
	}
	p, err := newPublisher(owner, DefaultConfig(map[string]string{"events": path}), claim, path+"_guard")
	if err != nil {
		t.Fatal(err)
	}
	if err := p.CheckReadiness(ctx); err != nil {
		t.Fatalf("publisher grants did not support readiness: %v", err)
	}
	if err := p.WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
		return tx.Publish("events", testEvent(t))
	}); err != nil {
		t.Fatalf("publisher grants did not support atomic topic publication: %v", err)
	}
	assertDenied(reader, "UPDATE "+owners+" SET generation = 2u WHERE replica = '"+slot.key()+"';")
	assertDenied(writer, "UPSERT INTO "+guards+" (replica, revision) VALUES ('acl-check', 1u);")
	verifyFence(t, ctx, owner, p, uid, 1, 1)
}
