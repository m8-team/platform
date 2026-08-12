import {describe, expect, it, vi} from 'vitest';
import {z} from 'zod';

import {defineQuery} from './definition';
import {DuplicateQueryError, UnknownQueryError} from './errors';
import {normalizeQueryInput, QueryRegistry} from './registry';

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
  it('registers and looks up the original definition', () => {
    const definition = query();
    const registry = new QueryRegistry();

    expect(registry.register(definition)).toBe(definition);
    expect(registry.get(definition.id)).toBe(definition);
    expect(registry.require(definition.id)).toBe(definition);
  });

  it('rejects duplicate IDs', () => {
    expect(() => new QueryRegistry([query(), query()])).toThrow(DuplicateQueryError);
  });

  it('rejects unknown IDs', () => {
    expect(() => new QueryRegistry().require('missing')).toThrow(UnknownQueryError);
  });

  it('validates input and output and creates an ID-prefixed key', async () => {
    const registry = new QueryRegistry([query()]);
    expect(() => registry.queryKey('test.get', {})).toThrow();
    expect(registry.queryKey('test.get', {value: 'a'})).toEqual([
      'query', 'test.get', null, null, null, null, null, 'test', 'a',
    ]);
    await expect(registry.execute('test.get', {value: 'a'}, new AbortController().signal))
      .resolves.toEqual({value: 'a'});

    const invalidOutput = query(vi.fn(async () => ({value: 1} as never)));
    await expect(new QueryRegistry([invalidOutput]).execute(
      'test.get', {value: 'a'}, new AbortController().signal,
    )).rejects.toThrow();
  });

  it('isolates cache keys by tenant, resource scope and actor', () => {
    const registry = new QueryRegistry([query()]);
    const first = registry.queryKey('test.get', {value: 'a'}, {
      tenantId: 'tenant-a', organizationId: 'org-a', actor: {id: 'user-a'},
    });
    const second = registry.queryKey('test.get', {value: 'a'}, {
      tenantId: 'tenant-b', organizationId: 'org-a', actor: {id: 'user-a'},
    });
    expect(first).not.toEqual(second);
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
