package seeder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	identityv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/identity/v1"
	ydb "github.com/m8platform/platform/iam/internal/adapter/out/ydb"
	"github.com/m8platform/platform/iam/internal/core"
	"github.com/m8platform/platform/iam/internal/foundation/config"
)

type Runner struct {
	store   *ydb.Client
	seedDir string
}

type Report struct {
	Files           []string
	Tenants         int
	Users           int
	Memberships     int
	Groups          int
	GroupMembers    int
	ServiceAccounts int
	OAuthClients    int
	Bindings        int
}

type Dataset struct {
	Tenants         []TenantSeed         `json:"tenants"`
	Users           []UserSeed           `json:"users"`
	Memberships     []MembershipSeed     `json:"memberships"`
	Groups          []GroupSeed          `json:"groups"`
	GroupMembers    []GroupMemberSeed    `json:"group_members"`
	ServiceAccounts []ServiceAccountSeed `json:"service_accounts"`
	OAuthClients    []OAuthClientSeed    `json:"oauth_clients"`
	Bindings        []BindingSeed        `json:"bindings"`
}

type TenantSeed struct {
	TenantID    string            `json:"tenant_id"`
	DisplayName string            `json:"display_name"`
	ExternalRef string            `json:"external_ref"`
	Labels      map[string]string `json:"labels"`
}

type UserSeed struct {
	UserID         string            `json:"user_id"`
	TenantID       string            `json:"tenant_id"`
	PrimaryEmail   string            `json:"primary_email"`
	DisplayName    string            `json:"display_name"`
	State          string            `json:"state"`
	GroupIDs       []string          `json:"group_ids"`
	Labels         map[string]string `json:"labels"`
	KeycloakUserID string            `json:"keycloak_user_id"`
}

type MembershipSeed struct {
	MembershipID string   `json:"membership_id"`
	TenantID     string   `json:"tenant_id"`
	UserID       string   `json:"user_id"`
	RoleIDs      []string `json:"role_ids"`
	State        string   `json:"state"`
}

