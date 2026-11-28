package event

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicoptions"
)

func TestIntegrationTopicReadiness(t *testing.T) {
	ctx, db, path := liveDatabase(t)
	createTestTopic(t, ctx, db, path, 4)
	createFixtureTopic(t, ctx, db, path+"_unbounded", topicoptions.CreateWithMinActivePartitions(4))
	for _, tc := range []struct {
		name, path string
		want       error
	}{
		{name: "compatible", path: path},
		{name: "missing bounds", path: path + "_unbounded", want: ErrInvalidTopic},
		{name: "missing topic", path: path + "_missing"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := livePublisher(t, ctx, db, path, DefaultConfig(map[string]string{"events": tc.path}), 10*time.Second)
			err := p.CheckReadiness(ctx)
			if tc.name == "missing topic" {
				if !ydb.IsOperationErrorSchemeError(err) {
					t.Fatalf("expected a metadata request failure, got %v", err)
				}
			} else if !errors.Is(err, tc.want) {
				t.Fatalf("readiness: %v, want %v", err, tc.want)
			}
			stopped, cancel := context.WithCancel(ctx)
			cancel()
			if err := p.CheckReadiness(stopped); !errors.Is(err, context.Canceled) {
				t.Fatalf("canceled readiness: %v", err)
			}
		})
	}
	cfg := DefaultConfig(map[string]string{"events": path})
	cfg.MinTopicRetention = 10 * 365 * 24 * time.Hour
	p := livePublisher(t, ctx, db, path, cfg, 8*time.Second)
	if err := p.CheckReadiness(ctx); !errors.Is(err, ErrInvalidTopic) {
		t.Fatalf("expected retention below minimum: %v", err)
	}
}

func TestRejectDuplicateTopicMappings(t *testing.T) {
	cfg := DefaultConfig(map[string]string{"resource": "/local/events", "alias": "/local/events"})
	if err := cfg.Validate(); !errors.Is(err, ErrInvalidConfiguration) {
		t.Fatalf("duplicate mappings: %v", err)
	}
}

func TestIntegrationFencingSchemaReadiness(t *testing.T) {
	ctx, db, path := liveDatabase(t)
	createTestTopic(t, ctx, db, path, 1)
	p := livePublisher(t, ctx, db, path, DefaultConfig(map[string]string{"events": path}), 8*time.Second)
	for _, tc := range []struct {
		name, columns, suffix string
	}{
		{name: "missing"},
		{name: "wrong key", columns: "replica Utf8 NOT NULL, pod_uid Utf8 NOT NULL, generation Uint64 NOT NULL, PRIMARY KEY (pod_uid)"},
		{name: "wrong type", columns: "replica Utf8 NOT NULL, pod_uid Utf8 NOT NULL, generation Int64 NOT NULL, PRIMARY KEY (replica)"},
		{name: "nullable", columns: "replica Utf8 NOT NULL, pod_uid Utf8, generation Uint64 NOT NULL, PRIMARY KEY (replica)"},
		{name: "ttl", columns: "replica Utf8 NOT NULL, pod_uid Utf8 NOT NULL, generation Uint64 NOT NULL, expires Timestamp, PRIMARY KEY (replica)", suffix: " WITH (TTL = Interval('P1D') ON expires)"},
		{name: "column table", columns: "replica Utf8 NOT NULL, pod_uid Utf8 NOT NULL, generation Uint64 NOT NULL, PRIMARY KEY (replica)", suffix: " WITH (STORE = COLUMN)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, alternate := liveDatabase(t)
			if tc.columns != "" {
				if err := db.Query().Exec(ctx, "CREATE TABLE `"+alternate+"` ("+tc.columns+")"+tc.suffix+";"); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := db.Query().Exec(ctx, "DROP TABLE `"+alternate+"`;"); err != nil {
						t.Error(err)
					}
				})
			}
			binding := p.ownership
			binding.slot.tablePath = alternate
			probe, err := newPublisher(p.owner, p.cfg, binding, p.guard)
			if err != nil {
				t.Fatal(err)
			}
			err = probe.CheckReadiness(ctx)
			if tc.columns == "" {
				if !ydb.IsOperationErrorSchemeError(err) {
					t.Fatalf("expected missing table, got %v", err)
				}
			} else if !errors.Is(err, ErrInvalidConfiguration) {
				t.Fatalf("expected schema rejection, got %v", err)
			}
		})
	}
	probe, err := newPublisher(p.owner, p.cfg, p.ownership, path+"_missing_guard")
	if err != nil {
		t.Fatal(err)
	}
	if err := probe.CheckReadiness(ctx); !ydb.IsOperationErrorSchemeError(err) {
		t.Fatalf("expected missing guard, got %v", err)
	}
	if err := db.Query().Exec(ctx, declareReplicaKey+"UPSERT INTO `"+p.guard+"` (replica, revision) VALUES ($replica, 0u);",
		query.WithParameters(ydb.ParamsBuilder().Param("$replica").Text(p.ownership.slot.key()).Build())); err != nil {
		t.Fatal(err)
	}
	if err := p.CheckReadiness(ctx); !errors.Is(err, ErrReplicaOwnership) {
		t.Fatalf("expected rejection of stored zero guard: %v", err)
	}
}
