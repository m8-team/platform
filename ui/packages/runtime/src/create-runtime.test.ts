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
  RouteCollisionError,
} from '@m8/core';
import {defineOperation} from '@m8/operation';
import {defineQuery} from '@m8/query';
import {describe, expect, it, vi} from 'vitest';
import {z} from 'zod';

import {createRuntime} from './create-runtime';
import {createActionHandlers} from './actions';
import {
  buildNextAppSpec,
  createRuntimeNextLoaders,
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

  it('composes every module permutation into the same runtime and NextAppSpec', async () => {
    const queryA = defineQuery({
      id: 'a.items.get',
      input: z.object({}),
      output: z.string(),
      execute: async () => 'a',
    });
    const queryB = defineQuery({
      id: 'b.items.get',
      input: z.object({}),
      output: z.string(),
      execute: async () => 'b',
    });
    const queryC = defineQuery({
      id: 'c.items.get',
      input: z.object({}),
      output: z.string(),
      execute: async () => 'c',
    });
    const operationA = defineOperation({
      id: 'a.items.create',
      input: z.object({}),
      output: z.string(),
      execute: async () => 'a',
    });
    const operationB = defineOperation({
      id: 'b.items.create',
      input: z.object({}),
      output: z.string(),
      execute: async () => 'b',
    });
    const operationC = defineOperation({
      id: 'c.items.create',
      input: z.object({}),
      output: z.string(),
      execute: async () => 'c',
    });
    const a = defineModule({
      id: 'a',
      title: 'A',
      queries: [queryA],
      operations: [operationA],
      routes: {
        '/a': {navigation: {label: 'A'}, page},
      },
    });
    const b = defineModule({
      id: 'b',
      title: 'B',
      queries: [queryB],
      operations: [operationB],
      routes: {
        '/b/[itemId]': {
          navigation: {label: 'B'},
          queries: {items: {query: queryA.id}},
          page,
        },
      },
    });
    const c = defineModule({
      id: 'c',
      title: 'C',
      queries: [queryC],
      operations: [operationC],
      routes: {
        '/c': {navigation: {label: 'C'}, page},
      },
    });
    const permutations = [
      [a, b, c],
      [a, c, b],
      [b, a, c],
      [b, c, a],
      [c, a, b],
      [c, b, a],
    ] as const;
    const baseSpec = {
      routes: {'/': {page}},
    };

    const snapshots = await Promise.all(permutations.map(async permutation => {
      const modules = defineModules(permutation);
      const runtime = createRuntime({modules});
      const spec = buildNextAppSpec(modules, {baseSpec});
      const queryIds = modules.getQueryContributions()
        .map(({contribution}) => contribution.id);
      const operationIds = modules.getOperationContributions()
        .map(({contribution}) => contribution.id);

      return {
        modules: modules.getModules().map(moduleDefinition => moduleDefinition.id),
        routes: modules.getRoutes().map(({moduleId, path}) => ({moduleId, path})),
        queries: queryIds.map(id => ({
          id,
          owner: modules.getQueryOwner(id),
          registered: runtime.queries.registry.has(id),
        })),
        operations: operationIds.map(id => ({
          id,
          owner: modules.getOperationOwner(id),
          registered: runtime.operations.registry.get(id)?.id,
        })),
        navigation: await runtime.navigation.getItems(),
        spec,
      };
    }));

    expect(snapshots[0]?.modules).toEqual(['a', 'b', 'c']);
    expect(Object.keys(snapshots[0]?.spec.routes ?? {}))
      .toEqual(['/', '/a', '/b/[itemId]', '/c']);
    for (const snapshot of snapshots.slice(1)) {
      expect(snapshot).toEqual(snapshots[0]);
    }
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

  it('composes route params with a custom json-render loader', async () => {
    const loadExample = vi.fn(async (
      params: Record<string, string | string[]>,
    ) => ({example: {id: params.exampleId}}));
    const modules = defineModules([defineModule({
      id: 'example',
      title: 'Example',
      routes: {
        '/examples/[exampleId]': {
          loader: 'loadExample',
          page,
        },
      },
    })]);
    const {getPageData} = createNextApp({
      spec: buildNextAppSpec(modules),
      loaders: createRuntimeNextLoaders({loadExample}),
    });
    const data = await getPageData({
      params: Promise.resolve({slug: ['examples', 'ex_1']}),
    });

    expect(loadExample).toHaveBeenCalledWith({exampleId: 'ex_1'});
    expect(data?.initialState).toMatchObject({
      example: {id: 'ex_1'},
      __runtime: {params: {exampleId: 'ex_1'}},
    });
  });

  it('rejects platform and module route collisions', () => {
    const modules = defineModules([defineModule({
      id: 'projects',
      title: 'Projects',
      routes: {'/projects/[projectId]': {page}},
    })]);

    expect(() => buildNextAppSpec(modules, {
      baseSpec: {
        routes: {'/projects/[id]': {page}},
      },
    })).toThrow(RouteCollisionError);
  });

  it('rejects exact platform and module route collisions', () => {
    const modules = defineModules([defineModule({
      id: 'projects',
      title: 'Projects',
      routes: {'/projects/': {page}},
    })]);

    expect(() => buildNextAppSpec(modules, {
      baseSpec: {
        routes: {'/projects': {page}},
      },
    })).toThrow(RouteCollisionError);
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

  it('protects the runtime namespace from custom loader data', async () => {
    const modules = defineModules([defineModule({
      id: 'example',
      title: 'Example',
      routes: {
        '/examples/[exampleId]': {loader: 'loadExample', page},
      },
    })]);
    const {getPageData} = createNextApp({
      spec: buildNextAppSpec(modules),
      loaders: createRuntimeNextLoaders({
        loadExample: async () => ({__runtime: {tampered: true}}),
      }),
    });

    await expect(getPageData({
      params: Promise.resolve({slug: ['examples', 'ex_1']}),
    })).rejects.toThrow(/reserved/);
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
