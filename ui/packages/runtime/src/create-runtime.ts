import type {ModuleDefinition, RuntimeContext, ModuleRegistry} from '@m8/core';
import {OperationRegistry, OperationRuntime, type OperationRuntimeAdapters} from '@m8/operation';
import {QueryRegistry, QueryRuntime} from '@m8/query';

type QueryDefinitions = NonNullable<ConstructorParameters<typeof QueryRegistry>[0]>;
type QueryDefinition = QueryDefinitions[number];
type OperationDefinitions = NonNullable<ConstructorParameters<typeof OperationRegistry>[0]>;
type OperationDefinition = OperationDefinitions[number];

function isDefinition(value: unknown): value is Readonly<{
  id: string;
  input: {parse(value: unknown): unknown};
  output: {parse(value: unknown): unknown};
  execute: (...args: never[]) => unknown;
}> {
  return typeof value === 'object' && value !== null &&
    'id' in value && typeof value.id === 'string' &&
    'input' in value && typeof value.input === 'object' && value.input !== null &&
    'parse' in value.input && typeof value.input.parse === 'function' &&
    'output' in value && typeof value.output === 'object' && value.output !== null &&
    'parse' in value.output && typeof value.output.parse === 'function' &&
    'execute' in value && typeof value.execute === 'function';
}

function collectDefinitions<TDefinition>(
  contributions: readonly unknown[],
  kind: 'query' | 'operation',
): readonly TDefinition[] {
  if (!contributions.every(isDefinition)) {
    throw new TypeError(`Invalid ${kind} contribution.`);
  }
  return contributions as readonly TDefinition[];
}

export interface RuntimeAuthorizationAdapter {
  can(input: {permission: string; context: RuntimeContext; signal?: AbortSignal}): boolean | Promise<boolean>;
}

export interface CreateRuntimeOptions {
  readonly modules: ModuleRegistry<readonly ModuleDefinition[]>;
  readonly adapters?: Omit<OperationRuntimeAdapters, 'authorization'> & {authorization?: RuntimeAuthorizationAdapter};
}

export function createRuntime(options: CreateRuntimeOptions) {
  const queryDefinitions = options.modules.getModules().flatMap(module => module.queries ?? []);
  const operationDefinitions = options.modules.getModules().flatMap(module => module.operations ?? []);
  const queryRegistry = new QueryRegistry(collectDefinitions<QueryDefinition>(queryDefinitions, 'query'));
  const operationRegistry = new OperationRegistry(collectDefinitions<OperationDefinition>(operationDefinitions, 'operation'));
  const operationAdapters: OperationRuntimeAdapters = {
    ...options.adapters,
    authorization: options.adapters?.authorization ? {
      check: async input => options.adapters!.authorization!.can(input),
    } : undefined,
  };

  return {
    modules: options.modules,
    queries: new QueryRuntime(queryRegistry),
    operations: new OperationRuntime(operationRegistry, operationAdapters),
    authorization: options.adapters?.authorization,
    navigation: {
      async getItems(context: RuntimeContext = {}) {
        const items = [];
        for (const {moduleId, path, route} of options.modules.getRoutes()) {
          if (!route.navigation || route.navigation.hidden) continue;
          if (route.access?.permission && options.adapters?.authorization &&
              !await options.adapters.authorization.can({permission: route.access.permission, context})) continue;
          items.push({...route.navigation, moduleId, path});
        }
        return items.sort((a, b) => (a.order ?? 0) - (b.order ?? 0));
      },
    },
  } as const;
}

export type Runtime = ReturnType<typeof createRuntime>;
