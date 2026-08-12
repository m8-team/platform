import type {
  LoaderFn,
  NextAppSpec,
  NextRouteSpec,
} from '@json-render/next';
import type {ActionBinding, Spec} from '@json-render/core';
import type {ModuleContribution, ModuleRegistry} from '@m8/core';

import {createRuntimeState, withRuntimeState} from './state';

export const RUNTIME_ROUTE_LOADER = '__m8RouteParams';

export const runtimeRouteLoader: LoaderFn = params =>
  createRuntimeState({params});

export const runtimeNextLoaders: Readonly<Record<string, LoaderFn>> =
  Object.freeze({
    [RUNTIME_ROUTE_LOADER]: runtimeRouteLoader,
  });

export interface BuildNextAppSpecOptions {
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

export function buildNextAppSpec<
  TQuery extends ModuleContribution,
  TOperation extends ModuleContribution,
>(
  modules: ModuleRegistry<TQuery, TOperation>,
  options: BuildNextAppSpecOptions = {},
): NextAppSpec {
  const routes: NextAppSpec['routes'] = {};

  for (const {path, route} of modules.getRoutes()) {
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

    const composedRoute: NextRouteSpec = {
      ...nativeRoute,
      layout: nativeRoute.layout ?? options.defaultLayout,
      // Named route params come from json-render's matched-route loader. A
      // route with its own loader remains fully owned by json-render and may
      // project params itself when its UI bindings need them.
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
    routes[path] = composedRoute;
  }

  return {
    routes,
    metadata: options.metadata,
    layouts: options.layouts,
    state: options.state,
  };
}
