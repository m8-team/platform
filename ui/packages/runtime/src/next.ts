import type {NextAppSpec} from '@json-render/next';
import type {ModuleRegistry, M8ModuleDefinition} from '@m8/core';
import {joinRoute} from '@m8/core';

export function buildNextAppSpec(
  modules: ModuleRegistry<readonly M8ModuleDefinition[]>,
  options: {
    metadata?: NextAppSpec['metadata'];
    layouts?: NextAppSpec['layouts'];
    state?: NextAppSpec['state'];
  } = {},
): NextAppSpec {
  const routes: NextAppSpec['routes'] = {};
  for (const moduleDefinition of modules.getModules()) {
    for (const [routePath, route] of Object.entries(moduleDefinition.routes ?? {})) {
      const {navigation: _navigation, access, availability: _availability, queries, ...nativeRoute} = route;
      const page = nativeRoute.page;
      if (!page) continue;
      const boundaryId = '__m8_route_runtime';
      routes[joinRoute(moduleDefinition.basePath, routePath)] = {
        ...nativeRoute,
        page: {
          ...page,
          root: boundaryId,
          elements: {
            ...page.elements,
            [boundaryId]: {
              type: '__M8RouteRuntime',
              props: {bindings: queries ?? {}, access, path: joinRoute(moduleDefinition.basePath, routePath)},
              children: [page.root],
            },
          },
        },
      } as NextAppSpec['routes'][string];
    }
  }
  return {routes, ...options};
}
