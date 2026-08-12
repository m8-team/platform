import type {ModuleRegistry, RuntimeContext} from '@m8/core';
import {
  OperationRegistry,
  OperationRuntime,
  type OperationRuntimeAdapters,
  type RegisteredOperationDefinition,
} from '@m8/operation';
import {
  QueryRegistry,
  QueryRuntime,
  type RegisteredQueryDefinition,
} from '@m8/query';

export interface RuntimeAuthorizationAdapter {
  can(input: {
    permission: string;
    context: RuntimeContext;
    signal?: AbortSignal;
  }): boolean | Promise<boolean>;
}

export interface RuntimeNavigationItem {
  readonly moduleId: string;
  readonly moduleTitle: string;
  readonly path: string;
  readonly label: string;
  readonly icon?: string;
}

export interface RuntimeNavigation {
  getItems(context?: RuntimeContext): Promise<readonly RuntimeNavigationItem[]>;
}

export interface Runtime<
  TQuery extends RegisteredQueryDefinition = RegisteredQueryDefinition,
  TOperation extends RegisteredOperationDefinition = RegisteredOperationDefinition,
> {
  readonly modules: ModuleRegistry<TQuery, TOperation>;
  readonly queries: QueryRuntime;
  readonly operations: OperationRuntime;
  readonly authorization?: RuntimeAuthorizationAdapter;
  readonly navigation: RuntimeNavigation;
}

export interface CreateRuntimeOptions<
  TQuery extends RegisteredQueryDefinition = RegisteredQueryDefinition,
  TOperation extends RegisteredOperationDefinition = RegisteredOperationDefinition,
> {
  readonly modules: ModuleRegistry<TQuery, TOperation>;
  readonly adapters?: Omit<OperationRuntimeAdapters, 'authorization'> & {
    authorization?: RuntimeAuthorizationAdapter;
  };
}

function compareText(left: string, right: string): number {
  if (left < right) return -1;
  if (left > right) return 1;
  return 0;
}

export function createRuntime<
  TQuery extends RegisteredQueryDefinition,
  TOperation extends RegisteredOperationDefinition,
>(options: CreateRuntimeOptions<TQuery, TOperation>): Runtime<TQuery, TOperation> {
  const queryDefinitions = options.modules.getQueryContributions()
    .map(({contribution}) => contribution);
  const operationDefinitions = options.modules.getOperationContributions()
    .map(({contribution}) => contribution);
  const queryRegistry = new QueryRegistry(queryDefinitions);
  const operationRegistry = new OperationRegistry(operationDefinitions);
  const authorization = options.adapters?.authorization;
  const operationAdapters: OperationRuntimeAdapters = {
    ...options.adapters,
    authorization: authorization ? {
      check: input => Promise.resolve(authorization.can(input)),
    } : undefined,
  };

  return {
    modules: options.modules,
    queries: new QueryRuntime(queryRegistry),
    operations: new OperationRuntime(operationRegistry, operationAdapters),
    authorization,
    navigation: {
      async getItems(context: RuntimeContext = {}) {
        const items: RuntimeNavigationItem[] = [];
        for (const {moduleId, path, route} of options.modules.getRoutes()) {
          if (!route.navigation || route.navigation.hidden) continue;
          if (route.access?.permission) {
            const allowed = authorization
              ? await authorization.can({
                  permission: route.access.permission,
                  context,
                })
              : false;
            if (!allowed) continue;
          }

          const moduleDefinition = options.modules.getModule(moduleId);
          items.push({
            moduleId,
            moduleTitle: moduleDefinition?.title ?? moduleId,
            path,
            label: route.navigation.label,
            icon: route.navigation.icon,
          });
        }
        return items.sort((left, right) =>
          compareText(left.moduleTitle, right.moduleTitle) ||
          compareText(left.moduleId, right.moduleId) ||
          compareText(left.path, right.path) ||
          compareText(left.label, right.label));
      },
    },
  };
}
