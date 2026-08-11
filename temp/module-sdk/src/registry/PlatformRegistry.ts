import type {
  M8Capability,
  M8FeatureFlag,
  M8Metadata,
  M8ModuleId,
  M8MountPointId,
  M8Permission,
  M8ScopeId,
  M8SlotId,
} from '../primitives';

export type M8PlatformRegistrySnapshot = {
  uiApiVersion: string;
  revision?: string;
  generatedAt?: string;
  scopes: M8ScopeDefinition[];
  mountPoints: M8MountPointDefinition[];
  slots: M8SlotDefinition[];
  modules: M8InstalledModuleDefinition[];
  metadata?: M8Metadata;
};

export type M8ScopeDefinition = {
  id: M8ScopeId;
  title: string;
  description?: string;
  level: number;
  paramName?: string;
  parentScopeId?: M8ScopeId;
  metadata?: M8Metadata;
};

export type M8MountPointDefinition = {
  id: M8MountPointId;
  scopeId: M8ScopeId;
  pathTemplate: string;
  title?: string;
  order?: number;
  metadata?: M8Metadata;
};

export type M8SlotDefinition = {
  id: M8SlotId;
  title: string;
  description?: string;
  scopeId?: M8ScopeId;
  order?: number;
  metadata?: M8Metadata;
};

export type M8InstalledModuleDefinition = {
  id: M8ModuleId;
  enabled: boolean;
  title?: string;
  version?: string;
  moduleApiVersion?: string;
  uiEntry?: string | null;
  manifestUrl?: string | null;
  requiredCapabilities?: M8Capability[];
  requiredPermissions?: M8Permission[];
  requiredFeatureFlags?: M8FeatureFlag[];
  metadata?: M8Metadata;
};

export function definePlatformRegistry<TRegistry extends M8PlatformRegistrySnapshot>(
  registry: TRegistry,
): TRegistry {
  return registry;
}

export function findMountPoint(
  registry: M8PlatformRegistrySnapshot,
  mountPointId: M8MountPointId,
): M8MountPointDefinition | undefined {
  return registry.mountPoints.find((mountPoint) => mountPoint.id === mountPointId);
}

export function findScope(
  registry: M8PlatformRegistrySnapshot,
  scopeId: M8ScopeId,
): M8ScopeDefinition | undefined {
  return registry.scopes.find((scope) => scope.id === scopeId);
}

export function findSlot(
  registry: M8PlatformRegistrySnapshot,
  slotId: M8SlotId,
): M8SlotDefinition | undefined {
  return registry.slots.find((slot) => slot.id === slotId);
}

export function findInstalledModule(
  registry: M8PlatformRegistrySnapshot,
  moduleId: M8ModuleId,
): M8InstalledModuleDefinition | undefined {
  return registry.modules.find((module) => module.id === moduleId);
}
