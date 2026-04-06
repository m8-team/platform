package spicedb

import (
	"errors"
	"fmt"
	"strings"

	spicedbv1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	identityv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/identity/v1"
)

const (
	userObjectType           = "user"
	federatedUserObjectType  = "federated_user"
	serviceAccountObjectType = "service_account"
	groupObjectType          = "group"
	tenantObjectType         = "tenant"
	projectObjectType        = "project"
	supportCaseObjectType    = "support_case"

	groupMemberRelation = "member"
)

func objectRefForResource(resource *authzv1.ResourceRef) (*spicedbv1.ObjectReference, error) {
	if resource == nil {
		return nil, errors.New("spicedb resource is nil")
	}
	objectType, err := objectTypeForResource(resource.GetType())
	if err != nil {
		return nil, err
	}
	return &spicedbv1.ObjectReference{
		ObjectType: objectType,
		ObjectId:   resource.GetId(),
	}, nil
}

func objectTypeForResource(resourceType authzv1.ResourceType) (string, error) {
	switch resourceType {
	case authzv1.ResourceType_RESOURCE_TYPE_TENANT:
		return tenantObjectType, nil
	case authzv1.ResourceType_RESOURCE_TYPE_PROJECT:
		return projectObjectType, nil
	case authzv1.ResourceType_RESOURCE_TYPE_SUPPORT_CASE:
		return supportCaseObjectType, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedResourceType, resourceType.String())
	}
}

func subjectRefForCheck(subject *authzv1.SubjectRef) (*spicedbv1.SubjectReference, error) {
	if subject == nil {
		return nil, errors.New("spicedb subject is nil")
	}
	objectType, err := objectTypeForSubject(subject.GetType())
	if err != nil {
		return nil, err
	}
	return &spicedbv1.SubjectReference{
		Object: &spicedbv1.ObjectReference{
			ObjectType: objectType,
			ObjectId:   subject.GetId(),
		},
	}, nil
}

func subjectReferencesForBinding(binding *authzv1.AccessBinding) ([]*spicedbv1.SubjectReference, error) {
	if binding == nil || binding.GetSubject() == nil {
		return nil, errors.New("spicedb binding subject is nil")
	}
	subject := binding.GetSubject()
	objectType, err := objectTypeForSubject(subject.GetType())
	if err != nil {
		return nil, err
	}

	base := &spicedbv1.SubjectReference{
		Object: &spicedbv1.ObjectReference{
			ObjectType: objectType,
			ObjectId:   subject.GetId(),
		},
	}
	if subject.GetType() != authzv1.SubjectType_SUBJECT_TYPE_GROUP {
		return []*spicedbv1.SubjectReference{base}, nil
	}

	member := &spicedbv1.SubjectReference{
		Object:           base.GetObject(),
		OptionalRelation: groupMemberRelation,
	}
	return []*spicedbv1.SubjectReference{base, member}, nil
}

func objectTypeForSubject(subjectType authzv1.SubjectType) (string, error) {
	switch subjectType {
	case authzv1.SubjectType_SUBJECT_TYPE_USER_ACCOUNT:
		return userObjectType, nil
	case authzv1.SubjectType_SUBJECT_TYPE_SERVICE_ACCOUNT:
		return serviceAccountObjectType, nil
	case authzv1.SubjectType_SUBJECT_TYPE_GROUP:
		return groupObjectType, nil
	case authzv1.SubjectType_SUBJECT_TYPE_FEDERATED_USER:
		return federatedUserObjectType, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedSubjectType, subjectType.String())
	}
}

func objectTypeForGroupMemberSubject(raw string) (string, error) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "subject_type_user_account", "user_account":
		return userObjectType, nil
	case "subject_type_service_account", "service_account":
		return serviceAccountObjectType, nil
	case "subject_type_federated_user", "federated_user":
		return federatedUserObjectType, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedSubjectType, raw)
	}
}

func permissionNameForCheck(resource *authzv1.ResourceRef, permission string) (string, error) {
	if resource == nil {
		return "", errors.New("spicedb resource is nil")
	}

	switch resource.GetType() {
	case authzv1.ResourceType_RESOURCE_TYPE_TENANT:
		switch permission {
		case "tenant.manage", "manage":
			return "manage", nil
		}
	case authzv1.ResourceType_RESOURCE_TYPE_PROJECT:
		switch permission {
		case "project.read", "read":
			return "read", nil
		case "project.write", "write":
			return "write", nil
		}
	case authzv1.ResourceType_RESOURCE_TYPE_SUPPORT_CASE:
		switch permission {
		case "support.case.read", "read":
			return "read", nil
		}
	}

	return "", fmt.Errorf("%w: %s on %s", ErrUnsupportedPermission, permission, resource.GetType().String())
}

func relationshipsForBinding(binding *authzv1.AccessBinding) ([]*spicedbv1.Relationship, error) {
	if binding == nil {
		return nil, errors.New("spicedb binding is nil")
	}

	resource, err := objectRefForResource(binding.GetResource())
	if err != nil {
		return nil, err
	}
	relations, err := relationNamesForBinding(binding.GetResource().GetType(), binding.GetRoleId())
	if err != nil {
		return nil, err
	}
	subjects, err := subjectReferencesForBinding(binding)
	if err != nil {
		return nil, err
	}
	caveat, err := caveatForBinding(binding)
	if err != nil {
		return nil, err
	}

	relationships := make([]*spicedbv1.Relationship, 0, len(relations)*len(subjects))
	for _, relation := range relations {
		for _, subject := range subjects {
			relationships = append(relationships, &spicedbv1.Relationship{
				Resource:          resource,
				Relation:          relation,
				Subject:           subject,
				OptionalCaveat:    caveat,
				OptionalExpiresAt: binding.GetExpiresAt(),
			})
		}
	}
	return relationships, nil
}

