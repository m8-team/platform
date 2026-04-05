package spicedb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	spicedbv1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	authzed "github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	identityv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/identity/v1"
	ydb "github.com/m8platform/platform/iam/internal/adapter/out/ydb"
	"github.com/m8platform/platform/iam/internal/foundation/config"
	foundationstore "github.com/m8platform/platform/iam/internal/foundation/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

var ErrNotConfigured = errors.New("spicedb endpoint is not configured")
var ErrNotImplemented = errors.New("spicedb runtime integration is pending")
var ErrUnsupportedSubjectType = errors.New("spicedb subject type is not supported")
var ErrUnsupportedResourceType = errors.New("spicedb resource type is not supported")
var ErrUnsupportedPermission = errors.New("spicedb permission is not supported")
var ErrUnsupportedRoleBinding = errors.New("spicedb role binding is not supported for resource type")
var ErrUnsupportedCaveatContext = errors.New("spicedb caveat context requires an explicit caveat schema")

type SyncReport struct {
	GroupMembers  int
	Resources     int
	Bindings      int
	Relationships int
}

type resourceSyncInput struct {
	resource *authzv1.ResourceRef
	bindings []*authzv1.AccessBinding
}

type Client struct {
	cfg config.SpiceDBConfig

	mu       sync.RWMutex
	client   *authzed.Client
	zedToken string
}

func NewClient(cfg config.SpiceDBConfig) *Client {
	return &Client{cfg: cfg}
}

