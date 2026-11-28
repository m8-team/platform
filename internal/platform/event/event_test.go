package event

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	eventv1 "github.com/m8-team/platform/internal/proto/m8/platform/event/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func testEvent(t testing.TB) *eventv1.Event {
	t.Helper()
	payload, err := anypb.New(wrapperspb.String("sample"))
	if err != nil {
		t.Fatal(err)
	}
	return &eventv1.Event{
		Id: "34b169b7-3969-47f1-bc60-18f8df84212b", Type: "m8.resource.created.v1",
		Source: "resource-manager", AggregateId: "resource-1", AggregateType: "resource",
		AggregateRevision: 1, OccurredAt: timestamppb.New(time.Unix(1, 0)), Payload: payload,
	}
}

func TestEncodeBatch(t *testing.T) {
	cfg := DefaultConfig(map[string]string{"resource": "/local/events/resource.v1"})
	ctx := context.Background()
	e := testEvent(t)
	next := proto.CloneOf(e)
	next.Id = "44b169b7-3969-47f1-bc60-18f8df84212b"
	msgs, _, err := encodeBatch(ctx, cfg, []*eventv1.Event{e, next})
	if err != nil || len(msgs) != 2 {
		t.Fatalf("batch: %v", err)
	}
	for i, want := range []*eventv1.Event{e, next} {
		data, err := io.ReadAll(msgs[i].Data)
		if err != nil {
			t.Fatal(err)
		}
		got := new(eventv1.Event)
		if err := proto.Unmarshal(data, got); err != nil || !proto.Equal(want, got) {
			t.Fatalf("round trip: %v", err)
		}
		if msgs[i].Key != routingKey(want) || msgs[i].SeqNo != 0 {
			t.Fatal("invalid routing or sequence")
		}
	}
	if msgs, _, err := encodeBatch(ctx, cfg, nil); err != nil || len(msgs) != 0 {
		t.Fatalf("empty batch: %v", err)
	}
}

func TestInvalidEvents(t *testing.T) {
	cfg := DefaultConfig(map[string]string{"resource": "/local/events/resource.v1"})
	tests := map[string]func(*eventv1.Event){
		"empty ID":         func(e *eventv1.Event) { e.Id = "" },
		"uppercase ID":     func(e *eventv1.Event) { e.Id = strings.ToUpper(e.Id) },
		"zero ID":          func(e *eventv1.Event) { e.Id = "00000000-0000-0000-0000-000000000000" },
		"uppercase tenant": func(e *eventv1.Event) { e.TenantId = strings.ToUpper(e.Id) },
		"zero tenant":      func(e *eventv1.Event) { e.TenantId = "00000000-0000-0000-0000-000000000000" },
		"source":           func(e *eventv1.Event) { e.Source = "" },
		"aggregate":        func(e *eventv1.Event) { e.AggregateId = "" },
		"metadata key":     func(e *eventv1.Event) { e.Metadata = map[string]string{strings.Repeat("x", 129): ""} },
		"metadata value":   func(e *eventv1.Event) { e.Metadata = map[string]string{"key": strings.Repeat("x", 1025)} },
		"invalid typed payload": func(e *eventv1.Event) {
			e.Payload = &anypb.Any{TypeUrl: "type.googleapis.com/m8.platform.event.v1.Event"}
		},
		"id":        func(e *eventv1.Event) { e.Id = "bad" },
		"type":      func(e *eventv1.Event) { e.Type = "unversioned" },
		"revision":  func(e *eventv1.Event) { e.AggregateRevision = 0 },
		"tenant":    func(e *eventv1.Event) { e.TenantId = "bad" },
		"time":      func(e *eventv1.Event) { e.OccurredAt = nil },
		"payload":   func(e *eventv1.Event) { e.Payload = nil },
		"malformed": func(e *eventv1.Event) { e.Payload.Value = []byte{0xff} },
		"unknown":   func(e *eventv1.Event) { e.Payload.TypeUrl = "type.googleapis.com/unknown.Message" },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			e := testEvent(t)
			mutate(e)
			if _, _, err := encodeBatch(context.Background(), cfg, []*eventv1.Event{e}); !errors.Is(err, ErrInvalidEvent) {
				t.Fatalf("expected invalid event, got %v", err)
			}
		})
	}
	for _, events := range [][]*eventv1.Event{{nil}, {testEvent(t), testEvent(t)}} {
		if _, _, err := encodeBatch(context.Background(), cfg, events); !errors.Is(err, ErrInvalidEvent) {
			t.Fatalf("expected invalid event, got %v", err)
		}
	}
}

