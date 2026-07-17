import type {ComponentType, ReactNode} from 'react';
import type {
  M8Capability,
  M8ComponentIconProps,
  M8FeatureFlag,
  M8Metadata,
  M8ModuleId,
  M8ModuleKind,
  M8ModuleLifecycle,
  M8MountPointId,
  M8Permission,
} from '../primitives';
import type {M8RouteContribution} from '../routes/RouteContribution';
import type {M8NavigationContribution} from '../navigation/NavContribution';
import type {M8WidgetContribution} from '../widgets/WidgetContribution';
import type {M8ActionContribution} from '../actions/ActionContribution';
import type {M8SearchContribution} from '../search/SearchContribution';
import type {M8BreadcrumbContribution} from '../breadcrumbs/BreadcrumbContribution';
import type {M8ModuleRuntimeContext} from '../runtime/ModuleRuntimeContext';

export type M8ModuleManifest = {
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

export type M8ProviderContribution = {
  id: string;
  component: ComponentType<{
    children: ReactNode;
    runtime: M8ModuleRuntimeContext;
  }>;
  order?: number;
  global?: boolean;
  metadata?: M8Metadata;
};

export function defineModuleManifest<TManifest extends M8ModuleManifest>(
  manifest: TManifest,
): TManifest {
  return manifest;
}
