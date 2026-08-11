export type ModuleRegistryErrorCode =
  | 'MODULE_REGISTRY_ERROR'
  | 'MODULE_DUPLICATE'
  | 'QUERY_DUPLICATE'
  | 'OPERATION_DUPLICATE'
  | 'MODULE_DEPENDENCY_MISSING'
  | 'MODULE_DEPENDENCY_CYCLE'
  | 'ROUTE_COLLISION'
  | 'MODULE_BASE_PATH_COLLISION';

export class ModuleRegistryError extends Error {
  override readonly name: string = 'ModuleRegistryError';
  readonly code: ModuleRegistryErrorCode = 'MODULE_REGISTRY_ERROR';

  constructor(message: string) {
    super(message);
  }
}

export class DuplicateModuleError extends ModuleRegistryError {
  override readonly name = 'DuplicateModuleError';
  override readonly code = 'MODULE_DUPLICATE';

  constructor(readonly moduleId: string) {
    super(`Duplicate module id: "${moduleId}"`);
  }
}

export class DuplicateQueryError extends ModuleRegistryError {
  override readonly name = 'DuplicateQueryError';
  override readonly code = 'QUERY_DUPLICATE';

  constructor(
    readonly queryId: string,
    readonly firstModuleId: string,
    readonly secondModuleId: string,
  ) {
    super(
      `Duplicate query id: "${queryId}". Declared by modules "${firstModuleId}" and "${secondModuleId}".`,
    );
  }
}

export class DuplicateOperationError extends ModuleRegistryError {
  override readonly name = 'DuplicateOperationError';
  override readonly code = 'OPERATION_DUPLICATE';

  constructor(
    readonly operationId: string,
    readonly firstModuleId: string,
    readonly secondModuleId: string,
  ) {
    super(
      `Duplicate operation id: "${operationId}". Declared by modules "${firstModuleId}" and "${secondModuleId}".`,
    );
  }
}

export class MissingModuleDependencyError extends ModuleRegistryError {
  override readonly name = 'MissingModuleDependencyError';
  override readonly code = 'MODULE_DEPENDENCY_MISSING';

  constructor(
    readonly moduleId: string,
    readonly dependencyId: string,
  ) {
    super(`Module "${moduleId}" requires missing module "${dependencyId}".`);
  }
}

export class CircularModuleDependencyError extends ModuleRegistryError {
  override readonly name = 'CircularModuleDependencyError';
  override readonly code = 'MODULE_DEPENDENCY_CYCLE';
  readonly path: readonly string[];

  constructor(path: readonly string[]) {
    super(`Circular module dependency: ${path.join(' -> ')}`);
    this.path = Object.freeze([...path]);
  }
}

export class RouteCollisionError extends ModuleRegistryError {
  override readonly name = 'RouteCollisionError';
  override readonly code = 'ROUTE_COLLISION';

  constructor(
    readonly route: string,
    readonly firstModuleId: string,
    readonly secondModuleId: string,
  ) {
    super(
      `Route collision: "${route}". Declared by modules "${firstModuleId}" and "${secondModuleId}".`,
    );
  }
}

export class BasePathCollisionError extends ModuleRegistryError {
  override readonly name = 'BasePathCollisionError';
  override readonly code = 'MODULE_BASE_PATH_COLLISION';

  constructor(
    readonly basePath: string,
    readonly firstModuleId: string,
    readonly secondModuleId: string,
  ) {
    super(
      `Module basePath collision: "${basePath}". Used by modules "${firstModuleId}" and "${secondModuleId}".`,
    );
  }
}
