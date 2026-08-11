import type {
  ModuleDefinition,
  OperationDefinitionRef,
  QueryDefinitionRef,
} from './types';
import {
  BasePathCollisionError,
  CircularModuleDependencyError,
  DuplicateModuleError,
  DuplicateOperationError,
  DuplicateQueryError,
  MissingModuleDependencyError,
  RouteCollisionError,
  UnknownRouteOperationError,
  UnknownRouteQueryError,
  InvalidModuleNamespaceError,
} from './errors';
import {canonicalizeRoute, joinRoute, normalizePath} from './routes';

export class ModuleRegistry<
  const TModules extends readonly ModuleDefinition[],
> {
  private readonly modulesById = new Map<string, ModuleDefinition>();

  private readonly queriesById = new Map<
    string,
    {moduleId: string; definition: QueryDefinitionRef}
  >();

  private readonly operationsById = new Map<
    string,
    {moduleId: string; definition: OperationDefinitionRef}
  >();

  constructor(private readonly modules: TModules) {
    this.validate();
  }

  getModules(): TModules {
    return this.modules;
  }

  getModule(moduleId: string): ModuleDefinition | undefined {
    return this.modulesById.get(moduleId);
  }

  getQuery(queryId: string): QueryDefinitionRef | undefined {
    return this.queriesById.get(queryId)?.definition;
  }

  getQueries(): readonly QueryDefinitionRef[] {
    return [...this.queriesById.values()].map(value => value.definition);
  }

  getOperations(): readonly OperationDefinitionRef[] {
    return [...this.operationsById.values()].map(value => value.definition);
  }

  getRoutes(): ReadonlyArray<Readonly<{moduleId: string; path: string; route: import('./types').RouteSpec}>> {
    return this.modules.flatMap(moduleDefinition =>
      Object.entries(moduleDefinition.routes ?? {}).map(([path, route]) => ({
        moduleId: moduleDefinition.id,
        path: joinRoute(moduleDefinition.basePath, path),
        route,
      })),
    );
  }

  getOperation(operationId: string): OperationDefinitionRef | undefined {
    return this.operationsById.get(operationId)?.definition;
  }

  private validate(): void {
    this.validateModules();
    this.validateDependencies();
    this.validateCircularDependencies();
    this.validateQueries();
    this.validateOperations();
    this.validateRoutes();
    this.validateRouteReferences();
  }

  private validateModules(): void {
    const basePaths = new Map<string, string>();

    for (const moduleDefinition of this.modules) {
      if (this.modulesById.has(moduleDefinition.id)) {
        throw new DuplicateModuleError(moduleDefinition.id);
      }

      this.modulesById.set(moduleDefinition.id, moduleDefinition);

      const normalizedBasePath = normalizePath(moduleDefinition.basePath);
      const existingModuleId = basePaths.get(normalizedBasePath);

      if (existingModuleId) {
        throw new BasePathCollisionError(
          normalizedBasePath,
          existingModuleId,
          moduleDefinition.id,
        );
      }

      basePaths.set(normalizedBasePath, moduleDefinition.id);
    }
  }

  private validateDependencies(): void {
    for (const moduleDefinition of this.modules) {
      for (const dependencyId of moduleDefinition.dependencies?.required ?? []) {
        if (!this.modulesById.has(dependencyId)) {
          throw new MissingModuleDependencyError(moduleDefinition.id, dependencyId);
        }
      }
    }
  }

  private validateCircularDependencies(): void {
    const visited = new Set<string>();
    const visiting = new Set<string>();
    const stack: string[] = [];

    const visit = (moduleId: string): void => {
      if (visiting.has(moduleId)) {
        const startIndex = stack.indexOf(moduleId);
        const cycle = stack.slice(startIndex).concat(moduleId);
        throw new CircularModuleDependencyError(cycle);
      }

      if (visited.has(moduleId)) {
        return;
      }

      visiting.add(moduleId);
      stack.push(moduleId);

      const moduleDefinition = this.modulesById.get(moduleId);
      for (const dependencyId of moduleDefinition?.dependencies?.required ?? []) {
        visit(dependencyId);
      }

      stack.pop();
      visiting.delete(moduleId);
      visited.add(moduleId);
    };

    for (const moduleDefinition of this.modules) {
      visit(moduleDefinition.id);
    }
  }

  private validateQueries(): void {
    for (const moduleDefinition of this.modules) {
      for (const query of moduleDefinition.queries ?? []) {
        const existing = this.queriesById.get(query.id);

        if (existing) {
          throw new DuplicateQueryError(
            query.id,
            existing.moduleId,
            moduleDefinition.id,
          );
        }
        if (!query.id.startsWith(`${moduleDefinition.id}.`)) {
          throw new InvalidModuleNamespaceError(moduleDefinition.id, query.id);
        }

        this.queriesById.set(query.id, {
          moduleId: moduleDefinition.id,
          definition: query,
        });
      }
    }
  }

  private validateOperations(): void {
    for (const moduleDefinition of this.modules) {
      for (const operation of moduleDefinition.operations ?? []) {
        const existing = this.operationsById.get(operation.id);

        if (existing) {
          throw new DuplicateOperationError(
            operation.id,
            existing.moduleId,
            moduleDefinition.id,
          );
        }
        if (!operation.id.startsWith(`${moduleDefinition.id}.`)) {
          throw new InvalidModuleNamespaceError(moduleDefinition.id, operation.id);
        }

        this.operationsById.set(operation.id, {
          moduleId: moduleDefinition.id,
          definition: operation,
        });
      }
    }
  }

  private validateRoutes(): void {
    const routes = new Map<string, string>();

    for (const moduleDefinition of this.modules) {
      for (const routePath of Object.keys(moduleDefinition.routes ?? {})) {
        const fullPath = joinRoute(moduleDefinition.basePath, routePath);
        const canonicalPath = canonicalizeRoute(fullPath);
        const existingModuleId = routes.get(canonicalPath);

        if (existingModuleId) {
          throw new RouteCollisionError(
            fullPath,
            existingModuleId,
            moduleDefinition.id,
          );
        }

        routes.set(canonicalPath, moduleDefinition.id);
      }
    }
  }

  private validateRouteReferences(): void {
    for (const {path, route} of this.getRoutes()) {
      for (const binding of Object.values(route.queries ?? {})) {
        if (!this.queriesById.has(binding.query)) throw new UnknownRouteQueryError(path, binding.query);
      }
      const visit = (value: unknown): void => {
        if (Array.isArray(value)) return value.forEach(visit);
        if (value === null || typeof value !== 'object') return;
        const record = value as Record<string, unknown>;
        if (record.action === 'executeOperation' && record.params && typeof record.params === 'object') {
          const operationId = (record.params as Record<string, unknown>).operation;
          if (typeof operationId === 'string' && !this.operationsById.has(operationId)) {
            throw new UnknownRouteOperationError(path, operationId);
          }
        }
        Object.values(record).forEach(visit);
      };
      visit(route.page);
    }
  }
}

export function defineModules<
  const TModules extends readonly ModuleDefinition[],
>(modules: TModules): ModuleRegistry<TModules> {
  return new ModuleRegistry(modules);
}
