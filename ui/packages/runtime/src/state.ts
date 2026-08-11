import type {RuntimeContext} from '@m8/core';

export interface RuntimeSystemState {
  readonly params: Readonly<Record<string, string | string[]>>;
  readonly context: RuntimeContext;
  readonly queries: Readonly<Record<string, unknown>>;
}

export function createRuntimeState(options: {
  readonly params?: Readonly<Record<string, string | string[]>>;
  readonly context?: RuntimeContext;
} = {}): {readonly __runtime: RuntimeSystemState} {
  return {
    __runtime: {
      params: options.params ?? {},
      context: options.context ?? {},
      queries: {},
    },
  };
}

export function withRuntimeState(
  state: Readonly<Record<string, unknown>> | undefined,
  options: Parameters<typeof createRuntimeState>[0] = {},
): Readonly<Record<string, unknown>> {
  if (state && Object.hasOwn(state, '__runtime')) {
    throw new TypeError('UI state must not define the reserved "__runtime" namespace.');
  }
  return {...state, ...createRuntimeState(options)};
}
