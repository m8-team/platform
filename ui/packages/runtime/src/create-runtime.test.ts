import {
  createStateStore,
  resolveAction,
  resolveActionParam,
  resolveBindings,
  resolveElementProps,
} from '@json-render/core';
import {createNextApp} from '@json-render/next/server';
import {
  defineModule,
  defineModules,
  selectEnabledModules,
} from '@m8/core';
import {defineOperation, UnknownOperationError} from '@m8/operation';
import {defineQuery, UnknownQueryError} from '@m8/query';
import {describe, expect, it, vi} from 'vitest';
import {z} from 'zod';

import {createRuntime} from './create-runtime';
import {createActionHandlers} from './actions';
import {
  buildNextAppSpec,
  runtimeNextLoaders,
} from './next';
import {createRuntimeState, withRuntimeState} from './state';

const page = {
  root: 'root',
  elements: {root: {type: 'Text', props: {}, children: []}},
};

describe('runtime composition', () => {
  it('registers typed module routes, queries and operations without guessing', async () => {
    const query = defineQuery({
      id: 'example.items.list',
      input: z.object({}),
      output: z.array(z.string()),
      queryKey: () => ['items'],
      execute: async () => ['one'],
    });
    const operation = defineOperation({
      id: 'example.items.create',
      input: z.object({}),
      output: z.object({ok: z.boolean()}),
      execute: async () => ({ok: true}),
    });
    const exampleModule = defineModule({
      id: 'example',
      title: 'Example',
      queries: [query],
      operations: [operation],
      routes: {
        '/example': {
          queries: {items: {query: query.id}},
          page,
        },
      },
    });
    const modules = defineModules([exampleModule]);
    const runtime = createRuntime({modules});

    await expect(runtime.queries.execute(
      query.id,
      {},
      new AbortController().signal,
    )).resolves.toEqual(['one']);
    await expect(runtime.operations.execute(operation.id, {}))
      .resolves.toEqual({ok: true});
    expect(buildNextAppSpec(modules).routes['/example']).toBeDefined();
  });

  it('bridges a json-render action to OperationRuntime', async () => {
    const execute = vi.fn(async ({input}: {input: {id: string}}) => input);
    const operation = defineOperation({
      id: 'example.items.delete',
      input: z.object({id: z.string()}),
      output: z.object({id: z.string()}),
      execute,
    });
    const modules = defineModules([defineModule({
      id: 'example',
      title: 'Example',
      operations: [operation],
    })]);
    const context = {actor: {id: 'usr_1'}};
    const handlers = createActionHandlers({
      operations: createRuntime({modules}).operations,
      getContext: () => context,
    });

    await expect(handlers.executeOperation({
      operation: operation.id,
      input: {id: 'item_1'},
    })).resolves.toEqual({id: 'item_1'});
    expect(execute).toHaveBeenCalledWith(expect.objectContaining({context}));
  });

  it('composes a dependency-sorted fixture into one final NextAppSpec', async () => {
    const baseQuery = defineQuery({
      id: 'base.items.get',
      input: z.object({}),
      output: z.string(),
      execute: async () => 'base',
    });
    const featureQuery = defineQuery({
      id: 'feature.items.get',
      input: z.object({}),
      output: z.string(),
      execute: async () => 'feature',
    });
    const baseOperation = defineOperation({
      id: 'base.items.create',
      input: z.object({}),
      output: z.string(),
      execute: async () => 'created',
    });
    const featureOperation = defineOperation({
      id: 'feature.items.delete',
      input: z.object({}),
      output: z.string(),
      execute: async () => 'deleted',
    });
    const baseModule = defineModule({
      id: 'base',
      title: 'Base',
      queries: [baseQuery],
      operations: [baseOperation],
      routes: {'/base': {page}},
    });
    const featureModule = defineModule({
      id: 'feature',
      title: 'Feature',
      dependencies: {required: ['base']},
      queries: [featureQuery],
      operations: [featureOperation],
      routes: {'/feature': {page}},
    });
    const optionalModule = defineModule({
      id: 'optional',
      title: 'Optional',
      dependencies: {optional: ['base']},
      routes: {'/optional': {page}},
    });
    const modules = defineModules([
      featureModule,
      optionalModule,
      baseModule,
    ]);
    const runtime = createRuntime({modules});
    const spec = buildNextAppSpec(modules);

    expect(modules.getModules().map(module => module.id))
      .toEqual(['base', 'feature', 'optional']);
    expect(Object.keys(spec.routes))
      .toEqual(['/base', '/feature', '/optional']);
    expect(runtime.queries.registry.has(baseQuery.id)).toBe(true);
    expect(runtime.queries.registry.has(featureQuery.id)).toBe(true);
    expect(runtime.operations.registry.get(baseOperation.id)).toBe(baseOperation);
    expect(runtime.operations.registry.get(featureOperation.id))
      .toBe(featureOperation);
  });

  it('does not register queries or operations from disabled modules', () => {
    const featureQuery = defineQuery({
      id: 'feature.items.get',
      input: z.object({}),
      output: z.string(),
      execute: async () => 'feature',
    });
    const featureOperation = defineOperation({
      id: 'feature.items.delete',
      input: z.object({}),
      output: z.string(),
      execute: async () => 'deleted',
    });
    const baseModule = defineModule({id: 'base', title: 'Base'});
    const featureModule = defineModule({
      id: 'feature',
      title: 'Feature',
      dependencies: {required: ['base']},
      queries: [featureQuery],
      operations: [featureOperation],
    });
    const enabled = selectEnabledModules([baseModule, featureModule], {
      enabledModuleIds: ['base'],
    });
    const runtime = createRuntime({modules: defineModules(enabled)});

    expect(() => runtime.queries.registry.require(featureQuery.id))
      .toThrow(UnknownQueryError);
    expect(() => runtime.operations.registry.require(featureOperation.id))
      .toThrow(UnknownOperationError);
  });

  it('includes or excludes application navigation after authorization', async () => {
    const modules = defineModules([defineModule({
      id: 'example',
      title: 'Example',
      routes: {
        '/example': {
          access: {permission: 'example.read'},
          navigation: {label: 'Overview'},
          page,
        },
      },
    })]);
    const runtime = createRuntime({
      modules,
      adapters: {
        authorization: {
          can: ({context, permission}) =>
            context.permissions?.includes(permission) ?? false,
        },
      },
    });

    await expect(runtime.navigation.getItems({permissions: ['example.read']}))
      .resolves.toMatchObject([{moduleId: 'example', path: '/example'}]);
    await expect(runtime.navigation.getItems({permissions: []}))
      .resolves.toHaveLength(0);
    await expect(createRuntime({modules}).navigation.getItems())
      .resolves.toHaveLength(0);
  });

  it('receives canonical named route params from the json-render loader', async () => {
    const modules = defineModules([defineModule({
      id: 'example',
      title: 'Example',
      routes: {'/examples/[exampleId]': {page}},
    })]);
    const {getPageData} = createNextApp({
      spec: buildNextAppSpec(modules),
      loaders: runtimeNextLoaders,
    });
    const data = await getPageData({
      params: Promise.resolve({slug: ['examples', 'ex_1']}),
    });

    expect(data?.initialState).toMatchObject({
      __runtime: {params: {exampleId: 'ex_1'}},
    });
  });
});

