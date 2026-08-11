import type {M8RuntimeContext} from '@m8/core';

import type {M8OperationConfirmation} from './definition';
import {
  MissingConfirmationAdapterError,
  OperationAuthorizationError,
  OperationConfirmationDeclinedError,
} from './errors';
import {OperationRegistry} from './registry';

export interface AuthorizationAdapter {
  check(input: {
    operationId: string;
    permission: string;
    context: M8RuntimeContext;
    signal: AbortSignal;
  }): Promise<boolean>;
}

export interface ConfirmationAdapter {
  confirm(input: {
    operationId: string;
    confirmation: M8OperationConfirmation;
    context: M8RuntimeContext;
    signal: AbortSignal;
  }): Promise<boolean>;
}

export interface OperationAuditEvent {
  readonly operationId: string;
  readonly phase: 'started' | 'succeeded' | 'failed' | 'cancelled';
  readonly context: M8RuntimeContext;
  readonly error?: unknown;
}

export interface OperationAuditAdapter {
  record(event: OperationAuditEvent): Promise<void> | void;
}

export interface QueryInvalidationAdapter {
  invalidate(queryId: string): Promise<void> | void;
}

export interface OperationRuntimeAdapters {
  readonly authorization?: AuthorizationAdapter;
  readonly confirmation?: ConfirmationAdapter;
  readonly audit?: OperationAuditAdapter;
  readonly queryInvalidation?: QueryInvalidationAdapter;
}

export interface ExecuteOperationOptions {
  readonly signal?: AbortSignal;
  readonly context?: M8RuntimeContext;
}

const defaultConfirmation: M8OperationConfirmation = {
  title: 'Confirm operation?',
  description: 'This action may be destructive.',
  confirmLabel: 'Confirm',
};

export class M8OperationRuntime {
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

    if (definition.requiredPermission) {
      const allowed = await this.adapters.authorization?.check({
        operationId,
        permission: definition.requiredPermission,
        context,
        signal,
      }) ?? false;
      if (!allowed) throw new OperationAuthorizationError(operationId);
    }

    const confirmation = definition.confirmation ??
      (definition.destructive ? defaultConfirmation : undefined);
    if (confirmation) {
      if (!this.adapters.confirmation) throw new MissingConfirmationAdapterError(operationId);
      const confirmed = await this.adapters.confirmation.confirm({
        operationId,
        confirmation,
        context,
        signal,
      });
      if (!confirmed) throw new OperationConfirmationDeclinedError(operationId);
    }

    const parsedInput = definition.input.parse(input);
    await this.adapters.audit?.record({operationId, phase: 'started', context});

    try {
      const rawOutput = await definition.execute({input: parsedInput, signal, context});
      const output = definition.output.parse(rawOutput);
      for (const queryId of definition.invalidate ?? []) {
        await this.adapters.queryInvalidation?.invalidate(queryId);
      }
      await this.adapters.audit?.record({operationId, phase: 'succeeded', context});
      return output;
    } catch (error) {
      await this.adapters.audit?.record({
        operationId,
        phase: signal.aborted ? 'cancelled' : 'failed',
        context,
        error,
      });
      throw error;
    }
  }
}
