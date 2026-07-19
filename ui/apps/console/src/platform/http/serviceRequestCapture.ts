const MAXIMUM_BODY_PREVIEW_BYTES = 32 * 1024
const MAXIMUM_REQUEST_BODY_BYTES = 8 * 1024
const MAXIMUM_COLLECTION_ENTRIES = 50
const MAXIMUM_VALUES_PER_PARAMETER = 5
const MAXIMUM_VALUE_LENGTH = 1024
const RESPONSE_CAPTURE_TIMEOUT_MS = 5000

export type ServiceRequestValues = Readonly<Record<string, readonly string[]>>
type MutableServiceRequestValues = Record<string, string[]>

export interface RequestPreview {
  method: string
  url: string
  parameters: ServiceRequestValues
  headers: ServiceRequestValues
  body?: unknown
}

export interface PreparedResponseBodyCapture {
  response?: Response
  omitted?: unknown
}

export function createRequestPreview(input: RequestInfo | URL, init: RequestInit): RequestPreview {
  const url = resolveURL(input)

  return {
    method: resolveMethod(input, init),
    url: url.pathname,
    parameters: collectParameters(url.searchParams),
    headers: collectRequestHeaders(input, init),
    body: sanitizeRequestBody(init.body),
  }
}

export function collectHeaders(headers: Headers): ServiceRequestValues {
  const result = createValuesCollection()
  let entries = 0

  for (const [key, value] of headers) {
    if (entries >= MAXIMUM_COLLECTION_ENTRIES) {
      result.__truncated__ = ['Additional headers omitted']
      break
    }

    result[key] = [sanitizeHeaderValue(key, value)]
    entries += 1
  }

  return result
}

export function prepareResponseBodyCapture(
  method: string,
  response: Response,
): PreparedResponseBodyCapture {
  if (!responseCanHaveBody(method, response.status)) return {}
  if (response.type === 'opaque') return {omitted: '[RESPONSE_BODY_OMITTED: opaque response]'}

  const contentType = response.headers.get('content-type')?.split(';', 1)[0].trim().toLowerCase() ?? ''
  if (!isJSONContentType(contentType)) {
    return {omitted: `[RESPONSE_BODY_OMITTED: ${contentType || 'missing content type'}]`}
  }

  const contentLength = Number(response.headers.get('content-length'))
  if (Number.isFinite(contentLength) && contentLength > MAXIMUM_BODY_PREVIEW_BYTES) {
    return {omitted: bodySizeLimitMessage('RESPONSE_BODY')}
  }

  try {
    return {response: response.clone()}
  } catch {
    return {omitted: '[RESPONSE_BODY_UNAVAILABLE]'}
  }
}

export async function captureResponseBody(response: Response): Promise<unknown> {
  try {
    const text = await readLimitedResponseText(response)
    if (text.startsWith('[RESPONSE_BODY_OMITTED:')) return text
    if (!text) return undefined

    try {
      return redact(JSON.parse(text) as unknown)
    } catch {
      return limitValue(sanitizeSensitiveText(text), MAXIMUM_BODY_PREVIEW_BYTES)
    }
  } catch {
    return '[RESPONSE_BODY_UNAVAILABLE]'
  }
}

export function sanitizeError(error: unknown) {
  const message = error instanceof Error ? error.message : String(error)
  return sanitizeSensitiveText(limitValue(message))
}

function resolveURL(input: RequestInfo | URL) {
  const value = input instanceof Request ? input.url : input.toString()
  return new URL(value, window.location.origin)
}

function resolveMethod(input: RequestInfo | URL, init: RequestInit) {
  return (init.method ?? (input instanceof Request ? input.method : 'GET')).toUpperCase()
}