describe('json-render integration', () => {
  it('uses the system state namespace for params, context and query projections', () => {
    const initial = createRuntimeState({
      params: {projectId: 'prj_1'},
      context: {organizationId: 'org_1'},
    });
    const state = {
      __runtime: {
        ...initial.__runtime,
        queryResults: {project: {data: {id: 'prj_1'}}},
      },
    };

    expect(resolveElementProps({
      projectId: {$state: '/__runtime/params/projectId'},
      organizationId: {$state: '/__runtime/context/organizationId'},
      resultId: {$state: '/__runtime/queryResults/project/data/id'},
    }, {stateModel: state})).toEqual({
      projectId: 'prj_1',
      organizationId: 'org_1',
      resultId: 'prj_1',
    });
  });

  it('delegates bindings and nested operation input resolution to json-render', () => {
    const store = createStateStore({
      filters: {search: ''},
      project: {id: 'prj_1', version: '3'},
    });
    const bindings = resolveBindings(
      {value: {$bindState: '/filters/search'}},
      {stateModel: store.getSnapshot()},
    );
    store.set(bindings?.value ?? '', 'updated');

    const input = resolveActionParam({
      projectId: {$state: '/project/id'},
      version: {$state: '/project/version'},
    }, {stateModel: store.getSnapshot()});
    const action = resolveAction({
      action: 'executeOperation',
      params: {operation: 'projects.delete', input},
    }, store.getSnapshot());

    expect(store.get('/filters/search')).toBe('updated');
    expect(action.params.input).toEqual({projectId: 'prj_1', version: '3'});
    expect(resolveAction({
      action: 'navigate',
      params: {href: '/resource-manager/projects'},
    }, store.getSnapshot())).toMatchObject({
      action: 'navigate',
      params: {href: '/resource-manager/projects'},
    });
  });

  it('protects the runtime namespace from route-owned initial state', () => {
    expect(() => withRuntimeState({__runtime: {}})).toThrow(/reserved/);
  });

  it.each([
    {
      id: 'binding',
      element: {
        type: 'TextInput',
        props: {value: {$bindState: '/__runtime/queryResults/project'}},
        children: [],
      },
    },
    {
      id: 'action',
      element: {
        type: 'Button',
        props: {},
        on: {
          press: {
            action: 'setState',
            params: {
              statePath: '/__runtime/queryResults/project',
              value: null,
            },
          },
        },
        children: [],
      },
    },
    {
      id: 'lifecycle',
      element: {
        type: 'Button',
        props: {},
        on: {
          press: {
            action: 'executeOperation',
            onSuccess: {
              set: {'/__runtime/queryResults/project': null},
            },
          },
        },
        children: [],
      },
    },
  ])('rejects UI $id writes to the read-only runtime projection', ({element}) => {
    const modules = defineModules([defineModule({
      id: 'example',
      title: 'Example',
      routes: {
        '/example': {
          page: {root: 'root', elements: {root: element}},
        },
      },
    })]);

    expect(() => buildNextAppSpec(modules)).toThrow(/cannot write reserved/);
  });

  it('rejects repeated two-way bindings into projected query rows', () => {
    const modules = defineModules([defineModule({
      id: 'example',
      title: 'Example',
      routes: {
        '/example': {
          page: {
            root: 'rows',
            elements: {
              rows: {
                type: 'Stack',
                props: {},
                repeat: {
                  statePath: '/__runtime/queryResults/projects/data/items',
                },
                children: ['name'],
              },
              name: {
                type: 'TextInput',
                props: {value: {$bindItem: 'name'}},
                children: [],
              },
            },
          },
        },
      },
    })]);

    expect(() => buildNextAppSpec(modules)).toThrow(/cannot write reserved/);
  });
});
