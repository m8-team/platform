export type ModuleRegistryErrorCode =
  | 'MODULE_REGISTRY_ERROR'
  | 'MODULE_DUPLICATE'
  | 'ROUTE_INVALID'
  | 'ROUTE_COLLISION'
  | 'QUERY_REFERENCE_UNKNOWN'
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

export class InvalidRouteError extends ModuleRegistryError {
  override readonly name = 'InvalidRouteError';
  override readonly code = 'ROUTE_INVALID';

  constructor(
    readonly route: string,
    readonly ownerId: string,
  ) {
    super(`Invalid route "${route}" declared by "${ownerId}".`);
  }
}

export class UnknownModuleQueryReferenceError extends ModuleRegistryError {
  override readonly name = 'UnknownModuleQueryReferenceError';
  override readonly code = 'QUERY_REFERENCE_UNKNOWN';

  constructor(
    readonly moduleId: string,
    readonly route: string,
    readonly queryId: string,
  ) {
    super(
      `Route "${route}" of module "${moduleId}" references unknown query ` +
      `"${queryId}".`,
    );
  }
}
