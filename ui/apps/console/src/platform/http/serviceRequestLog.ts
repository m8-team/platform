export interface ServiceRequestRecord {
  id: string
  service: string
  method: string
  url: string
  parameters: Record<string, string[]>
  requestHeaders: Record<string, string[]>
  requestBody?: unknown
  responseHeaders?: Record<string, string[]>
  responseBody?: unknown
  responseBodyPending?: boolean
  startedAt: string
  durationMs?: number
  status?: number
  error?: string
  pending: boolean
}

let records: ServiceRequestRecord[] = []
const listeners = new Set<() => void>()
const maximumBodyPreviewBytes = 32 * 1024
const maximumCollectionEntries = 50
const maximumValueLength = 1024

export const isServiceRequestLoggingEnabled =
  import.meta.env.DEV || import.meta.env.VITE_ENABLE_REQUEST_CONSOLE === 'true'

export const serviceRequestLog = {
  subscribe(listener: () => void) {
    listeners.add(listener)
    return () => listeners.delete(listener)
  },
  getSnapshot() {
    return records
  },
  clear() {
    records = []
    emitChange()
  },
}

export async function loggedFetch(service: string, input: RequestInfo | URL, init: RequestInit = {}) {
  if (!isServiceRequestLoggingEnabled) return fetch(input, init)

  const startedAt = performance.now()
  const url = resolveURL(input)
  const id = crypto.randomUUID()
  const record: ServiceRequestRecord = {
    id,
    service,
    method: init.method?.toUpperCase() ?? (input instanceof Request ? input.method : 'GET'),
    url: url.pathname,
    parameters: collectParameters(url.searchParams),
    requestHeaders: collectRequestHeaders(input, init),
    requestBody: sanitizeRequestBody(init.body),
    startedAt: new Date().toISOString(),
    pending: true,
  }
  records = [record, ...records].slice(0, 100)
  emitChange()

  try {
    const response = await fetch(input, init)
    const responseCapture = prepareResponseCapture(record.method, response)
    updateRecord(id, {
      status: response.status,
      responseHeaders: collectHeaders(response.headers),
      responseBody: responseCapture.omitted,
      responseBodyPending: Boolean(responseCapture.response),
      durationMs: Math.round(performance.now() - startedAt),
      pending: false,
    })
    if (responseCapture.response) {
      void captureResponseBody(responseCapture.response).then((responseBody) => {
        updateRecord(id, {responseBody, responseBodyPending: false})
      })
    }
    return response
  } catch (error) {
    updateRecord(id, {
      error: sanitizeSensitiveText(limitValue(error instanceof Error ? error.message : String(error))),
      durationMs: Math.round(performance.now() - startedAt),
      pending: false,
    })
    throw error
  }
}

function resolveURL(input: RequestInfo | URL) {
  const value = input instanceof Request ? input.url : input.toString()
  return new URL(value, window.location.origin)
}

function collectParameters(searchParams: URLSearchParams) {
  const result: Record<string, string[]> = {}
  let entries = 0
  for (const [key, value] of searchParams) {
    if (entries >= maximumCollectionEntries) {
      result.__truncated__ = ['Additional parameters omitted']
      break
    }
    entries += 1
    const currentValues = result[key] ?? []
    if (currentValues.length >= 5) continue
    const safeValue = isSensitiveKey(key) ? '[REDACTED]' : limitValue(value)
    result[key] = [...currentValues, safeValue]
  }
  return result
}

function collectRequestHeaders(input: RequestInfo | URL, init: RequestInit) {
  const headers = init.headers ?? (input instanceof Request ? input.headers : undefined)
  return collectHeaders(new Headers(headers))
}

function collectHeaders(headers: Headers) {
  const result: Record<string, string[]> = {}
  let entries = 0
  for (const [key, value] of headers) {
    if (entries >= maximumCollectionEntries) {
      result.__truncated__ = ['Additional headers omitted']
      break
    }
    result[key] = [sanitizeHeaderValue(key, value)]
    entries += 1
  }
  return result
}

function sanitizeHeaderValue(key: string, value: string) {
  if (isSensitiveKey(key)) return '[REDACTED]'
  if (key.toLowerCase() === 'location') return sanitizeLocation(value)
  if (!isSafeHeaderKey(key)) return '[OMITTED]'
  return limitValue(sanitizeSensitiveText(value))
}

function sanitizeLocation(value: string) {
  try {
    const location = new URL(value, window.location.origin)
    return `${location.origin}${location.pathname}`
  } catch {
    return '[OMITTED]'
  }
}

function isSafeHeaderKey(key: string) {
  return /^(accept|accept-language|content-type|content-length|cache-control|pragma|expires|date|etag|last-modified|if-match|if-none-match|retry-after|vary|allow|server-timing|traceparent|x-request-id|x-correlation-id|x-trace-id|grpc-status|grpc-message|access-control-(?:allow|expose)-.*)$/i.test(
    key,
  )
}

