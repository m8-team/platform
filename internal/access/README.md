# M8 Access

## Responsibility

Owns permissions, authorization relationships, authorization model revisions and access decision checks.

## Owns

- Permission definition
- Role
- RoleBinding
- Relationship
- AuthorizationModel revision
- CheckPermission decision policy

## Does Not Own

- Users and identities
- Authentication sessions
- Organization, workspace and project hierarchy
- Audit event persistence
- Provider-specific provisioning state

## Main APIs

- CheckPermission
- BatchCheckPermissions
- ExplainAccessDecision
- ListEffectivePermissions
- ListSubjectsWithAccess
- CreatePermissionDefinition
- CreateRole
- CreateRoleBinding
- PublishAuthorizationModel

## Events Published

- AccessRelationshipCreated
- AccessRelationshipDeleted
- AccessRelationshipChanged
- AuthorizationModelPublished

## Events Consumed

- UserDisabled
- ProjectSuspended
- WorkspaceSuspended

## Module Configuration

The Access module is composed through `access.Module(config)`. Permission checks fail closed by default; fail-open behavior must be explicitly configured by the caller for a non-critical operation with a decision reference.
