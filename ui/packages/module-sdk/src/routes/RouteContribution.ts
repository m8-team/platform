import type {ComponentType, LazyExoticComponent} from 'react';
import type {
  M8FeatureFlag,
  M8MaybePromise,
  M8Metadata,
  M8Permission,
  M8ScopeId,
} from '../primitives';
import type {M8ModuleRuntimeContext} from '../runtime/ModuleRuntimeContext';

export type M8RouteContribution = {
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

export type M8LazyRouteComponent =
  | ComponentType<any>
  | LazyExoticComponent<ComponentType<any>>
  | (() => Promise<{default: ComponentType<any>}>);

export type M8RouteLoader = (ctx: M8RouteLoaderContext) => M8MaybePromise<unknown>;

export type M8RouteLoaderContext = {
  runtime: M8ModuleRuntimeContext;
  params: Record<string, string>;
  search: Record<string, unknown>;
};

export type M8SearchSchema<TValue = unknown> = {
  parse: (value: unknown) => TValue;
  serialize?: (value: TValue) => unknown;
};
