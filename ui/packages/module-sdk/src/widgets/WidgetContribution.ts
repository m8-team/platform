import type {ComponentType, LazyExoticComponent} from 'react';
import type {
  M8FeatureFlag,
  M8MaybePromise,
  M8Metadata,
  M8Permission,
  M8ScopeId,
  M8SlotId,
} from '../primitives';
import type {M8ModuleRuntimeContext} from '../runtime/ModuleRuntimeContext';

export type M8WidgetContribution = {
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

export type M8LazyWidgetComponent =
  | ComponentType<M8WidgetProps>
  | LazyExoticComponent<ComponentType<M8WidgetProps>>
  | (() => Promise<{default: ComponentType<M8WidgetProps>}>);

export type M8WidgetProps = {
  runtime: M8ModuleRuntimeContext;
  slotId: M8SlotId;
  params?: Record<string, string>;
};

export type M8WidgetResolver = (ctx: M8WidgetResolverContext) => M8MaybePromise<M8WidgetContribution[]>;

export type M8WidgetResolverContext = {
  runtime: M8ModuleRuntimeContext;
  slotId: M8SlotId;
};
