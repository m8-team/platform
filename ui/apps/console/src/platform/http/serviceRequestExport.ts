import type {ServiceRequestRecord} from './serviceRequestLog'

/**
 * Exports the entire safely captured diagnostic snapshot of one request.
 * Values were redacted and bounded by serviceRequestCapture *before* logging.
 * Missing/omitted values are represented explicitly, never fetched again.
 */
export function formatServiceRequestForCopy(record: ServiceRequestRecord): string {
  const requestFinished = !record.pending
  const responseCaptured = requestFinished && !record.responseBodyPending

  return JSON.stringify({
    id: record.id,
    service: record.service,
    request: {
      method: record.method,
      url: record.origin ? `${record.origin}${record.url}` : record.url,
      parameters: record.parameters,
      headers: record.requestHeaders,
      body: record.requestBody ?? null,
    },
    response: {
      status: record.status ?? null,
      headers: record.responseHeaders ?? null,
      body: responseCaptured ? record.responseBody ?? null : '[RESPONSE_BODY_PENDING]',
    },
    timing: {
      startedAt: record.startedAt,
      durationMs: record.durationMs ?? null,
    },
    error: record.error ?? null,
    pending: record.pending || Boolean(record.responseBodyPending),
    capture: {
      sensitiveValuesRedacted: true,
      limitedByCapturePolicy: true,
      note: 'This is the complete captured diagnostic record. Secrets are redacted and large/non-JSON bodies or headers may be truncated or omitted.',
    },
  }, null, 2)
}
