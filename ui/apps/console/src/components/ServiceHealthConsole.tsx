import {useCallback, useEffect, useRef, useState} from 'react'
import {ArrowRotateRight, Xmark} from '@gravity-ui/icons'
import {Button, Icon, Label, Text} from '@gravity-ui/uikit'

import type {Translate, TranslationKey} from '../i18n'
import {
  fetchHealthProbe,
  healthProbes,
  monitoredServices,
} from '../platform/health/serviceHealth'
import type {
  HealthDisplayStatus,
  HealthProbePath,
  HealthProbeResult,
} from '../platform/health/serviceHealth'

type ServiceProbeResults = Partial<Record<HealthProbePath, HealthProbeResult>>
type ServiceHealthResults = Record<string, ServiceProbeResults>

const pollingIntervalMs = 15_000

const statusKeys: Record<HealthDisplayStatus, TranslationKey> = {
  HEALTHY: 'health.status.HEALTHY',
  DEGRADED: 'health.status.DEGRADED',
  UNHEALTHY: 'health.status.UNHEALTHY',
  UNKNOWN: 'health.status.UNKNOWN',
  UNREACHABLE: 'health.status.UNREACHABLE',
  CHECKING: 'health.status.CHECKING',
}

const statusThemes: Record<HealthDisplayStatus, 'success' | 'warning' | 'danger' | 'normal' | 'info'> = {
  HEALTHY: 'success',
  DEGRADED: 'warning',
  UNHEALTHY: 'danger',
  UNKNOWN: 'normal',
  UNREACHABLE: 'danger',
  CHECKING: 'info',
}

function StatusBadge({status, t}: {status: HealthDisplayStatus; t: Translate}) {
  return <Label theme={statusThemes[status]}>{t(statusKeys[status])}</Label>
}

function formatTime(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

function ProbeCard({probe, result, t}: {
  probe: (typeof healthProbes)[number]
  result?: HealthProbeResult
  t: Translate
}) {
  const status = result?.status ?? 'CHECKING'
  return (
    <section className="m8-health-probe">
      <div className="m8-health-probe__heading">
        <div>
          <Text variant="body-2" className="m8-health-probe__title">{t(probe.titleKey)}</Text>
          <Text variant="caption-2" color="secondary">{probe.path}</Text>
        </div>
        <StatusBadge status={status} t={t} />
      </div>
      {result ? (
        <>
          <div className="m8-health-probe__meta">
            {result.httpStatus !== undefined ? <Text variant="caption-2">HTTP {result.httpStatus}</Text> : null}
            <Text variant="caption-2" color="secondary">
              {result.durationMs} {t('health.milliseconds')}
            </Text>
          </div>
          {result.error ? (
            <Text variant="caption-2" color="danger" className="m8-health-probe__error">
              {t('health.transportError')}: {result.error}
            </Text>
          ) : null}
          {result.snapshot?.results?.length ? (
            <div className="m8-health-probe__checks">
              <Text variant="caption-2" color="secondary">{t('health.checks')}</Text>
              {result.snapshot.results.map((check, index) => (
                <div key={`${check.name}-${index}`} className="m8-health-probe__check">
                  <div className="m8-health-probe__check-heading">
                    <Text variant="body-2">{check.name}</Text>
                    <StatusBadge status={check.status} t={t} />
                  </div>
                  {check.target ? (
                    <Text variant="caption-2" color="secondary">
                      {[check.target.module, check.target.kind, check.target.name].filter(Boolean).join(' · ')}
                    </Text>
                  ) : null}
                  <Text variant="caption-2" color="secondary">
                    {check.criticality === 'REQUIRED'
                      ? t('health.required')
                      : check.criticality === 'OPTIONAL'
                        ? t('health.optional')
                        : ''}
                    {typeof check.latency === 'number'
                      ? ` · ${(check.latency / 1_000_000).toFixed(1)} ${t('health.milliseconds')}`
                      : ''}
                  </Text>
                  {check.message ? <Text variant="caption-2">{check.message}</Text> : null}
                  {check.error ? <Text variant="caption-2" color="danger">{check.error}</Text> : null}
                </div>
              ))}
            </div>
          ) : null}
        </>
      ) : null}
    </section>
  )
}

export function ServiceHealthConsole({open, t, onClose}: {
  open: boolean
  t: Translate
  onClose: () => void
}) {
  const [results, setResults] = useState<ServiceHealthResults>({})
  const [refreshing, setRefreshing] = useState(false)
  const activeRequest = useRef<AbortController | null>(null)

  const refresh = useCallback(async () => {
    activeRequest.current?.abort()
    const controller = new AbortController()
    activeRequest.current = controller
    setRefreshing(true)
    const observations = await Promise.all(
      monitoredServices.flatMap((service) =>
        healthProbes.map(async (probe) => ({
          serviceId: service.id,
          result: await fetchHealthProbe(service, probe.path, controller.signal),
        })),
      ),
    )
    if (controller.signal.aborted) return

    const next: ServiceHealthResults = {}
    for (const {serviceId, result} of observations) {
      if (!next[serviceId]) next[serviceId] = {}
      next[serviceId][result.path] = result
    }
    setResults(next)
    setRefreshing(false)
  }, [])

  useEffect(() => {
    if (!open) return
    void refresh()
    const timer = window.setInterval(() => { void refresh() }, pollingIntervalMs)
    return () => {
      window.clearInterval(timer)
      activeRequest.current?.abort()
    }
  }, [open, refresh])

  return (
    <div className="m8-health-console">
      <div className="m8-request-console__header">
        <div>
          <Text as="h2" variant="header-1">{t('health.title')}</Text>
          <Text variant="body-2" color="secondary">{t('health.description')}</Text>
        </div>
        <div className="m8-request-console__header-actions">
          <Button
            view="flat"
            size="m"
            loading={refreshing}
            aria-label={t('health.refresh')}
            title={t('health.refresh')}
            onClick={() => { void refresh() }}
          >
            <Icon data={ArrowRotateRight} size={18} />
          </Button>
          <Button view="flat" size="m" aria-label={t('health.close')} title={t('health.close')} onClick={onClose}>
            <Icon data={Xmark} size={18} />
          </Button>
        </div>
      </div>
      <Text variant="caption-2" color="secondary">
        {t('health.autoRefresh')}
      </Text>
      <div className="m8-health-services">
        {monitoredServices.map((service) => {
          const probes = results[service.id]
          const readiness = probes?.['/readyz']
          const status: HealthDisplayStatus = readiness?.status ?? 'CHECKING'

          return (
            <article key={service.id} className="m8-health-service">
              <div className="m8-health-service__heading">
                <div className="m8-health-service__title">
                  <Text as="h3" variant="subheader-2">{service.name}</Text>
                  <Text variant="caption-2" color="secondary">{service.id}</Text>
                </div>
                <StatusBadge status={status} t={t} />
              </div>
              <Text variant="caption-2" color="secondary">
                {t('health.overallReadiness')}
                {readiness?.snapshot ? ` · ${t('health.lastChecked')}: ${formatTime(readiness.snapshot.checked_at)}` : ''}
              </Text>
              <Text variant="caption-2" color="secondary" className="m8-health-service__url">
                {service.baseUrl}
              </Text>
              <div className="m8-health-service__probes">
                {healthProbes.map((probe) => (
                  <ProbeCard key={probe.path} probe={probe} result={probes?.[probe.path]} t={t} />
                ))}
              </div>
            </article>
          )
        })}
      </div>
    </div>
  )
}
