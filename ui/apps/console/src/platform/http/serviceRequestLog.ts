export interface ServiceRequestRecord {
  id: string
  service: string
  method: string
  url: string
  parameters: Record<string, string[]>
  body?: unknown
  startedAt: string
  durationMs?: number
  status?: number
  error?: string
  pending: boolean
}

let records: ServiceRequestRecord[] = []
const listeners = new Set<() => void>()

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
    body: sanitizeBody(init.body),
    startedAt: new Date().toISOString(),
    pending: true,
  }
  records = [record, ...records].slice(0, 100)
  emitChange()

  try {
    const response = await fetch(input, init)
    updateRecord(id, {
      status: response.status,
      durationMs: Math.round(performance.now() - startedAt),
      pending: false,
    })
    return response
  } catch (error) {
    updateRecord(id, {
      error: error instanceof Error ? error.message : String(error),
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
  searchParams.forEach((value, key) => {
    const safeValue = isSensitiveKey(key) ? '[REDACTED]' : value
    result[key] = [...(result[key] ?? []), safeValue]
  })
  return result
}

function sanitizeBody(body: BodyInit | null | undefined): unknown {
  if (!body) return undefined
  if (body instanceof URLSearchParams) return collectParameters(body)
  if (body instanceof FormData) return '[FORM_DATA_OMITTED]'
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

function redact(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(redact)
  if (!value || typeof value !== 'object') return value
  return Object.fromEntries(
    Object.entries(value).map(([key, item]) => [
      key,
      isSensitiveKey(key) ? '[REDACTED]' : redact(item),
    ]),
  )
}

function isSensitiveKey(key: string) {
  return /authorization|token|password|secret|cookie|api[-_]?key/i.test(key)
}

function updateRecord(id: string, update: Partial<ServiceRequestRecord>) {
  records = records.map((record) => (record.id === id ? {...record, ...update} : record))
  emitChange()
}

function emitChange() {
  listeners.forEach((listener) => listener())
}
