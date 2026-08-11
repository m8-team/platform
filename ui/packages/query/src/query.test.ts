import {describe, expect, it, vi} from 'vitest';
import {z} from 'zod';

import {defineQuery} from './definition';
import {DuplicateQueryError} from './errors';
import {normalizeQueryInput, QueryRegistry} from './registry';
import {resolveInput} from './resolve-input';

function query(execute = vi.fn(async ({input}: {input: {value: string}}) => ({value: input.value}))) {
  return defineQuery({
    id: 'test.get',
    input: z.object({value: z.string()}),
    output: z.object({value: z.string()}),
    queryKey: input => ['test', input.value],
    execute,
  });
}

describe('QueryRegistry', () => {
  it('rejects duplicate IDs', () => {
    expect(() => new QueryRegistry([query(), query()])).toThrow(DuplicateQueryError);
  });

  it('validates input and output and creates an ID-prefixed key', async () => {
    const registry = new QueryRegistry([query()]);
    expect(() => registry.queryKey('test.get', {})).toThrow();
    expect(registry.queryKey('test.get', {value: 'a'})).toEqual(['test.get', 'test', 'a']);
    await expect(registry.execute('test.get', {value: 'a'}, new AbortController().signal))
      .resolves.toEqual({value: 'a'});

    const invalidOutput = query(vi.fn(async () => ({value: 1} as never)));
    await expect(new QueryRegistry([invalidOutput]).execute(
      'test.get', {value: 'a'}, new AbortController().signal,
    )).rejects.toThrow();
  });

  it('forwards AbortSignal', async () => {
    const execute = vi.fn(async ({signal}: {input: {value: string}; signal: AbortSignal}) => {
      expect(signal).toBe(controller.signal);
      return {value: 'a'};
    });
    const controller = new AbortController();
    await new QueryRegistry([query(execute)]).execute('test.get', {value: 'a'}, controller.signal);
    expect(execute).toHaveBeenCalledOnce();
  });

  it('forwards runtime context', async () => {
    const context = {actor: {id: 'usr_1'}};
    const execute = vi.fn(async ({input, context: received}: {input: {value: string}; context: typeof context}) => {
      expect(received).toBe(context);
      return input;
    });
    await new QueryRegistry([query(execute)]).execute('test.get', {value: 'a'}, new AbortController().signal, context);
  });

  it('drops null unset optional fields and preserves nullable fields', () => {
    const schema = z.object({optional: z.string().optional(), nullable: z.string().nullable(), search: z.string()});
    expect(normalizeQueryInput(schema, {optional: null, nullable: null, search: ''}))
      .toEqual({nullable: null, search: ''});
    expect(schema.parse(normalizeQueryInput(schema, {optional: null, nullable: null, search: ''})))
      .toEqual({nullable: null, search: ''});
  });
});

describe('resolveInput', () => {
  const sources = {
    state: {filters: {search: 'needle'}},
    params: {projectId: 'prj_1'},
    context: {organization: {id: 'org_1'}},
    queries: {organization: {data: {id: 'org_2'}}},
  };

  it('resolves state, params and context recursively', () => {
    expect(resolveInput({
      search: {$state: '/filters/search'},
      projectId: {$param: '/projectId'},
      organizationId: {$context: '/organization/id'},
      queryOrganizationId: {$query: '/organization/data/id'},
      literal: {$literal: {$state: '/filters/search'}},
    }, sources)).toEqual({search: 'needle', projectId: 'prj_1', organizationId: 'org_1', queryOrganizationId: 'org_2', literal: {$state: '/filters/search'}});
  });

  it('returns undefined for missing paths', () => {
    expect(resolveInput({$state: '/missing'}, sources)).toBeUndefined();
  });
});
