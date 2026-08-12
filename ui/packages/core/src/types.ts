import type {NextRouteSpec} from '@json-render/next';
import type {PropExpression} from '@json-render/core';

export interface ModuleContribution {
  readonly id: string;
}

export interface QueryBinding {
  readonly query: string;
  readonly input?: Readonly<Record<string, PropExpression>>;
  readonly enabled?: PropExpression<boolean>;
}

export interface Navigation {
  readonly label: string;
  readonly icon?: string;
  readonly order?: number;
  readonly hidden?: boolean;
}

export interface RouteAccess {
  readonly permission?: string;
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

export interface ModuleDefinition<
  TQuery extends ModuleContribution = ModuleContribution,
  TOperation extends ModuleContribution = ModuleContribution,
> {
  readonly id: string;
  readonly title: string;
  readonly icon?: string;
  readonly order?: number;
  readonly dependencies?: ModuleDependencyDefinition;
  readonly routes?: Readonly<Record<`/${string}`, ModuleRouteSpec>>;
  readonly queries?: readonly TQuery[];
  readonly operations?: readonly TOperation[];
}

export interface OwnedModuleContribution<TContribution extends ModuleContribution> {
  readonly moduleId: string;
  readonly contribution: TContribution;
}
