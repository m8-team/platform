import {describe, expect, it, vi} from 'vitest';
import {z} from 'zod';

import {defineOperation} from './definition';
import {
  DuplicateOperationError,
  OperationAuthorizationError,
  OperationConfirmationDeclinedError,
} from './errors';
import {OperationRegistry} from './registry';
import {M8OperationRuntime, type OperationAuditEvent} from './runtime';

function operation(execute = vi.fn(async ({input}: {input: {id: string}}) => ({id: input.id}))) {
  return defineOperation({
    id: 'test.delete',
    input: z.object({id: z.string()}),
    output: z.object({id: z.string()}),
    requiredPermission: 'test.delete',
    destructive: true,
    confirmation: {title: 'Delete?'},
    invalidate: ['test.list'],
    execute,
  });
}

describe('M8OperationRuntime', () => {
  it('rejects duplicate IDs', () => {
    expect(() => new OperationRegistry([operation(), operation()])).toThrow(DuplicateOperationError);
  });

  it('denies unauthorized execution', async () => {
    const runtime = new M8OperationRuntime(new OperationRegistry([operation()]), {
      authorization: {check: async () => false},
    });
    await expect(runtime.execute('test.delete', {id: '1'})).rejects.toBeInstanceOf(OperationAuthorizationError);
  });

  it('stops when confirmation is declined', async () => {
    const execute = vi.fn(async () => ({id: '1'}));
    const runtime = new M8OperationRuntime(new OperationRegistry([operation(execute)]), {
      authorization: {check: async () => true},
      confirmation: {confirm: async () => false},
    });
    await expect(runtime.execute('test.delete', {id: '1'})).rejects.toBeInstanceOf(OperationConfirmationDeclinedError);
    expect(execute).not.toHaveBeenCalled();
  });

  it('validates, executes, audits and invalidates after success', async () => {
    const events: OperationAuditEvent[] = [];
    const invalidate = vi.fn(async () => undefined);
    const execute = vi.fn(async ({input, signal}: {input: {id: string}; signal: AbortSignal}) => {
      expect(signal).toBe(controller.signal);
      return input;
    });
    const controller = new AbortController();
    const runtime = new M8OperationRuntime(new OperationRegistry([operation(execute)]), {
      authorization: {check: async () => true},
      confirmation: {confirm: async () => true},
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
    const runtime = new M8OperationRuntime(new OperationRegistry([failing]), {
      authorization: {check: async () => true},
      confirmation: {confirm: async () => true},
      audit: {record: event => { events.push(event); }},
      queryInvalidation: {invalidate},
    });
    await expect(runtime.execute('test.delete', {id: '1'})).rejects.toThrow('failed');
    expect(invalidate).not.toHaveBeenCalled();
    expect(events.map(event => event.phase)).toEqual(['started', 'failed']);
  });

  it('validates input and output', async () => {
    const runtime = new M8OperationRuntime(new OperationRegistry([operation()]), {
      authorization: {check: async () => true},
      confirmation: {confirm: async () => true},
    });
    await expect(runtime.execute('test.delete', {})).rejects.toThrow();
  });
});
