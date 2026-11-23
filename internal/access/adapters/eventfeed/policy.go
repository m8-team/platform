package eventfeed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/m8-team/platform/internal/access/app/policyjournal"
	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/domain"
)

// policyEnvelope is the stable Kafka/Outbox wire format for a certified
// tenant PolicyJournal event. Only the journal, never Kafka, is authoritative
// for the mutation and its complete projection job manifest.
type policyEnvelope struct {
	EventID string `json:"event_id"`
	TenantID string `json:"tenant_id"`
	Kind projection.Kind `json:"projection_kind"`
	AggregateID string `json:"aggregate_id"`
	ProjectionRevision int64 `json:"projection_revision"`
	PolicyRevision uint64 `json:"policy_revision"`
}

func EncodePolicyEvent(e policyjournal.Event) ([]byte, error) {
	if err := e.Validate(); err != nil {
		return nil, ErrInvalidEnvelope
	}
	value := policyEnvelope{
		EventID: e.Projection.EventID, TenantID: string(e.Projection.Key.TenantID),
		Kind: e.Projection.Key.Kind, AggregateID: e.Projection.Key.AggregateID,
		ProjectionRevision: e.Projection.Revision, PolicyRevision: e.PolicyRevision,
	}
	encoded, err := json.Marshal(value)
	if err != nil || len(encoded) > maxEventBytes {
		return nil, ErrInvalidEnvelope
	}
	return encoded, nil
}

type PolicyFeed struct{ receiver Receiver }

var _ policyjournal.Feed = (*PolicyFeed)(nil)

func NewPolicyKafka(receiver Receiver) (*PolicyFeed, error) {
	if receiver == nil {
		return nil, ErrInvalidEnvelope
	}
	return &PolicyFeed{receiver: receiver}, nil
}
func NewPolicyOutbox(receiver Receiver) (*PolicyFeed, error) {
	return NewPolicyKafka(receiver)
}

type policyDelivery struct {
	event policyjournal.Event
	ack func(context.Context) error
}
func (d policyDelivery) Event() policyjournal.Event { return d.event }
func (d policyDelivery) Ack(ctx context.Context) error { return d.ack(ctx) }

func(f *PolicyFeed) Next(ctx context.Context)(policyjournal.Delivery,error){
	if f==nil||f.receiver==nil{return nil,ErrInvalidEnvelope}
	if err:=ctx.Err();err!=nil{return nil,err}
	message,err:=f.receiver.Receive(ctx)
	if err!=nil{return nil,err}
	if message.Acknowledge==nil||len(message.Payload)==0||len(message.Payload)>maxEventBytes{
		return nil,ErrInvalidEnvelope
	}
	dec:=json.NewDecoder(bytes.NewReader(message.Payload))
	dec.DisallowUnknownFields()
	var value policyEnvelope
	if err:=dec.Decode(&value);err!=nil{return nil,fmt.Errorf("%w: malformed policy event",ErrInvalidEnvelope)}
	var extra any
	if err:=dec.Decode(&extra);err!=io.EOF{return nil,ErrInvalidEnvelope}
	tenant,err:=domain.NewTenantID(value.TenantID)
	if err!=nil{return nil,ErrInvalidEnvelope}
	event:=policyjournal.Event{PolicyRevision:value.PolicyRevision,Projection:projection.Event{
		EventID:value.EventID,Key:projection.Key{TenantID:tenant,Kind:value.Kind,AggregateID:value.AggregateID},
		Revision:value.ProjectionRevision,
	}}
	if err:=event.Validate();err!=nil{return nil,ErrInvalidEnvelope}
	return policyDelivery{event:event,ack:message.Acknowledge},nil
}
