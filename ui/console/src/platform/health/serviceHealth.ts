/**
 * The console's health probes are separate from the business API.
 * Add future services here once their reverse proxy routes are configured.
 */
export const monitoredServices = [
  {
    id: 'resource-manager',
    name: 'Resource Manager',
    baseUrl: (import.meta.env.VITE_RESOURCE_MANAGER_HEALTH_URL || '/health/resource-manager').replace(/\/+$/, ''),
  },
] as const

export const healthProbes = [
  {path: '/livez', titleKey: 'health.probe.liveness'},
  {path: '/readyz', titleKey: 'health.probe.readiness'},
  {path: '/startupz', titleKey: 'health.probe.startup'},
  {path: '/healthz', titleKey: 'health.probe.diagnostics'},
] as const

export type HealthProbePath = (typeof healthProbes)[number]['path']
export type HealthStatus = 'HEALTHY' | 'DEGRADED' | 'UNHEALTHY' | 'UNKNOWN'
export type HealthDisplayStatus = HealthStatus | 'UNREACHABLE' | 'CHECKING'

export interface HealthCheck {
  name: string
  status: HealthStatus
  message?: string
  error?: string
  /** Go time.Duration is encoded as nanoseconds. */
  latency?: number
  checked_at?: string
  target?: {kind?: string; name?: string; module?: string}
  criticality?: 'REQUIRED' | 'OPTIONAL'
}

export interface HealthSnapshot {
  status: HealthStatus
  kind: string
  checked_at: string
  results?: HealthCheck[]
}

export interface HealthProbeResult {
  path: HealthProbePath
  status: HealthDisplayStatus
  httpStatus?: number
  durationMs: number
  checkedAt: string
  snapshot?: HealthSnapshot
  error?: string
}

const validStatuses: readonly string[] = ['HEALTHY', 'DEGRADED', 'UNHEALTHY', 'UNKNOWN']

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
}

function isHealthSnapshot(value: unknown): value is HealthSnapshot {
  if (!isRecord(value) || !validStatuses.includes(String(value.status)) ||
      typeof value.kind !== 'string' || typeof value.checked_at !== 'string') {
    return false
  }
  if (value.results === undefined) return true
  return Array.isArray(value.results) && value.results.every((result: unknown) =>
    isRecord(result) && typeof result.name === 'string' &&
    validStatuses.includes(String(result.status)),
  )
}

/** Keep transport failures distinct from valid UNHEALTHY/UNKNOWN snapshots (HTTP 503). */
export async function fetchHealthProbe(
  service: (typeof monitoredServices)[number],
  path: HealthProbePath,
  signal: AbortSignal,
): Promise<HealthProbeResult> {
  const started = performance.now()
  const checkedAt = new Date().toISOString()

  try {
    const response = await fetch(`${service.baseUrl}${path}`, {
      method: 'GET',
      headers: {Accept: 'application/json'},
      cache: 'no-store',
      credentials: 'same-origin',
      signal: AbortSignal.any([signal, AbortSignal.timeout(6000)]),
    })
    if (response.status !== 200 && response.status !== 503) {
      throw new Error(`HTTP ${response.status}`)
    }

    const payload: unknown = await response.json()
    if (!isHealthSnapshot(payload)) {
      throw new Error('Invalid health response')
    }

    return {
      path,
      status: payload.status,
      httpStatus: response.status,
      durationMs: Math.round(performance.now() - started),
      checkedAt,
      snapshot: payload,
    }
  } catch (error) {
    return {
      path,
      status: 'UNREACHABLE',
      durationMs: Math.round(performance.now() - started),
      checkedAt,
      error: error instanceof Error ? error.message : String(error),
    }
  }
}
