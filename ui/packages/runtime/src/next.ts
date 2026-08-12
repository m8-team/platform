import type {
  LoaderFn,
  NextAppSpec,
  NextRouteSpec,
} from '@json-render/next';
import type {ActionBinding, Spec} from '@json-render/core';
import {
  canonicalizeRoute,
  InvalidRouteError,
  isValidRoute,
  normalizePath,
  RouteCollisionError,
  type ModuleContribution,
  type ModuleRegistry,
  type ModuleRouteSpec,
} from '@m8/core';

import {createRuntimeState, withRuntimeState} from './state';

export const RUNTIME_ROUTE_LOADER = '__m8RouteParams';

export const runtimeRouteLoader: LoaderFn = params =>
  createRuntimeState({params});

function compareText(left: string, right: string): number {
  if (left < right) return -1;
  if (left > right) return 1;
  return 0;
}

export function createRuntimeNextLoaders(
  loaders: Readonly<Record<string, LoaderFn>> = {},
): Readonly<Record<string, LoaderFn>> {
  if (Object.hasOwn(loaders, RUNTIME_ROUTE_LOADER)) {
    throw new TypeError(`Loader id "${RUNTIME_ROUTE_LOADER}" is reserved.`);
  }

  const composedLoaders = Object.fromEntries(
    Object.entries(loaders)
      .sort(([left], [right]) => compareText(left, right))
      .map(([loaderId, loader]) => [
        loaderId,
        async (params: Record<string, string | string[]>) => ({
          ...withRuntimeState(await loader(params), {params}),
        }),
      ]),
  );

  return Object.freeze({
    [RUNTIME_ROUTE_LOADER]: runtimeRouteLoader,
    ...composedLoaders,
  });
}

export const runtimeNextLoaders = createRuntimeNextLoaders();

