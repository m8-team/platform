package contracts

import (
	"context"
	"time"

	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	identityv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/identity/v1"
	"google.golang.org/protobuf/proto"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

type EventPublisher interface {
	PublishProto(ctx context.Context, topic string, msg proto.Message) error
}

type WorkflowStarter interface {
	StartWorkflow(ctx context.Context, workflowName string, workflowID string, input any) (string, error)
}

type AuthorizationRuntime interface {
	Check(ctx context.Context, req *authzv1.CheckAccessRequest) (*authzv1.AccessCheckResult, error)
	SyncResource(ctx context.Context, resource *authzv1.ResourceRef, bindings []*authzv1.AccessBinding) error
	WriteGroupMembership(ctx context.Context, tenantID string, member *identityv1.GroupMember) error
	DeleteGroupMembership(ctx context.Context, tenantID string, groupID string, subjectType string, subjectID string) error
}

type KeycloakClient interface {
	CreateConfidentialClient(ctx context.Context, tenantID string, clientID string, displayName string, serviceAccountsEnabled bool) (string, error)
	RotateClientSecret(ctx context.Context, clientID string) (string, string, error)
}
