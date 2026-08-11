import {describe, expect, it} from 'vitest';

import {defineModule} from './define-module';
import {
  BasePathCollisionError,
  CircularModuleDependencyError,
  DuplicateModuleError,
  DuplicateOperationError,
  DuplicateQueryError,
  MissingModuleDependencyError,
  RouteCollisionError,
} from './errors';
import {defineModules} from './module-registry';
import {joinRoute, normalizePath} from './routes';

const page = {root: 'root', elements: {root: {type: 'Text', props: {}, children: []}}};

function module(overrides: Record<string, unknown> = {}) {
  return defineModule({
    id: 'first',
    title: 'First',
    basePath: '/first',
    routes: {'/': {page}},
    ...overrides,
  });
}

describe('ModuleRegistry', () => {
  it('rejects duplicate module IDs', () => {
    expect(() => defineModules([module(), module()])).toThrow(DuplicateModuleError);
  });

  it('rejects duplicate normalized base paths', () => {
    expect(() => defineModules([
      module(),
      module({id: 'second', basePath: '/first/'}),
    ])).toThrow(BasePathCollisionError);
  });

  it.each([
    ['query', {queries: [{id: 'first.shared'}]}, DuplicateQueryError],
    ['operation', {operations: [{id: 'first.shared'}]}, DuplicateOperationError],
  ])('rejects duplicate %s IDs', (_kind, contribution, ErrorType) => {
    expect(() => defineModules([
      module(contribution),
      module({id: 'second', basePath: '/second', ...contribution}),
    ])).toThrow(ErrorType);
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
      module({id: 'second', basePath: '/second', dependencies: {required: ['first']}}),
    ])).toThrow(CircularModuleDependencyError);
  });

  it('rejects collisions after route joining', () => {
    expect(() => defineModules([
      module({basePath: '/', routes: {'/first': {page}}}),
      module({id: 'second', basePath: '/first', routes: {'/': {page}}}),
    ])).toThrow(RouteCollisionError);
  });

  it.each([
    ['/projects/[id]', '/projects/[projectId]'],
    ['/projects/[...slug]', '/projects/[...path]'],
    ['/projects/[[...slug]]', '/projects/[[...path]]'],
  ])('rejects canonical dynamic collision %s and %s', (first, second) => {
    expect(() => defineModules([
      module({basePath: '/', routes: {[first]: {page}, [second]: {page}}}),
    ])).toThrow(RouteCollisionError);
  });

  it('rejects an unknown route query', () => {
    expect(() => defineModules([module({routes: {'/': {page, queries: {data: {query: 'first.missing'}}}}})]))
      .toThrow(/unknown query/);
  });

  it('returns immutable module contributions', () => {
    const definition = module({queries: [{id: 'first.list'}]});
    expect(Object.isFrozen(definition)).toBe(true);
    expect(Object.isFrozen((definition as unknown as {queries: readonly unknown[]}).queries)).toBe(true);
  });
});

describe('route helpers', () => {
  it.each([
    ['', '/'],
    ['resource-manager/', '/resource-manager'],
    ['//a///b/', '/a/b'],
  ])('normalizes %s', (input, expected) => expect(normalizePath(input)).toBe(expected));

  it.each([
    ['/resource-manager', '/', '/resource-manager'],
    ['/', '/projects', '/projects'],
    ['/resource-manager/', '/projects', '/resource-manager/projects'],
  ])('joins %s and %s', (base, route, expected) => expect(joinRoute(base, route)).toBe(expected));
});