func relationNamesForBinding(resourceType authzv1.ResourceType, roleID string) ([]string, error) {
	switch resourceType {
	case authzv1.ResourceType_RESOURCE_TYPE_TENANT:
		switch roleID {
		case "tenant-owner":
			return []string{"owner"}, nil
		case "tenant-admin":
			return []string{"admin"}, nil
		}
	case authzv1.ResourceType_RESOURCE_TYPE_PROJECT:
		switch roleID {
		case "project-viewer":
			return []string{"viewer"}, nil
		case "project-editor":
			return []string{"editor"}, nil
		case "support-operator":
			return []string{"support"}, nil
		}
	case authzv1.ResourceType_RESOURCE_TYPE_SUPPORT_CASE:
		switch roleID {
		case "support-operator":
			return []string{"viewer"}, nil
		}
	}
	return nil, fmt.Errorf("%w: role=%s resource=%s", ErrUnsupportedRoleBinding, roleID, resourceType.String())
}

func mutableRelationsForResource(resourceType authzv1.ResourceType) []string {
	switch resourceType {
	case authzv1.ResourceType_RESOURCE_TYPE_TENANT:
		return []string{"owner", "admin"}
	case authzv1.ResourceType_RESOURCE_TYPE_PROJECT:
		return []string{"viewer", "editor", "support"}
	case authzv1.ResourceType_RESOURCE_TYPE_SUPPORT_CASE:
		return []string{"viewer"}
	default:
		return nil
	}
}

func tenantAnchorRelationship(resource *authzv1.ResourceRef) (*spicedbv1.Relationship, bool, error) {
	if resource == nil {
		return nil, false, errors.New("spicedb resource is nil")
	}
	switch resource.GetType() {
	case authzv1.ResourceType_RESOURCE_TYPE_PROJECT, authzv1.ResourceType_RESOURCE_TYPE_SUPPORT_CASE:
		resourceRef, err := objectRefForResource(resource)
		if err != nil {
			return nil, false, err
		}
		return &spicedbv1.Relationship{
			Resource: resourceRef,
			Relation: "tenant",
			Subject: &spicedbv1.SubjectReference{
				Object: &spicedbv1.ObjectReference{
					ObjectType: tenantObjectType,
					ObjectId:   resource.GetTenantId(),
				},
			},
		}, true, nil
	case authzv1.ResourceType_RESOURCE_TYPE_TENANT:
		return nil, false, nil
	default:
		return nil, false, fmt.Errorf("%w: %s", ErrUnsupportedResourceType, resource.GetType().String())
	}
}

func relationshipForGroupMember(member *identityv1.GroupMember) (*spicedbv1.Relationship, error) {
	if member == nil {
		return nil, errors.New("spicedb group member is nil")
	}
	subjectObjectType, err := objectTypeForGroupMemberSubject(member.GetSubjectType())
	if err != nil {
		return nil, err
	}
	return &spicedbv1.Relationship{
		Resource: &spicedbv1.ObjectReference{
			ObjectType: groupObjectType,
			ObjectId:   member.GetGroupId(),
		},
		Relation: groupMemberRelation,
		Subject: &spicedbv1.SubjectReference{
			Object: &spicedbv1.ObjectReference{
				ObjectType: subjectObjectType,
				ObjectId:   member.GetSubjectId(),
			},
		},
	}, nil
}

func caveatForBinding(binding *authzv1.AccessBinding) (*spicedbv1.ContextualizedCaveat, error) {
	if binding == nil || binding.GetCaveatContext() == nil {
		return nil, nil
	}
	if len(binding.GetCaveatContext().GetFields()) == 0 {
		return nil, nil
	}
	return nil, ErrUnsupportedCaveatContext
}

func touchRelationship(relationship *spicedbv1.Relationship) *spicedbv1.RelationshipUpdate {
	return &spicedbv1.RelationshipUpdate{
		Operation:    spicedbv1.RelationshipUpdate_OPERATION_TOUCH,
		Relationship: relationship,
	}
}

func decisionFromPermissionship(permissionship spicedbv1.CheckPermissionResponse_Permissionship) authzv1.PermissionDecision {
	switch permissionship {
	case spicedbv1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION:
		return authzv1.PermissionDecision_PERMISSION_DECISION_ALLOW
	case spicedbv1.CheckPermissionResponse_PERMISSIONSHIP_CONDITIONAL_PERMISSION:
		return authzv1.PermissionDecision_PERMISSION_DECISION_CONDITIONAL
	default:
		return authzv1.PermissionDecision_PERMISSION_DECISION_DENY
	}
}

func estimatedRelationshipCount(resource *authzv1.ResourceRef, bindings []*authzv1.AccessBinding) int {
	count := 0
	if resource != nil && resource.GetType() != authzv1.ResourceType_RESOURCE_TYPE_TENANT {
		count++
	}
	for _, binding := range bindings {
		if binding == nil {
			continue
		}
		relations, err := relationNamesForBinding(binding.GetResource().GetType(), binding.GetRoleId())
		if err != nil {
			continue
		}
		subjectMultiplier := 1
		if binding.GetSubject().GetType() == authzv1.SubjectType_SUBJECT_TYPE_GROUP {
			subjectMultiplier = 2
		}
		count += len(relations) * subjectMultiplier
	}
	return count
}

func resourceKey(resource *authzv1.ResourceRef) string {
	return fmt.Sprintf("%s:%s:%s", resource.GetType().String(), resource.GetTenantId(), resource.GetId())
}

func sameResource(left *authzv1.ResourceRef, right *authzv1.ResourceRef) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.GetType() == right.GetType() && left.GetTenantId() == right.GetTenantId() && left.GetId() == right.GetId()
}
