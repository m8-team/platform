import {describe, expect, it} from 'vitest';

import {defineModule} from './define-module';
import {
  CircularModuleDependencyError,
  DuplicateModuleError,
  MissingModuleDependencyError,
  RouteCollisionError,
} from './errors';
import {defineModules} from './module-registry';
import {normalizePath} from './routes';

const page = {root: 'root', elements: {root: {type: 'Text', props: {}, children: []}}};

function module(overrides: Record<string, unknown> = {}) {
  return defineModule({
    id: 'first',
    title: 'First',
    routes: {'/first': {page}},
    ...overrides,
  });
}

describe('ModuleRegistry', () => {
  it('rejects duplicate module IDs', () => {
    expect(() => defineModules([module(), module()])).toThrow(DuplicateModuleError);
  });

  it('rejects missing required dependencies but permits missing optional ones', () => {
    expect(() => defineModules([
      module({dependencies: {required: ['missing']}}),
    ])).toThrow(MissingModuleDependencyError);
    expect(() => defineModules([
      module({dependencies: {optional: ['missing']}}),
    ])).not.toThrow();
  });

  it('rejects circular required dependencies', () => {
    expect(() => defineModules([
      module({dependencies: {required: ['second']}}),
      module({id: 'second', routes: {'/second': {page}}, dependencies: {required: ['first']}}),
    ])).toThrow(CircularModuleDependencyError);
  });

  it('rejects normalized route collisions', () => {
    expect(() => defineModules([
      module({routes: {'/first': {page}}}),
      module({id: 'second', routes: {'/first/': {page}}}),
    ])).toThrow(RouteCollisionError);
  });

  it.each([
    ['/projects/[id]', '/projects/[projectId]'],
    ['/projects/[...slug]', '/projects/[...path]'],
    ['/projects/[[...slug]]', '/projects/[[...path]]'],
  ])('rejects canonical dynamic collision %s and %s', (first, second) => {
    expect(() => defineModules([
      module({routes: {[first]: {page}, [second]: {page}}}),
    ])).toThrow(RouteCollisionError);
  });

  it('returns modules in dependency-safe declaration order for a valid graph', () => {
    const registry = defineModules([
      module(),
      module({id: 'second', routes: {'/second': {page}}, dependencies: {required: ['first']}}),
    ]);
    expect(registry.getModules().map(item => item.id)).toEqual(['first', 'second']);
  });

  it('returns immutable module metadata', () => {
    const definition = defineModule({
      id: 'first', title: 'First', routes: {'/first': {page}},
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
  ])('normalizes %s', (input, expected) => expect(normalizePath(input)).toBe(expected));
});
