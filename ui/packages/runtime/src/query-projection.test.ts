import {createStateStore} from '@json-render/core';
import {describe, expect, it, vi} from 'vitest';

import {projectQueryResult} from './query-projection';

describe('projectQueryResult', () => {
  it('does not write an equivalent projection twice', () => {
    const store = createStateStore({__runtime: {queries: {}}});
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
    const store = createStateStore({__runtime: {queries: {}}});
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
});
