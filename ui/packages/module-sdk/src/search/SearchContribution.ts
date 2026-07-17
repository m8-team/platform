import type {ComponentType} from 'react';
import type {
  M8FeatureFlag,
  M8MaybePromise,
  M8Metadata,
  M8Permission,
} from '../primitives';
import type {M8ComponentIconProps} from '../primitives';
import type {M8ModuleRuntimeContext} from '../runtime/ModuleRuntimeContext';

export type M8SearchContribution = {
  id: string;
  title: string;
  scopes?: string[];
  requiredPermissions?: M8Permission[];
  requiredFeatureFlags?: M8FeatureFlag[];
  search: (ctx: M8SearchContext) => M8MaybePromise<M8SearchResult[]>;
  metadata?: M8Metadata;
};

export type M8SearchContext = {
  runtime: M8ModuleRuntimeContext;
  query: string;
  limit: number;
  scope?: string;
};

export type M8SearchResult = {
  id: string;
  title: string;
  subtitle?: string;
  to: string;
  icon?: ComponentType<M8ComponentIconProps>;
  metadata?: M8Metadata;
};
