import {
  CircularModuleDependencyError,
  DuplicateModuleError,
  DuplicateModuleOperationError,
  DuplicateModuleQueryError,
  MissingModuleDependencyError,
  RouteCollisionError,
  SelfModuleDependencyError,
} from './errors';
import {canonicalizeRoute, normalizePath} from './routes';
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

function compareModules(
  left: ModuleDefinition,
  right: ModuleDefinition,
): number {
  const order = (left.order ?? 0) - (right.order ?? 0);
  return order || left.id.localeCompare(right.id);
}

export class ModuleRegistry<
  TQuery extends ModuleContribution = ModuleContribution,
  TOperation extends ModuleContribution = ModuleContribution,
> {
  private readonly modulesById = new Map<
    string,
    ModuleDefinition<TQuery, TOperation>
  >();
  private readonly queryOwners = new Map<string, string>();
  private readonly operationOwners = new Map<string, string>();
  private readonly modules: readonly ModuleDefinition<TQuery, TOperation>[];
  private readonly routes: readonly OwnedModuleRoute[];
  private readonly queries: readonly OwnedModuleContribution<TQuery>[];
  private readonly operations: readonly OwnedModuleContribution<TOperation>[];

  constructor(modules: readonly ModuleDefinition<TQuery, TOperation>[]) {
    this.indexModules(modules);
    this.validateDependencies();
    this.modules = Object.freeze(this.sortModules());

    const ownership = this.indexOwnership();
    this.routes = Object.freeze(ownership.routes);
    this.queries = Object.freeze(ownership.queries);
    this.operations = Object.freeze(ownership.operations);
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

  private indexModules(
    modules: readonly ModuleDefinition<TQuery, TOperation>[],
  ): void {
    for (const moduleDefinition of modules) {
      if (this.modulesById.has(moduleDefinition.id)) {
        throw new DuplicateModuleError(moduleDefinition.id);
      }
      this.modulesById.set(moduleDefinition.id, moduleDefinition);
    }
  }

  private validateDependencies(): void {
    for (const moduleDefinition of this.modulesById.values()) {
      const dependencies = [
        ...(moduleDefinition.dependencies?.required ?? []),
        ...(moduleDefinition.dependencies?.optional ?? []),
      ];
      if (dependencies.includes(moduleDefinition.id)) {
        throw new SelfModuleDependencyError(moduleDefinition.id);
      }

      for (const dependencyId of moduleDefinition.dependencies?.required ?? []) {
        if (!this.modulesById.has(dependencyId)) {
          throw new MissingModuleDependencyError(moduleDefinition.id, dependencyId);
        }
      }
    }
  }

  private getPresentDependencies(
    moduleDefinition: ModuleDefinition<TQuery, TOperation>,
  ): readonly string[] {
    const dependencies = new Set(moduleDefinition.dependencies?.required ?? []);
    for (const dependencyId of moduleDefinition.dependencies?.optional ?? []) {
      if (this.modulesById.has(dependencyId)) {
        dependencies.add(dependencyId);
      }
    }
    return [...dependencies];
  }

  private sortModules(): ModuleDefinition<TQuery, TOperation>[] {
    const indegree = new Map<string, number>();
    const dependants = new Map<string, Set<string>>();

    for (const moduleDefinition of this.modulesById.values()) {
      const dependencies = this.getPresentDependencies(moduleDefinition);
      indegree.set(moduleDefinition.id, dependencies.length);
      for (const dependencyId of dependencies) {
        const dependencyDependants = dependants.get(dependencyId) ?? new Set<string>();
        dependencyDependants.add(moduleDefinition.id);
        dependants.set(dependencyId, dependencyDependants);
      }
    }

    const ready = [...this.modulesById.values()]
      .filter(moduleDefinition => indegree.get(moduleDefinition.id) === 0)
      .sort(compareModules);
    const sorted: ModuleDefinition<TQuery, TOperation>[] = [];

    while (ready.length > 0) {
      const moduleDefinition = ready.shift();
      if (!moduleDefinition) break;
      sorted.push(moduleDefinition);

      for (const dependantId of dependants.get(moduleDefinition.id) ?? []) {
        const nextIndegree = (indegree.get(dependantId) ?? 0) - 1;
        indegree.set(dependantId, nextIndegree);
        if (nextIndegree === 0) {
          const dependant = this.modulesById.get(dependantId);
          if (dependant) ready.push(dependant);
        }
      }
      ready.sort(compareModules);
    }

    if (sorted.length !== this.modulesById.size) {
      throw new CircularModuleDependencyError(this.findCircularDependency());
    }
    return sorted;
  }

  private findCircularDependency(): readonly string[] {
    const visited = new Set<string>();
    const visiting = new Set<string>();
    const path: string[] = [];

    const visit = (moduleId: string): readonly string[] | undefined => {
      if (visiting.has(moduleId)) {
        const cycleStart = path.indexOf(moduleId);
        return path.slice(cycleStart).concat(moduleId);
      }
      if (visited.has(moduleId)) return undefined;

      visiting.add(moduleId);
      path.push(moduleId);
      const moduleDefinition = this.modulesById.get(moduleId);
      if (moduleDefinition) {
        for (const dependencyId of [...this.getPresentDependencies(moduleDefinition)].sort()) {
          const cycle = visit(dependencyId);
          if (cycle) return cycle;
        }
      }
      path.pop();
      visiting.delete(moduleId);
      visited.add(moduleId);
      return undefined;
    };

    for (const moduleId of [...this.modulesById.keys()].sort()) {
      const cycle = visit(moduleId);
      if (cycle) return cycle;
    }
    return [];
  }

  private indexOwnership(): {
    routes: OwnedModuleRoute[];
    queries: OwnedModuleContribution<TQuery>[];
    operations: OwnedModuleContribution<TOperation>[];
  } {
    const routeOwners = new Map<string, string>();
    const routes: OwnedModuleRoute[] = [];
    const queries: OwnedModuleContribution<TQuery>[] = [];
    const operations: OwnedModuleContribution<TOperation>[] = [];

    for (const moduleDefinition of this.modules) {
      for (const [routePath, route] of Object.entries(moduleDefinition.routes ?? {})) {
        const path = normalizePath(routePath);
        const canonicalPath = canonicalizeRoute(path);
        const existingModuleId = routeOwners.get(canonicalPath);
        if (existingModuleId) {
          throw new RouteCollisionError(path, existingModuleId, moduleDefinition.id);
        }
        routeOwners.set(canonicalPath, moduleDefinition.id);
        routes.push(Object.freeze({moduleId: moduleDefinition.id, path, route}));
      }

      for (const contribution of moduleDefinition.queries ?? []) {
        const existingModuleId = this.queryOwners.get(contribution.id);
        if (existingModuleId) {
          throw new DuplicateModuleQueryError(
            contribution.id,
            existingModuleId,
            moduleDefinition.id,
          );
        }
        this.queryOwners.set(contribution.id, moduleDefinition.id);
        queries.push(Object.freeze({moduleId: moduleDefinition.id, contribution}));
      }

      for (const contribution of moduleDefinition.operations ?? []) {
        const existingModuleId = this.operationOwners.get(contribution.id);
        if (existingModuleId) {
          throw new DuplicateModuleOperationError(
            contribution.id,
            existingModuleId,
            moduleDefinition.id,
          );
        }
        this.operationOwners.set(contribution.id, moduleDefinition.id);
        operations.push(Object.freeze({moduleId: moduleDefinition.id, contribution}));
      }
    }

    return {routes, queries, operations};
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
