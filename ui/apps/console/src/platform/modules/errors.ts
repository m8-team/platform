export class ModuleRegistryError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'ModuleRegistryError';
  }
}

export class DuplicateModuleError extends ModuleRegistryError {
  constructor(moduleId: string) {
    super(`Duplicate module id: "${moduleId}"`);
    this.name = 'DuplicateModuleError';
  }
}

export class DuplicateQueryError extends ModuleRegistryError {
  constructor(queryId: string, firstModuleId: string, secondModuleId: string) {
    super(
      `Duplicate query id: "${queryId}". Declared by modules "${firstModuleId}" and "${secondModuleId}".`,
    );
    this.name = 'DuplicateQueryError';
  }
}

export class DuplicateOperationError extends ModuleRegistryError {
  constructor(operationId: string, firstModuleId: string, secondModuleId: string) {
    super(
      `Duplicate operation id: "${operationId}". Declared by modules "${firstModuleId}" and "${secondModuleId}".`,
    );
    this.name = 'DuplicateOperationError';
  }
}

export class MissingModuleDependencyError extends ModuleRegistryError {
  constructor(moduleId: string, dependencyId: string) {
    super(`Module "${moduleId}" requires missing module "${dependencyId}".`);
    this.name = 'MissingModuleDependencyError';
  }
}

export class CircularModuleDependencyError extends ModuleRegistryError {
  constructor(path: readonly string[]) {
    super(`Circular module dependency: ${path.join(' -> ')}`);
    this.name = 'CircularModuleDependencyError';
  }
}

export class RouteCollisionError extends ModuleRegistryError {
  constructor(route: string, firstModuleId: string, secondModuleId: string) {
    super(
      `Route collision: "${route}". Declared by modules "${firstModuleId}" and "${secondModuleId}".`,
    );
    this.name = 'RouteCollisionError';
  }
}

export class BasePathCollisionError extends ModuleRegistryError {
  constructor(basePath: string, firstModuleId: string, secondModuleId: string) {
    super(
      `Module basePath collision: "${basePath}". Used by modules "${firstModuleId}" and "${secondModuleId}".`,
    );
    this.name = 'BasePathCollisionError';
  }
}
