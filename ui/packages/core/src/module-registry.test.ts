import {describe, expect, it} from 'vitest';

import {defineModule} from './define-module';
import {
  DuplicateModuleError,
  DuplicateModuleOperationError,
  DuplicateModuleQueryError,
  InvalidRouteError,
  RouteCollisionError,
  UnknownModuleQueryReferenceError,
} from './errors';
import {defineModules, type ModuleRegistry} from './module-registry';
import {isValidRoute, normalizePath} from './routes';
import type {ModuleDefinition} from './types';

const page = {
  root: 'root',
  elements: {root: {type: 'Text', props: {}, children: []}},
};

function module(
  id: string,
  options: Omit<ModuleDefinition, 'id' | 'title'> = {},
): ModuleDefinition {
  return defineModule({
    id,
    title: id.toUpperCase(),
    routes: {[`/${id}`]: {page}},
    ...options,
  });
}

function snapshot(registry: ModuleRegistry) {
  const routes = registry.getRoutes();
  const queries = registry.getQueryContributions();
  const operations = registry.getOperationContributions();

  return {
    modules: registry.getModules().map(definition => definition.id),
    routes: routes.map(({moduleId, path}) => ({moduleId, path})),
    queries: queries.map(({moduleId, contribution}) => ({
      moduleId,
      id: contribution.id,
    })),
    operations: operations.map(({moduleId, contribution}) => ({
      moduleId,
      id: contribution.id,
    })),
    ownership: {
      routes: Object.fromEntries(routes.map(({path}) => [
        path,
        registry.getRouteOwner(path),
      ])),
      queries: Object.fromEntries(queries.map(({contribution}) => [
        contribution.id,
        registry.getQueryOwner(contribution.id),
      ])),
      operations: Object.fromEntries(operations.map(({contribution}) => [
        contribution.id,
        registry.getOperationOwner(contribution.id),
      ])),
    },
  };
}

describe('ModuleRegistry validation', () => {
  it('rejects duplicate module IDs', () => {
    expect(() => defineModules([module('a'), module('a')]))
      .toThrow(DuplicateModuleError);
  });

  it('rejects normalized route collisions and names both owners', () => {
    expect(() => defineModules([
      module('b', {routes: {'/projects/': {page}}}),
      module('a', {routes: {'/projects': {page}}}),
    ])).toThrowError(
      'Route collision: "/projects". Declared by modules "a" and "b".',
    );
  });

  it.each([
    ['/projects/[id]', '/projects/[projectId]'],
    ['/projects/[...slug]', '/projects/[...path]'],
    ['/projects/[[...slug]]', '/projects/[[...path]]'],
  ])('rejects canonical dynamic collision %s and %s', (first, second) => {
    expect(() => defineModules([
      module('a', {routes: {[first]: {page}}}),
      module('b', {routes: {[second]: {page}}}),
    ])).toThrow(RouteCollisionError);
  });

  it('rejects invalid route syntax', () => {
    const invalidModule = {
      id: 'invalid',
      title: 'Invalid',
      routes: {'projects': {page}},
    } as unknown as ModuleDefinition;

    expect(() => defineModules([invalidModule])).toThrow(InvalidRouteError);
  });

  it('rejects duplicate global query IDs', () => {
    expect(() => defineModules([
      module('b', {queries: [{id: 'projects.list'}]}),
      module('a', {queries: [{id: 'projects.list'}]}),
    ])).toThrowError(new DuplicateModuleQueryError('projects.list', 'a', 'b'));
  });

  it('rejects duplicate global operation IDs', () => {
    expect(() => defineModules([
      module('b', {operations: [{id: 'projects.delete'}]}),
      module('a', {operations: [{id: 'projects.delete'}]}),
    ])).toThrowError(
      new DuplicateModuleOperationError('projects.delete', 'a', 'b'),
    );
  });

  it('resolves cross-module query references from the complete registry', () => {
    const consumer = module('consumer', {
      routes: {
        '/consumer': {
          queries: {users: {query: 'identity.users.list'}},
          page,
        },
      },
    });
    const identity = module('identity', {
      queries: [{id: 'identity.users.list'}],
    });

    expect(() => defineModules([consumer, identity])).not.toThrow();
    expect(() => defineModules([identity, consumer])).not.toThrow();
  });

  it('reports an unknown query reference at its actual use site', () => {
    expect(() => defineModules([module('admin', {
      routes: {
        '/admin': {
          queries: {users: {query: 'identity.users.list'}},
          page,
        },
      },
    })])).toThrowError(new UnknownModuleQueryReferenceError(
      'admin',
      '/admin',
      'identity.users.list',
    ));
  });
});

