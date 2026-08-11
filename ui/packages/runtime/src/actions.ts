import type {RuntimeContext} from '@m8/core';
import type {OperationRuntime} from '@m8/operation';

function asRecord(value: unknown): Record<string, unknown> {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    throw new TypeError('Action parameters must be an object.');
  }
  return value as Record<string, unknown>;
}

function requiredString(value: unknown, field: string): string {
  if (typeof value !== 'string' || value.length === 0) {
    throw new TypeError(`Action parameter "${field}" must be a non-empty string.`);
  }
  return value;
}

export interface CreateActionHandlersOptions {
  readonly operations: OperationRuntime;
  readonly getContext: () => RuntimeContext;
}

export function createActionHandlers(options: CreateActionHandlersOptions) {
  return {
    executeOperation: async (params: Record<string, unknown>) => {
      const values = asRecord(params);
      return options.operations.execute(
        requiredString(values.operation, 'operation'),
        values.input ?? {},
        {context: options.getContext()},
      );
    },
  };
}
