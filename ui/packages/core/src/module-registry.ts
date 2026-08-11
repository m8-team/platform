import type {NextAppSpec} from '@json-render/next';
import type {
  M8ModuleDefinition,
  M8OperationDefinitionRef,
  M8QueryDefinitionRef,
} from './types';
import {
  BasePathCollisionError,
  CircularModuleDependencyError,
  DuplicateModuleError,
  DuplicateOperationError,
  DuplicateQueryError,
  MissingModuleDependencyError,
  RouteCollisionError,
} from './errors';
import {joinRoute, normalizePath} from './routes';
import {toNextRouteSpec} from './to-next-route';

export interface BuildNextAppSpecOptions {
  metadata?: NextAppSpec['metadata'];
  layouts?: NextAppSpec['layouts'];
  state?: NextAppSpec['state'];
}

export class ModuleRegistry<
  const TModules extends readonly M8ModuleDefinition[],
> {
  private readonly modulesById = new Map<string, M8ModuleDefinition>();

  private readonly queriesById = new Map<
    string,
    {moduleId: string; definition: M8QueryDefinitionRef}
  >();

  private readonly operationsById = new Map<
    string,
    {moduleId: string; definition: M8OperationDefinitionRef}
  >();

  constructor(private readonly modules: TModules) {
    this.validate();
  }

  getModules(): TModules {
    return this.modules;
  }

  getModule(moduleId: string): M8ModuleDefinition | undefined {
    return this.modulesById.get(moduleId);
  }

  getQuery(queryId: string): M8QueryDefinitionRef | undefined {
    return this.queriesById.get(queryId)?.definition;
  }

  getOperation(operationId: string): M8OperationDefinitionRef | undefined {
    return this.operationsById.get(operationId)?.definition;
  }

  buildNextAppSpec(options: BuildNextAppSpecOptions = {}): NextAppSpec {
    const routes: NextAppSpec['routes'] = {};

    for (const moduleDefinition of this.modules) {
      for (const [routePath, route] of Object.entries(moduleDefinition.routes ?? {})) {
        const fullPath = joinRoute(moduleDefinition.basePath, routePath);
        routes[fullPath] = toNextRouteSpec(route);
      }
    }

    const spec: NextAppSpec = {routes};

    if (options.metadata !== undefined) {
      spec.metadata = options.metadata;
    }
    if (options.layouts !== undefined) {
      spec.layouts = options.layouts;
    }
    if (options.state !== undefined) {
      spec.state = options.state;
    }

    return spec;
  }

  private validate(): void {
    this.validateModules();
    this.validateDependencies();
    this.validateCircularDependencies();
    this.validateQueries();
    this.validateOperations();
    this.validateRoutes();
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
        const existingModuleId = routes.get(fullPath);

        if (existingModuleId) {
          throw new RouteCollisionError(
            fullPath,
            existingModuleId,
            moduleDefinition.id,
          );
        }

        routes.set(fullPath, moduleDefinition.id);
      }
    }
  }
}

export function defineModules<
  const TModules extends readonly M8ModuleDefinition[],
>(modules: TModules): ModuleRegistry<TModules> {
  return new ModuleRegistry(modules);
}
