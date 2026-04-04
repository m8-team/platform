package core

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	identityv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/identity/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var ErrNotFound = errors.New("document not found")

type StoredDocument struct {
	ID        string
	TenantID  string
	Payload   []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DocumentStore interface {
	GetDocument(ctx context.Context, table string, id string) (StoredDocument, error)
	UpsertDocument(ctx context.Context, table string, doc StoredDocument) error
	DeleteDocument(ctx context.Context, table string, id string) error
	ListDocuments(ctx context.Context, table string, tenantID string, offset int, limit int) ([]StoredDocument, string, error)
}

type Cache interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

type EventPublisher interface {
	PublishProto(ctx context.Context, topic string, msg proto.Message) error
}

type Validator interface {
	Validate(message proto.Message) error
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

func EncodePageToken(offset int) string {
	if offset <= 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

func DecodePageToken(token string) int {
	if token == "" {
		return 0
	}
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return 0
	}
	offset, err := strconv.Atoi(string(decoded))
	if err != nil || offset < 0 {
		return 0
	}
	return offset
}

func MarshalProto(message proto.Message) ([]byte, error) {
	return protojson.MarshalOptions{UseProtoNames: true}.Marshal(message)
}

func UnmarshalProto(payload []byte, target proto.Message) error {
	return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(payload, target)
}

func LabelsFromMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func Timestamp(value time.Time) *timestamppb.Timestamp {
	if value.IsZero() {
		return nil
	}
	return timestamppb.New(value.UTC())
}

func JSONString(value any) string {
	payload, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(payload)
}
