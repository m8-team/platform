import type {StateStore} from '@json-render/core';

import {QUERY_RESULTS_STATE_PATH} from './state';

export interface QueryProjection {
  readonly status: 'pending' | 'error' | 'success';
  readonly data: unknown;
  readonly error: unknown;
  readonly fetching: boolean;
}

function sameError(left: unknown, right: unknown): boolean {
  if (left === right) return true;
  if (left === null || right === null ||
      typeof left !== 'object' || typeof right !== 'object') return false;
  if (!('name' in left) || !('message' in left) ||
      !('name' in right) || !('message' in right)) return false;
  return left.name === right.name && left.message === right.message;
}

function sameProjection(left: unknown, right: QueryProjection): boolean {
  return typeof left === 'object' && left !== null &&
    'status' in left && left.status === right.status &&
    'data' in left && left.data === right.data &&
    'error' in left && sameError(left.error, right.error) &&
    'fetching' in left && left.fetching === right.fetching;
}

export function projectQueryResult(
  store: Pick<StateStore, 'get' | 'set'>,
  name: string,
  projection: QueryProjection,
): void {
  // This is deliberately a one-way projection. Cache ownership, retries,
  // invalidation and freshness stay in TanStack Query.
  const path = `${QUERY_RESULTS_STATE_PATH}/${name}`;
  if (!sameProjection(store.get(path), projection)) {
    store.set(path, projection);
  }
}