function collectParameters(searchParams: URLSearchParams): ServiceRequestValues {
  const result = createValuesCollection()
  let entries = 0

  for (const [key, value] of searchParams) {
    if (entries >= MAXIMUM_COLLECTION_ENTRIES) {
      result.__truncated__ = ['Additional parameters omitted']
      break
    }

    entries += 1
    const currentValues = result[key] ?? []
    if (currentValues.length >= MAXIMUM_VALUES_PER_PARAMETER) continue

    const safeValue = isSensitiveParameterKey(key)
      ? '[REDACTED]'
      : limitValue(sanitizeSensitiveText(value))
    result[key] = [...currentValues, safeValue]
  }

  return result
}

function collectRequestHeaders(input: RequestInfo | URL, init: RequestInit) {
  const headers = init.headers ?? (input instanceof Request ? input.headers : undefined)
  return collectHeaders(new Headers(headers))
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
  if (body.length > MAXIMUM_REQUEST_BODY_BYTES) return bodySizeLimitMessage('BODY', MAXIMUM_REQUEST_BODY_BYTES)

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
  }, RESPONSE_CAPTURE_TIMEOUT_MS)

  try {
    while (true) {
      let chunk: ReadableStreamReadResult<Uint8Array<ArrayBufferLike>>
      try {
        chunk = await reader.read()
      } catch (error) {
        if (timedOut) return '[RESPONSE_BODY_OMITTED: capture timeout]'
        throw error
      }

      if (timedOut) return '[RESPONSE_BODY_OMITTED: capture timeout]'
      if (chunk.done) return text + decoder.decode()

      bytesRead += chunk.value.byteLength
      if (bytesRead > MAXIMUM_BODY_PREVIEW_BYTES) {
        void reader.cancel('response capture size limit')
        return bodySizeLimitMessage('RESPONSE_BODY')
      }
      text += decoder.decode(chunk.value, {stream: true})
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
    const items = value.slice(0, MAXIMUM_COLLECTION_ENTRIES).map((item) => redact(item, depth + 1))
    return value.length > items.length ? [...items, '[TRUNCATED: additional items omitted]'] : items
  }
  if (!value || typeof value !== 'object') return value

  const entries = Object.entries(value)
  const preview = Object.fromEntries(
    entries
      .slice(0, MAXIMUM_COLLECTION_ENTRIES)
      .map(([key, item]) => [key, isSensitiveKey(key) ? '[REDACTED]' : redact(item, depth + 1)]),
  )
  if (entries.length > MAXIMUM_COLLECTION_ENTRIES) preview.__truncated__ = 'Additional fields omitted'
  return preview
}

function isSensitiveKey(key: string) {
  const normalizedKey = key.replace(/([a-z\d])([A-Z])/g, '$1_$2').toLowerCase()

  return (
    /(?:^|[-_])(?:authorization|password|passwd|secret|cookie|credential|signature|session|otp|assertion|saml|csrf|baggage)(?:$|[-_])/.test(
      normalizedKey,
    ) ||
    /(?:^|[-_])(?:access|refresh|identity|id|auth)[-_]?token(?:$|[-_])/.test(normalizedKey) ||
    /(?:^|[-_])api[-_]?key(?:$|[-_])/.test(normalizedKey) ||
    /(?:^|[-_])client[-_]?data(?:$|[-_])/.test(normalizedKey) ||
    normalizedKey === 'token'
  )
}

function isSensitiveParameterKey(key: string) {
  return isSensitiveKey(key) || /^(?:authorization[-_]?code|auth[-_]?code|code)$/i.test(key)
}

function limitValue(value: string, maximumLength = MAXIMUM_VALUE_LENGTH) {
  return value.length > maximumLength ? `${value.slice(0, maximumLength)}…` : value
}

function bodySizeLimitMessage(kind: 'BODY' | 'RESPONSE_BODY', bytes = MAXIMUM_BODY_PREVIEW_BYTES) {
  return `[${kind}_OMITTED: exceeds ${bytes / 1024} KiB]`
}

function createValuesCollection(): MutableServiceRequestValues {
  return Object.create(null) as MutableServiceRequestValues
}
