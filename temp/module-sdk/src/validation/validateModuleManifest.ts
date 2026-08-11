import type {M8ModuleManifest} from '../manifest/ModuleManifest';
import type {M8PlatformRegistrySnapshot} from '../registry/PlatformRegistry';
import {findInstalledModule, findMountPoint, findScope, findSlot} from '../registry/PlatformRegistry';
import {createValidationResult, errorIssue, warningIssue} from './ValidationResult';
import type {M8ValidationIssue, M8ValidationResult} from './ValidationResult';

export type M8ManifestValidationOptions = {
  requireInstalledModule?: boolean;
  requireEnabledModule?: boolean;
  validateNavigationMountPoints?: boolean;
};

export function validateModuleManifest(
  manifest: M8ModuleManifest,
  registry: M8PlatformRegistrySnapshot,
  options: M8ManifestValidationOptions = {},
): M8ValidationResult {
  const issues: M8ValidationIssue[] = [];

  if (!manifest.id) {
    issues.push(errorIssue('manifest.id.required', 'Module manifest id is required.', 'id'));
  }

  if (!manifest.title) {
    issues.push(errorIssue('manifest.title.required', `Module ${manifest.id} title is required.`, 'title'));
  }

  if (!manifest.version) {
    issues.push(errorIssue('manifest.version.required', `Module ${manifest.id} version is required.`, 'version'));
  }

  if (!manifest.moduleApiVersion) {
    issues.push(errorIssue('manifest.moduleApiVersion.required', `Module ${manifest.id} moduleApiVersion is required.`, 'moduleApiVersion'));
  }

  if (!manifest.basePath) {
    issues.push(errorIssue('manifest.basePath.required', `Module ${manifest.id} basePath is required.`, 'basePath'));
  }

  if (manifest.basePath.startsWith('/')) {
    issues.push(warningIssue('manifest.basePath.absolute', `Module ${manifest.id} basePath should be a route segment without leading slash.`, 'basePath'));
  }

  if (!findMountPoint(registry, manifest.mountPointId)) {
    issues.push(errorIssue('manifest.mountPoint.unknown', `Module ${manifest.id} references unknown mountPointId: ${manifest.mountPointId}`, 'mountPointId'));
  }

  const installedModule = findInstalledModule(registry, manifest.id);

  if (options.requireInstalledModule && !installedModule) {
    issues.push(errorIssue('manifest.module.notInstalled', `Module ${manifest.id} is not listed in registry.modules.`, 'id'));
  }

  if (options.requireEnabledModule && installedModule && !installedModule.enabled) {
    issues.push(errorIssue('manifest.module.disabled', `Module ${manifest.id} is disabled in registry.modules.`, 'id'));
  }

  const routeIds = new Set<string>();
  manifest.routes.forEach((route, index) => {
    const path = `routes.${index}`;

    if (!route.id) {
      issues.push(errorIssue('route.id.required', `Route id is required in module ${manifest.id}.`, `${path}.id`));
    }

    if (routeIds.has(route.id)) {
      issues.push(errorIssue('route.id.duplicate', `Duplicate route id in module ${manifest.id}: ${route.id}`, `${path}.id`));
    }

    routeIds.add(route.id);

    if (!route.path && route.path !== '') {
      issues.push(errorIssue('route.path.required', `Route ${route.id} path is required.`, `${path}.path`));
    }

    for (const scopeId of route.requiredScopes ?? []) {
      if (!findScope(registry, scopeId)) {
        issues.push(errorIssue('route.scope.unknown', `Route ${route.id} references unknown scope: ${scopeId}`, `${path}.requiredScopes`));
      }
    }
  });

  const navIds = new Set<string>();
  walkNavigation(manifest.navigation, (navItem, path) => {
    if (!navItem.id) {
      issues.push(errorIssue('navigation.id.required', `Navigation item id is required in module ${manifest.id}.`, `${path}.id`));
    }

    if (navIds.has(navItem.id)) {
      issues.push(errorIssue('navigation.id.duplicate', `Duplicate navigation item id in module ${manifest.id}: ${navItem.id}`, `${path}.id`));
    }

    navIds.add(navItem.id);

    if (!navItem.title) {
      issues.push(errorIssue('navigation.title.required', `Navigation item ${navItem.id} title is required.`, `${path}.title`));
    }

    if (!navItem.to && navItem.to !== '') {
      issues.push(errorIssue('navigation.to.required', `Navigation item ${navItem.id} target is required.`, `${path}.to`));
    }

    const navMountPointId = navItem.mountPointId;
    if (options.validateNavigationMountPoints && navMountPointId && !findMountPoint(registry, navMountPointId)) {
      issues.push(errorIssue('navigation.mountPoint.unknown', `Navigation item ${navItem.id} references unknown mountPointId: ${navMountPointId}`, `${path}.mountPointId`));
    }
  });

  (manifest.widgets ?? []).forEach((widget, index) => {
    const path = `widgets.${index}`;

    if (!widget.id) {
      issues.push(errorIssue('widget.id.required', `Widget id is required in module ${manifest.id}.`, `${path}.id`));
    }

    if (!findSlot(registry, widget.slotId)) {
      issues.push(errorIssue('widget.slot.unknown', `Widget ${widget.id} references unknown slotId: ${widget.slotId}`, `${path}.slotId`));
    }

    for (const scopeId of widget.requiredScopes ?? []) {
      if (!findScope(registry, scopeId)) {
        issues.push(errorIssue('widget.scope.unknown', `Widget ${widget.id} references unknown scope: ${scopeId}`, `${path}.requiredScopes`));
      }
    }
  });

  return createValidationResult(issues);
}

function walkNavigation(
  items: M8ModuleManifest['navigation'],
  visitor: (item: M8ModuleManifest['navigation'][number], path: string) => void,
  basePath = 'navigation',
): void {
  items.forEach((item, index) => {
    const path = `${basePath}.${index}`;
    visitor(item, path);

    if (item.children?.length) {
      walkNavigation(item.children, visitor, `${path}.children`);
    }
  });
}
