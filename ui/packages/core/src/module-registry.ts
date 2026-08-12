import {
  DuplicateModuleError,
  DuplicateModuleOperationError,
  DuplicateModuleQueryError,
  InvalidRouteError,
  RouteCollisionError,
  UnknownModuleQueryReferenceError,
} from './errors';
import {canonicalizeRoute, isValidRoute, normalizePath} from './routes';
import type {
  ModuleContribution,
  ModuleDefinition,
  ModuleRouteSpec,
  OwnedModuleContribution,
} from './types';

export interface OwnedModuleRoute {
  readonly moduleId: string;
  readonly path: string;
  readonly route: ModuleRouteSpec;
}

interface CollectedModuleRoute extends OwnedModuleRoute {
  readonly canonicalPath: string;
}

interface CollectedContributions<
  TQuery extends ModuleContribution,
  TOperation extends ModuleContribution,
> {
  readonly routes: CollectedModuleRoute[];
  readonly queries: OwnedModuleContribution<TQuery>[];
  readonly operations: OwnedModuleContribution<TOperation>[];
}

interface ContributionOwnership {
  readonly routeOwners: Map<string, string>;
  readonly queryOwners: Map<string, string>;
  readonly operationOwners: Map<string, string>;
}

function compareText(left: string, right: string): number {
  if (left < right) return -1;
  if (left > right) return 1;
  return 0;
}

function indexModules<
  TQuery extends ModuleContribution,
  TOperation extends ModuleContribution,
>(
  definitions: readonly ModuleDefinition<TQuery, TOperation>[],
): readonly ModuleDefinition<TQuery, TOperation>[] {
  const modulesById = new Map<
    string,
    ModuleDefinition<TQuery, TOperation>
  >();

  for (const moduleDefinition of definitions) {
    if (modulesById.has(moduleDefinition.id)) {
      throw new DuplicateModuleError(moduleDefinition.id);
    }
    modulesById.set(moduleDefinition.id, moduleDefinition);
  }

  return [...modulesById.values()].sort((left, right) =>
    compareText(left.id, right.id));
}

function collectContributions<
  TQuery extends ModuleContribution,
  TOperation extends ModuleContribution,
>(
  modules: readonly ModuleDefinition<TQuery, TOperation>[],
): CollectedContributions<TQuery, TOperation> {
  const routes: CollectedModuleRoute[] = [];
  const queries: OwnedModuleContribution<TQuery>[] = [];
  const operations: OwnedModuleContribution<TOperation>[] = [];

  for (const moduleDefinition of modules) {
    for (const [routePath, route] of Object.entries(
      moduleDefinition.routes ?? {},
    )) {
      if (!isValidRoute(routePath)) {
        throw new InvalidRouteError(routePath, moduleDefinition.id);
      }
      const path = normalizePath(routePath);
      routes.push({
        moduleId: moduleDefinition.id,
        path,
        canonicalPath: canonicalizeRoute(path),
        route,
      });
    }

    for (const contribution of moduleDefinition.queries ?? []) {
      queries.push({moduleId: moduleDefinition.id, contribution});
    }

    for (const contribution of moduleDefinition.operations ?? []) {
      operations.push({moduleId: moduleDefinition.id, contribution});
    }
  }

  routes.sort((left, right) =>
    compareText(left.canonicalPath, right.canonicalPath) ||
    compareText(left.moduleId, right.moduleId) ||
    compareText(left.path, right.path));
  queries.sort((left, right) =>
    compareText(left.contribution.id, right.contribution.id) ||
    compareText(left.moduleId, right.moduleId));
  operations.sort((left, right) =>
    compareText(left.contribution.id, right.contribution.id) ||
    compareText(left.moduleId, right.moduleId));

  return {routes, queries, operations};
}

function validateContributions<
  TQuery extends ModuleContribution,
  TOperation extends ModuleContribution,
