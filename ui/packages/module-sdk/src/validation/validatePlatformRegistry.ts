import type {M8PlatformRegistrySnapshot} from '../registry/PlatformRegistry';
import {createValidationResult, errorIssue, warningIssue} from './ValidationResult';
import type {M8ValidationResult, M8ValidationIssue} from './ValidationResult';

export function validatePlatformRegistry(
  registry: M8PlatformRegistrySnapshot,
): M8ValidationResult {
  const issues: M8ValidationIssue[] = [];

  if (!registry.uiApiVersion) {
    issues.push(errorIssue('registry.uiApiVersion.required', 'Registry uiApiVersion is required.', 'uiApiVersion'));
  }

  const scopeIds = new Set<string>();
  registry.scopes.forEach((scope, index) => {
    if (!scope.id) {
      issues.push(errorIssue('scope.id.required', 'Scope id is required.', `scopes.${index}.id`));
    }

    if (scopeIds.has(scope.id)) {
      issues.push(errorIssue('scope.id.duplicate', `Duplicate scope id: ${scope.id}`, `scopes.${index}.id`));
    }

    scopeIds.add(scope.id);

    if (scope.parentScopeId && !scopeIds.has(scope.parentScopeId)) {
      const existsLater = registry.scopes.some((candidate) => candidate.id === scope.parentScopeId);
      if (!existsLater) {
        issues.push(errorIssue('scope.parent.unknown', `Scope ${scope.id} references unknown parentScopeId: ${scope.parentScopeId}`, `scopes.${index}.parentScopeId`));
      }
    }
  });

  const mountPointIds = new Set<string>();
  registry.mountPoints.forEach((mountPoint, index) => {
    if (!mountPoint.id) {
      issues.push(errorIssue('mountPoint.id.required', 'Mount point id is required.', `mountPoints.${index}.id`));
    }

    if (mountPointIds.has(mountPoint.id)) {
      issues.push(errorIssue('mountPoint.id.duplicate', `Duplicate mount point id: ${mountPoint.id}`, `mountPoints.${index}.id`));
    }

    mountPointIds.add(mountPoint.id);

    if (!scopeIds.has(mountPoint.scopeId)) {
      issues.push(errorIssue('mountPoint.scope.unknown', `Mount point ${mountPoint.id} references unknown scopeId: ${mountPoint.scopeId}`, `mountPoints.${index}.scopeId`));
    }

    if (!mountPoint.pathTemplate.startsWith('/')) {
      issues.push(warningIssue('mountPoint.pathTemplate.relative', `Mount point ${mountPoint.id} pathTemplate should start with /.`, `mountPoints.${index}.pathTemplate`));
    }
  });

  const slotIds = new Set<string>();
  registry.slots.forEach((slot, index) => {
    if (!slot.id) {
      issues.push(errorIssue('slot.id.required', 'Slot id is required.', `slots.${index}.id`));
    }

    if (slotIds.has(slot.id)) {
      issues.push(errorIssue('slot.id.duplicate', `Duplicate slot id: ${slot.id}`, `slots.${index}.id`));
    }

    slotIds.add(slot.id);

    if (slot.scopeId && !scopeIds.has(slot.scopeId)) {
      issues.push(errorIssue('slot.scope.unknown', `Slot ${slot.id} references unknown scopeId: ${slot.scopeId}`, `slots.${index}.scopeId`));
    }
  });

  const moduleIds = new Set<string>();
  registry.modules.forEach((module, index) => {
    if (!module.id) {
      issues.push(errorIssue('module.id.required', 'Installed module id is required.', `modules.${index}.id`));
    }

    if (moduleIds.has(module.id)) {
      issues.push(errorIssue('module.id.duplicate', `Duplicate installed module id: ${module.id}`, `modules.${index}.id`));
    }

    moduleIds.add(module.id);
  });

  return createValidationResult(issues);
}
