# M8 Resource Manager

## Responsibility

Owns organizations, workspaces, services, service environment assignment, resource catalog metadata and platform-level resource intent.

## Owns

- Organization
- Workspace
- Service
- Immutable environment assignment for each Service
- Resource catalog metadata
- Platform-level desired and actual state summary

## Does Not Own

- Users
- Authentication sessions
- Permissions and authorization relationships
- Provider-specific provisioning drivers
- Runtime clusters and placements

## Main APIs

- CreateOrganization
- GetOrganization
- UpdateOrganization
- DeleteOrganization
- CreateWorkspace
- GetWorkspace
- UpdateWorkspace
- DeleteWorkspace
- CreateService
- GetService
- UpdateService
- DeleteService

## Events Published

- OrganizationCreated
- OrganizationUpdated
- OrganizationDeleted
- WorkspaceCreated
- WorkspaceUpdated
- WorkspaceDeleted
- ServiceCreated
- ServiceUpdated
- ServiceDeleted

## Events Consumed

- None currently.

## Module Configuration

The Resource Manager module is composed through `resourcemanager.Module(config)` and validates its module-level configuration during Fx application construction.
