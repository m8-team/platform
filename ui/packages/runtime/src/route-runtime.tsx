'use client';

import {useEffect, useMemo, useState, type ReactNode} from 'react';
import {resolveElementProps, resolvePropValue} from '@json-render/core';
import {useStateStore} from '@json-render/react';
import {useRegisteredQuery} from '@m8/query/react';
import type {QueryBinding, RouteAccess} from '@m8/core';

import {useRuntime} from './provider';
import {projectQueryResult} from './query-projection';

function errorValue(error: unknown): unknown {
  return error instanceof Error ? {name: error.name, message: error.message} : error;
}

function RegisteredQueryBinding({name, binding}: {name: string; binding: QueryBinding}) {
  const {context} = useRuntime();
  const store = useStateStore();
  const state = store.state;
  const input = useMemo(
    () => resolveElementProps(binding.input ?? {}, {stateModel: state}),
    [binding.input, state],
  );
  const enabled = binding.enabled === undefined
    ? true
    : Boolean(resolvePropValue(binding.enabled, {stateModel: state}));
  const result = useRegisteredQuery(binding.query, input, enabled, context);

  useEffect(() => {
    projectQueryResult(store, name, {
      status: result.status,
      data: result.data,
      error: result.error ? errorValue(result.error) : null,
      fetching: result.isFetching,
    });
  }, [name, result.data, result.error, result.isFetching, result.status, store]);
  return null;
}

export function RouteRuntimeBoundary({bindings = {}, access, children}: {
  bindings?: Readonly<Record<string, QueryBinding>>;
  access?: RouteAccess;
  children?: ReactNode;
}) {
  const runtime = useRuntime();
  const store = useStateStore();
  const [allowed, setAllowed] = useState(access?.permission ? null : true);

  useEffect(() => {
    store.set('/__runtime/context', runtime.context);
  }, [runtime.context, store]);

  useEffect(() => {
    if (!access?.permission) {
      setAllowed(true);
      return;
    }
    const controller = new AbortController();
    void Promise.resolve(runtime.authorization?.can({
      permission: access.permission,
      context: runtime.context,
      signal: controller.signal,
    }) ?? false).then(value => {
      if (!controller.signal.aborted) setAllowed(value);
    });
    return () => controller.abort();
  }, [access?.permission, runtime.authorization, runtime.context]);

  if (allowed === null) return null;
  if (!allowed) return <div role="alert">Access denied</div>;
  return <>{Object.entries(bindings).map(([name, binding]) => (
    <RegisteredQueryBinding key={name} name={name} binding={binding} />
  ))}{children}</>;
}
