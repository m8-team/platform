import type {RuntimeContext} from '@m8/core';

import {
  MissingLongRunningOperationAdapterError,
  LongRunningOperationFailedError,
  OperationAuthorizationError,
} from './errors';
import {OperationRegistry} from './registry';
import type {LongRunningOperationAdapter} from './long-running';

export interface AuthorizationAdapter {
  check(input: {
    operationId: string;
    permission: string;
    context: RuntimeContext;
    signal: AbortSignal;
  }): Promise<boolean>;
}

export interface OperationAuditEvent {
  readonly operationId: string;
  readonly phase: 'started' | 'succeeded' | 'failed' | 'cancelled';
  readonly context: RuntimeContext;
  readonly error?: unknown;
}

export interface OperationAuditAdapter {
  record(event: OperationAuditEvent): Promise<void> | void;
}

export interface QueryInvalidationAdapter {
  invalidate(queryId: string): Promise<void> | void;
}

export interface RuntimeErrorReporter {
  report(input: {source: 'operation-audit' | 'operation-invalidation'; operationId: string; error: unknown}): void;
}

export interface OperationRuntimeAdapters {
  readonly authorization?: AuthorizationAdapter;
  readonly audit?: OperationAuditAdapter;
  readonly queryInvalidation?: QueryInvalidationAdapter;
  readonly longRunningOperations?: LongRunningOperationAdapter;
  readonly errorReporter?: RuntimeErrorReporter;
}

export interface ExecuteOperationOptions {
  readonly signal?: AbortSignal;
  readonly context?: RuntimeContext;
}

export class OperationRuntime {
  constructor(
    readonly registry: OperationRegistry,
    private readonly adapters: OperationRuntimeAdapters = {},
  ) {}

  async execute(
    operationId: string,
    input: unknown,
    options: ExecuteOperationOptions = {},
  ): Promise<unknown> {
    const definition = this.registry.require(operationId);
    const signal = options.signal ?? new AbortController().signal;
    const context = options.context ?? {};
    signal.throwIfAborted();
    const parsedInput = definition.input.parse(input);

    if (definition.requiredPermission) {
      const allowed = await this.adapters.authorization?.check({
        operationId,
        permission: definition.requiredPermission,
        context,
        signal,
      }) ?? false;
      if (!allowed) throw new OperationAuthorizationError(operationId);
    }

    await this.runEffect(operationId, 'operation-audit', () =>
      this.adapters.audit?.record({operationId, phase: 'started', context}));

    let output: unknown;
    try {
      const rawOutput = await definition.execute({input: parsedInput, signal, context});
      output = definition.output.parse(rawOutput);
      if (definition.mode === 'long-running') {
        const operationIdentifier = (output as {operationId?: unknown}).operationId;
        if (typeof operationIdentifier !== 'string') {
          throw new TypeError(`Long-running operation "${operationId}" did not return operationId.`);
        }
        if (!this.adapters.longRunningOperations) throw new MissingLongRunningOperationAdapterError(operationId);
        const terminal = await this.adapters.longRunningOperations.wait(operationIdentifier, {signal});
        if (terminal.status === 'FAILED' || terminal.status === 'CANCELLED') {
          throw new LongRunningOperationFailedError(operationIdentifier, terminal.status, terminal.error?.message);
        }
        if (terminal.status !== 'SUCCEEDED') {
          throw new TypeError(`Long-running operation adapter returned non-terminal status "${terminal.status}".`);
        }
      }
    } catch (error) {
      await this.runEffect(operationId, 'operation-audit', () => this.adapters.audit?.record({
        operationId,
        phase: signal.aborted ? 'cancelled' : 'failed',
        context,
        error,
      }));
      throw error;
    }

    for (const queryId of definition.invalidate ?? []) {
      await this.runEffect(operationId, 'operation-invalidation', () =>
        this.adapters.queryInvalidation?.invalidate(queryId));
    }
    await this.runEffect(operationId, 'operation-audit', () =>
      this.adapters.audit?.record({operationId, phase: 'succeeded', context}));
    return output;
  }

  private async runEffect(
    operationId: string,
    source: 'operation-audit' | 'operation-invalidation',
    effect: () => Promise<void> | void | undefined,
  ): Promise<void> {
    try {
      await effect();
    } catch (error) {
      this.adapters.errorReporter?.report({source, operationId, error});
    }
  }
}