>(
  contributions: CollectedContributions<TQuery, TOperation>,
): ContributionOwnership {
  const routeOwners = new Map<string, string>();
  const queryOwners = new Map<string, string>();
  const operationOwners = new Map<string, string>();

  for (const route of contributions.routes) {
    const existingModuleId = routeOwners.get(route.canonicalPath);
    if (existingModuleId) {
      throw new RouteCollisionError(
        route.path,
        existingModuleId,
        route.moduleId,
      );
    }
    routeOwners.set(route.canonicalPath, route.moduleId);
  }

  for (const {moduleId, contribution} of contributions.queries) {
    const existingModuleId = queryOwners.get(contribution.id);
    if (existingModuleId) {
      throw new DuplicateModuleQueryError(
        contribution.id,
        existingModuleId,
        moduleId,
      );
    }
    queryOwners.set(contribution.id, moduleId);
  }

  for (const {moduleId, contribution} of contributions.operations) {
    const existingModuleId = operationOwners.get(contribution.id);
    if (existingModuleId) {
      throw new DuplicateModuleOperationError(
        contribution.id,
        existingModuleId,
        moduleId,
      );
    }
    operationOwners.set(contribution.id, moduleId);
  }

  for (const {moduleId, path, route} of contributions.routes) {
    for (const binding of Object.values(route.queries ?? {})) {
      if (!queryOwners.has(binding.query)) {
        throw new UnknownModuleQueryReferenceError(
          moduleId,
          path,
          binding.query,
        );
      }
    }
  }

  return {routeOwners, queryOwners, operationOwners};
}

export class ModuleRegistry<
  TQuery extends ModuleContribution = ModuleContribution,
  TOperation extends ModuleContribution = ModuleContribution,
> {
  private readonly modulesById: Map<
    string,
    ModuleDefinition<TQuery, TOperation>
  >;
  private readonly routeOwners: Map<string, string>;
  private readonly queryOwners: Map<string, string>;
  private readonly operationOwners: Map<string, string>;
  private readonly modules: readonly ModuleDefinition<TQuery, TOperation>[];
  private readonly routes: readonly OwnedModuleRoute[];
  private readonly queries: readonly OwnedModuleContribution<TQuery>[];
  private readonly operations: readonly OwnedModuleContribution<TOperation>[];

  constructor(definitions: readonly ModuleDefinition<TQuery, TOperation>[]) {
    const modules = indexModules(definitions);
    const contributions = collectContributions(modules);
    const ownership = validateContributions(contributions);

    this.modules = Object.freeze([...modules]);
    this.modulesById = new Map(modules.map(moduleDefinition => [
      moduleDefinition.id,
      moduleDefinition,
    ]));
    this.routeOwners = ownership.routeOwners;
    this.queryOwners = ownership.queryOwners;
    this.operationOwners = ownership.operationOwners;
    this.routes = Object.freeze(contributions.routes.map(({
      canonicalPath: _canonicalPath,
      ...route
    }) => Object.freeze(route)));
    this.queries = Object.freeze(contributions.queries.map(contribution =>
      Object.freeze(contribution)));
    this.operations = Object.freeze(contributions.operations.map(contribution =>
      Object.freeze(contribution)));
  }

  getModules(): readonly ModuleDefinition<TQuery, TOperation>[] {
    return this.modules;
  }

  getModule(moduleId: string): ModuleDefinition<TQuery, TOperation> | undefined {
    return this.modulesById.get(moduleId);
  }

  getRoutes(): readonly OwnedModuleRoute[] {
    return this.routes;
  }

  getRouteOwner(routePath: string): string | undefined {
    return this.routeOwners.get(canonicalizeRoute(routePath));
  }

  getQueryContributions(): readonly OwnedModuleContribution<TQuery>[] {
    return this.queries;
  }

  getOperationContributions(): readonly OwnedModuleContribution<TOperation>[] {
    return this.operations;
  }

  getQueryOwner(queryId: string): string | undefined {
    return this.queryOwners.get(queryId);
  }

  getOperationOwner(operationId: string): string | undefined {
    return this.operationOwners.get(operationId);
  }
}

export function defineModules<
  const TQuery extends ModuleContribution = never,
  const TOperation extends ModuleContribution = never,
>(
  modules: readonly ModuleDefinition<TQuery, TOperation>[],
): ModuleRegistry<TQuery, TOperation> {
  return new ModuleRegistry(modules);
}
