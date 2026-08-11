import type {NextAppSpec} from '@json-render/next';

export type NextRouteSpec = NextAppSpec['routes'][string];

export type M8InputValue =
  | null
  | string
  | number
  | boolean
  | M8InputValue[]
  | {[key: string]: M8InputValue}
  | {$state: string}
  | {$param: string}
  | {$context: string};

export interface M8QueryBinding {
  /** Query ID from Query Registry. */
  query: string;
  input?: Record<string, M8InputValue>;
  enabled?: M8InputValue;
}

export interface M8RouteNavigation {
  label: string;
  icon?: string;
  order?: number;
  hidden?: boolean;
}

export interface M8RouteAccess {
  permission?: string;
  feature?: string;
}

export type M8RouteSpec = NextRouteSpec & {
  navigation?: M8RouteNavigation;
  access?: M8RouteAccess;
  queries?: Record<string, M8QueryBinding>;
};

export interface M8ModuleDependencyDefinition {
  required?: readonly string[];
  optional?: readonly string[];
}

export interface M8ModuleAccess {
  permission?: string;
  feature?: string;
  editions?: readonly string[];
}

export interface M8QueryDefinitionRef {
  id: string;
}

export interface M8OperationDefinitionRef {
  id: string;
}

export interface ModuleDefinition {
  id: string;
  title: string;
  basePath: `/${string}`;
  icon?: string;
  order?: number;
  access?: M8ModuleAccess;
  dependencies?: M8ModuleDependencyDefinition;
  routes?: Record<`/${string}`, M8RouteSpec>;
  queries?: readonly M8QueryDefinitionRef[];
  operations?: readonly M8OperationDefinitionRef[];
}
