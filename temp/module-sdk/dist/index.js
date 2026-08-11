// src/registry/PlatformRegistry.ts
function definePlatformRegistry(registry) {
  return registry;
}
function findMountPoint(registry, mountPointId) {
  return registry.mountPoints.find((mountPoint) => mountPoint.id === mountPointId);
}
function findScope(registry, scopeId) {
  return registry.scopes.find((scope) => scope.id === scopeId);
}
function findSlot(registry, slotId) {
  return registry.slots.find((slot) => slot.id === slotId);
}
function findInstalledModule(registry, moduleId) {
  return registry.modules.find((module) => module.id === moduleId);
}

// src/manifest/ModuleManifest.ts
function defineModuleManifest(manifest) {
  return manifest;
}

// src/validation/isModuleVisible.ts
function isModuleVisible(manifest, runtime) {
  if (manifest.requiredPermissions?.length && !runtime.permissions.hasAll(manifest.requiredPermissions)) {
    return false;
  }
  if (manifest.requiredFeatureFlags?.length) {
    return manifest.requiredFeatureFlags.every((flag) => runtime.featureFlags.enabled(flag));
  }
  return true;
}

// src/navigation/buildNavigation.ts
function buildNavigation(manifests, runtime) {
  return manifests.filter((manifest) => isModuleVisible(manifest, runtime)).flatMap((manifest) => filterNavigationItems(manifest.navigation, runtime)).sort(sortByOrder);
}
function filterNavigationItems(items, runtime) {
  return items.filter((item) => isNavigationItemVisible(item, runtime)).map((item) => {
    const children = item.children ? filterNavigationItems(item.children, runtime).sort(sortByOrder) : void 0;
    if (!children) {
      return item;
    }
    return {
      ...item,
      children
    };
  }).sort(sortByOrder);
}
function isNavigationItemVisible(item, runtime) {
  if (item.requiredPermissions?.length && !runtime.permissions.hasAll(item.requiredPermissions)) {
    return false;
  }
  if (item.requiredFeatureFlags?.length) {
    return item.requiredFeatureFlags.every((flag) => runtime.featureFlags.enabled(flag));
  }
  return true;
}
function sortByOrder(a, b) {
  const orderDiff = (a.order ?? 0) - (b.order ?? 0);
  if (orderDiff !== 0) {
    return orderDiff;
  }
  return (a.title ?? "").localeCompare(b.title ?? "");
}

// src/runtime/createScopeRuntime.ts
function createScopeRuntime(current) {
  return {
    current,
    get(scopeId) {
      return current[scopeId];
    },
    has(scopeId) {
      return Boolean(current[scopeId]);
    },
    require(scopeId) {
      const value = current[scopeId];
      if (!value) {
        throw new Error(`Scope value is required: ${scopeId}`);
      }
      return value;
    }
  };
}

// src/utils/path.ts
function normalizePath(path) {
  if (!path) {
    return "";
  }
  const trimmed = path.trim();
  if (!trimmed || trimmed === "/") {
    return "";
  }
  return trimmed.startsWith("/") ? trimmed : `/${trimmed}`;
}
function joinPaths(...parts) {
  const result = parts.map((part) => normalizePath(part)).filter(Boolean).join("");
  return result || "/";
}
function interpolatePathTemplate(template, params) {
  return template.replace(/:([A-Za-z0-9_]+)/g, (_, key) => {
    const value = params[key];
    if (value === void 0 || value === null || value === "") {
      throw new Error(`Missing required path parameter: ${key}`);
    }
    return encodeURIComponent(String(value));
  });
}
function stripTrailingSlash(path) {
  if (path === "/") {
    return path;
  }
  return path.endsWith("/") ? path.slice(0, -1) : path;
}

// src/runtime/createRouterRuntime.ts
function createRouterRuntime(input) {
  return {
    navigate: input.navigate,
    buildPath: (buildInput) => buildPathFromRegistry(input.registry, input.scopes, buildInput)
  };
}
function buildPathFromRegistry(registry, scopes, input) {
  const mountPoint = findMountPoint(registry, input.mountPointId);
  if (!mountPoint) {
    throw new Error(`Unknown mount point: ${input.mountPointId}`);
  }
  const params = {
    ...scopes.current,
    ...input.params
  };
  const mountedPath = interpolatePathTemplate(mountPoint.pathTemplate, params);
  return joinPaths(mountedPath, input.moduleBasePath, input.relativePath);
}

// src/module/RemoteModule.ts
function defineRemoteModule(remoteModule) {
  return remoteModule;
}

// src/module/defineModule.ts
function defineModule(module) {
  return {
    ...module,
    getManifest: () => module.manifest
  };
}