export interface BuildNextAppSpecOptions {
  readonly baseSpec?: NextAppSpec;
  readonly metadata?: NextAppSpec['metadata'];
  readonly layouts?: NextAppSpec['layouts'];
  readonly state?: NextAppSpec['state'];
  readonly defaultLayout?: string;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function isRuntimeStatePath(value: unknown): value is string {
  return typeof value === 'string' &&
    (value === '/__runtime' || value.startsWith('/__runtime/'));
}

function rejectRuntimeStateWrite(
  path: string,
  routePath: string,
  elementId: string,
): never {
  throw new TypeError(
    `Route "${routePath}" element "${elementId}" cannot write reserved ` +
    `runtime state path "${path}".`,
  );
}

function validatePropBindings(
  value: unknown,
  routePath: string,
  elementId: string,
): void {
  if (Array.isArray(value)) {
    for (const item of value) {
      validatePropBindings(item, routePath, elementId);
    }
    return;
  }
  if (!isRecord(value)) return;

  if (isRuntimeStatePath(value.$bindState)) {
    rejectRuntimeStateWrite(value.$bindState, routePath, elementId);
  }
  for (const nested of Object.values(value)) {
    validatePropBindings(nested, routePath, elementId);
  }
}

function validateActionBinding(
  binding: ActionBinding,
  routePath: string,
  elementId: string,
): void {
  if (['setState', 'pushState', 'removeState'].includes(binding.action)) {
    const statePath = binding.params?.statePath;
    if (isRuntimeStatePath(statePath)) {
      rejectRuntimeStateWrite(statePath, routePath, elementId);
    }
  }

  for (const lifecycle of [binding.onSuccess, binding.onError]) {
    if (!lifecycle || !('set' in lifecycle)) continue;
    for (const statePath of Object.keys(lifecycle.set)) {
      if (isRuntimeStatePath(statePath)) {
        rejectRuntimeStateWrite(statePath, routePath, elementId);
      }
    }
  }
}

function validateActionBindings(
  bindings: Record<string, ActionBinding | ActionBinding[]> | undefined,
  routePath: string,
  elementId: string,
): void {
  for (const binding of Object.values(bindings ?? {})) {
    for (const candidate of Array.isArray(binding) ? binding : [binding]) {
      validateActionBinding(candidate, routePath, elementId);
    }
  }
}

function runtimeItemPath(
  value: unknown,
  repeatBasePath: string | undefined,
): string | undefined {
  if (!repeatBasePath || !isRuntimeStatePath(repeatBasePath) || !isRecord(value) ||
      typeof value.$item !== 'string') return undefined;
  return `${repeatBasePath}/${value.$item}`.replace(/\/$/, '');
}

function validateRepeatedStateWrites(
  spec: Spec,
  routePath: string,
  elementId: string,
  repeatBasePath?: string,
  visited = new Set<string>(),
): void {
  const visitKey = `${elementId}\0${repeatBasePath ?? ''}`;
  if (visited.has(visitKey)) return;
  visited.add(visitKey);

  const element = spec.elements[elementId];
  if (!element) return;

  for (const value of Object.values(element.props)) {
    if (isRecord(value) && typeof value.$bindItem === 'string' &&
        repeatBasePath && isRuntimeStatePath(repeatBasePath)) {
      rejectRuntimeStateWrite(
        `${repeatBasePath}/${value.$bindItem}`.replace(/\/$/, ''),
        routePath,
        elementId,
      );
    }
  }

  for (const bindings of [element.on, element.watch]) {
    for (const binding of Object.values(bindings ?? {})) {
      for (const candidate of Array.isArray(binding) ? binding : [binding]) {
        if (!['setState', 'pushState', 'removeState'].includes(candidate.action)) {
          continue;
        }
        const statePath = runtimeItemPath(
          candidate.params?.statePath,
          repeatBasePath,
        );
        if (statePath) rejectRuntimeStateWrite(statePath, routePath, elementId);
      }
    }
  }

  const childRepeatBasePath = element.repeat?.statePath ?? repeatBasePath;
  for (const childId of element.children ?? []) {
    validateRepeatedStateWrites(
      spec,
      routePath,
      childId,
      childRepeatBasePath,
      visited,
    );
  }
}

function validateRuntimeStateOwnership(spec: Spec, routePath: string): void {
  for (const [elementId, element] of Object.entries(spec.elements)) {
    validatePropBindings(element.props, routePath, elementId);
    validateActionBindings(element.on, routePath, elementId);
    validateActionBindings(element.watch, routePath, elementId);
  }
  validateRepeatedStateWrites(spec, routePath, spec.root);
}

interface RouteInput {
  readonly ownerId: string;
  readonly path: string;
  readonly canonicalPath: string;
}

interface BaseRouteInput extends RouteInput {
  readonly route: NextRouteSpec;
  readonly moduleRoute: false;
}

interface ModuleRouteInput extends RouteInput {
  readonly route: ModuleRouteSpec;
  readonly moduleRoute: true;
}

type ComposedRouteInput = BaseRouteInput | ModuleRouteInput;

function collectBaseRoutes(
  routes: NextAppSpec['routes'] | undefined,
): BaseRouteInput[] {
  return Object.entries(routes ?? {}).map(([routePath, route]) => {
    if (!isValidRoute(routePath)) {
      throw new InvalidRouteError(routePath, 'platform');
    }
    const path = normalizePath(routePath);
    return {
      ownerId: 'platform',
      path,
      canonicalPath: canonicalizeRoute(path),
      route,
      moduleRoute: false as const,
    };
  }).sort((left, right) =>
    compareText(left.canonicalPath, right.canonicalPath) ||
    compareText(left.path, right.path));
}

function validateRouteCollisions(routes: readonly ComposedRouteInput[]): void {
  const owners = new Map<string, string>();
  for (const route of routes) {
    const existingOwnerId = owners.get(route.canonicalPath);
    if (existingOwnerId) {
      throw new RouteCollisionError(
        route.path,
        existingOwnerId,
        route.ownerId,
      );
    }
    owners.set(route.canonicalPath, route.ownerId);
  }
}

function composeModuleRoute(
  path: string,
  route: ModuleRouteSpec,
  defaultLayout: string | undefined,
): NextRouteSpec {
  const {
    navigation: _navigation,
    access,
    queries,
    ...nativeRoute
  } = route;
  validateRuntimeStateOwnership(nativeRoute.page, path);
  const boundaryId = '__route_runtime';
  if (Object.hasOwn(nativeRoute.page.elements, boundaryId)) {
    throw new Error(
      `Route "${path}" uses reserved element id "${boundaryId}".`,
    );
  }

  return {
    ...nativeRoute,
    layout: nativeRoute.layout ?? defaultLayout,
    // @json-render/next accepts one loader per route. Custom loaders are
    // wrapped by createRuntimeNextLoaders so their data and M8 route params
    // are merged into the same initial state.
    loader: nativeRoute.loader ?? RUNTIME_ROUTE_LOADER,
    page: {
      ...nativeRoute.page,
      state: withRuntimeState(nativeRoute.page.state),
      root: boundaryId,
      elements: {
        ...nativeRoute.page.elements,
        [boundaryId]: {
          type: '__RouteRuntime',
          props: {bindings: queries ?? {}, access},
          children: [nativeRoute.page.root],
        },
      },
    },
  };
}

export function buildNextAppSpec<
  TQuery extends ModuleContribution,
  TOperation extends ModuleContribution,
>(
  modules: ModuleRegistry<TQuery, TOperation>,
  options: BuildNextAppSpecOptions = {},
): NextAppSpec {
  const baseSpec = options.baseSpec;
  const baseRoutes = collectBaseRoutes(baseSpec?.routes);
  const moduleRoutes: ModuleRouteInput[] = modules.getRoutes().map(({
    moduleId,
    path,
    route,
  }) => ({
    ownerId: moduleId,
    path,
    canonicalPath: canonicalizeRoute(path),
    route,
    moduleRoute: true,
  }));
  const collectedRoutes = [...baseRoutes, ...moduleRoutes];

  validateRouteCollisions(collectedRoutes);

  const routes = Object.fromEntries(collectedRoutes
    .map(route => {
      if (!route.moduleRoute) {
        validateRuntimeStateOwnership(route.route.page, route.path);
      }
      return [
        route.path,
        route.moduleRoute
          ? composeModuleRoute(
              route.path,
              route.route,
              options.defaultLayout,
            )
          : route.route,
      ] as const;
    })
    .sort(([left], [right]) => compareText(left, right)));

  return {
    routes,
    metadata: options.metadata ?? baseSpec?.metadata,
    layouts: options.layouts ?? baseSpec?.layouts,
    state: options.state ?? baseSpec?.state,
  };
}
