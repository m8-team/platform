import {describe, expect, it} from 'vitest';

import {defineModule} from './define-module';
import {
  CircularModuleDependencyError,
  DuplicateModuleError,
  DuplicateModuleOperationError,
  DuplicateModuleQueryError,
  MissingModuleDependencyError,
  RouteCollisionError,
  SelfModuleDependencyError,
} from './errors';
import {defineModules} from './module-registry';
import {normalizePath} from './routes';
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

describe('ModuleRegistry dependencies', () => {
  it('rejects duplicate module IDs', () => {
    expect(() => defineModules([module('a'), module('a')]))
      .toThrow(DuplicateModuleError);
  });

  it('rejects a missing required dependency', () => {
    expect(() => defineModules([
      module('a', {dependencies: {required: ['missing']}}),
    ])).toThrowError('Module "a" requires missing module "missing".');
  });

  it.each(['required', 'optional'] as const)(
    'rejects a self %s dependency',
    dependencyKind => {
      expect(() => defineModules([
        module('a', {dependencies: {[dependencyKind]: ['a']}}),
      ])).toThrow(SelfModuleDependencyError);
    },
  );

  it('sorts a dependant before its input dependency', () => {
    const registry = defineModules([
      module('a', {dependencies: {required: ['b']}}),
      module('b'),
    ]);

    expect(registry.getModules().map(item => item.id)).toEqual(['b', 'a']);
  });

  it('sorts a multi-level graph independently of registration order', () => {
    const registry = defineModules([
      module('c', {dependencies: {required: ['b']}}),
      module('a'),
      module('b', {dependencies: {required: ['a']}}),
    ]);

    expect(registry.getModules().map(item => item.id)).toEqual(['a', 'b', 'c']);
  });

  it('sorts independent branches deterministically after their dependency', () => {
    const registry = defineModules([
      module('c', {dependencies: {required: ['a']}}),
      module('b', {dependencies: {required: ['a']}}),
      module('a'),
    ]);

    expect(registry.getModules().map(item => item.id)).toEqual(['a', 'b', 'c']);
  });

  it('ignores an absent optional dependency', () => {
    const registry = defineModules([
      module('a', {dependencies: {optional: ['b']}}),
    ]);

    expect(registry.getModules().map(item => item.id)).toEqual(['a']);
  });

  it('orders a present optional dependency before its dependant', () => {
    const registry = defineModules([
      module('a', {dependencies: {optional: ['b']}}),
      module('b'),
    ]);

    expect(registry.getModules().map(item => item.id)).toEqual(['b', 'a']);
  });

  it('reports a readable dependency cycle', () => {
    expect(() => defineModules([
      module('b', {dependencies: {required: ['a']}}),
      module('a', {dependencies: {required: ['b']}}),
    ])).toThrowError(new CircularModuleDependencyError(['a', 'b', 'a']));
  });
});

describe('ModuleRegistry ownership', () => {
  it('rejects normalized route collisions and names both owners', () => {
    expect(() => defineModules([
      module('a', {routes: {'/projects': {page}}}),
      module('b', {routes: {'/projects/': {page}}}),
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

  it('rejects duplicate global query IDs', () => {
    expect(() => defineModules([
      module('a', {queries: [{id: 'projects.list'}]}),
      module('b', {queries: [{id: 'projects.list'}]}),
    ])).toThrowError(new DuplicateModuleQueryError('projects.list', 'a', 'b'));
  });

  it('rejects duplicate global operation IDs', () => {
    expect(() => defineModules([
      module('a', {operations: [{id: 'projects.delete'}]}),
      module('b', {operations: [{id: 'projects.delete'}]}),
    ])).toThrowError(
      new DuplicateModuleOperationError('projects.delete', 'a', 'b'),
    );
  });

  it('retains simple route, query and operation ownership metadata', () => {
    const query = {id: 'projects.list'};
    const operation = {id: 'projects.delete'};
    const registry = defineModules([
      module('a', {queries: [query], operations: [operation]}),
    ]);

    expect(registry.getRoutes()[0]).toMatchObject({moduleId: 'a', path: '/a'});
    expect(registry.getQueryOwner(query.id)).toBe('a');
    expect(registry.getOperationOwner(operation.id)).toBe('a');
    expect(registry.getQueryContributions()).toEqual([
      {moduleId: 'a', contribution: query},
    ]);
    expect(registry.getOperationContributions()).toEqual([
      {moduleId: 'a', contribution: operation},
    ]);
  });

  it('returns immutable module metadata', () => {
    const definition = defineModule({
      id: 'a',
      title: 'A',
      routes: {'/a': {page}},
      dependencies: {required: ['base']},
    });

    expect(Object.isFrozen(definition)).toBe(true);
    expect(Object.isFrozen(definition.dependencies?.required)).toBe(true);
  });
});

describe('route helpers', () => {
  it.each([
    ['', '/'],
    ['resource-manager/', '/resource-manager'],
    ['//a///b/', '/a/b'],
  ])('normalizes %s', (input, expected) =>
    expect(normalizePath(input)).toBe(expected));
});