type GroupSeed struct {
	GroupID     string            `json:"group_id"`
	TenantID    string            `json:"tenant_id"`
	DisplayName string            `json:"display_name"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
}

type GroupMemberSeed struct {
	GroupID     string `json:"group_id"`
	SubjectID   string `json:"subject_id"`
	SubjectType string `json:"subject_type"`
}

type ServiceAccountSeed struct {
	ServiceAccountID string `json:"service_account_id"`
	TenantID         string `json:"tenant_id"`
	DisplayName      string `json:"display_name"`
	Description      string `json:"description"`
	Disabled         bool   `json:"disabled"`
	KeycloakClientID string `json:"keycloak_client_id"`
	OperationID      string `json:"operation_id"`
}

type OAuthClientSeed struct {
	OAuthClientID          string   `json:"oauth_client_id"`
	TenantID               string   `json:"tenant_id"`
	DisplayName            string   `json:"display_name"`
	ClientType             string   `json:"client_type"`
	RedirectURIs           []string `json:"redirect_uris"`
	Scopes                 []string `json:"scopes"`
	ServiceAccountsEnabled bool     `json:"service_accounts_enabled"`
	KeycloakClientID       string   `json:"keycloak_client_id"`
}

type BindingSeed struct {
	BindingID string            `json:"binding_id"`
	Subject   SubjectSeed       `json:"subject"`
	Resource  ResourceSeed      `json:"resource"`
	RoleID    string            `json:"role_id"`
	Reason    string            `json:"reason"`
	Labels    map[string]string `json:"labels"`
}

type SubjectSeed struct {
	Type     string `json:"type"`
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
}

type ResourceSeed struct {
	Type     string `json:"type"`
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
}

func New(cfg config.YDBConfig, seedDir string) (*Runner, error) {
	if cfg.DSN == "" {
		return nil, errors.New("IAM_YDB_DSN is required for seeder")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	store, err := ydb.Open(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Runner{store: store, seedDir: seedDir}, nil
}

func (r *Runner) Close(ctx context.Context) error {
	if r == nil || r.store == nil {
		return nil
	}
	return r.store.Close(ctx)
}

func (r *Runner) Run(ctx context.Context) (*Report, error) {
	data, files, err := LoadDatasets(r.seedDir)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	report := &Report{Files: files}

	for _, tenant := range data.Tenants {
		item := &identityv1.Tenant{
			TenantId:    tenant.TenantID,
			DisplayName: tenant.DisplayName,
			ExternalRef: tenant.ExternalRef,
			Labels:      core.LabelsFromMap(tenant.Labels),
			CreatedAt:   core.Timestamp(now),
			UpdatedAt:   core.Timestamp(now),
		}
		if err := core.SaveProto(ctx, r.store, ydb.TableTenants, item.GetTenantId(), item.GetTenantId(), item, now); err != nil {
			return nil, fmt.Errorf("tenant %s: %w", tenant.TenantID, err)
		}
		report.Tenants++
	}

	for _, user := range data.Users {
		state, err := parseUserState(user.State)
		if err != nil {
			return nil, fmt.Errorf("user %s: %w", user.UserID, err)
		}
		item := &identityv1.User{
			UserId:         user.UserID,
			TenantId:       user.TenantID,
			PrimaryEmail:   strings.ToLower(user.PrimaryEmail),
			DisplayName:    user.DisplayName,
			State:          state,
			GroupIds:       uniqueSortedStrings(user.GroupIDs),
			Labels:         core.LabelsFromMap(user.Labels),
			CreatedAt:      core.Timestamp(now),
			UpdatedAt:      core.Timestamp(now),
			KeycloakUserId: user.KeycloakUserID,
		}
		if err := core.SaveProto(ctx, r.store, ydb.TableUsers, item.GetUserId(), item.GetTenantId(), item, now); err != nil {
			return nil, fmt.Errorf("user %s: %w", user.UserID, err)
		}
		report.Users++
	}

	for _, membership := range data.Memberships {
		state, err := parseMembershipState(membership.State)
		if err != nil {
			return nil, fmt.Errorf("membership %s: %w", membership.MembershipID, err)
		}
		item := &identityv1.Membership{
			MembershipId: membership.MembershipID,
			TenantId:     membership.TenantID,
			UserId:       membership.UserID,
			RoleIds:      uniqueSortedStrings(membership.RoleIDs),
			State:        state,
			CreatedAt:    core.Timestamp(now),
			UpdatedAt:    core.Timestamp(now),
		}
		if err := core.SaveProto(ctx, r.store, ydb.TableMemberships, item.GetMembershipId(), item.GetTenantId(), item, now); err != nil {
			return nil, fmt.Errorf("membership %s: %w", membership.MembershipID, err)
		}
		report.Memberships++
	}

	for _, group := range data.Groups {
		item := &identityv1.Group{
			GroupId:     group.GroupID,
			TenantId:    group.TenantID,
			DisplayName: group.DisplayName,
			Description: group.Description,
			Labels:      core.LabelsFromMap(group.Labels),
			CreatedAt:   core.Timestamp(now),
			UpdatedAt:   core.Timestamp(now),
		}
		if err := core.SaveProto(ctx, r.store, ydb.TableGroups, item.GetGroupId(), item.GetTenantId(), item, now); err != nil {
			return nil, fmt.Errorf("group %s: %w", group.GroupID, err)
		}
		report.Groups++
	}

	groupTenants := make(map[string]string, len(data.Groups))
	for _, group := range data.Groups {
		groupTenants[group.GroupID] = group.TenantID
	}
	for _, member := range data.GroupMembers {
		tenantID, ok := groupTenants[member.GroupID]
		if !ok {
			return nil, fmt.Errorf("group member %s/%s: unknown group", member.GroupID, member.SubjectID)
		}
		item := &identityv1.GroupMember{
			GroupId:     member.GroupID,
			SubjectId:   member.SubjectID,
			SubjectType: member.SubjectType,
			CreatedAt:   core.Timestamp(now),
		}
		memberID := fmt.Sprintf("%s:%s:%s", member.GroupID, member.SubjectType, member.SubjectID)
		if err := core.SaveProto(ctx, r.store, ydb.TableGroupMembers, memberID, tenantID, item, now); err != nil {
			return nil, fmt.Errorf("group member %s: %w", memberID, err)
		}
		report.GroupMembers++
	}

	for _, account := range data.ServiceAccounts {
		item := &identityv1.ServiceAccount{
			ServiceAccountId: account.ServiceAccountID,
			TenantId:         account.TenantID,
			DisplayName:      account.DisplayName,
			Description:      account.Description,
			Disabled:         account.Disabled,
			KeycloakClientId: account.KeycloakClientID,
			OperationId:      account.OperationID,
			CreatedAt:        core.Timestamp(now),
			UpdatedAt:        core.Timestamp(now),
		}
		if err := core.SaveProto(ctx, r.store, ydb.TableServiceAccounts, item.GetServiceAccountId(), item.GetTenantId(), item, now); err != nil {
			return nil, fmt.Errorf("service account %s: %w", account.ServiceAccountID, err)
		}
		report.ServiceAccounts++
	}

	for _, client := range data.OAuthClients {
		clientType, err := parseOAuthClientType(client.ClientType)
		if err != nil {
			return nil, fmt.Errorf("oauth client %s: %w", client.OAuthClientID, err)
		}
		item := &identityv1.OAuthClient{
			OauthClientId:          client.OAuthClientID,
			TenantId:               client.TenantID,
			DisplayName:            client.DisplayName,
			ClientType:             clientType,
			RedirectUris:           uniqueSortedStrings(client.RedirectURIs),
			Scopes:                 uniqueSortedStrings(client.Scopes),
			ServiceAccountsEnabled: client.ServiceAccountsEnabled,
			KeycloakClientId:       client.KeycloakClientID,
			CreatedAt:              core.Timestamp(now),
			UpdatedAt:              core.Timestamp(now),
		}
		if err := core.SaveProto(ctx, r.store, ydb.TableOAuthClients, item.GetOauthClientId(), item.GetTenantId(), item, now); err != nil {
			return nil, fmt.Errorf("oauth client %s: %w", client.OAuthClientID, err)
		}
		report.OAuthClients++
	}

	for _, binding := range data.Bindings {
		subjectType, err := parseSubjectType(binding.Subject.Type)
		if err != nil {
			return nil, fmt.Errorf("binding %s subject: %w", binding.BindingID, err)
		}
		resourceType, err := parseResourceType(binding.Resource.Type)
		if err != nil {
			return nil, fmt.Errorf("binding %s resource: %w", binding.BindingID, err)
		}
		item := &authzv1.AccessBinding{
			BindingId: binding.BindingID,
			Subject: &authzv1.SubjectRef{
				Type:     subjectType,
				Id:       binding.Subject.ID,
				TenantId: binding.Subject.TenantID,
			},
			Resource: &authzv1.ResourceRef{
				Type:     resourceType,
				Id:       binding.Resource.ID,
				TenantId: binding.Resource.TenantID,
			},
			RoleId:    binding.RoleID,
			Reason:    binding.Reason,
			Labels:    core.LabelsFromMap(binding.Labels),
			CreatedAt: core.Timestamp(now),
		}
		if err := core.SaveProto(ctx, r.store, ydb.TableBindingOperations, item.GetBindingId(), item.GetResource().GetTenantId(), item, now); err != nil {
			return nil, fmt.Errorf("binding %s: %w", binding.BindingID, err)
		}
		report.Bindings++
	}

	return report, nil
}

func LoadDatasets(seedDir string) (Dataset, []string, error) {
	entries, err := os.ReadDir(seedDir)
	if err != nil {
		return Dataset{}, nil, err
	}

	files := make([]string, 0, len(entries))
	data := Dataset{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)
	if len(files) == 0 {
		return Dataset{}, nil, fmt.Errorf("no seed files found in %s", seedDir)
	}

	for _, name := range files {
		path := filepath.Join(seedDir, name)
		payload, err := os.ReadFile(path)
		if err != nil {
			return Dataset{}, nil, err
		}
		var partial Dataset
		decoder := json.NewDecoder(strings.NewReader(string(payload)))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&partial); err != nil {
			return Dataset{}, nil, fmt.Errorf("%s: %w", name, err)
		}
		data.Tenants = append(data.Tenants, partial.Tenants...)
		data.Users = append(data.Users, partial.Users...)
		data.Memberships = append(data.Memberships, partial.Memberships...)
		data.Groups = append(data.Groups, partial.Groups...)
		data.GroupMembers = append(data.GroupMembers, partial.GroupMembers...)
		data.ServiceAccounts = append(data.ServiceAccounts, partial.ServiceAccounts...)
		data.OAuthClients = append(data.OAuthClients, partial.OAuthClients...)
		data.Bindings = append(data.Bindings, partial.Bindings...)
	}

	if err := data.normalize(); err != nil {
		return Dataset{}, nil, err
	}
	return data, files, nil
}

func (d *Dataset) normalize() error {
	groupTenants := make(map[string]string, len(d.Groups))
	for _, group := range d.Groups {
		groupTenants[group.GroupID] = group.TenantID
	}

	inferredGroupIDs := make(map[string][]string)
	for _, member := range d.GroupMembers {
		tenantID, ok := groupTenants[member.GroupID]
		if !ok {
			return fmt.Errorf("group member %s/%s references unknown group", member.GroupID, member.SubjectID)
		}
		if member.SubjectType != "SUBJECT_TYPE_USER_ACCOUNT" {
			continue
		}
		inferredGroupIDs[member.SubjectID] = append(inferredGroupIDs[member.SubjectID], member.GroupID)
		_ = tenantID
	}

	for i := range d.Users {
		user := &d.Users[i]
		merged := uniqueSortedStrings(append(user.GroupIDs, inferredGroupIDs[user.UserID]...))
		for _, groupID := range merged {
			groupTenantID, ok := groupTenants[groupID]
			if !ok {
				return fmt.Errorf("user %s references unknown group %s", user.UserID, groupID)
			}
			if groupTenantID != user.TenantID {
				return fmt.Errorf("user %s group %s belongs to tenant %s, expected %s", user.UserID, groupID, groupTenantID, user.TenantID)
			}
		}
		user.GroupIDs = merged
	}
	return nil
}

func parseUserState(raw string) (identityv1.UserState, error) {
	if raw == "" {
		return identityv1.UserState_USER_STATE_ACTIVE, nil
	}
	value, ok := identityv1.UserState_value[raw]
	if !ok {
		return identityv1.UserState_USER_STATE_UNSPECIFIED, fmt.Errorf("unknown user state %q", raw)
	}
	return identityv1.UserState(value), nil
}

func parseMembershipState(raw string) (identityv1.MembershipState, error) {
	if raw == "" {
		return identityv1.MembershipState_MEMBERSHIP_STATE_ACTIVE, nil
	}
	value, ok := identityv1.MembershipState_value[raw]
	if !ok {
		return identityv1.MembershipState_MEMBERSHIP_STATE_UNSPECIFIED, fmt.Errorf("unknown membership state %q", raw)
	}
	return identityv1.MembershipState(value), nil
}

func parseOAuthClientType(raw string) (identityv1.OAuthClientType, error) {
	if raw == "" {
		return identityv1.OAuthClientType_OAUTH_CLIENT_TYPE_CONFIDENTIAL, nil
	}
	value, ok := identityv1.OAuthClientType_value[raw]
	if !ok {
		return identityv1.OAuthClientType_OAUTH_CLIENT_TYPE_UNSPECIFIED, fmt.Errorf("unknown oauth client type %q", raw)
	}
	return identityv1.OAuthClientType(value), nil
}

func parseSubjectType(raw string) (authzv1.SubjectType, error) {
	value, ok := authzv1.SubjectType_value[raw]
	if !ok {
		return authzv1.SubjectType_SUBJECT_TYPE_UNSPECIFIED, fmt.Errorf("unknown subject type %q", raw)
	}
	return authzv1.SubjectType(value), nil
}

func parseResourceType(raw string) (authzv1.ResourceType, error) {
	value, ok := authzv1.ResourceType_value[raw]
	if !ok {
		return authzv1.ResourceType_RESOURCE_TYPE_UNSPECIFIED, fmt.Errorf("unknown resource type %q", raw)
	}
	return authzv1.ResourceType(value), nil
}

func uniqueSortedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	unique := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := unique[value]; ok {
			continue
		}
		unique[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