func (c *Client) Close() error {
	if c == nil {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client == nil {
		return nil
	}
	err := c.client.Close()
	c.client = nil
	return err
}

func (c *Client) ApplySchema(ctx context.Context, schema string) error {
	client, err := c.ensureClient()
	if err != nil {
		return err
	}
	if strings.TrimSpace(schema) == "" {
		return errors.New("spicedb schema is empty")
	}

	resp, err := client.WriteSchema(ctx, &spicedbv1.WriteSchemaRequest{
		Schema: schema,
	})
	if err != nil {
		return err
	}
	c.storeToken(resp.GetWrittenAt())
	return nil
}

func (c *Client) ApplySchemaFile(ctx context.Context, path string) error {
	payload, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return c.ApplySchema(ctx, string(payload))
}

func (c *Client) Check(ctx context.Context, req *authzv1.CheckAccessRequest) (*authzv1.AccessCheckResult, error) {
	client, err := c.ensureClient()
	if err != nil {
		return nil, err
	}

	resource, err := objectRefForResource(req.GetResource())
	if err != nil {
		return nil, err
	}
	subject, err := subjectRefForCheck(req.GetSubject())
	if err != nil {
		return nil, err
	}
	permission, err := permissionNameForCheck(req.GetResource(), req.GetPermission())
	if err != nil {
		return nil, err
	}

	resp, err := client.CheckPermission(ctx, &spicedbv1.CheckPermissionRequest{
		Consistency: c.consistency(),
		Resource:    resource,
		Permission:  permission,
		Subject:     subject,
		Context:     req.GetCaveatContext(),
	})
	if err != nil {
		return nil, err
	}

	c.storeToken(resp.GetCheckedAt())
	result := &authzv1.AccessCheckResult{
		Decision:   decisionFromPermissionship(resp.GetPermissionship()),
		Permission: req.GetPermission(),
		CacheHit:   false,
		ZedToken:   resp.GetCheckedAt().GetToken(),
	}
	if info := resp.GetPartialCaveatInfo(); info != nil {
		result.CaveatExpressions = append(result.CaveatExpressions, info.GetMissingRequiredContext()...)
	}
	return result, nil
}

func (c *Client) SyncResource(ctx context.Context, resource *authzv1.ResourceRef, bindings []*authzv1.AccessBinding) error {
	client, err := c.ensureClient()
	if err != nil {
		return err
	}
	if resource == nil {
		return errors.New("spicedb resource sync requires a resource")
	}

	if err := c.deleteResourceBindings(ctx, client, resource); err != nil {
		return err
	}

	updates := make([]*spicedbv1.RelationshipUpdate, 0, estimatedRelationshipCount(resource, bindings))
	if tenantRelationship, ok, err := tenantAnchorRelationship(resource); err != nil {
		return err
	} else if ok {
		updates = append(updates, touchRelationship(tenantRelationship))
	}

	for _, binding := range bindings {
		if binding == nil {
			continue
		}
		if !sameResource(binding.GetResource(), resource) {
			return fmt.Errorf("binding %s does not belong to resource %s/%s", binding.GetBindingId(), resource.GetType().String(), resource.GetId())
		}
		relationships, err := relationshipsForBinding(binding)
		if err != nil {
			return err
		}
		for _, relationship := range relationships {
			updates = append(updates, touchRelationship(relationship))
		}
	}

	if len(updates) == 0 {
		return nil
	}

	resp, err := client.WriteRelationships(ctx, &spicedbv1.WriteRelationshipsRequest{
		Updates: updates,
	})
	if err != nil {
		return err
	}
	c.storeToken(resp.GetWrittenAt())
	return nil
}

func (c *Client) WriteGroupMembership(ctx context.Context, _ string, member *identityv1.GroupMember) error {
	client, err := c.ensureClient()
	if err != nil {
		return err
	}
	if member == nil {
		return errors.New("spicedb group membership is nil")
	}

	relationship, err := relationshipForGroupMember(member)
	if err != nil {
		return err
	}
	resp, err := client.WriteRelationships(ctx, &spicedbv1.WriteRelationshipsRequest{
		Updates: []*spicedbv1.RelationshipUpdate{touchRelationship(relationship)},
	})
	if err != nil {
		return err
	}
	c.storeToken(resp.GetWrittenAt())
	return nil
}

func (c *Client) DeleteGroupMembership(ctx context.Context, _ string, groupID string, subjectType string, subjectID string) error {
	client, err := c.ensureClient()
	if err != nil {
		return err
	}

	subjectObjectType, err := objectTypeForGroupMemberSubject(subjectType)
	if err != nil {
		return err
	}
	resp, err := client.DeleteRelationships(ctx, &spicedbv1.DeleteRelationshipsRequest{
		RelationshipFilter: &spicedbv1.RelationshipFilter{
			ResourceType:       groupObjectType,
			OptionalResourceId: groupID,
			OptionalRelation:   groupMemberRelation,
			OptionalSubjectFilter: &spicedbv1.SubjectFilter{
				SubjectType:       subjectObjectType,
				OptionalSubjectId: subjectID,
			},
		},
	})
	if err != nil {
		return err
	}
	c.storeToken(resp.GetDeletedAt())
	return nil
}

func (c *Client) SyncSnapshot(ctx context.Context, store foundationstore.DocumentStore) (*SyncReport, error) {
	if store == nil {
		return nil, errors.New("spicedb snapshot sync requires a document store")
	}

	groupMembers, err := loadAllProto(ctx, store, ydb.TableGroupMembers, func() *identityv1.GroupMember {
		return &identityv1.GroupMember{}
	})
	if err != nil {
		return nil, err
	}
	bindings, err := loadAllProto(ctx, store, ydb.TableBindingOperations, func() *authzv1.AccessBinding {
		return &authzv1.AccessBinding{}
	})
	if err != nil {
		return nil, err
	}

	report := &SyncReport{}
	for _, member := range groupMembers {
		if err := c.WriteGroupMembership(ctx, "", member); err != nil {
			return nil, err
		}
		report.GroupMembers++
		report.Relationships++
	}

	grouped := make(map[string]*resourceSyncInput, len(bindings))
	keys := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		resource := binding.GetResource()
		if resource == nil {
			continue
		}
		key := resourceKey(resource)
		entry, ok := grouped[key]
		if !ok {
			entry = &resourceSyncInput{resource: resource}
			grouped[key] = entry
			keys = append(keys, key)
		}
		entry.bindings = append(entry.bindings, binding)
	}
	sort.Strings(keys)

	for _, key := range keys {
		entry := grouped[key]
		if err := c.SyncResource(ctx, entry.resource, entry.bindings); err != nil {
			return nil, err
		}
		report.Resources++
		report.Bindings += len(entry.bindings)
		report.Relationships += estimatedRelationshipCount(entry.resource, entry.bindings)
	}

	return report, nil
}

