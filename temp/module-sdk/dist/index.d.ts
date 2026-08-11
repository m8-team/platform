import { ComponentType, LazyExoticComponent, ReactNode } from 'react';

type M8Id = string;
type M8ModuleId = string;
type M8ScopeId = string;
type M8MountPointId = string;
type M8SlotId = string;
type M8Permission = string;
type M8Capability = string;
type M8FeatureFlag = string;
type M8ModuleKind = string;
type M8ModuleLifecycle = string;
type M8Metadata = Record<string, unknown>;
type M8Dictionary<TValue = unknown> = Record<string, TValue>;
type M8MaybePromise<T> = T | Promise<T>;
type M8ComponentIconProps = {
    className?: string;
    width?: number | string;
    height?: number | string;
};

type M8PlatformRegistrySnapshot = {
    uiApiVersion: string;
    revision?: string;
    generatedAt?: string;
    scopes: M8ScopeDefinition[];
    mountPoints: M8MountPointDefinition[];
    slots: M8SlotDefinition[];
    modules: M8InstalledModuleDefinition[];
    metadata?: M8Metadata;
};
type M8ScopeDefinition = {
    id: M8ScopeId;
    title: string;
    description?: string;
    level: number;
    paramName?: string;
    parentScopeId?: M8ScopeId;
    metadata?: M8Metadata;
};
type M8MountPointDefinition = {
    id: M8MountPointId;
    scopeId: M8ScopeId;
    pathTemplate: string;
    title?: string;
    order?: number;
    metadata?: M8Metadata;
};
type M8SlotDefinition = {
    id: M8SlotId;
    title: string;
    description?: string;
    scopeId?: M8ScopeId;
    order?: number;
    metadata?: M8Metadata;
};
type M8InstalledModuleDefinition = {
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
declare function definePlatformRegistry<TRegistry extends M8PlatformRegistrySnapshot>(registry: TRegistry): TRegistry;
declare function findMountPoint(registry: M8PlatformRegistrySnapshot, mountPointId: M8MountPointId): M8MountPointDefinition | undefined;
declare function findScope(registry: M8PlatformRegistrySnapshot, scopeId: M8ScopeId): M8ScopeDefinition | undefined;
declare function findSlot(registry: M8PlatformRegistrySnapshot, slotId: M8SlotId): M8SlotDefinition | undefined;
declare function findInstalledModule(registry: M8PlatformRegistrySnapshot, moduleId: M8ModuleId): M8InstalledModuleDefinition | undefined;

type M8ModuleRuntimeContext<TApiRuntime extends M8ApiRuntime = M8ApiRuntime, TStoreRuntime extends M8StoreRuntime = M8StoreRuntime> = {
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
type M8ScopeRuntime = {
    current: Record<M8ScopeId, string | undefined>;
    get: (scopeId: M8ScopeId) => string | undefined;
    has: (scopeId: M8ScopeId) => boolean;
    require: (scopeId: M8ScopeId) => string;
};
type M8AuthRuntime = {
    userId?: string;
    username?: string;
    email?: string;
    isAuthenticated: boolean;
    logout: () => Promise<void>;
    metadata?: M8Metadata;
};
type M8PermissionRuntime = {
    has: (permission: M8Permission) => boolean;
    hasAny: (permissions: M8Permission[]) => boolean;
    hasAll: (permissions: M8Permission[]) => boolean;
};
type M8FeatureFlagRuntime = {
    enabled: (flag: M8FeatureFlag) => boolean;
};
type M8ApiRuntime = Record<string, unknown>;
type M8QueryRuntime = {
    queryClient: unknown;
};
type M8StoreRuntime = Record<string, unknown>;
type M8RouterRuntime = {
    navigate: (to: string, options?: unknown) => Promise<void> | void;
    buildPath: (input: M8BuildPathInput) => string;
};
type M8BuildPathInput = {
    mountPointId: M8MountPointId;
    moduleBasePath?: string;
    relativePath?: string;
    params?: Record<string, string | number | undefined>;
};
type M8NotificationRuntime = {
    success: (message: string) => void;
    error: (message: string) => void;
    warning: (message: string) => void;
    info: (message: string) => void;
};
type M8TelemetryRuntime = {
    event: (name: string, attributes?: Record<string, unknown>) => void;
    error: (error: unknown, attributes?: Record<string, unknown>) => void;
};
type M8ModuleRegistryRuntime = {
    isInstalled: (moduleId: M8ModuleId) => boolean;
    isEnabled: (moduleId: M8ModuleId) => boolean;
    getManifest: (moduleId: M8ModuleId) => M8ModuleManifest | undefined;
    getManifests: () => M8ModuleManifest[];
};

type M8RouteContribution = {
    id: string;
    path: string;
    title?: string;
    component: M8LazyRouteComponent;
    pendingComponent?: M8LazyRouteComponent;
    errorComponent?: M8LazyRouteComponent;
    notFoundComponent?: M8LazyRouteComponent;
    requiredPermissions?: M8Permission[];
    requiredFeatureFlags?: M8FeatureFlag[];
    requiredScopes?: M8ScopeId[];
    hidden?: boolean;
    loader?: M8RouteLoader;
    searchSchema?: M8SearchSchema;
    metadata?: M8Metadata;
};
type M8LazyRouteComponent = ComponentType<any> | LazyExoticComponent<ComponentType<any>> | (() => Promise<{
    default: ComponentType<any>;
}>);
type M8RouteLoader = (ctx: M8RouteLoaderContext) => M8MaybePromise<unknown>;
type M8RouteLoaderContext = {
    runtime: M8ModuleRuntimeContext;
    params: Record<string, string>;
    search: Record<string, unknown>;
};
type M8SearchSchema<TValue = unknown> = {
    parse: (value: unknown) => TValue;
    serialize?: (value: TValue) => unknown;
};

type M8NavigationContribution = {
    id: string;
    parentId?: string;
    title: string;
    description?: string;
    to: string;
    mountPointId?: M8MountPointId;
    icon?: ComponentType<M8ComponentIconProps>;
    order?: number;
    requiredPermissions?: M8Permission[];
    requiredFeatureFlags?: M8FeatureFlag[];
    badge?: M8NavigationBadge;
    exact?: boolean;
    children?: M8NavigationContribution[];
    metadata?: M8Metadata;
};
type M8NavigationBadge = {
    text: string;
    tone?: string;
};

type M8WidgetContribution = {
    id: string;
    slotId: M8SlotId;
    title: string;
    description?: string;
    component: M8LazyWidgetComponent;
    order?: number;
    requiredPermissions?: M8Permission[];
    requiredFeatureFlags?: M8FeatureFlag[];
    requiredScopes?: M8ScopeId[];
    metadata?: M8Metadata;
};
type M8LazyWidgetComponent = ComponentType<M8WidgetProps> | LazyExoticComponent<ComponentType<M8WidgetProps>> | (() => Promise<{
    default: ComponentType<M8WidgetProps>;
}>);
type M8WidgetProps = {
    runtime: M8ModuleRuntimeContext;
    slotId: M8SlotId;
    params?: Record<string, string>;
};
type M8WidgetResolver = (ctx: M8WidgetResolverContext) => M8MaybePromise<M8WidgetContribution[]>;
type M8WidgetResolverContext = {
    runtime: M8ModuleRuntimeContext;
    slotId: M8SlotId;
};

type M8ActionScope = 'global' | 'command-palette' | 'entity' | 'table-row' | 'page-header' | 'context-menu' | string;
type M8ActionContribution<TEntity = unknown> = {
    id: string;
    title: string;
    description?: string;
    scope: M8ActionScope;
    icon?: ComponentType<M8ComponentIconProps>;
    order?: number;
    requiredPermissions?: M8Permission[];
    requiredFeatureFlags?: M8FeatureFlag[];
    run: (ctx: M8ActionContext<TEntity>) => M8MaybePromise<void>;
    metadata?: M8Metadata;
};
type M8ActionContext<TEntity = unknown> = {
    runtime: M8ModuleRuntimeContext;
    entity?: TEntity;
    params?: Record<string, string>;
};

type M8SearchContribution = {
    id: string;
    title: string;
    scopes?: string[];
    requiredPermissions?: M8Permission[];
    requiredFeatureFlags?: M8FeatureFlag[];
    search: (ctx: M8SearchContext) => M8MaybePromise<M8SearchResult[]>;
    metadata?: M8Metadata;
};
type M8SearchContext = {
    runtime: M8ModuleRuntimeContext;
    query: string;
    limit: number;
    scope?: string;
};
type M8SearchResult = {
    id: string;
    title: string;
    subtitle?: string;
    to: string;
    icon?: ComponentType<M8ComponentIconProps>;
    metadata?: M8Metadata;
};

type M8BreadcrumbContribution = {
    resolve: (ctx: M8BreadcrumbContext) => M8MaybePromise<M8BreadcrumbItem[]>;
};
type M8BreadcrumbContext = {
    runtime: M8ModuleRuntimeContext;
    params: Record<string, string>;
    pathname: string;
};
type M8BreadcrumbItem = {
    title: string;
    to?: string;
};

type M8ModuleManifest = {
    id: M8ModuleId;
    title: string;
    description?: string;
    version: string;
    moduleApiVersion: string;
    kind?: M8ModuleKind;
    lifecycle?: M8ModuleLifecycle;
    basePath: string;
    mountPointId: M8MountPointId;
    order?: number;
    icon?: ComponentType<M8ComponentIconProps>;
    requiredCapabilities?: M8Capability[];
    requiredPermissions?: M8Permission[];
    requiredFeatureFlags?: M8FeatureFlag[];
    routes: M8RouteContribution[];
    navigation: M8NavigationContribution[];
    widgets?: M8WidgetContribution[];
    providers?: M8ProviderContribution[];
    actions?: M8ActionContribution[];
    search?: M8SearchContribution[];
    breadcrumbs?: M8BreadcrumbContribution;
    queryNamespace?: string;
    metadata?: M8Metadata;
};
type M8ProviderContribution = {
    id: string;
    component: ComponentType<{
        children: ReactNode;
        runtime: M8ModuleRuntimeContext;
    }>;
    order?: number;
    global?: boolean;
    metadata?: M8Metadata;
};
declare function defineModuleManifest<TManifest extends M8ModuleManifest>(manifest: TManifest): TManifest;

declare function buildNavigation(manifests: M8ModuleManifest[], runtime: M8ModuleRuntimeContext): M8NavigationContribution[];
declare function filterNavigationItems(items: M8NavigationContribution[], runtime: M8ModuleRuntimeContext): M8NavigationContribution[];
declare function isNavigationItemVisible(item: M8NavigationContribution, runtime: M8ModuleRuntimeContext): boolean;

declare function createScopeRuntime(current: Record<M8ScopeId, string | undefined>): M8ScopeRuntime;

declare function createRouterRuntime(input: {
    registry: M8PlatformRegistrySnapshot;
    scopes: M8ScopeRuntime;
    navigate: M8RouterRuntime['navigate'];
}): M8RouterRuntime;
declare function buildPathFromRegistry(registry: M8PlatformRegistrySnapshot, scopes: M8ScopeRuntime, input: M8BuildPathInput): string;

type M8RemoteModule = {
    getManifest: () => M8MaybePromise<M8ModuleManifest>;
    initialize?: (ctx: M8ModuleRuntimeContext) => M8MaybePromise<void>;
    dispose?: () => M8MaybePromise<void>;
};
declare function defineRemoteModule<TRemoteModule extends M8RemoteModule>(remoteModule: TRemoteModule): TRemoteModule;

type M8ModuleDefinition<TManifest extends M8ModuleManifest = M8ModuleManifest> = {
    manifest: TManifest;
    initialize?: (ctx: M8ModuleRuntimeContext) => M8MaybePromise<void>;
    dispose?: () => M8MaybePromise<void>;
};
declare function defineModule<TManifest extends M8ModuleManifest>(module: M8ModuleDefinition<TManifest>): M8ModuleDefinition<TManifest> & M8RemoteModule;

type M8ValidationSeverity = 'error' | 'warning';
type M8ValidationIssue = {
    severity: M8ValidationSeverity;
    code: string;
    message: string;
    path?: string;
};
type M8ValidationResult = {
    valid: boolean;
    issues: M8ValidationIssue[];
    errors: M8ValidationIssue[];
    warnings: M8ValidationIssue[];
};
declare function createValidationResult(issues: M8ValidationIssue[]): M8ValidationResult;
declare function errorIssue(code: string, message: string, path?: string): M8ValidationIssue;
declare function warningIssue(code: string, message: string, path?: string): M8ValidationIssue;

declare function validatePlatformRegistry(registry: M8PlatformRegistrySnapshot): M8ValidationResult;

type M8ManifestValidationOptions = {
    requireInstalledModule?: boolean;
    requireEnabledModule?: boolean;
    validateNavigationMountPoints?: boolean;
};
declare function validateModuleManifest(manifest: M8ModuleManifest, registry: M8PlatformRegistrySnapshot, options?: M8ManifestValidationOptions): M8ValidationResult;

declare function isModuleVisible(manifest: M8ModuleManifest, runtime: M8ModuleRuntimeContext): boolean;

declare function normalizePath(path: string | undefined | null): string;
declare function joinPaths(...parts: Array<string | undefined | null>): string;
declare function interpolatePathTemplate(template: string, params: Record<string, string | number | undefined>): string;
declare function stripTrailingSlash(path: string): string;

declare function compact<TValue>(items: Array<TValue | undefined | null | false>): TValue[];
declare function unique<TValue>(items: TValue[]): TValue[];

export { type M8ActionContext, type M8ActionContribution, type M8ActionScope, type M8ApiRuntime, type M8AuthRuntime, type M8BreadcrumbContext, type M8BreadcrumbContribution, type M8BreadcrumbItem, type M8BuildPathInput, type M8Capability, type M8ComponentIconProps, type M8Dictionary, type M8FeatureFlag, type M8FeatureFlagRuntime, type M8Id, type M8InstalledModuleDefinition, type M8LazyRouteComponent, type M8LazyWidgetComponent, type M8ManifestValidationOptions, type M8MaybePromise, type M8Metadata, type M8ModuleDefinition, type M8ModuleId, type M8ModuleKind, type M8ModuleLifecycle, type M8ModuleManifest, type M8ModuleRegistryRuntime, type M8ModuleRuntimeContext, type M8MountPointDefinition, type M8MountPointId, type M8NavigationBadge, type M8NavigationContribution, type M8NotificationRuntime, type M8Permission, type M8PermissionRuntime, type M8PlatformRegistrySnapshot, type M8ProviderContribution, type M8QueryRuntime, type M8RemoteModule, type M8RouteContribution, type M8RouteLoader, type M8RouteLoaderContext, type M8RouterRuntime, type M8ScopeDefinition, type M8ScopeId, type M8ScopeRuntime, type M8SearchContext, type M8SearchContribution, type M8SearchResult, type M8SearchSchema, type M8SlotDefinition, type M8SlotId, type M8StoreRuntime, type M8TelemetryRuntime, type M8ValidationIssue, type M8ValidationResult, type M8ValidationSeverity, type M8WidgetContribution, type M8WidgetProps, type M8WidgetResolver, type M8WidgetResolverContext, buildNavigation, buildPathFromRegistry, compact, createRouterRuntime, createScopeRuntime, createValidationResult, defineModule, defineModuleManifest, definePlatformRegistry, defineRemoteModule, errorIssue, filterNavigationItems, findInstalledModule, findMountPoint, findScope, findSlot, interpolatePathTemplate, isModuleVisible, isNavigationItemVisible, joinPaths, normalizePath, stripTrailingSlash, unique, validateModuleManifest, validatePlatformRegistry, warningIssue };
