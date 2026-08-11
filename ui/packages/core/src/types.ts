import type {NextRouteSpec} from '@json-render/next';

export type M8ExpressionValue =
  | null
  | string
  | number
  | boolean
  | readonly M8ExpressionValue[]
  | {readonly [key: string]: M8ExpressionValue}
  | {$state: string}
  | {$param: string}
  | {$query: string}
  | {$context: string}
  | {$literal: M8ExpressionValue};

export interface M8QueryBinding {
  /** Query ID from Query Registry. */
  query: string;
  input?: Readonly<Record<string, M8ExpressionValue>>;
  enabled?: M8ExpressionValue;
}

export interface M8Navigation {
  label: string;
  icon?: string;
  order?: number;
  hidden?: boolean;
}

export interface M8RouteAccess {
  permission?: string;
}

export interface M8Availability {
  feature?: string;
  editions?: readonly string[];
}

export type M8RouteSpec = NextRouteSpec & {
  navigation?: M8Navigation;
  access?: M8RouteAccess;
  availability?: M8Availability;
  queries?: Readonly<Record<string, M8QueryBinding>>;
};

export interface M8ModuleDependencyDefinition {
  required?: readonly string[];
  optional?: readonly string[];
}

export interface M8QueryDefinitionRef {
  readonly id: string;
}

export interface M8OperationDefinitionRef {
  readonly id: string;
}

export interface M8RuntimeContext {
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

export interface M8ModuleDefinition {
  readonly id: string;
  readonly title: string;
  readonly basePath: `/${string}`;
  readonly icon?: string;
  readonly order?: number;
  readonly availability?: M8Availability;
  readonly dependencies?: M8ModuleDependencyDefinition;
  readonly routes?: Readonly<Record<`/${string}`, M8RouteSpec>>;
  readonly queries?: readonly M8QueryDefinitionRef[];
  readonly operations?: readonly M8OperationDefinitionRef[];
}

/** @deprecated Use M8ModuleDefinition. */
export type ModuleDefinition = M8ModuleDefinition;
