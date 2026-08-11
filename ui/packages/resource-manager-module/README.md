# M8 Resource Manager UI Module

## Responsibility

Owns the Resource Manager Console contribution: organizations and projects routes, query definitions, operations and its API adapter.

## Owns

- `/resource-manager` routes
- Organization and project query definitions
- Create and delete project operations
- Resource Manager API client boundary

## Does Not Own

- Console application composition
- Shared json-render module contracts
- Component renderer implementations

## Integration

Install the workspace package in Console and pass `resourceManagerModule` to the Console module registry.
