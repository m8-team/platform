import type {NextRouteSpec} from '@json-render/next';

export type ExpressionValue =
  | null
  | string
  | number
  | boolean
  | readonly ExpressionValue[]
  | {readonly [key: string]: ExpressionValue}
  | {$state: string}
  | {$param: string}
  | {$query: string}
  | {$context: string}
  | {$literal: ExpressionValue};

export interface QueryBinding {
  /** Query ID from Query Registry. */
  query: string;
  input?: Readonly<Record<string, ExpressionValue>>;
  enabled?: ExpressionValue;
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

export type RouteSpec = NextRouteSpec & {
  navigation?: Navigation;
  access?: RouteAccess;
  queries?: Readonly<Record<string, QueryBinding>>;
};

export interface ModuleDependencyDefinition {
  required?: readonly string[];
  optional?: readonly string[];
}

export interface QueryDefinitionRef {
  readonly id: string;
}

export interface OperationDefinitionRef {
  readonly id: string;
}

export interface RuntimeContext {
  actor?: Readonly<{id: string; displayName?: string}>;
  organization?: Readonly<{id: string}>;
  workspace?: Readonly<{id: string}>;
  project?: Readonly<{id: string}>;
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
  readonly routes?: Readonly<Record<`/${string}`, RouteSpec>>;
  readonly queries?: readonly QueryDefinitionRef[];
  readonly operations?: readonly OperationDefinitionRef[];
}
