import type {ModuleDefinition, ModuleRouteSpec} from './types';
import {
  CircularModuleDependencyError,
  DuplicateModuleError,
  MissingModuleDependencyError,
  RouteCollisionError,
  InvalidModuleNamespaceError,
} from './errors';
import {canonicalizeRoute, normalizePath} from './routes';

export class ModuleRegistry<
  const TModules extends readonly ModuleDefinition[],
> {
  private readonly modulesById = new Map<string, ModuleDefinition>();

  constructor(private readonly modules: TModules) {
    this.validate();
  }

  getModules(): TModules {
    return this.modules;
  }

  getModule(moduleId: string): ModuleDefinition | undefined {
    return this.modulesById.get(moduleId);
  }

  getRoutes(): ReadonlyArray<Readonly<{moduleId: string; path: string; route: ModuleRouteSpec}>> {
    return this.modules.flatMap(moduleDefinition =>
      Object.entries(moduleDefinition.routes ?? {}).map(([path, route]) => ({
        moduleId: moduleDefinition.id,
        path: normalizePath(path),
        route,
      })),
    );
  }

  private validate(): void {
    this.validateModules();
    this.validateDependencies();
    this.validateCircularDependencies();
    this.validateContributionNamespaces();
    this.validateRoutes();
  }

  private validateModules(): void {
    for (const moduleDefinition of this.modules) {
      if (this.modulesById.has(moduleDefinition.id)) {
        throw new DuplicateModuleError(moduleDefinition.id);
      }

      this.modulesById.set(moduleDefinition.id, moduleDefinition);
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

  private validateContributionNamespaces(): void {
    for (const moduleDefinition of this.modules) {
      for (const query of moduleDefinition.queries ?? []) {
        if (!query.id.startsWith(`${moduleDefinition.id}.`)) {
          throw new InvalidModuleNamespaceError(moduleDefinition.id, query.id);
        }
      }
      for (const operation of moduleDefinition.operations ?? []) {
        if (!operation.id.startsWith(`${moduleDefinition.id}.`)) {
          throw new InvalidModuleNamespaceError(moduleDefinition.id, operation.id);
        }
      }
    }
  }

  private validateRoutes(): void {
    const routes = new Map<string, string>();

    for (const moduleDefinition of this.modules) {
      for (const routePath of Object.keys(moduleDefinition.routes ?? {})) {
        const fullPath = normalizePath(routePath);
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

}

export function defineModules<
  const TModules extends readonly ModuleDefinition[],
>(modules: TModules): ModuleRegistry<TModules> {
  return new ModuleRegistry(modules);
}
