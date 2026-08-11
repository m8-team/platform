import {describe, expect, it, vi} from 'vitest';
import {z} from 'zod';

import {defineOperation} from './definition';
import {
  DuplicateOperationError,
  OperationAuthorizationError,
  UnknownOperationError,
} from './errors';
import {OperationRegistry} from './registry';
import {OperationRuntime, type OperationAuditEvent} from './runtime';

function operation(execute = vi.fn(async ({input}: {input: {id: string}}) => ({id: input.id}))) {
  return defineOperation({
    id: 'test.delete',
    input: z.object({id: z.string()}),
    output: z.object({id: z.string()}),
    requiredPermission: 'test.delete',
    invalidate: ['test.list'],
    execute,
  });
}

describe('OperationRuntime', () => {
  it('rejects duplicate IDs', () => {
    expect(() => new OperationRegistry([operation(), operation()])).toThrow(DuplicateOperationError);
  });

  it('rejects unknown IDs', () => {
    expect(() => new OperationRegistry().require('missing')).toThrow(UnknownOperationError);
  });

  it('denies unauthorized execution', async () => {
    const runtime = new OperationRuntime(new OperationRegistry([operation()]), {
      authorization: {check: async () => false},
    });
    await expect(runtime.execute('test.delete', {id: '1'})).rejects.toBeInstanceOf(OperationAuthorizationError);
  });

  it('validates, executes, audits and invalidates after success', async () => {
    const events: OperationAuditEvent[] = [];
    const invalidate = vi.fn(async () => undefined);
    const execute = vi.fn(async ({input, signal}: {input: {id: string}; signal: AbortSignal}) => {
      expect(signal).toBe(controller.signal);
      return input;
    });
    const controller = new AbortController();
    const runtime = new OperationRuntime(new OperationRegistry([operation(execute)]), {
      authorization: {check: async () => true},
      audit: {record: event => { events.push(event); }},
      queryInvalidation: {invalidate},
    });

    await expect(runtime.execute('test.delete', {id: '1'}, {signal: controller.signal}))
      .resolves.toEqual({id: '1'});
    expect(invalidate).toHaveBeenCalledWith('test.list');
    expect(events.map(event => event.phase)).toEqual(['started', 'succeeded']);
  });

  it('does not invalidate after failure and records failure', async () => {
    const invalidate = vi.fn();
    const events: OperationAuditEvent[] = [];
    const failing = operation(vi.fn(async () => { throw new Error('failed'); }));
    const runtime = new OperationRuntime(new OperationRegistry([failing]), {
      authorization: {check: async () => true},
      audit: {record: event => { events.push(event); }},
      queryInvalidation: {invalidate},
    });
    await expect(runtime.execute('test.delete', {id: '1'})).rejects.toThrow('failed');
    expect(invalidate).not.toHaveBeenCalled();
    expect(events.map(event => event.phase)).toEqual(['started', 'failed']);
  });

  it('validates input and output', async () => {
    const runtime = new OperationRuntime(new OperationRegistry([operation()]), {
      authorization: {check: async () => true},
    });
    await expect(runtime.execute('test.delete', {})).rejects.toThrow();

    const invalidOutput = operation(vi.fn(async () => ({id: 1} as never)));
    const outputRuntime = new OperationRuntime(new OperationRegistry([invalidOutput]), {
      authorization: {check: async () => true},
    });
    await expect(outputRuntime.execute('test.delete', {id: '1'})).rejects.toThrow();
  });

  it.each(['invalidation', 'audit'] as const)('%s failure does not change successful mutation result', async kind => {
    const report = vi.fn();
    const runtime = new OperationRuntime(new OperationRegistry([operation()]), {
      authorization: {check: async () => true},
      queryInvalidation: {invalidate: async () => { if (kind === 'invalidation') throw new Error('cache'); }},
      audit: {record: async event => { if (kind === 'audit' && event.phase === 'succeeded') throw new Error('audit'); }},
      errorReporter: {report},
    });
    await expect(runtime.execute('test.delete', {id: '1'})).resolves.toEqual({id: '1'});
    expect(report).toHaveBeenCalledOnce();
  });

  it.each([
    ['SUCCEEDED', true],
    ['FAILED', false],
    ['CANCELLED', false],
  ] as const)('long-running %s invalidates only on success', async (status, invalidates) => {
    const invalidate = vi.fn();
    const definition = defineOperation({
      id: 'test.create', mode: 'long-running',
      input: z.object({}), output: z.object({operationId: z.string()}),
      invalidate: ['test.list'],
      execute: async () => ({operationId: 'op_1'}),
    });
    const controller = new AbortController();
    const wait = vi.fn(async (_id: string, options?: {signal?: AbortSignal}) => {
      expect(options?.signal).toBe(controller.signal);
      return {id: 'op_1', status};
    });
    const runtime = new OperationRuntime(new OperationRegistry([definition]), {
      longRunningOperations: {wait}, queryInvalidation: {invalidate},
    });
    const result = runtime.execute('test.create', {}, {signal: controller.signal});
    if (status === 'SUCCEEDED') await expect(result).resolves.toEqual({operationId: 'op_1'});
    else await expect(result).rejects.toThrow();
    expect(invalidate).toHaveBeenCalledTimes(invalidates ? 1 : 0);
  });
});
