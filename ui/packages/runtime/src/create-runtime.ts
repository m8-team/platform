import type {ModuleDefinition, RuntimeContext, ModuleRegistry} from '@m8/core';
import {OperationRegistry, OperationRuntime, type OperationRuntimeAdapters} from '@m8/operation';
import {QueryRegistry, QueryRuntime} from '@m8/query';
import {buildNextAppSpec} from './next';

export interface RuntimeAuthorizationAdapter {
  can(input: {permission: string; context: RuntimeContext; signal?: AbortSignal}): boolean | Promise<boolean>;
}

export interface CreateRuntimeOptions {
  readonly modules: ModuleRegistry<readonly ModuleDefinition[]>;
  readonly catalog?: Readonly<Record<string, unknown>>;
  readonly adapters?: Omit<OperationRuntimeAdapters, 'authorization'> & {authorization?: RuntimeAuthorizationAdapter};
}

export function createRuntime(options: CreateRuntimeOptions) {
  if (options.catalog) {
    for (const {path, route} of options.modules.getRoutes()) {
      for (const element of Object.values(route.page?.elements ?? {})) {
        if (!options.catalog[element.type]) {
          throw new Error(`Route "${path}" references unknown component "${element.type}".`);
        }
      }
    }
  }
  const queryRegistry = new QueryRegistry(options.modules.getQueries() as ConstructorParameters<typeof QueryRegistry>[0]);
  const operationRegistry = new OperationRegistry(options.modules.getOperations() as ConstructorParameters<typeof OperationRegistry>[0]);
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
    appSpec: buildNextAppSpec(options.modules),
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
