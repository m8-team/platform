import type {M8ModuleDefinition, M8RuntimeContext, ModuleRegistry} from '@m8/core';
import {OperationRegistry, M8OperationRuntime, type OperationRuntimeAdapters} from '@m8/operation';
import {QueryRegistry, M8QueryRuntime} from '@m8/query';
import {buildNextAppSpec} from './next';

export interface RuntimeAuthorizationAdapter {
  can(input: {permission: string; context: M8RuntimeContext; signal?: AbortSignal}): boolean | Promise<boolean>;
}

export interface CreateM8RuntimeOptions {
  readonly modules: ModuleRegistry<readonly M8ModuleDefinition[]>;
  readonly catalog?: Readonly<Record<string, unknown>>;
  readonly adapters?: Omit<OperationRuntimeAdapters, 'authorization'> & {authorization?: RuntimeAuthorizationAdapter};
}

export function createM8Runtime(options: CreateM8RuntimeOptions) {
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
    queries: new M8QueryRuntime(queryRegistry),
    operations: new M8OperationRuntime(operationRegistry, operationAdapters),
    appSpec: buildNextAppSpec(options.modules),
    authorization: options.adapters?.authorization,
    navigation: {
      async getItems(context: M8RuntimeContext = {}) {
        const items = [];
        for (const {moduleId, path, route} of options.modules.getRoutes()) {
          if (!route.navigation || route.navigation.hidden) continue;
          const moduleDefinition = options.modules.getModule(moduleId)!;
          if (moduleDefinition.availability?.feature && !context.features?.includes(moduleDefinition.availability.feature)) continue;
          if (route.availability?.feature && !context.features?.includes(route.availability.feature)) continue;
          const editions = route.availability?.editions ?? moduleDefinition.availability?.editions;
          if (editions?.length && (!context.edition || !editions.includes(context.edition))) continue;
          if (route.access?.permission && options.adapters?.authorization &&
              !await options.adapters.authorization.can({permission: route.access.permission, context})) continue;
          items.push({...route.navigation, moduleId, path});
        }
        return items.sort((a, b) => (a.order ?? 0) - (b.order ?? 0));
      },
    },
  } as const;
}

export type M8Runtime = ReturnType<typeof createM8Runtime>;
