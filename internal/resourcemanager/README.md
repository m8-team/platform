# M8 Resource Manager

## Responsibility

Owns organizations, workspaces, projects, resource catalog metadata and platform-level resource intent.

## Owns

- Organization
- Workspace
- Project
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
- CreateProject
- GetProject
- UpdateProject
- DeleteProject

## Events Published

- OrganizationCreated
- OrganizationUpdated
- OrganizationDeleted
- WorkspaceCreated
- WorkspaceUpdated
- WorkspaceDeleted
- ProjectCreated
- ProjectUpdated
- ProjectDeleted

## Events Consumed

- None currently.

## Module Configuration

The Resource Manager module is composed through `resourcemanager.Module(config)` and validates its module-level configuration during Fx application construction.
