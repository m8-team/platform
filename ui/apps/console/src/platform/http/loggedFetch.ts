import {
  captureResponseBody,
  collectHeaders,
  createRequestPreview,
  prepareResponseBodyCapture,
  sanitizeError,
} from './serviceRequestCapture'
import {prependServiceRequest, updateServiceRequest} from './serviceRequestLog'
import type {ServiceRequestRecordUpdate} from './serviceRequestLog'

interface ActiveRequestLog {
  id: string
  method: string
  startedAt: number
}

export const isServiceRequestLoggingEnabled =
  import.meta.env.DEV || import.meta.env.VITE_ENABLE_REQUEST_CONSOLE === 'true'

export async function loggedFetch(service: string, input: RequestInfo | URL, init: RequestInit = {}) {
  if (!isServiceRequestLoggingEnabled) return fetch(input, init)

  const activeRequest = tryStartRequestLog(service, input, init)

  let response: Response
  try {
    response = await fetch(input, init)
  } catch (error) {
    if (activeRequest) failRequestLog(activeRequest, error)
    throw error
  }

  if (activeRequest) completeRequestLog(activeRequest, response)
  return response
}

function tryStartRequestLog(
  service: string,
  input: RequestInfo | URL,
  init: RequestInit,
): ActiveRequestLog | undefined {
  try {
    const startedAt = performance.now()
    const request = createRequestPreview(input, init)
    const id = crypto.randomUUID()

    prependServiceRequest({
      id,
      service,
      method: request.method,
      url: request.url,
      parameters: request.parameters,
      requestHeaders: request.headers,
      requestBody: request.body,
      startedAt: new Date().toISOString(),
      pending: true,
    })

    return {id, method: request.method, startedAt}
  } catch {
    return undefined
  }
}

function completeRequestLog(activeRequest: ActiveRequestLog, response: Response) {
  try {
    const responseCapture = prepareResponseBodyCapture(activeRequest.method, response)

    updateServiceRequest(activeRequest.id, {
      status: response.status,
      responseHeaders: collectHeaders(response.headers),
      responseBody: responseCapture.omitted,
      responseBodyPending: Boolean(responseCapture.response),
      durationMs: elapsedMilliseconds(activeRequest.startedAt),
      pending: false,
    })

    if (responseCapture.response) {
      void captureResponseBody(responseCapture.response).then((responseBody) => {
        tryUpdateRequest(activeRequest.id, {responseBody, responseBodyPending: false})
      })
    }
  } catch {
    try {
      updateServiceRequest(activeRequest.id, {
        status: response.status,
        responseBody: '[RESPONSE_BODY_UNAVAILABLE]',
        responseBodyPending: false,
        durationMs: elapsedMilliseconds(activeRequest.startedAt),
        pending: false,
      })
    } catch {
      // A successful response must be returned even if its diagnostics fail.
    }
  }
}

function failRequestLog(activeRequest: ActiveRequestLog, error: unknown) {
  try {
    updateServiceRequest(activeRequest.id, {
      error: sanitizeError(error),
      durationMs: elapsedMilliseconds(activeRequest.startedAt),
      pending: false,
    })
  } catch {
    // Logging failures must not replace the original fetch error.
  }
}

function tryUpdateRequest(id: string, update: ServiceRequestRecordUpdate) {
  try {
    updateServiceRequest(id, update)
  } catch {
    // Response capture is best effort and has no effect on the caller.
  }
}

function elapsedMilliseconds(startedAt: number) {
  return Math.round(performance.now() - startedAt)
}
