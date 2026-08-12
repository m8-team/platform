export type ModuleRegistryErrorCode =
  | 'MODULE_REGISTRY_ERROR'
  | 'MODULE_DUPLICATE'
  | 'MODULE_DEPENDENCY_MISSING'
  | 'MODULE_DEPENDENCY_SELF'
  | 'MODULE_DEPENDENCY_CYCLE'
  | 'ROUTE_COLLISION'
  | 'QUERY_DUPLICATE'
  | 'OPERATION_DUPLICATE';

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
    super(`Duplicate module id "${moduleId}".`);
  }
}

export class SelfModuleDependencyError extends ModuleRegistryError {
  override readonly name = 'SelfModuleDependencyError';
  override readonly code = 'MODULE_DEPENDENCY_SELF';

  constructor(readonly moduleId: string) {
    super(`Module "${moduleId}" cannot depend on itself.`);
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
    super(`Circular module dependency detected: ${path.join(' -> ')}`);
    this.path = Object.freeze([...path]);
  }
}

abstract class DuplicateContributionError extends ModuleRegistryError {
  constructor(
    kind: 'query' | 'operation',
    readonly contributionId: string,
    readonly firstModuleId: string,
    readonly secondModuleId: string,
  ) {
    super(
      `Duplicate ${kind} id "${contributionId}" declared by modules ` +
      `"${firstModuleId}" and "${secondModuleId}".`,
    );
  }
}

export class DuplicateModuleQueryError extends DuplicateContributionError {
  override readonly name = 'DuplicateModuleQueryError';
  override readonly code = 'QUERY_DUPLICATE';

  constructor(queryId: string, firstModuleId: string, secondModuleId: string) {
    super('query', queryId, firstModuleId, secondModuleId);
  }
}

export class DuplicateModuleOperationError extends DuplicateContributionError {
  override readonly name = 'DuplicateModuleOperationError';
  override readonly code = 'OPERATION_DUPLICATE';

  constructor(operationId: string, firstModuleId: string, secondModuleId: string) {
    super('operation', operationId, firstModuleId, secondModuleId);
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
