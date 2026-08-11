import {describe, expect, it} from 'vitest';
import {createStateStore, resolveAction, resolveBindings, resolveElementProps} from '@json-render/core';
import {z} from 'zod';
import {defineModule, defineModules} from '@m8/core';
import {defineQuery} from '@m8/query';
import {defineOperation} from '@m8/operation';

import {createRuntime} from './create-runtime';
import {buildNextAppSpec} from './next';
import {createRuntimeState, withRuntimeState} from './state';

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

describe('json-render integration', () => {
  it('uses the system state namespace for params, context and query projections', () => {
    const initial = createRuntimeState({
        params: {projectId: 'prj_1'},
        context: {organizationId: 'org_1'},
      });
    const state = {
      __runtime: {...initial.__runtime, queries: {project: {data: {id: 'prj_1'}}}},
    };
    expect(resolveElementProps({
      projectId: {$state: '/__runtime/params/projectId'},
      organizationId: {$state: '/__runtime/context/organizationId'},
      resultId: {$state: '/__runtime/queries/project/data/id'},
    }, {stateModel: state})).toEqual({
      projectId: 'prj_1', organizationId: 'org_1', resultId: 'prj_1',
    });
  });

  it('delegates two-way bindings and operation input resolution to json-render', () => {
    const store = createStateStore({filters: {search: ''}, form: {id: 'prj_1'}});
    const props = {value: {$bindState: '/filters/search'}};
    const bindings = resolveBindings(props, {stateModel: store.getSnapshot()});
    store.set(bindings?.value ?? '', 'updated');
    expect(store.get('/filters/search')).toBe('updated');
    expect(resolveAction({
      action: 'executeOperation',
      params: {operation: 'projects.delete', input: {$state: '/form'}},
    }, store.getSnapshot()).params.input).toEqual({id: 'prj_1'});
    expect(resolveAction({
      action: 'navigate', params: {href: '/resource-manager/projects'},
    }, store.getSnapshot())).toMatchObject({
      action: 'navigate', params: {href: '/resource-manager/projects'},
    });
  });

  it('protects the runtime namespace from route-owned state', () => {
    expect(() => withRuntimeState({__runtime: {}})).toThrow(/reserved/);
  });
});
