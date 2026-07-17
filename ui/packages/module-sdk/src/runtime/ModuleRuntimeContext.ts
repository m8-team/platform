 import type {M8PlatformRegistrySnapshot} from '../registry/PlatformRegistry';
import type {
  M8FeatureFlag,
  M8Metadata,
  M8ModuleId,
  M8MountPointId,
  M8Permission,
  M8ScopeId,
} from '../primitives';
import type {M8ModuleManifest} from '../manifest/ModuleManifest';

export type M8ModuleRuntimeContext<
  TApiRuntime extends M8ApiRuntime = M8ApiRuntime,
  TStoreRuntime extends M8StoreRuntime = M8StoreRuntime,
> = {
  registry: M8PlatformRegistrySnapshot;
  scopes: M8ScopeRuntime;
  auth: M8AuthRuntime;
  permissions: M8PermissionRuntime;
  featureFlags: M8FeatureFlagRuntime;
  api: TApiRuntime;
  query: M8QueryRuntime;
  store: TStoreRuntime;
  router: M8RouterRuntime;
  notifications: M8NotificationRuntime;
  telemetry: M8TelemetryRuntime;
  modules: M8ModuleRegistryRuntime;
};

export type M8ScopeRuntime = {
  current: Record<M8ScopeId, string | undefined>;
  get: (scopeId: M8ScopeId) => string | undefined;
  has: (scopeId: M8ScopeId) => boolean;
  require: (scopeId: M8ScopeId) => string;
};

export type M8AuthRuntime = {
  userId?: string;
  username?: string;
  email?: string;
  isAuthenticated: boolean;
  logout: () => Promise<void>;
  metadata?: M8Metadata;
};

export type M8PermissionRuntime = {
  has: (permission: M8Permission) => boolean;
  hasAny: (permissions: M8Permission[]) => boolean;
  hasAll: (permissions: M8Permission[]) => boolean;
};

export type M8FeatureFlagRuntime = {
  enabled: (flag: M8FeatureFlag) => boolean;
};

export type M8ApiRuntime = Record<string, unknown>;

export type M8QueryRuntime = {
  queryClient: unknown;
};

export type M8StoreRuntime = Record<string, unknown>;

export type M8RouterRuntime = {
  navigate: (to: string, options?: unknown) => Promise<void> | void;
  buildPath: (input: M8BuildPathInput) => string;
};

export type M8BuildPathInput = {
  mountPointId: M8MountPointId;
  moduleBasePath?: string;
  relativePath?: string;
  params?: Record<string, string | number | undefined>;
};

export type M8NotificationRuntime = {
  success: (message: string) => void;
  error: (message: string) => void;
  warning: (message: string) => void;
  info: (message: string) => void;
};

export type M8TelemetryRuntime = {
  event: (name: string, attributes?: Record<string, unknown>) => void;
  error: (error: unknown, attributes?: Record<string, unknown>) => void;
};

export type M8ModuleRegistryRuntime = {
  isInstalled: (moduleId: M8ModuleId) => boolean;
  isEnabled: (moduleId: M8ModuleId) => boolean;
  getManifest: (moduleId: M8ModuleId) => M8ModuleManifest | undefined;
  getManifests: () => M8ModuleManifest[];
};
