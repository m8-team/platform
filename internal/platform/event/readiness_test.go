package event

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3"
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
}

func TestRejectDuplicateTopicMappings(t *testing.T) {
	cfg := DefaultConfig(map[string]string{"resource": "/local/events", "alias": "/local/events"})
	if err := cfg.Validate(); !errors.Is(err, ErrInvalidConfiguration) {
		t.Fatalf("duplicate mappings: %v", err)
	}
}
