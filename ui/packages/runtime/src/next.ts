import type {NextAppSpec} from '@json-render/next';
import type {ModuleRegistry, ModuleDefinition} from '@m8/core';
import {normalizePath} from '@m8/core';
import {withRuntimeState} from './state';

export function buildNextAppSpec(
  modules: ModuleRegistry<readonly ModuleDefinition[]>,
  options: {
    metadata?: NextAppSpec['metadata'];
    layouts?: NextAppSpec['layouts'];
    state?: NextAppSpec['state'];
  } = {},
): NextAppSpec {
  const routes: NextAppSpec['routes'] = {};
  for (const moduleDefinition of modules.getModules()) {
    for (const [routePath, route] of Object.entries(moduleDefinition.routes ?? {})) {
      const {navigation: _navigation, access, queries, ...nativeRoute} = route;
      const page = nativeRoute.page;
      if (!page) continue;
      const boundaryId = '__route_runtime';
      routes[normalizePath(routePath)] = {
        ...nativeRoute,
        page: {
          ...page,
          state: withRuntimeState(page.state),
          root: boundaryId,
          elements: {
            ...page.elements,
            [boundaryId]: {
              type: '__RouteRuntime',
              props: {bindings: queries ?? {}, access},
              children: [page.root],
            },
          },
        },
      } as NextAppSpec['routes'][string];
    }
  }
  return {routes, ...options};
}
