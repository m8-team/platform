import type {NextAppSpec} from '@json-render/next';

export type NextRouteSpec = NextAppSpec['routes'][string];

export type InputValue =
  | null
  | string
  | number
  | boolean
  | InputValue[]
  | {
  [key: string]: InputValue;
}
  | {
  $state: string;
}
  | {
  $param: string;
}
  | {
  $context: string;
};

export interface QueryBinding {
  /**
   * Query ID from Query Registry.
   *
   * Example:
   * resource-manager.projects.list
   */
  query: string;

  /**
   * Declarative query input.
   */
  input?: Record<string, InputValue>;

  /**
   * Optional conditional execution.
   */
  enabled?: InputValue;
}

export interface RouteNavigation {
  label: string;

  icon?: string;

  order?: number;

  hidden?: boolean;
}

export interface RouteAccess {
  permission?: string;

  feature?: string;
}

export type RouteSpec =
  NextRouteSpec & {
  /**
   *  navigation metadata.
   */
  navigation?: RouteNavigation;

  /**
   * Route-level authorization.
   */
  access?: RouteAccess;

  /**
   * Reactive queries required by route.
   */
  queries?: Record<
    string,
    QueryBinding
  >;
};

export interface ModuleDependencyDefinition {
  required?: readonly string[];

  optional?: readonly string[];
}

export interface ModuleAccess {
  /**
   * Module-level permission.
   */
  permission?: string;

  /**
   * Feature flag / installed capability.
   */
  feature?: string;

  /**
   * Optional edition gating.
   */
  editions?: readonly string[];
}

/**
 * Minimal contract required from registered
 * query definitions.
 */
export interface QueryDefinitionRef {
  id: string;
}

/**
 * Minimal contract required from registered
 * operation definitions.
 */
export interface OperationDefinitionRef {
  id: string;
}

export interface ModuleDefinition {
  /**
   * Stable module identifier.
   *
   * Examples:
   * resource-manager
   * iam
   * gateway
   */
  id: string;

  /**
   * Human readable module name.
   */
  title: string;

  /**
   * URL prefix.
   *
   * Example:
   * /resource-manager
   */
  basePath: `/${string}`;

  icon?: string;

  order?: number;

  access?: ModuleAccess;

  dependencies?: ModuleDependencyDefinition;

  /**
   * Routes are relative to basePath.
   *
   * Examples:
   * /
   * /projects
   * /projects/[projectId]
   */
  routes: Record<
    `/${string}`,
    RouteSpec
  >;

  queries?:
    readonly QueryDefinitionRef[];

  operations?:
    readonly OperationDefinitionRef[];
}
