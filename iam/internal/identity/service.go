package identity

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	eventsv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/events/v1"
	identityv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/identity/v1"
	"github.com/m8platform/platform/iam/internal/config"
	"github.com/m8platform/platform/iam/internal/core"
	"github.com/m8platform/platform/iam/internal/storage/ydb"
	"github.com/m8platform/platform/iam/internal/temporalx"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type Service struct {
	identityv1.UnimplementedIdentityServiceServer
	identityv1.UnimplementedOAuthFacadeServiceServer

	store     core.DocumentStore
	publisher core.EventPublisher
	workflows core.WorkflowStarter
	runtime   core.AuthorizationRuntime
	keycloak  core.KeycloakClient
	logger    *zap.Logger
	now       func() time.Time
	topics    config.TopicsConfig
}

func NewService(store core.DocumentStore, publisher core.EventPublisher, workflows core.WorkflowStarter, runtime core.AuthorizationRuntime, keycloak core.KeycloakClient, logger *zap.Logger, cfg config.Config) *Service {
	return &Service{
		store:     store,
		publisher: publisher,
		workflows: workflows,
		runtime:   runtime,
		keycloak:  keycloak,
		logger:    logger,
		now:       time.Now,
		topics:    cfg.Topics,
	}
}

func (s *Service) GetUser(ctx context.Context, req *identityv1.GetUserRequest) (*identityv1.User, error) {
	user := &identityv1.User{}
	if err := core.LoadProto(ctx, s.store, ydb.TableUsers, req.GetUserId(), user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) ListUsers(ctx context.Context, req *identityv1.ListUsersRequest) (*identityv1.ListUsersResponse, error) {
	users, next, err := core.ListProto(ctx, s.store, ydb.TableUsers, req.GetTenantId(), int(req.GetPageSize()), req.GetPageToken(), func() *identityv1.User {
		return &identityv1.User{}
	})
	if err != nil {
		return nil, err
	}
	if query := strings.TrimSpace(strings.ToLower(req.GetQuery())); query != "" {
		filtered := make([]*identityv1.User, 0, len(users))
		for _, user := range users {
			if strings.Contains(strings.ToLower(user.GetDisplayName()), query) || strings.Contains(strings.ToLower(user.GetPrimaryEmail()), query) {
				filtered = append(filtered, user)
			}
		}
		users = filtered
	}
	return &identityv1.ListUsersResponse{Users: users, NextPageToken: next}, nil
}

func (s *Service) CreateUser(ctx context.Context, req *identityv1.CreateUserRequest) (*identityv1.User, error) {
	now := s.now()
	user := &identityv1.User{
		UserId:       fmt.Sprintf("user-%d", now.UnixNano()),
		TenantId:     req.GetTenantId(),
		PrimaryEmail: strings.ToLower(req.GetPrimaryEmail()),
		DisplayName:  req.GetDisplayName(),
		State:        identityv1.UserState_USER_STATE_ACTIVE,
		Labels:       core.LabelsFromMap(req.GetLabels()),
		CreatedAt:    core.Timestamp(now),
		UpdatedAt:    core.Timestamp(now),
	}
	if err := core.SaveProto(ctx, s.store, ydb.TableUsers, user.GetUserId(), user.GetTenantId(), user, now); err != nil {
		return nil, err
	}
	operation := core.NewOperation(now, user.GetTenantId(), "create_user", "user", user.GetUserId())
	if err := core.PersistOperation(ctx, s.store, operation, now); err != nil {
		return nil, err
	}
	audit := core.NewAuditEvent(now, user.GetTenantId(), "user.created", req.GetPerformedBy(), operation.GetOperationId(), "create user")
	if err := core.PersistAuditEvent(ctx, s.store, audit, now); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) UpdateUser(ctx context.Context, req *identityv1.UpdateUserRequest) (*identityv1.User, error) {
	current := &identityv1.User{}
	if err := core.LoadProto(ctx, s.store, ydb.TableUsers, req.GetUser().GetUserId(), current); err != nil {
		return nil, err
	}
	applyUserMask(current, req.GetUser(), req.GetUpdateMask())
	now := s.now()
	current.UpdatedAt = core.Timestamp(now)
	if err := core.SaveProto(ctx, s.store, ydb.TableUsers, current.GetUserId(), current.GetTenantId(), current, now); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) DisableUser(ctx context.Context, req *identityv1.DisableUserRequest) (*identityv1.MutateIdentityResponse, error) {
	user := &identityv1.User{}
	if err := core.LoadProto(ctx, s.store, ydb.TableUsers, req.GetUserId(), user); err != nil {
		return nil, err
	}
	now := s.now()
	user.State = identityv1.UserState_USER_STATE_DISABLED
	user.UpdatedAt = core.Timestamp(now)
	if err := core.SaveProto(ctx, s.store, ydb.TableUsers, user.GetUserId(), user.GetTenantId(), user, now); err != nil {
		return nil, err
	}
	operation := core.NewOperation(now, user.GetTenantId(), "disable_user", "user", user.GetUserId())
	if err := core.PersistOperation(ctx, s.store, operation, now); err != nil {
		return nil, err
	}
	audit := core.NewAuditEvent(now, user.GetTenantId(), "user.disabled", req.GetPerformedBy(), operation.GetOperationId(), req.GetReason())
	if err := core.PersistAuditEvent(ctx, s.store, audit, now); err != nil {
		return nil, err
	}
	if err := s.publisher.PublishProto(ctx, s.topics.IdentityUsers, &eventsv1.UserDisabled{
		Meta: &eventsv1.EventMeta{
			EventId:       operation.GetOperationId(),
			OccurredAt:    core.Timestamp(now),
			CorrelationId: operation.GetOperationId(),
			TenantId:      user.GetTenantId(),
		},
		UserId: user.GetUserId(),
		Reason: req.GetReason(),
	}); err != nil {
		return nil, err
	}
	return &identityv1.MutateIdentityResponse{OperationId: operation.GetOperationId()}, nil
}

func (s *Service) GetTenant(ctx context.Context, req *identityv1.GetTenantRequest) (*identityv1.Tenant, error) {
	tenant := &identityv1.Tenant{}
	if err := core.LoadProto(ctx, s.store, ydb.TableTenants, req.GetTenantId(), tenant); err != nil {
		return nil, err
	}
	return tenant, nil
}

func (s *Service) ListTenants(ctx context.Context, req *identityv1.ListTenantsRequest) (*identityv1.ListTenantsResponse, error) {
	tenants, next, err := core.ListProto(ctx, s.store, ydb.TableTenants, "", int(req.GetPageSize()), req.GetPageToken(), func() *identityv1.Tenant {
		return &identityv1.Tenant{}
	})
	if err != nil {
		return nil, err
	}
	if query := strings.TrimSpace(strings.ToLower(req.GetQuery())); query != "" {
		filtered := make([]*identityv1.Tenant, 0, len(tenants))
		for _, tenant := range tenants {
			if strings.Contains(strings.ToLower(tenant.GetDisplayName()), query) || strings.Contains(strings.ToLower(tenant.GetTenantId()), query) {
				filtered = append(filtered, tenant)
			}
		}
		tenants = filtered
	}
	return &identityv1.ListTenantsResponse{Tenants: tenants, NextPageToken: next}, nil
}

func (s *Service) CreateTenant(ctx context.Context, req *identityv1.CreateTenantRequest) (*identityv1.Tenant, error) {
	now := s.now()
	tenant := &identityv1.Tenant{
		TenantId:    req.GetTenantId(),
		DisplayName: req.GetDisplayName(),
		ExternalRef: req.GetExternalRef(),
		Labels:      core.LabelsFromMap(req.GetLabels()),
		CreatedAt:   core.Timestamp(now),
		UpdatedAt:   core.Timestamp(now),
	}
	if err := core.SaveProto(ctx, s.store, ydb.TableTenants, tenant.GetTenantId(), tenant.GetTenantId(), tenant, now); err != nil {
		return nil, err
	}
	operation := core.NewOperation(now, tenant.GetTenantId(), "create_tenant", "tenant", tenant.GetTenantId())
	if err := core.PersistOperation(ctx, s.store, operation, now); err != nil {
		return nil, err
	}
	if err := core.PersistAuditEvent(ctx, s.store, core.NewAuditEvent(now, tenant.GetTenantId(), "tenant.created", req.GetPerformedBy(), operation.GetOperationId(), "create tenant"), now); err != nil {
		return nil, err
	}
	if err := s.publisher.PublishProto(ctx, s.topics.IdentityMemberships, &eventsv1.TenantCreated{
		Meta: &eventsv1.EventMeta{
			EventId:       operation.GetOperationId(),
			OccurredAt:    core.Timestamp(now),
			CorrelationId: operation.GetOperationId(),
			TenantId:      tenant.GetTenantId(),
		},
		Tenant: tenant,
	}); err != nil {
		return nil, err
	}
	return tenant, nil
}

func (s *Service) UpdateTenant(ctx context.Context, req *identityv1.UpdateTenantRequest) (*identityv1.Tenant, error) {
	current := &identityv1.Tenant{}
	if err := core.LoadProto(ctx, s.store, ydb.TableTenants, req.GetTenant().GetTenantId(), current); err != nil {
		return nil, err
	}
	applyTenantMask(current, req.GetTenant(), req.GetUpdateMask())
	now := s.now()
	current.UpdatedAt = core.Timestamp(now)
	if err := core.SaveProto(ctx, s.store, ydb.TableTenants, current.GetTenantId(), current.GetTenantId(), current, now); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) ListMemberships(ctx context.Context, req *identityv1.ListMembershipsRequest) (*identityv1.ListMembershipsResponse, error) {
	memberships, next, err := core.ListProto(ctx, s.store, ydb.TableMemberships, req.GetTenantId(), int(req.GetPageSize()), req.GetPageToken(), func() *identityv1.Membership {
		return &identityv1.Membership{}
	})
	if err != nil {
		return nil, err
	}
	if req.GetUserId() != "" {
		filtered := make([]*identityv1.Membership, 0, len(memberships))
		for _, membership := range memberships {
			if membership.GetUserId() == req.GetUserId() {
				filtered = append(filtered, membership)
			}
		}
		memberships = filtered
	}
	return &identityv1.ListMembershipsResponse{Memberships: memberships, NextPageToken: next}, nil
}

func (s *Service) CreateMembership(ctx context.Context, req *identityv1.CreateMembershipRequest) (*identityv1.Membership, error) {
	now := s.now()
	membership := &identityv1.Membership{
		MembershipId: req.GetMembershipId(),
		TenantId:     req.GetTenantId(),
		UserId:       req.GetUserId(),
		RoleIds:      slices.Clone(req.GetRoleIds()),
		State:        identityv1.MembershipState_MEMBERSHIP_STATE_ACTIVE,
		CreatedAt:    core.Timestamp(now),
		UpdatedAt:    core.Timestamp(now),
	}
	if err := core.SaveProto(ctx, s.store, ydb.TableMemberships, membership.GetMembershipId(), membership.GetTenantId(), membership, now); err != nil {
		return nil, err
	}
	if err := s.publisher.PublishProto(ctx, s.topics.IdentityMemberships, &eventsv1.MembershipCreated{
		Meta: &eventsv1.EventMeta{
			EventId:       req.GetRequestId(),
			OccurredAt:    core.Timestamp(now),
			CorrelationId: req.GetRequestId(),
			TenantId:      membership.GetTenantId(),
		},
		Membership: membership,
	}); err != nil {
		return nil, err
	}
	return membership, nil
}

func (s *Service) DeleteMembership(ctx context.Context, req *identityv1.DeleteMembershipRequest) (*identityv1.MutateIdentityResponse, error) {
	membership := &identityv1.Membership{}
	if err := core.LoadProto(ctx, s.store, ydb.TableMemberships, req.GetMembershipId(), membership); err != nil {
		return nil, err
	}
	if err := s.store.DeleteDocument(ctx, ydb.TableMemberships, req.GetMembershipId()); err != nil {
		return nil, err
	}
	now := s.now()
	if err := s.publisher.PublishProto(ctx, s.topics.IdentityMemberships, &eventsv1.MembershipDeleted{
		Meta: &eventsv1.EventMeta{
			EventId:       req.GetRequestId(),
			OccurredAt:    core.Timestamp(now),
			CorrelationId: req.GetRequestId(),
			TenantId:      membership.GetTenantId(),
		},
		MembershipId: membership.GetMembershipId(),
		UserId:       membership.GetUserId(),
	}); err != nil {
		return nil, err
	}
	return &identityv1.MutateIdentityResponse{OperationId: req.GetRequestId()}, nil
}

func (s *Service) GetGroup(ctx context.Context, req *identityv1.GetGroupRequest) (*identityv1.Group, error) {
	group := &identityv1.Group{}
	if err := core.LoadProto(ctx, s.store, ydb.TableGroups, req.GetGroupId(), group); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *Service) ListGroups(ctx context.Context, req *identityv1.ListGroupsRequest) (*identityv1.ListGroupsResponse, error) {
	groups, next, err := core.ListProto(ctx, s.store, ydb.TableGroups, req.GetTenantId(), int(req.GetPageSize()), req.GetPageToken(), func() *identityv1.Group {
		return &identityv1.Group{}
	})
	if err != nil {
		return nil, err
	}
	if query := strings.TrimSpace(strings.ToLower(req.GetQuery())); query != "" {
		filtered := make([]*identityv1.Group, 0, len(groups))
		for _, group := range groups {
			if strings.Contains(strings.ToLower(group.GetDisplayName()), query) {
				filtered = append(filtered, group)
			}
		}
		groups = filtered
	}
	return &identityv1.ListGroupsResponse{Groups: groups, NextPageToken: next}, nil
}

func (s *Service) CreateGroup(ctx context.Context, req *identityv1.CreateGroupRequest) (*identityv1.Group, error) {
	now := s.now()
	group := &identityv1.Group{
		GroupId:     req.GetGroupId(),
		TenantId:    req.GetTenantId(),
		DisplayName: req.GetDisplayName(),
		Description: req.GetDescription(),
		Labels:      core.LabelsFromMap(req.GetLabels()),
		CreatedAt:   core.Timestamp(now),
		UpdatedAt:   core.Timestamp(now),
	}
	if err := core.SaveProto(ctx, s.store, ydb.TableGroups, group.GetGroupId(), group.GetTenantId(), group, now); err != nil {
		return nil, err
	}
	if err := s.publisher.PublishProto(ctx, s.topics.IdentityGroups, &eventsv1.GroupCreated{
		Meta: &eventsv1.EventMeta{
			EventId:       req.GetRequestId(),
			OccurredAt:    core.Timestamp(now),
			CorrelationId: req.GetRequestId(),
			TenantId:      group.GetTenantId(),
		},
		Group: group,
	}); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *Service) UpdateGroup(ctx context.Context, req *identityv1.UpdateGroupRequest) (*identityv1.Group, error) {
	current := &identityv1.Group{}
	if err := core.LoadProto(ctx, s.store, ydb.TableGroups, req.GetGroup().GetGroupId(), current); err != nil {
		return nil, err
	}
	applyGroupMask(current, req.GetGroup(), req.GetUpdateMask())
	now := s.now()
	current.UpdatedAt = core.Timestamp(now)
	if err := core.SaveProto(ctx, s.store, ydb.TableGroups, current.GetGroupId(), current.GetTenantId(), current, now); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) DeleteGroup(ctx context.Context, req *identityv1.DeleteGroupRequest) (*identityv1.MutateIdentityResponse, error) {
	group := &identityv1.Group{}
	if err := core.LoadProto(ctx, s.store, ydb.TableGroups, req.GetGroupId(), group); err != nil {
		return nil, err
	}
	if err := s.store.DeleteDocument(ctx, ydb.TableGroups, req.GetGroupId()); err != nil {
		return nil, err
	}
	return &identityv1.MutateIdentityResponse{OperationId: req.GetRequestId()}, nil
}

func (s *Service) AddGroupMember(ctx context.Context, req *identityv1.AddGroupMemberRequest) (*identityv1.GroupMember, error) {
	now := s.now()
	member := &identityv1.GroupMember{
		GroupId:     req.GetGroupId(),
		SubjectId:   req.GetSubjectId(),
		SubjectType: req.GetSubjectType(),
		CreatedAt:   core.Timestamp(now),
	}
	memberID := groupMemberID(req.GetGroupId(), req.GetSubjectType(), req.GetSubjectId())
	group := &identityv1.Group{}
	if err := core.LoadProto(ctx, s.store, ydb.TableGroups, req.GetGroupId(), group); err != nil {
		return nil, err
	}
	if err := core.SaveProto(ctx, s.store, ydb.TableGroupMembers, memberID, group.GetTenantId(), member, now); err != nil {
		return nil, err
	}
	if s.runtime != nil {
		if err := s.runtime.WriteGroupMembership(ctx, group.GetTenantId(), member); err != nil {
			return nil, err
		}
	}
	if err := s.publisher.PublishProto(ctx, s.topics.IdentityGroups, &eventsv1.GroupMemberAdded{
		Meta: &eventsv1.EventMeta{
			EventId:       req.GetRequestId(),
			OccurredAt:    core.Timestamp(now),
			CorrelationId: req.GetRequestId(),
			TenantId:      group.GetTenantId(),
		},
		Member: member,
	}); err != nil {
		return nil, err
	}
	return member, nil
}

func (s *Service) RemoveGroupMember(ctx context.Context, req *identityv1.RemoveGroupMemberRequest) (*identityv1.MutateIdentityResponse, error) {
	group := &identityv1.Group{}
	if err := core.LoadProto(ctx, s.store, ydb.TableGroups, req.GetGroupId(), group); err != nil {
		return nil, err
	}
	if err := s.store.DeleteDocument(ctx, ydb.TableGroupMembers, groupMemberID(req.GetGroupId(), req.GetSubjectType(), req.GetSubjectId())); err != nil {
		return nil, err
	}
	if s.runtime != nil {
		if err := s.runtime.DeleteGroupMembership(ctx, group.GetTenantId(), req.GetGroupId(), req.GetSubjectType(), req.GetSubjectId()); err != nil {
			return nil, err
		}
	}
	now := s.now()
	if err := s.publisher.PublishProto(ctx, s.topics.IdentityGroups, &eventsv1.GroupMemberRemoved{
		Meta: &eventsv1.EventMeta{
			EventId:       req.GetRequestId(),
			OccurredAt:    core.Timestamp(now),
			CorrelationId: req.GetRequestId(),
			TenantId:      group.GetTenantId(),
		},
		GroupId:     req.GetGroupId(),
		SubjectId:   req.GetSubjectId(),
		SubjectType: req.GetSubjectType(),
	}); err != nil {
		return nil, err
	}
	return &identityv1.MutateIdentityResponse{OperationId: req.GetRequestId()}, nil
}

func (s *Service) GetServiceAccount(ctx context.Context, req *identityv1.GetServiceAccountRequest) (*identityv1.ServiceAccount, error) {
	account := &identityv1.ServiceAccount{}
	if err := core.LoadProto(ctx, s.store, ydb.TableServiceAccounts, req.GetServiceAccountId(), account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *Service) ListServiceAccounts(ctx context.Context, req *identityv1.ListServiceAccountsRequest) (*identityv1.ListServiceAccountsResponse, error) {
	accounts, next, err := core.ListProto(ctx, s.store, ydb.TableServiceAccounts, req.GetTenantId(), int(req.GetPageSize()), req.GetPageToken(), func() *identityv1.ServiceAccount {
		return &identityv1.ServiceAccount{}
	})
	if err != nil {
		return nil, err
	}
	return &identityv1.ListServiceAccountsResponse{ServiceAccounts: accounts, NextPageToken: next}, nil
}

func (s *Service) CreateServiceAccount(ctx context.Context, req *identityv1.CreateServiceAccountRequest) (*identityv1.ServiceAccount, error) {
	now := s.now()
	operationID := fmt.Sprintf("op-sa-%d", now.UnixNano())
	account := &identityv1.ServiceAccount{
		ServiceAccountId: req.GetServiceAccountId(),
		TenantId:         req.GetTenantId(),
		DisplayName:      req.GetDisplayName(),
		Description:      req.GetDescription(),
		Disabled:         false,
		OperationId:      operationID,
		CreatedAt:        core.Timestamp(now),
		UpdatedAt:        core.Timestamp(now),
	}
	if s.keycloak != nil {
		if keycloakClientID, err := s.keycloak.CreateConfidentialClient(ctx, req.GetTenantId(), req.GetServiceAccountId(), req.GetDisplayName(), true); err == nil {
			account.KeycloakClientId = keycloakClientID
		} else {
			s.logger.Warn("keycloak client creation skipped", zap.Error(err))
		}
	}
	if err := core.SaveProto(ctx, s.store, ydb.TableServiceAccounts, account.GetServiceAccountId(), account.GetTenantId(), account, now); err != nil {
		return nil, err
	}
	if s.workflows != nil {
		if _, err := s.workflows.StartWorkflow(ctx, temporalx.CreateServiceAccountWorkflowName, operationID, temporalx.CreateServiceAccountInput{
			ServiceAccountID: account.GetServiceAccountId(),
			TenantID:         account.GetTenantId(),
			DisplayName:      account.GetDisplayName(),
			Description:      account.GetDescription(),
			RequestedBy:      req.GetPerformedBy(),
		}); err != nil {
			s.logger.Warn("create service account workflow start failed", zap.Error(err))
		}
	}
	if err := s.publisher.PublishProto(ctx, s.topics.ServiceAccounts, &eventsv1.ServiceAccountCreated{
		Meta: &eventsv1.EventMeta{
			EventId:       operationID,
			OccurredAt:    core.Timestamp(now),
			CorrelationId: operationID,
			TenantId:      account.GetTenantId(),
		},
		ServiceAccount: account,
	}); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *Service) UpdateServiceAccount(ctx context.Context, req *identityv1.UpdateServiceAccountRequest) (*identityv1.ServiceAccount, error) {
	current := &identityv1.ServiceAccount{}
	if err := core.LoadProto(ctx, s.store, ydb.TableServiceAccounts, req.GetServiceAccount().GetServiceAccountId(), current); err != nil {
		return nil, err
	}
	applyServiceAccountMask(current, req.GetServiceAccount(), req.GetUpdateMask())
	now := s.now()
	current.UpdatedAt = core.Timestamp(now)
	if err := core.SaveProto(ctx, s.store, ydb.TableServiceAccounts, current.GetServiceAccountId(), current.GetTenantId(), current, now); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) DeleteServiceAccount(ctx context.Context, req *identityv1.DeleteServiceAccountRequest) (*identityv1.MutateIdentityResponse, error) {
	if err := s.store.DeleteDocument(ctx, ydb.TableServiceAccounts, req.GetServiceAccountId()); err != nil {
		return nil, err
	}
	return &identityv1.MutateIdentityResponse{OperationId: req.GetRequestId()}, nil
}

func (s *Service) GetOAuthClient(ctx context.Context, req *identityv1.GetOAuthClientRequest) (*identityv1.OAuthClient, error) {
	client := &identityv1.OAuthClient{}
	if err := core.LoadProto(ctx, s.store, ydb.TableOAuthClients, req.GetOauthClientId(), client); err != nil {
		return nil, err
	}
	return client, nil
}

func (s *Service) ListOAuthClients(ctx context.Context, req *identityv1.ListOAuthClientsRequest) (*identityv1.ListOAuthClientsResponse, error) {
	clients, next, err := core.ListProto(ctx, s.store, ydb.TableOAuthClients, req.GetTenantId(), int(req.GetPageSize()), req.GetPageToken(), func() *identityv1.OAuthClient {
		return &identityv1.OAuthClient{}
	})
	if err != nil {
		return nil, err
	}
	return &identityv1.ListOAuthClientsResponse{OauthClients: clients, NextPageToken: next}, nil
}

func (s *Service) CreateOAuthClient(ctx context.Context, req *identityv1.CreateOAuthClientRequest) (*identityv1.OAuthClient, error) {
	now := s.now()
	client := &identityv1.OAuthClient{
		OauthClientId:          req.GetOauthClientId(),
		TenantId:               req.GetTenantId(),
		DisplayName:            req.GetDisplayName(),
		ClientType:             req.GetClientType(),
		RedirectUris:           slices.Clone(req.GetRedirectUris()),
		Scopes:                 slices.Clone(req.GetScopes()),
		ServiceAccountsEnabled: req.GetServiceAccountsEnabled(),
		CreatedAt:              core.Timestamp(now),
		UpdatedAt:              core.Timestamp(now),
	}
	if s.keycloak != nil {
		if keycloakClientID, err := s.keycloak.CreateConfidentialClient(ctx, client.GetTenantId(), client.GetOauthClientId(), client.GetDisplayName(), client.GetServiceAccountsEnabled()); err == nil {
			client.KeycloakClientId = keycloakClientID
		} else {
			s.logger.Warn("oauth keycloak client creation skipped", zap.Error(err))
		}
	}
	if err := core.SaveProto(ctx, s.store, ydb.TableOAuthClients, client.GetOauthClientId(), client.GetTenantId(), client, now); err != nil {
		return nil, err
	}
	if err := s.publisher.PublishProto(ctx, s.topics.OAuthClients, &eventsv1.OAuthClientCreated{
		Meta: &eventsv1.EventMeta{
			EventId:       req.GetRequestId(),
			OccurredAt:    core.Timestamp(now),
			CorrelationId: req.GetRequestId(),
			TenantId:      client.GetTenantId(),
		},
		OauthClient: client,
	}); err != nil {
		return nil, err
	}
	return client, nil
}

func (s *Service) UpdateOAuthClient(ctx context.Context, req *identityv1.UpdateOAuthClientRequest) (*identityv1.OAuthClient, error) {
	current := &identityv1.OAuthClient{}
	if err := core.LoadProto(ctx, s.store, ydb.TableOAuthClients, req.GetOauthClient().GetOauthClientId(), current); err != nil {
		return nil, err
	}
	applyOAuthClientMask(current, req.GetOauthClient(), req.GetUpdateMask())
	now := s.now()
	current.UpdatedAt = core.Timestamp(now)
	if err := core.SaveProto(ctx, s.store, ydb.TableOAuthClients, current.GetOauthClientId(), current.GetTenantId(), current, now); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) DeleteOAuthClient(ctx context.Context, req *identityv1.DeleteOAuthClientRequest) (*identityv1.MutateIdentityResponse, error) {
	if err := s.store.DeleteDocument(ctx, ydb.TableOAuthClients, req.GetOauthClientId()); err != nil {
		return nil, err
	}
	return &identityv1.MutateIdentityResponse{OperationId: req.GetRequestId()}, nil
}

func (s *Service) RotateClientSecret(ctx context.Context, req *identityv1.RotateClientSecretRequest) (*identityv1.RotateClientSecretResponse, error) {
	now := s.now()
	operationID := fmt.Sprintf("rotate-%d", now.UnixNano())
	secretRef := fmt.Sprintf("vault://oauth/%s/%s", "clients", req.GetOauthClientId())
	if s.keycloak != nil {
		if _, ref, err := s.keycloak.RotateClientSecret(ctx, req.GetOauthClientId()); err == nil {
			secretRef = ref
		} else {
			s.logger.Warn("client secret rotation fallback", zap.Error(err))
		}
	}
	if s.workflows != nil {
		if _, err := s.workflows.StartWorkflow(ctx, temporalx.RotateClientSecretWorkflowName, operationID, temporalx.RotateClientSecretInput{
			OAuthClientID: req.GetOauthClientId(),
			RequestedBy:   req.GetPerformedBy(),
			Reason:        req.GetReason(),
		}); err != nil {
			s.logger.Warn("rotate client secret workflow start failed", zap.Error(err))
		}
	}
	return &identityv1.RotateClientSecretResponse{OperationId: operationID, SecretRef: secretRef}, nil
}

func groupMemberID(groupID string, subjectType string, subjectID string) string {
	return fmt.Sprintf("%s:%s:%s", groupID, subjectType, subjectID)
}

func applyUserMask(target *identityv1.User, patch *identityv1.User, mask *fieldmaskpb.FieldMask) {
	if len(mask.GetPaths()) == 0 {
		*target = *patch
		return
	}
	for _, path := range mask.GetPaths() {
		switch path {
		case "primary_email":
			target.PrimaryEmail = patch.GetPrimaryEmail()
		case "display_name":
			target.DisplayName = patch.GetDisplayName()
		case "labels":
			target.Labels = core.LabelsFromMap(patch.GetLabels())
		case "group_ids":
			target.GroupIds = slices.Clone(patch.GetGroupIds())
		case "state":
			target.State = patch.GetState()
		}
	}
}

func applyTenantMask(target *identityv1.Tenant, patch *identityv1.Tenant, mask *fieldmaskpb.FieldMask) {
	if len(mask.GetPaths()) == 0 {
		*target = *patch
		return
	}
	for _, path := range mask.GetPaths() {
		switch path {
		case "display_name":
			target.DisplayName = patch.GetDisplayName()
		case "external_ref":
			target.ExternalRef = patch.GetExternalRef()
		case "labels":
			target.Labels = core.LabelsFromMap(patch.GetLabels())
		}
	}
}

func applyGroupMask(target *identityv1.Group, patch *identityv1.Group, mask *fieldmaskpb.FieldMask) {
	if len(mask.GetPaths()) == 0 {
		*target = *patch
		return
	}
	for _, path := range mask.GetPaths() {
		switch path {
		case "display_name":
			target.DisplayName = patch.GetDisplayName()
		case "description":
			target.Description = patch.GetDescription()
		case "labels":
			target.Labels = core.LabelsFromMap(patch.GetLabels())
		}
	}
}

func applyServiceAccountMask(target *identityv1.ServiceAccount, patch *identityv1.ServiceAccount, mask *fieldmaskpb.FieldMask) {
	if len(mask.GetPaths()) == 0 {
		*target = *patch
		return
	}
	for _, path := range mask.GetPaths() {
		switch path {
		case "display_name":
			target.DisplayName = patch.GetDisplayName()
		case "description":
			target.Description = patch.GetDescription()
		case "disabled":
			target.Disabled = patch.GetDisabled()
		}
	}
}

func applyOAuthClientMask(target *identityv1.OAuthClient, patch *identityv1.OAuthClient, mask *fieldmaskpb.FieldMask) {
	if len(mask.GetPaths()) == 0 {
		*target = *patch
		return
	}
	for _, path := range mask.GetPaths() {
		switch path {
		case "display_name":
			target.DisplayName = patch.GetDisplayName()
		case "client_type":
			target.ClientType = patch.GetClientType()
		case "redirect_uris":
			target.RedirectUris = slices.Clone(patch.GetRedirectUris())
		case "scopes":
			target.Scopes = slices.Clone(patch.GetScopes())
		case "service_accounts_enabled":
			target.ServiceAccountsEnabled = patch.GetServiceAccountsEnabled()
		}
	}
}

func ensureID(value string) string {
	if value != "" {
		return value
	}
	return uuid.NewString()
}
