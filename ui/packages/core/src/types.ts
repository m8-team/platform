import type {NextRouteSpec} from '@json-render/next';
import type {PropExpression} from '@json-render/core';
import type {z} from 'zod';

export interface QueryBinding {
  readonly query: string;
  readonly input?: Readonly<Record<string, PropExpression>>;
  readonly enabled?: PropExpression<boolean>;
}

export interface Navigation {
  label: string;
  icon?: string;
  order?: number;
  hidden?: boolean;
}

export interface RouteAccess {
  permission?: string;
}

export type ModuleRouteSpec = NextRouteSpec & {
  navigation?: Navigation;
  access?: RouteAccess;
  queries?: Readonly<Record<string, QueryBinding>>;
};

export interface ModuleDependencyDefinition {
  required?: readonly string[];
  optional?: readonly string[];
}

export interface ModuleQueryContribution {
  readonly id: string;
  readonly input: z.ZodType;
  readonly output: z.ZodType;
  readonly queryKey?: (input: never, context: RuntimeContext) => readonly unknown[];
  readonly execute: (execution: {
    readonly input: never;
    readonly signal: AbortSignal;
    readonly context: RuntimeContext;
  }) => Promise<unknown>;
}

export interface ModuleOperationContribution {
  readonly id: string;
  readonly mode?: 'immediate' | 'long-running';
  readonly input: z.ZodType;
  readonly output: z.ZodType;
  readonly requiredPermission?: string;
  readonly invalidate?: readonly string[];
  readonly execute: (execution: {
    readonly input: never;
    readonly signal: AbortSignal;
    readonly context: RuntimeContext;
  }) => Promise<unknown>;
}

export interface RuntimeContext {
  actor?: Readonly<{id: string; displayName?: string}>;
  tenantId?: string;
  organizationId?: string;
  workspaceId?: string;
  projectId?: string;
  module?: Readonly<{id: string}>;
  permissions?: readonly string[];
  features?: readonly string[];
  edition?: string;
  readonly [key: string]: unknown;
}

export interface ModuleDefinition {
  readonly id: string;
  readonly title: string;
  readonly icon?: string;
  readonly order?: number;
  readonly dependencies?: ModuleDependencyDefinition;
  readonly routes?: Readonly<Record<`/${string}`, ModuleRouteSpec>>;
  readonly queries?: readonly ModuleQueryContribution[];
  readonly operations?: readonly ModuleOperationContribution[];
}
