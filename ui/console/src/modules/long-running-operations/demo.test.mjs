import assert from 'node:assert/strict'
import {afterEach, test} from 'node:test'
import {
  advanceDemoOperation,
  cancelDemoOperation,
  getDemoOperation,
  resetOperationsDemo,
  retryDemoOperation,
} from './demo.ts'

afterEach(() => resetOperationsDemo())

test('Retry creates a new operation and preserves original failure', () => {
  const retryId = retryDemoOperation('op-2026-1038')
  assert.equal(retryId, 'op-2026-1043')
  const retried = getDemoOperation(retryId)
  assert.equal(retried?.status, 'Queued')
  assert.equal(retried?.attempt, 3)
  assert.equal(retried?.parentOperationId, 'op-2026-1038')
  assert.equal(getDemoOperation('op-2026-1038')?.status, 'Failed')
})

test('Cancellation is requested and confirmed separately', () => {
  assert.equal(cancelDemoOperation('op-2026-1042'), true)
  assert.equal(getDemoOperation('op-2026-1042')?.status, 'CancelRequested')
  assert.equal(advanceDemoOperation('op-2026-1042'), true)
  assert.equal(getDemoOperation('op-2026-1042')?.status, 'Cancelled')
})

test('Terminal operations cannot be cancelled or advanced', () => {
  assert.equal(cancelDemoOperation('op-2026-1039'), false)
  assert.equal(advanceDemoOperation('op-2026-1039'), false)
  assert.equal(retryDemoOperation('op-2026-1039'), null)
})

test('Queued operation advances without skipping to success', () => {
  const id = 'op-2026-1037'
  assert.equal(getDemoOperation(id)?.status, 'Queued')
  assert.equal(advanceDemoOperation(id), true)
  assert.equal(getDemoOperation(id)?.status, 'Running')
  assert.equal(getDemoOperation(id)?.progress, 24)
  for (let i=0; i<4; i++) advanceDemoOperation(id)
  assert.equal(getDemoOperation(id)?.status, 'Succeeded')
  assert.equal(getDemoOperation(id)?.progress, 100)
})

test('Reset restores seeded demo operations', () => {
  cancelDemoOperation('op-2026-1042')
  resetOperationsDemo()
  assert.equal(getDemoOperation('op-2026-1042')?.status, 'Running')
})