// src/validation/ValidationResult.ts
function createValidationResult(issues) {
  const errors = issues.filter((issue) => issue.severity === "error");
  const warnings = issues.filter((issue) => issue.severity === "warning");
  return {
    valid: errors.length === 0,
    issues,
    errors,
    warnings
  };
}
function errorIssue(code, message, path) {
  return {
    severity: "error",
    code,
    message,
    ...path ? { path } : {}
  };
}
function warningIssue(code, message, path) {
  return {
    severity: "warning",
    code,
    message,
    ...path ? { path } : {}
  };
}

// src/validation/validatePlatformRegistry.ts
function validatePlatformRegistry(registry) {
  const issues = [];
  if (!registry.uiApiVersion) {
    issues.push(errorIssue("registry.uiApiVersion.required", "Registry uiApiVersion is required.", "uiApiVersion"));
  }
  const scopeIds = /* @__PURE__ */ new Set();
  registry.scopes.forEach((scope, index) => {
    if (!scope.id) {
      issues.push(errorIssue("scope.id.required", "Scope id is required.", `scopes.${index}.id`));
    }
    if (scopeIds.has(scope.id)) {
      issues.push(errorIssue("scope.id.duplicate", `Duplicate scope id: ${scope.id}`, `scopes.${index}.id`));
    }
    scopeIds.add(scope.id);
    if (scope.parentScopeId && !scopeIds.has(scope.parentScopeId)) {
      const existsLater = registry.scopes.some((candidate) => candidate.id === scope.parentScopeId);
      if (!existsLater) {
        issues.push(errorIssue("scope.parent.unknown", `Scope ${scope.id} references unknown parentScopeId: ${scope.parentScopeId}`, `scopes.${index}.parentScopeId`));
      }
    }
  });
  const mountPointIds = /* @__PURE__ */ new Set();
  registry.mountPoints.forEach((mountPoint, index) => {
    if (!mountPoint.id) {
      issues.push(errorIssue("mountPoint.id.required", "Mount point id is required.", `mountPoints.${index}.id`));
    }
    if (mountPointIds.has(mountPoint.id)) {
      issues.push(errorIssue("mountPoint.id.duplicate", `Duplicate mount point id: ${mountPoint.id}`, `mountPoints.${index}.id`));
    }
    mountPointIds.add(mountPoint.id);
    if (!scopeIds.has(mountPoint.scopeId)) {
      issues.push(errorIssue("mountPoint.scope.unknown", `Mount point ${mountPoint.id} references unknown scopeId: ${mountPoint.scopeId}`, `mountPoints.${index}.scopeId`));
    }
    if (!mountPoint.pathTemplate.startsWith("/")) {
      issues.push(warningIssue("mountPoint.pathTemplate.relative", `Mount point ${mountPoint.id} pathTemplate should start with /.`, `mountPoints.${index}.pathTemplate`));
    }
  });
  const slotIds = /* @__PURE__ */ new Set();
  registry.slots.forEach((slot, index) => {
    if (!slot.id) {
      issues.push(errorIssue("slot.id.required", "Slot id is required.", `slots.${index}.id`));
    }
    if (slotIds.has(slot.id)) {
      issues.push(errorIssue("slot.id.duplicate", `Duplicate slot id: ${slot.id}`, `slots.${index}.id`));
    }
    slotIds.add(slot.id);
    if (slot.scopeId && !scopeIds.has(slot.scopeId)) {
      issues.push(errorIssue("slot.scope.unknown", `Slot ${slot.id} references unknown scopeId: ${slot.scopeId}`, `slots.${index}.scopeId`));
    }
  });
  const moduleIds = /* @__PURE__ */ new Set();
  registry.modules.forEach((module, index) => {
    if (!module.id) {
      issues.push(errorIssue("module.id.required", "Installed module id is required.", `modules.${index}.id`));
    }
    if (moduleIds.has(module.id)) {
      issues.push(errorIssue("module.id.duplicate", `Duplicate installed module id: ${module.id}`, `modules.${index}.id`));
    }
    moduleIds.add(module.id);
  });
  return createValidationResult(issues);
}

