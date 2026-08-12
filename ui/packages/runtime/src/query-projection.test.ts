import {createStateStore} from '@json-render/core';
import {QueryClient} from '@tanstack/react-query';
import {describe, expect, it, vi} from 'vitest';

import {projectQueryResult} from './query-projection';

describe('projectQueryResult', () => {
  it('does not write an equivalent projection twice', () => {
    const store = createStateStore({__runtime: {queryResults: {}}});
    const set = vi.spyOn(store, 'set');
    const first = {
      status: 'success' as const,
      data: {id: 'prj_1'},
      error: null,
      fetching: false,
    };

    projectQueryResult(store, 'project', first);
    projectQueryResult(store, 'project', {...first});

    expect(set).toHaveBeenCalledOnce();
  });

  it('writes when a stable projection field changes', () => {
    const store = createStateStore({__runtime: {queryResults: {}}});
    const set = vi.spyOn(store, 'set');
    const projection = {
      status: 'pending' as const,
      data: undefined,
      error: null,
      fetching: true,
    };

    projectQueryResult(store, 'project', projection);
    projectQueryResult(store, 'project', {...projection, fetching: false});

    expect(set).toHaveBeenCalledTimes(2);
  });

  it('cannot mutate the TanStack Query source of truth through its UI projection', () => {
    const queryClient = new QueryClient();
    const queryKey = ['query', 'projects.get', 'prj_1'];
    const source = {id: 'prj_1', version: '1'};
    queryClient.setQueryData(queryKey, source);
    const store = createStateStore({__runtime: {queryResults: {}}});

    projectQueryResult(store, 'project', {
      status: 'success',
      data: queryClient.getQueryData(queryKey),
      error: null,
      fetching: false,
    });
    store.set('/__runtime/queryResults/project/data/version', '2');

    expect(queryClient.getQueryData(queryKey)).toEqual(source);
    expect(store.get('/__runtime/queryResults/project/data/version')).toBe('2');
  });
});
