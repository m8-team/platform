import type {ComponentType} from 'react';
import type {
  M8FeatureFlag,
  M8MaybePromise,
  M8Metadata,
  M8Permission,
} from '../primitives';
import type {M8ComponentIconProps} from '../primitives';
import type {M8ModuleRuntimeContext} from '../runtime/ModuleRuntimeContext';

export type M8ActionScope =
  | 'global'
  | 'command-palette'
  | 'entity'
  | 'table-row'
  | 'page-header'
  | 'context-menu'
  | string;

export type M8ActionContribution<TEntity = unknown> = {
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

export type M8ActionContext<TEntity = unknown> = {
  runtime: M8ModuleRuntimeContext;
  entity?: TEntity;
  params?: Record<string, string>;
};
