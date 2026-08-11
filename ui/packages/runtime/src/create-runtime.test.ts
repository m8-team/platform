import {describe, expect, it} from 'vitest';
import {z} from 'zod';
import {defineModule, defineModules} from '@m8/core';
import {defineQuery} from '@m8/query';
import {defineOperation} from '@m8/operation';

import {createRuntime} from './create-runtime';
import {buildNextAppSpec} from './next';

describe('runtime composition', () => {
  it('automatically registers module route, query and operation', async () => {
    const query = defineQuery({
      id: 'example.items.list', input: z.object({}), output: z.array(z.string()),
      queryKey: () => ['items'], execute: async () => ['one'],
    });
    const operation = defineOperation({
      id: 'example.items.create', input: z.object({}), output: z.object({ok: z.boolean()}),
      execute: async () => ({ok: true}),
    });
    const module = defineModule({
      id: 'example', title: 'Example',
      queries: [query], operations: [operation],
      routes: {'/example': {
        queries: {items: {query: query.id}},
        page: {root: 'root', elements: {root: {type: 'Text', props: {}, children: []}}},
      }},
    });
    const modules = defineModules([module]);
    const runtime = createRuntime({modules});

    await expect(runtime.queries.execute(query.id, {}, new AbortController().signal)).resolves.toEqual(['one']);
    await expect(runtime.operations.execute(operation.id, {})).resolves.toEqual({ok: true});
    expect(buildNextAppSpec(modules).routes['/example']).toBeDefined();
  });
});
