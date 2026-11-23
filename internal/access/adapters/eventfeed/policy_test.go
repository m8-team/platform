package eventfeed

import (
	"context"
	"errors"
	"testing"

	"github.com/m8-team/platform/internal/access/app/policyjournal"
	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/domain"
)

func testPolicyEvent(t *testing.T)policyjournal.Event{
	t.Helper()
	tenant,err:=domain.NewTenantID("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	if err!=nil{t.Fatal(err)}
	return policyjournal.Event{
		PolicyRevision:11,Projection:projection.Event{
			EventID:"profile/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/11111111-1111-4111-8111-111111111111/2",
			Key:projection.Key{TenantID:tenant,Kind:projection.KindProfileAssignment,
				AggregateID:"11111111-1111-4111-8111-111111111111"},Revision:2,
		},
	}
}
func TestProfilePolicyKafkaEnvelopeRoundTripAndDeferredAck(t *testing.T){
	want:=testPolicyEvent(t)
	encoded,err:=EncodePolicyEvent(want)
	if err!=nil{t.Fatal(err)}
	ack:=0
	receiver:=&fakeReceiver{msg:Message{Payload:encoded,Acknowledge:func(context.Context)error{ack++;return nil}}}
	feed,err:=NewPolicyKafka(receiver)
	if err!=nil{t.Fatal(err)}
	delivery,err:=feed.Next(context.Background())
	if err!=nil{t.Fatal(err)}
	if got:=delivery.Event();got!=want{t.Fatalf("changed PolicyRevision or aggregate identity %+v",got)}
	if ack!=0{t.Fatal("ACK before policy graph certification")}
	if err:=delivery.Ack(context.Background());err!=nil||ack!=1{t.Fatalf("missing broker ACK: %v",err)}
}
func TestProfilePolicyFeedRejectsBadOrUntrustedEnvelope(t *testing.T){
	valid,err:=EncodePolicyEvent(testPolicyEvent(t))
	if err!=nil{t.Fatal(err)}
	cases:=[]string{
		"{}",
		"not-json",
		string(valid[:len(valid)-1]),
		string(valid[:len(valid)-1]) + `,"tenant_header":"forged"}`,
		`{"event_id":"evt","tenant_id":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa","projection_kind":"PROFILE_ASSIGNMENT","aggregate_id":"11111111-1111-4111-8111-111111111111","projection_revision":2,"policy_revision":0}`,
	}
	for _,payload:=range cases{
		ack:=0
		feed,_:=NewPolicyKafka(&fakeReceiver{msg:Message{Payload:[]byte(payload),Acknowledge:func(context.Context)error{ack++;return nil}}})
		if _,err:=feed.Next(context.Background());!errors.Is(err,ErrInvalidEnvelope){
			t.Fatalf("untrusted payload was accepted (%q): %v",payload,err)
		}
		if ack!=0{t.Fatal("ACKed an invalid policy event")}
	}
	if _,err:=NewPolicyKafka(nil);!errors.Is(err,ErrInvalidEnvelope){t.Fatal(err)}
}
