package policyjournal

import (
    "context"
    "errors"

    "github.com/m8-team/platform/internal/access/app/projection"
)

var ErrInvalidPolicyDelivery=errors.New("invalid policy projection event or delivery")

// Event is emitted by the SAME authoritative Access/YDB transaction as the
// policy Head and all fanout Jobs. A role change has one policyRevision and
// one per-binding projection Event with its own per-key revision.
type Event struct {
    PolicyRevision uint64
    Projection projection.Event
}
func (e Event)Validate()error{
    if e.PolicyRevision==0 || e.Projection.Validate()!=nil || (e.Projection.Key.Kind!=projection.KindGrant && e.Projection.Key.Kind!=projection.KindRoleBinding){
        return ErrInvalidPolicyDelivery
    }
    return nil
}

// JobCommitter verifies the certified YDB checkpoint and job manifest after
// the underlying projector has finished the official SpiceDB SDK write.
type JobCommitter interface {
    CommitPolicyJob(context.Context,uint64,projection.Key)error
}
type Projector interface {
    Process(context.Context,projection.Event)error
}
type Worker struct {
    project Projector
    journal JobCommitter
}
func NewWorker(project Projector,journal JobCommitter)(*Worker,error){
    if project==nil || journal==nil{return nil,ErrInvalidPolicyDelivery}
    return &Worker{project:project,journal:journal},nil
}
func (w *Worker)Process(ctx context.Context,event Event)error{
    if w==nil || w.project==nil || w.journal==nil{return ErrInvalidPolicyDelivery}
    if err:=event.Validate();err!=nil{return err}
    if err:=w.project.Process(ctx,event.Projection);err!=nil{return err}
    // ACK is forbidden until this second YDB transaction confirms the Access
    // manifest job from the current lease-fenced graph checkpoint.
    return w.journal.CommitPolicyJob(ctx,event.PolicyRevision,event.Projection.Key)
}
type Delivery interface {
    Event()Event
    Ack(context.Context)error
}
type Feed interface{
    Next(context.Context)(Delivery,error)
}
func (w *Worker)Run(ctx context.Context,feed Feed)error{
    if w==nil || feed==nil{return ErrInvalidPolicyDelivery}
    for {
        if err:=ctx.Err();err!=nil{return err}
        msg,err:=feed.Next(ctx)
        if err!=nil{return err}
        if msg==nil{return ErrInvalidPolicyDelivery}
        if err:=w.Process(ctx,msg.Event());err!=nil{return err}
        if err:=msg.Ack(ctx);err!=nil{return err}
    }
}