func (c *Client) ensureClient() (*authzed.Client, error) {
	if c == nil {
		return nil, ErrNotConfigured
	}
	if strings.TrimSpace(c.cfg.Endpoint) == "" {
		return nil, ErrNotConfigured
	}

	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()
	if client != nil {
		return client, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		return c.client, nil
	}

	opts, err := c.dialOptions()
	if err != nil {
		return nil, err
	}
	client, err = authzed.NewClient(c.cfg.Endpoint, opts...)
	if err != nil {
		return nil, err
	}
	c.client = client
	return client, nil
}

func (c *Client) dialOptions() ([]grpc.DialOption, error) {
	token := strings.TrimSpace(c.cfg.Token)
	if token == "" {
		token = strings.TrimSpace(c.cfg.PreSharedKey)
	}

	options := make([]grpc.DialOption, 0, 2)
	if c.cfg.Insecure {
		options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if token != "" {
			options = append(options, grpcutil.WithInsecureBearerToken(token))
		}
		return options, nil
	}

	systemCerts, err := grpcutil.WithSystemCerts(grpcutil.VerifyCA)
	if err != nil {
		return nil, err
	}
	options = append(options, systemCerts)
	if token != "" {
		options = append(options, grpcutil.WithBearerToken(token))
	}
	return options, nil
}

func (c *Client) consistency() *spicedbv1.Consistency {
	consistency := strings.TrimSpace(strings.ToLower(c.cfg.Consistency))
	switch consistency {
	case "minimize_latency":
		return &spicedbv1.Consistency{
			Requirement: &spicedbv1.Consistency_MinimizeLatency{MinimizeLatency: true},
		}
	case "fully_consistent":
		return &spicedbv1.Consistency{
			Requirement: &spicedbv1.Consistency_FullyConsistent{FullyConsistent: true},
		}
	case "", "at_least_as_fresh":
		if token := c.currentToken(); token != "" {
			return &spicedbv1.Consistency{
				Requirement: &spicedbv1.Consistency_AtLeastAsFresh{
					AtLeastAsFresh: &spicedbv1.ZedToken{Token: token},
				},
			}
		}
		return &spicedbv1.Consistency{
			Requirement: &spicedbv1.Consistency_FullyConsistent{FullyConsistent: true},
		}
	default:
		return &spicedbv1.Consistency{
			Requirement: &spicedbv1.Consistency_FullyConsistent{FullyConsistent: true},
		}
	}
}

func (c *Client) currentToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.zedToken
}

func (c *Client) storeToken(token *spicedbv1.ZedToken) {
	if c == nil || token == nil || strings.TrimSpace(token.GetToken()) == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.zedToken = token.GetToken()
}

func (c *Client) deleteResourceBindings(ctx context.Context, client *authzed.Client, resource *authzv1.ResourceRef) error {
	objectType, err := objectTypeForResource(resource.GetType())
	if err != nil {
		return err
	}

	for _, relation := range mutableRelationsForResource(resource.GetType()) {
		resp, err := client.DeleteRelationships(ctx, &spicedbv1.DeleteRelationshipsRequest{
			RelationshipFilter: &spicedbv1.RelationshipFilter{
				ResourceType:       objectType,
				OptionalResourceId: resource.GetId(),
				OptionalRelation:   relation,
			},
		})
		if err != nil {
			return err
		}
		c.storeToken(resp.GetDeletedAt())
	}
	return nil
}

func loadAllProto[T proto.Message](ctx context.Context, store foundationstore.DocumentStore, table string, newItem func() T) ([]T, error) {
	pageToken := ""
	items := make([]T, 0)

	for {
		page, nextToken, err := foundationstore.ListProto(ctx, store, table, "", 1000, pageToken, newItem)
		if err != nil {
			return nil, err
		}
		items = append(items, page...)
		if nextToken == "" {
			return items, nil
		}
		pageToken = nextToken
	}
}
