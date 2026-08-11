import type {M8ExpressionValue} from '@m8/core';

export interface M8InputResolutionContext {
  readonly state?: unknown;
  readonly params?: unknown;
  readonly context?: unknown;
  readonly queries?: unknown;
}

function readPath(source: unknown, path: string): unknown {
  const segments = path.replace(/^\//, '').split('/').filter(Boolean);
  let current = source;
  for (const segment of segments) {
    if (current === null || typeof current !== 'object') return undefined;
    current = (current as Record<string, unknown>)[segment.replace(/~1/g, '/').replace(/~0/g, '~')];
  }
  return current;
}

function isBinding(value: unknown, key: '$state' | '$param' | '$query' | '$context'): value is Record<typeof key, string> {
  return value !== null && typeof value === 'object' &&
    Object.keys(value).length === 1 && typeof (value as Record<string, unknown>)[key] === 'string';
}

export function resolveInput(
  value: M8ExpressionValue,
  sources: M8InputResolutionContext,
): unknown {
  if (isBinding(value, '$state')) return readPath(sources.state, value.$state);
  if (isBinding(value, '$param')) return readPath(sources.params, value.$param);
  if (isBinding(value, '$query')) return readPath(sources.queries, value.$query);
  if (isBinding(value, '$context')) return readPath(sources.context, value.$context);
  if (value !== null && typeof value === 'object' && '$literal' in value) {
    return (value as {$literal: unknown}).$literal;
  }
  if (Array.isArray(value)) return value.map(item => resolveInput(item, sources));
  if (value !== null && typeof value === 'object') {
    return Object.fromEntries(
      Object.entries(value).map(([key, item]) => [key, resolveInput(item, sources)]),
    );
  }
  return value;
}
