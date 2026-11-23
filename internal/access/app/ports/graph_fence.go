package ports

import (
	"context"
	"errors"
	"strings"

	"github.com/m8-team/platform/internal/access/domain"
)

var ErrInvalidGraphFence = errors.New("invalid access projection graph fence")

// GraphFence is a monotonically increasing write generation for one Access
// projection aggregate. Only an authoritative distributed lease may issue it.
type GraphFence struct {
	TenantID    domain.TenantID
	Kind        string
	AggregateID string
	Epoch       uint64
}

func (f GraphFence) Validate() error {
	if err := f.TenantID.Validate(); err != nil {
		return err
	}
	if (f.Kind != "ROLE_BINDING" && f.Kind != "ACCESS_GRANT" && f.Kind != "RM_ORGANIZATION" && f.Kind != "RM_WORKSPACE" && f.Kind != "RM_ENVIRONMENT" && f.Kind != "RM_WORKLOAD_SERVICE") || strings.TrimSpace(f.AggregateID) == "" || f.Epoch == 0 {
		return ErrInvalidGraphFence
	}
	return nil
}

// GraphFenceInstaller atomically replaces an older graph guard with the current
// generation. Implementations must reject attempts to install an older epoch.
type GraphFenceInstaller interface {
	EstablishFence(context.Context, GraphFence) (string, error)
}

type graphFenceContextKey struct{}

// WithGraphFence attaches a server-issued fencing generation to graph writes.
// It must never be populated from caller-controlled HTTP or Kafka metadata.
func WithGraphFence(ctx context.Context, fence GraphFence) (context.Context, error) {
	if err := fence.Validate(); err != nil {
		return nil, err
	}
	return context.WithValue(ctx, graphFenceContextKey{}, fence), nil
}

func GraphFenceFromContext(ctx context.Context) (GraphFence, bool) {
	f, ok := ctx.Value(graphFenceContextKey{}).(GraphFence)
	return f, ok
}

// GraphProjectionReader reads the actual graph tuples owned by one aggregate.
// It is mandatory for reconciliation after checkpoint/graph commit ambiguity:
// a checkpoint alone cannot list tuples written before a failed YDB commit.
type GraphProjectionReader interface {
	ReadProjectionEdges(ctx context.Context, fence GraphFence) ([]domain.Relationship, error)
}

// OwnedObject identifies the only graph node a fenced writer may mutate.
// RM marker ownership is separate from Access grants and role bindings.
func (f GraphFence) OwnedObject() (string,string,error) {
    if err:=f.Validate();err!=nil{return "","",err}
    switch f.Kind {
    case "ROLE_BINDING":return "access_binding",f.AggregateID,nil
    case "ACCESS_GRANT":return "access_grant",f.AggregateID,nil
    case "RM_ORGANIZATION":return "rm_projection","organization:"+f.AggregateID,nil
    case "RM_WORKSPACE":return "rm_projection","workspace:"+f.AggregateID,nil
    case "RM_ENVIRONMENT":return "rm_projection","environment:"+f.AggregateID,nil
    case "RM_WORKLOAD_SERVICE":return "rm_projection","workload_service:"+f.AggregateID,nil
    default:return "","",ErrInvalidGraphFence
    }
}
