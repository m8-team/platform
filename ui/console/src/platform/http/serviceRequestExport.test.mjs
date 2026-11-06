import assert from 'node:assert/strict'
import test from 'node:test'

import {formatServiceRequestForCopy} from './serviceRequestExport.ts'

function request(overrides = {}) {
  return {
    id: 'req-123',
    service: 'resource-manager',
    method: 'POST',
    origin: 'http://localhost:8080',
    url: '/resource-manager/v1/resources',
    parameters: {tenantId: ['tenant-a'], access_token: ['[REDACTED]']},
    requestHeaders: {accept: ['application/json'], authorization: ['[REDACTED]']},
    requestBody: {resource: {organization: {name: 'Example'}}},
    responseHeaders: {'content-type': ['application/json']},
    responseBody: {code: 5, message: 'resource not found'},
    startedAt: '2026-10-27T12:40:36.000Z',
    durationMs: 14,
    status: 404,
    pending: false,
    responseBodyPending: false,
    ...overrides,
  }
}

test('copies all captured fields for a completed HTTP error', () => {
  const result = JSON.parse(formatServiceRequestForCopy(request()))
  assert.equal(result.service, 'resource-manager')
  assert.equal(result.request.method, 'POST')
  assert.equal(result.request.url, 'http://localhost:8080/resource-manager/v1/resources')
  assert.deepEqual(result.request.parameters, {tenantId: ['tenant-a'], access_token: ['[REDACTED]']})
  assert.deepEqual(result.request.headers.authorization, ['[REDACTED]'])
  assert.equal(result.request.body.resource.organization.name, 'Example')
  assert.equal(result.response.status, 404)
  assert.equal(result.response.body.message, 'resource not found')
  assert.deepEqual(result.response.headers['content-type'], ['application/json'])
  assert.equal(result.timing.durationMs, 14)
  assert.equal(result.error, null)
  assert.equal(result.pending, false)
  assert.equal(result.capture.sensitiveValuesRedacted, true)
})

test('preserves explicit omission markers, and does not invent omitted data', () => {
  const result = JSON.parse(formatServiceRequestForCopy(request({
    requestBody: '[BODY_OMITTED: exceeds 8 KiB]',
    responseBody: '[RESPONSE_BODY_OMITTED: exceeds 32 KiB]',
  })))
  assert.equal(result.request.body, '[BODY_OMITTED: exceeds 8 KiB]')
  assert.equal(result.response.body, '[RESPONSE_BODY_OMITTED: exceeds 32 KiB]')
})

test('represents pending/error response and missing fields explicitly', () => {
  const result = JSON.parse(formatServiceRequestForCopy(request({
    origin: undefined,
    pending: true,
    responseBodyPending: true,
    status: undefined,
    responseHeaders: undefined,
    requestBody: undefined,
    responseBody: undefined,
    durationMs: undefined,
    error: 'Network failed',
  })))
  assert.equal(result.request.url, '/resource-manager/v1/resources')
  assert.equal(result.request.body, null)
  assert.equal(result.response.status, null)
  assert.equal(result.response.body, '[RESPONSE_BODY_PENDING]')
  assert.equal(result.timing.durationMs, null)
  assert.equal(result.error, 'Network failed')
  assert.equal(result.pending, true)
})