function sanitizeRequestBody(body: BodyInit | null | undefined): unknown {
  if (!body) return undefined
  if (body instanceof URLSearchParams) return collectParameters(body)
  if (body instanceof FormData) return '[FORM_DATA_OMITTED]'
  if (body instanceof Blob) return `[BLOB_BODY_OMITTED: ${body.size} bytes]`
  if (typeof body !== 'string' || body.length === 0) return undefined
  if (body.length > 8192) return '[BODY_OMITTED: exceeds 8 KiB]'

  try {
    return redact(JSON.parse(body) as unknown)
  } catch {
    if (body.includes('=')) {
      const parameters = new URLSearchParams(body)
      if ([...parameters.keys()].length > 0) return collectParameters(parameters)
    }
    return '[NON_JSON_BODY_OMITTED]'
  }
}

function responseCanHaveBody(method: string, status: number) {
  return method !== 'HEAD' && status !== 204 && status !== 205 && status !== 304
}

interface PreparedResponseCapture {
  response?: Response
  omitted?: unknown
}

function prepareResponseCapture(method: string, response: Response): PreparedResponseCapture {
  if (!responseCanHaveBody(method, response.status)) return {}
  if (response.type === 'opaque') return {omitted: '[RESPONSE_BODY_OMITTED: opaque response]'}

  const contentType = response.headers.get('content-type')?.split(';', 1)[0].trim().toLowerCase() ?? ''
  if (!isJSONContentType(contentType)) {
    return {omitted: `[RESPONSE_BODY_OMITTED: ${contentType || 'missing content type'}]`}
  }

  const contentLength = Number(response.headers.get('content-length'))
  if (Number.isFinite(contentLength) && contentLength > maximumBodyPreviewBytes) {
    return {omitted: `[RESPONSE_BODY_OMITTED: exceeds ${maximumBodyPreviewBytes / 1024} KiB]`}
  }

  try {
    return {response: response.clone()}
  } catch {
    return {omitted: '[RESPONSE_BODY_UNAVAILABLE]'}
  }
}

async function captureResponseBody(response: Response): Promise<unknown> {
  try {
    const text = await readLimitedResponseText(response)
    if (text.startsWith('[RESPONSE_BODY_OMITTED:')) return text
    if (!text) return undefined
    try {
      return redact(JSON.parse(text) as unknown)
    } catch {
      return limitValue(sanitizeSensitiveText(text), maximumBodyPreviewBytes)
    }
  } catch {
    return '[RESPONSE_BODY_UNAVAILABLE]'
  }
}

async function readLimitedResponseText(response: Response) {
  const reader = response.body?.getReader()
  if (!reader) return ''

  const decoder = new TextDecoder()
  let bytesRead = 0
  let text = ''
  let timedOut = false
  const timeout = window.setTimeout(() => {
    timedOut = true
    void reader.cancel('response capture timeout')
  }, 5000)

  try {
    while (true) {
      const {done, value} = await reader.read()
      if (timedOut) return '[RESPONSE_BODY_OMITTED: capture timeout]'
      if (done) return text + decoder.decode()
      bytesRead += value.byteLength
      if (bytesRead > maximumBodyPreviewBytes) {
        void reader.cancel('response capture size limit')
        return `[RESPONSE_BODY_OMITTED: exceeds ${maximumBodyPreviewBytes / 1024} KiB]`
      }
      text += decoder.decode(value, {stream: true})
    }
  } finally {
    window.clearTimeout(timeout)
  }
}

function isJSONContentType(contentType: string) {
  return contentType === 'application/json' || contentType.endsWith('+json')
}

function sanitizeSensitiveText(value: string) {
  return value
    .replace(/\bBearer\s+[^\s,;]+/gi, 'Bearer [REDACTED]')
    .replace(/\beyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b/g, '[REDACTED_JWT]')
    .replace(
      /((?:authorization|access[-_]?token|refresh[-_]?token|password|secret|cookie|api[-_]?key|credential|signature|session|otp|assertion)[\s"'=:]+)([^\s,;"'}&]+)/gi,
      '$1[REDACTED]',
    )
}

function redact(value: unknown, depth = 0): unknown {
  if (depth >= 8) return '[TRUNCATED: maximum depth]'
  if (typeof value === 'string') return limitValue(sanitizeSensitiveText(value))
  if (Array.isArray(value)) {
    const items = value.slice(0, maximumCollectionEntries).map((item) => redact(item, depth + 1))
    return value.length > items.length ? [...items, '[TRUNCATED: additional items omitted]'] : items
  }
  if (!value || typeof value !== 'object') return value

  const entries = Object.entries(value)
  const preview = Object.fromEntries(
    entries
      .slice(0, maximumCollectionEntries)
      .map(([key, item]) => [key, isSensitiveKey(key) ? '[REDACTED]' : redact(item, depth + 1)]),
  )
  if (entries.length > maximumCollectionEntries) preview.__truncated__ = 'Additional fields omitted'
  return preview
}

function isSensitiveKey(key: string) {
  return /authorization|token|password|secret|cookie|api[-_]?key|credential|signature|session|otp|assertion|saml|client[-_]?data|csrf|baggage|(^|[-_])code($|[-_])/i.test(
    key,
  )
}

function limitValue(value: string, maximumLength = maximumValueLength) {
  return value.length > maximumLength ? `${value.slice(0, maximumLength)}…` : value
}

function updateRecord(id: string, update: Partial<ServiceRequestRecord>) {
  records = records.map((record) => (record.id === id ? {...record, ...update} : record))
  emitChange()
}

function emitChange() {
  listeners.forEach((listener) => listener())
}
