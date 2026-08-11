export class OperationRuntimeError extends Error {
  override readonly name: string = 'OperationRuntimeError';
}

export class DuplicateOperationError extends OperationRuntimeError {
  override readonly name = 'DuplicateOperationError';
  constructor(readonly operationId: string) {
    super(`Duplicate operation id: "${operationId}".`);
  }
}

export class UnknownOperationError extends OperationRuntimeError {
  override readonly name = 'UnknownOperationError';
  constructor(readonly operationId: string) {
    super(`Unknown operation id: "${operationId}".`);
  }
}

export class OperationAuthorizationError extends OperationRuntimeError {
  override readonly name = 'OperationAuthorizationError';
  constructor(readonly operationId: string) {
    super(`Operation "${operationId}" is not authorized.`);
  }
}

export class OperationConfirmationDeclinedError extends OperationRuntimeError {
  override readonly name = 'OperationConfirmationDeclinedError';
  constructor(readonly operationId: string) {
    super(`Operation "${operationId}" was not confirmed.`);
  }
}

export class MissingConfirmationAdapterError extends OperationRuntimeError {
  override readonly name = 'MissingConfirmationAdapterError';
  constructor(readonly operationId: string) {
    super(`Operation "${operationId}" requires a confirmation adapter.`);
  }
}

export class MissingLongRunningOperationAdapterError extends OperationRuntimeError {
  override readonly name = 'MissingLongRunningOperationAdapterError';
  constructor(readonly operationId: string) {
    super(`Long-running operation "${operationId}" requires an adapter.`);
  }
}

export class LongRunningOperationFailedError extends OperationRuntimeError {
  override readonly name = 'LongRunningOperationFailedError';
  constructor(readonly operationId: string, readonly status: 'FAILED' | 'CANCELLED', message?: string) {
    super(message ?? `Long-running operation "${operationId}" ended with ${status}.`);
  }
}