func TestLimitsAndCancellation(t *testing.T) {
	cfg := DefaultConfig(map[string]string{"resource": "/local/events/resource.v1"})
	e := testEvent(t)
	next := proto.CloneOf(e)
	next.Id = "44b169b7-3969-47f1-bc60-18f8df84212b"
	cfg.MaxBatchEvents = 1
	if _, _, err := encodeBatch(context.Background(), cfg, []*eventv1.Event{e, next}); !errors.Is(err, ErrBatchTooLarge) {
		t.Fatal(err)
	}
	cfg.MaxBatchEvents = 2
	cfg.MaxEventBytes = proto.Size(e) - 1
	if _, _, err := encodeBatch(context.Background(), cfg, []*eventv1.Event{e}); !errors.Is(err, ErrEventTooLarge) {
		t.Fatal(err)
	}
	cfg.MaxEventBytes = proto.Size(e)
	cfg.MaxBatchBytes = proto.Size(e) + proto.Size(next) - 1
	if _, _, err := encodeBatch(context.Background(), cfg, []*eventv1.Event{e, next}); !errors.Is(err, ErrBatchTooLarge) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := encodeBatch(ctx, cfg, []*eventv1.Event{e}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestRoutingKey(t *testing.T) {
	a, b := testEvent(t), testEvent(t)
	a.AggregateType, a.AggregateId = "a/b", "c"
	b.AggregateType, b.AggregateId = "a", "b/c"
	if routingKey(a) == routingKey(b) || routingKey(a) != routingKey(proto.CloneOf(a)) {
		t.Fatal("routing key collision or instability")
	}
	b = proto.CloneOf(a)
	b.TenantId = b.Id
	if routingKey(a) == routingKey(b) || routingKey(a)[0] != 0 {
		t.Fatal("tenant must be a distinct length-prefixed component")
	}
}

func TestConfig(t *testing.T) {
	for _, path := range []string{"", "relative", "/local/../bad", "/local/a b", "/"} {
		if err := DefaultConfig(map[string]string{"resource": path}).Validate(); !errors.Is(err, ErrInvalidConfiguration) {
			t.Fatalf("accepted %q: %v", path, err)
		}
	}
	if err := DefaultConfig(map[string]string{"resource": "/local/events/resource.v1"}).Validate(); err != nil {
		t.Fatal(err)
	}
	if _, err := NewPublisher(nil, Config{}, PodOwnership{}, ""); !errors.Is(err, ErrInvalidConfiguration) {
		t.Fatal(err)
	}
}

func BenchmarkEncodeBatch(b *testing.B) {
	for _, n := range []int{1, 10, 50, 100} {
		for _, size := range []int{256, 1024, 16384, 131072} {
			if n*(size+1024) > 4<<20 {
				continue
			}
			b.Run(fmt.Sprintf("batch_%d/payload_%d", n, size), func(b *testing.B) {
				cfg := DefaultConfig(map[string]string{"resource": "/local/events/resource.v1"})
				events := make([]*eventv1.Event, n)
				for i := range events {
					e := testEvent(b)
					e.Id = fmt.Sprintf("34b169b7-3969-47f1-bc60-%012x", i)
					var err error
					e.Payload, err = anypb.New(wrapperspb.String(strings.Repeat("x", size)))
					if err != nil {
						b.Fatal(err)
					}
					events[i] = e
				}
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if _, _, err := encodeBatch(context.Background(), cfg, events); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

// newSampleCreated is an example of an event owner's typed construction path.
// A call passing *wrapperspb.Int64Value fails to compile.
func newSampleCreated(attrs Attributes, payload *wrapperspb.StringValue) (*eventv1.Event, error) {
	attrs.Type = "m8.sample.created.v1"
	return New(attrs, payload)
}

func sampleAttributes(t testing.TB) Attributes {
	e := testEvent(t)
	return Attributes{
		ID: e.Id, Type: e.Type, Source: e.Source, AggregateID: e.AggregateId,
		AggregateType: e.AggregateType, AggregateRevision: e.AggregateRevision,
		OccurredAt: e.OccurredAt.AsTime(),
	}
}

func TestNew(t *testing.T) {
	attrs := sampleAttributes(t)
	attrs.Type = "m8.incorrect.v1"
	attrs.Metadata = map[string]string{"test": "original"}
	payload := wrapperspb.String("original")
	got, err := newSampleCreated(attrs, payload)
	if err != nil {
		t.Fatal(err)
	}
	again, err := newSampleCreated(attrs, payload)
	if err != nil || !proto.Equal(got, again) {
		t.Fatalf("construction changed caller identity: %v", err)
	}
	attrs.Metadata["test"] = "changed"
	payload.Value = "changed"
	unpacked := new(wrapperspb.StringValue)
	if err := got.Payload.UnmarshalTo(unpacked); err != nil || unpacked.Value != "original" ||
		got.Id != attrs.ID || got.Metadata["test"] != "original" || got.Type != "m8.sample.created.v1" {
		t.Fatalf("invalid typed construction: %v", err)
	}
	for _, invalid := range []proto.Message{nil, (*wrapperspb.StringValue)(nil), &eventv1.Event{}} {
		if _, err := New(attrs, invalid); !errors.Is(err, ErrInvalidEvent) {
			t.Fatalf("accepted invalid payload: %v", err)
		}
	}
	for _, mutate := range []func(*Attributes){
		func(a *Attributes) { a.OccurredAt = time.Time{} },
		func(a *Attributes) { a.OccurredAt = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC) },
		func(a *Attributes) { a.AggregateRevision = 0 },
		func(a *Attributes) { a.Source = "" },
	} {
		a := attrs
		mutate(&a)
		if _, err := New(a, payload); !errors.Is(err, ErrInvalidEvent) {
			t.Fatalf("accepted invalid attributes: %v", err)
		}
	}
}

func TestOccurredAtValidation(t *testing.T) {
	for _, tc := range []struct {
		name  string
		stamp *timestamppb.Timestamp
		valid bool
	}{
		{name: "absent"},
		{name: "invalid", stamp: &timestamppb.Timestamp{Seconds: 253402300800}},
		{name: "Go zero time", stamp: timestamppb.New(time.Time{})},
		{name: "Unix epoch", stamp: timestamppb.New(time.Unix(0, 0)), valid: true},
		{name: "historical", stamp: timestamppb.New(time.Date(1850, 1, 1, 0, 0, 0, 0, time.UTC)), valid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := testEvent(t)
			e.OccurredAt = tc.stamp
			_, _, err := encodeBatch(context.Background(), DefaultConfig(nil), []*eventv1.Event{e})
			if (err == nil) != tc.valid || err != nil && !errors.Is(err, ErrInvalidEvent) {
				t.Fatalf("publication validation: %v", err)
			}
			// Attributes always supplies a time.Time, so absence is represented by its zero value.
			attrs := sampleAttributes(t)
			attrs.OccurredAt = time.Time{}
			if tc.stamp != nil {
				attrs.OccurredAt = tc.stamp.AsTime()
			}
			_, err = newSampleCreated(attrs, wrapperspb.String("sample"))
			if (err == nil) != tc.valid || err != nil && !errors.Is(err, ErrInvalidEvent) {
				t.Fatalf("constructor validation: %v", err)
			}
		})
	}
}
