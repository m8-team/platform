package eventfeed

import (
	"context"
	"errors"
	"testing"
)

type fakeReceiver struct {
	msg Message
	err error
}

func (r *fakeReceiver) Receive(context.Context) (Message, error) { return r.msg, r.err }
func TestDecodeKafkaAndOutboxEnvelope(t *testing.T) {
	for _, factory := range []func(Receiver) (*Feed, error){NewKafka, NewOutbox} {
		ack := 0
		r := &fakeReceiver{msg: Message{Payload: []byte(`{"event_id":"evt-1","tenant_id":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa","projection_kind":"ACCESS_GRANT","aggregate_id":"11111111-1111-4111-8111-111111111111","projection_revision":3}`), Acknowledge: func(context.Context) error { ack++; return nil }}}
		feed, err := factory(r)
		if err != nil {
			t.Fatal(err)
		}
		delivery, err := feed.Next(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if delivery.Event().Revision != 3 || delivery.Event().Key.Kind != "ACCESS_GRANT" {
			t.Fatal("bad decoded event")
		}
		if ack != 0 {
			t.Fatal("premature ack")
		}
		if err := delivery.Ack(context.Background()); err != nil {
			t.Fatal(err)
		}
		if ack != 1 {
			t.Fatal("no ack")
		}
	}
}
func TestRejectInvalidEnvelopeWithoutAck(t *testing.T) {
	for _, raw := range []string{`{}`, `{"tenant_id":"not-a-tenant"}`, `{"event_id":"x","tenant_id":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa","projection_kind":"ACCESS_GRANT","aggregate_id":"11111111-1111-4111-8111-111111111111","projection_revision":1,"untrusted":true}`} {
		ack := 0
		r := &fakeReceiver{msg: Message{Payload: []byte(raw), Acknowledge: func(context.Context) error { ack++; return nil }}}
		feed, _ := NewKafka(r)
		_, err := feed.Next(context.Background())
		if !errors.Is(err, ErrInvalidEnvelope) {
			t.Fatalf("wanted invalid envelope for %s: %v", raw, err)
		}
		if ack != 0 {
			t.Fatal("ack invalid payload")
		}
	}
}