// src/validation/validateModuleManifest.ts
function validateModuleManifest(manifest, registry, options = {}) {
  const issues = [];
  if (!manifest.id) {
    issues.push(errorIssue("manifest.id.required", "Module manifest id is required.", "id"));
  }
  if (!manifest.title) {
    issues.push(errorIssue("manifest.title.required", `Module ${manifest.id} title is required.`, "title"));
  }
  if (!manifest.version) {
    issues.push(errorIssue("manifest.version.required", `Module ${manifest.id} version is required.`, "version"));
  }
  if (!manifest.moduleApiVersion) {
    issues.push(errorIssue("manifest.moduleApiVersion.required", `Module ${manifest.id} moduleApiVersion is required.`, "moduleApiVersion"));
  }
  if (!manifest.basePath) {
    issues.push(errorIssue("manifest.basePath.required", `Module ${manifest.id} basePath is required.`, "basePath"));
  }
  if (manifest.basePath.startsWith("/")) {
    issues.push(warningIssue("manifest.basePath.absolute", `Module ${manifest.id} basePath should be a route segment without leading slash.`, "basePath"));
  }
  if (!findMountPoint(registry, manifest.mountPointId)) {
    issues.push(errorIssue("manifest.mountPoint.unknown", `Module ${manifest.id} references unknown mountPointId: ${manifest.mountPointId}`, "mountPointId"));
  }
  const installedModule = findInstalledModule(registry, manifest.id);
  if (options.requireInstalledModule && !installedModule) {
    issues.push(errorIssue("manifest.module.notInstalled", `Module ${manifest.id} is not listed in registry.modules.`, "id"));
  }
  if (options.requireEnabledModule && installedModule && !installedModule.enabled) {
    issues.push(errorIssue("manifest.module.disabled", `Module ${manifest.id} is disabled in registry.modules.`, "id"));
  }
  const routeIds = /* @__PURE__ */ new Set();
  manifest.routes.forEach((route, index) => {
    const path = `routes.${index}`;
    if (!route.id) {
      issues.push(errorIssue("route.id.required", `Route id is required in module ${manifest.id}.`, `${path}.id`));
    }
    if (routeIds.has(route.id)) {
      issues.push(errorIssue("route.id.duplicate", `Duplicate route id in module ${manifest.id}: ${route.id}`, `${path}.id`));
    }
    routeIds.add(route.id);
    if (!route.path && route.path !== "") {
      issues.push(errorIssue("route.path.required", `Route ${route.id} path is required.`, `${path}.path`));
    }
    for (const scopeId of route.requiredScopes ?? []) {
      if (!findScope(registry, scopeId)) {
        issues.push(errorIssue("route.scope.unknown", `Route ${route.id} references unknown scope: ${scopeId}`, `${path}.requiredScopes`));
      }
    }
  });
  const navIds = /* @__PURE__ */ new Set();
  walkNavigation(manifest.navigation, (navItem, path) => {
    if (!navItem.id) {
      issues.push(errorIssue("navigation.id.required", `Navigation item id is required in module ${manifest.id}.`, `${path}.id`));
    }
    if (navIds.has(navItem.id)) {
      issues.push(errorIssue("navigation.id.duplicate", `Duplicate navigation item id in module ${manifest.id}: ${navItem.id}`, `${path}.id`));
    }
    navIds.add(navItem.id);
    if (!navItem.title) {
      issues.push(errorIssue("navigation.title.required", `Navigation item ${navItem.id} title is required.`, `${path}.title`));
    }
    if (!navItem.to && navItem.to !== "") {
      issues.push(errorIssue("navigation.to.required", `Navigation item ${navItem.id} target is required.`, `${path}.to`));
    }
    const navMountPointId = navItem.mountPointId;
    if (options.validateNavigationMountPoints && navMountPointId && !findMountPoint(registry, navMountPointId)) {
      issues.push(errorIssue("navigation.mountPoint.unknown", `Navigation item ${navItem.id} references unknown mountPointId: ${navMountPointId}`, `${path}.mountPointId`));
    }
  });
  (manifest.widgets ?? []).forEach((widget, index) => {
    const path = `widgets.${index}`;
    if (!widget.id) {
      issues.push(errorIssue("widget.id.required", `Widget id is required in module ${manifest.id}.`, `${path}.id`));
    }
    if (!findSlot(registry, widget.slotId)) {
      issues.push(errorIssue("widget.slot.unknown", `Widget ${widget.id} references unknown slotId: ${widget.slotId}`, `${path}.slotId`));
    }
    for (const scopeId of widget.requiredScopes ?? []) {
      if (!findScope(registry, scopeId)) {
        issues.push(errorIssue("widget.scope.unknown", `Widget ${widget.id} references unknown scope: ${scopeId}`, `${path}.requiredScopes`));
      }
    }
  });
  return createValidationResult(issues);
}
function walkNavigation(items, visitor, basePath = "navigation") {
  items.forEach((item, index) => {
    const path = `${basePath}.${index}`;
    visitor(item, path);
    if (item.children?.length) {
      walkNavigation(item.children, visitor, `${path}.children`);
    }
  });
}

// src/utils/object.ts
function compact(items) {
  return items.filter(Boolean);
}
function unique(items) {
  return Array.from(new Set(items));
}
export {
  buildNavigation,
  buildPathFromRegistry,
  compact,
  createRouterRuntime,
  createScopeRuntime,
  createValidationResult,
  defineModule,
  defineModuleManifest,
  definePlatformRegistry,
  defineRemoteModule,
  errorIssue,
  filterNavigationItems,
  findInstalledModule,
  findMountPoint,
  findScope,
  findSlot,
  interpolatePathTemplate,
  isModuleVisible,
  isNavigationItemVisible,
  joinPaths,
  normalizePath,
  stripTrailingSlash,
  unique,
  validateModuleManifest,
  validatePlatformRegistry,
  warningIssue
};
