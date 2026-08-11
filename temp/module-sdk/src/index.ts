export type * from './primitives';

export type * from './registry/PlatformRegistry';
export {
  definePlatformRegistry,
  findInstalledModule,
  findMountPoint,
  findScope,
  findSlot,
} from './registry/PlatformRegistry';

export type * from './manifest/ModuleManifest';
export {defineModuleManifest} from './manifest/ModuleManifest';

export type * from './routes/RouteContribution';
export type * from './navigation/NavContribution';
export {buildNavigation, filterNavigationItems, isNavigationItemVisible} from './navigation/buildNavigation';

export type * from './widgets/WidgetContribution';
export type * from './actions/ActionContribution';
export type * from './search/SearchContribution';
export type * from './breadcrumbs/BreadcrumbContribution';

export type * from './runtime/ModuleRuntimeContext';
export {createScopeRuntime} from './runtime/createScopeRuntime';
export {buildPathFromRegistry, createRouterRuntime} from './runtime/createRouterRuntime';

export type * from './module/RemoteModule';
export {defineRemoteModule} from './module/RemoteModule';
export type * from './module/defineModule';
export {defineModule} from './module/defineModule';

export type * from './validation/ValidationResult';
export {createValidationResult, errorIssue, warningIssue} from './validation/ValidationResult';
export {validatePlatformRegistry} from './validation/validatePlatformRegistry';
export type {M8ManifestValidationOptions} from './validation/validateModuleManifest';
export {validateModuleManifest} from './validation/validateModuleManifest';
export {isModuleVisible} from './validation/isModuleVisible';

export {interpolatePathTemplate, joinPaths, normalizePath, stripTrailingSlash} from './utils/path';
export {compact, unique} from './utils/object';
