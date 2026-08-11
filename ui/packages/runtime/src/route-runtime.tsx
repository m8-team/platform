'use client';

import {useEffect, useMemo, useState, type ReactNode} from 'react';
import {useStateStore} from '@json-render/react';
import {useM8Query} from '@m8/query/react';
import {resolveInput, type M8InputResolutionContext} from '@m8/query';
import type {M8QueryBinding, M8RouteAccess, M8RuntimeContext} from '@m8/core';
import {usePathname} from 'next/navigation';

import {useM8Runtime} from './provider';

function errorValue(error: unknown): unknown {
  return error instanceof Error ? {name: error.name, message: error.message} : error;
}

function QueryBinding({name, binding, sources}: {
  name: string;
  binding: M8QueryBinding;
  sources: M8InputResolutionContext;
}) {
  const {context} = useM8Runtime();
  const store = useStateStore();
  const input = useMemo(() => resolveInput(binding.input ?? {}, sources), [binding.input, sources]);
  const enabled = binding.enabled === undefined ? true : Boolean(resolveInput(binding.enabled, sources));
  const result = useM8Query(binding.query, input, enabled, context);
  useEffect(() => {
    store.update({
      [`/queries/${name}/data`]: result.data,
      [`/queries/${name}/loading`]: result.isLoading || result.isFetching,
      [`/queries/${name}/error`]: result.error ? errorValue(result.error) : null,
    });
  }, [name, result.data, result.error, result.isFetching, result.isLoading, store]);
  return null;
}

function resolveRouteParams(pattern: string, pathname: string): Record<string, string | string[]> {
  const segments = pattern.split('/').filter(Boolean);
  const values = pathname.split('/').filter(Boolean);
  const result: Record<string, string | string[]> = {};
  segments.forEach((segment, index) => {
    const optionalCatchAll = segment.match(/^\[\[\.\.\.(.+)\]\]$/);
    const catchAll = segment.match(/^\[\.\.\.(.+)\]$/);
    const dynamic = segment.match(/^\[(.+)\]$/);
    if (optionalCatchAll) result[optionalCatchAll[1]!] = values.slice(index);
    else if (catchAll) result[catchAll[1]!] = values.slice(index);
    else if (dynamic && values[index] !== undefined) result[dynamic[1]!] = values[index]!;
  });
  return result;
}

export function RouteRuntimeBoundary({bindings = {}, access, path = '/', children}: {
  bindings?: Readonly<Record<string, M8QueryBinding>>;
  access?: M8RouteAccess;
  path?: string;
  children?: ReactNode;
}) {
  const runtime = useM8Runtime();
  const store = useStateStore();
  const pathname = usePathname();
  const params = useMemo(() => resolveRouteParams(path, pathname), [path, pathname]);
  const [allowed, setAllowed] = useState(access?.permission ? null : true);
  const state = store.state;
  const sources = useMemo(() => ({
    state,
    params,
    queries: store.get('/queries'),
    context: runtime.context,
  }), [params, runtime.context, state, store]);

  useEffect(() => {
    if (!access?.permission) return setAllowed(true);
    const controller = new AbortController();
    void Promise.resolve(runtime.authorization?.can({
      permission: access.permission,
      context: runtime.context,
      signal: controller.signal,
    }) ?? false).then(value => { if (!controller.signal.aborted) setAllowed(value); });
    return () => controller.abort();
  }, [access?.permission, runtime.authorization, runtime.context]);

  if (allowed === null) return null;
  if (!allowed) return <div role="alert">Access denied</div>;
  return <>{Object.entries(bindings).map(([name, binding]) => (
    <QueryBinding key={name} name={name} binding={binding} sources={sources} />
  ))}{children}</>;
}