describe('ModuleRegistry ownership', () => {
  it('retains route, query and operation owner mappings', () => {
    const query = {id: 'projects.list'};
    const operation = {id: 'projects.delete'};
    const registry = defineModules([
      module('projects', {queries: [query], operations: [operation]}),
    ]);

    expect(registry.getRoutes()[0]).toMatchObject({
      moduleId: 'projects',
      path: '/projects',
    });
    expect(registry.getRouteOwner('/projects/')).toBe('projects');
    expect(registry.getQueryOwner(query.id)).toBe('projects');
    expect(registry.getOperationOwner(operation.id)).toBe('projects');
    expect(registry.getQueryContributions()).toEqual([
      {moduleId: 'projects', contribution: query},
    ]);
    expect(registry.getOperationContributions()).toEqual([
      {moduleId: 'projects', contribution: operation},
    ]);
  });

  it('returns immutable module and contribution collections', () => {
    const definition = defineModule({
      id: 'a',
      title: 'A',
      routes: {'/a': {page}},
      queries: [{id: 'a.query'}],
      operations: [{id: 'a.operation'}],
    });
    const registry = defineModules([definition]);

    expect(Object.isFrozen(definition)).toBe(true);
    expect(Object.isFrozen(definition.routes)).toBe(true);
    expect(Object.isFrozen(definition.queries)).toBe(true);
    expect(Object.isFrozen(definition.operations)).toBe(true);
    expect(Object.isFrozen(registry.getModules())).toBe(true);
    expect(Object.isFrozen(registry.getRoutes())).toBe(true);
  });
});

describe('registration order independence', () => {
  const queryA = {id: 'a.items.list'};
  const queryB = {id: 'b.items.list'};
  const queryC = {id: 'c.items.list'};
  const operationA = {id: 'a.items.create'};
  const operationB = {id: 'b.items.create'};
  const operationC = {id: 'c.items.create'};
  const a = module('a', {
    queries: [queryA],
    operations: [operationA],
  });
  const b = module('b', {
    routes: {
      '/b/[itemId]': {
        queries: {items: {query: queryA.id}},
        page,
      },
    },
    queries: [queryB],
    operations: [operationB],
  });
  const c = module('c', {
    queries: [queryC],
    operations: [operationC],
  });
  const permutations = [
    [a, b, c],
    [a, c, b],
    [b, a, c],
    [b, c, a],
    [c, a, b],
    [c, b, a],
  ] as const;

  it('builds the same modules, contributions and ownership for every permutation', () => {
    const expected = snapshot(defineModules(permutations[0]));

    for (const permutation of permutations) {
      expect(snapshot(defineModules(permutation))).toEqual(expected);
    }
  });
});

describe('route helpers', () => {
  it.each([
    ['', '/'],
    ['resource-manager/', '/resource-manager'],
    ['//a///b/', '/a/b'],
  ])('normalizes %s', (input, expected) =>
    expect(normalizePath(input)).toBe(expected));

  it.each([
    '/',
    '/projects',
    '/projects/[projectId]',
    '/docs/[...path]',
    '/settings/[[...path]]',
  ])('accepts valid route %s', route => {
    expect(isValidRoute(route)).toBe(true);
  });

  it.each([
    'projects',
    '/projects/[id',
    '/docs/[...path]/edit',
    '/projects?tab=all',
  ])('rejects invalid route %s', route => {
    expect(isValidRoute(route)).toBe(false);
  });
});
