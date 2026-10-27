import {useCallback, useEffect, useRef, useState} from 'react'
import {ArrowRotateRight, Xmark} from '@gravity-ui/icons'
import {Button, DefinitionList, Icon, Label, Text} from '@gravity-ui/uikit'

import {JsonPreview} from './JsonPreview'
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
type MonitoredService = (typeof monitoredServices)[number]

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

function StatusLabel({status, t}: {status: HealthDisplayStatus; t: Translate}) {
  return <Label theme={statusThemes[status]}>{t(statusKeys[status])}</Label>
}

function formatTime(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

/** Services retain API Debug's compact expandable summary. Probes are a DefinitionList. */
function ServiceHealthItem({service, probes, t}: {
  service: MonitoredService
  probes?: ServiceProbeResults
  t: Translate
}) {
  const [expanded, setExpanded] = useState(false)
  const readiness = probes?.['/readyz']
  const status: HealthDisplayStatus = readiness?.status ?? 'CHECKING'

  return (
    <details
      className="m8-request-console__record m8-health-service"
      open={expanded}
      onToggle={(event) => setExpanded(event.currentTarget.open)}
    >
      <summary className="m8-request-console__summary m8-health-service__summary">
        <Text ellipsis>{service.name}</Text>
        <StatusLabel status={status} t={t} />
      </summary>
      {expanded ? (
        <div className="m8-health-service__details">
          <Text variant="caption-2" color="secondary" className="m8-health-service__url">
            {service.id} · {service.baseUrl}
          </Text>
          <DefinitionList
            className="m8-health-service__probes"
            direction="horizontal"
            nameMaxWidth={230}
            responsive
          >
            {healthProbes.map((probe) => {
              const result = probes?.[probe.path]
              const probeStatus: HealthDisplayStatus = result?.status ?? 'CHECKING'
              const checkedAt = result?.snapshot?.checked_at ?? result?.checkedAt

              return (
                <DefinitionList.Item
                  key={probe.path}
                  name={
                    <div className="m8-health-probe__term">
                      <Text variant="body-2">{t(probe.titleKey)}</Text>
                      <StatusLabel status={probeStatus} t={t} />
                    </div>
                  }
                >
                  <div className="m8-health-probe__value">
                    <Text variant="caption-2" color="secondary" className="m8-health-probe__time">
                      {checkedAt ? formatTime(checkedAt) : '—'}
                    </Text>
                    <ProbeDetails result={result} t={t} />
                  </div>
                </DefinitionList.Item>
              )
            })}
          </DefinitionList>
        </div>
      ) : null}
    </details>
  )
}

/** Diagnostic metadata follows the same DefinitionList pattern as API Debug. */
function ProbeDetails({result, t}: {
  result?: HealthProbeResult
  t: Translate
}) {
  return (
    <DefinitionList
      className="m8-request-console__definitions m8-health-probe__definitions"
      direction="horizontal"
      nameMaxWidth={135}
    >
      <DefinitionList.Item name={t('requestConsole.responseStatus')}>
        <ResponseStatusLabel result={result} t={t} />
      </DefinitionList.Item>
      <DefinitionList.Item name={t('requestConsole.duration')}>
        {result ? result.durationMs + ' ' + t('health.milliseconds') : '—'}
      </DefinitionList.Item>
      {result?.error ? (
        <DefinitionList.Item name={t('health.transportError')}>
          <Text color="danger" className="m8-health-probe__error">{result.error}</Text>
        </DefinitionList.Item>
      ) : null}
      <DefinitionList.Item name={t('health.checks')}>
        {result?.snapshot?.results?.length ? (
          <div className="m8-health-probe__checks">
            {result.snapshot.results.map((check, index) => (
              <div className="m8-health-probe__check" key={check.name + '-' + index}>
                <div className="m8-health-probe__check-heading">
                  <Text variant="body-2">{check.name}</Text>
                  <StatusLabel status={check.status} t={t} />
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
                    ? ' · ' + (check.latency / 1_000_000).toFixed(1) + ' ' + t('health.milliseconds')
                    : ''}
                </Text>
                {check.message ? <Text variant="caption-2">{check.message}</Text> : null}
                {check.error ? <Text variant="caption-2" color="danger">{check.error}</Text> : null}
              </div>
            ))}
          </div>
        ) : (
          <Text variant="caption-2" color="secondary">{t('health.noChecks')}</Text>
        )}
      </DefinitionList.Item>
      {result?.snapshot ? (
        <DefinitionList.Item name={t('health.rawResponse')}>
          <JsonPreview
            value={result.snapshot}
            copyText={t('resource.copy')}
            copiedText={t('resource.copied')}
            openText={t('requestConsole.openJson')}
            overlayTitle={t('requestConsole.jsonPreview')}
            closeText={t('requestConsole.closeJson')}
          />
        </DefinitionList.Item>
      ) : null}
    </DefinitionList>
  )
}

function ResponseStatusLabel({result, t}: {result?: HealthProbeResult; t: Translate}) {
  if (!result) return <Label theme="info" loading>{t('requestConsole.pending')}</Label>
  if (result.httpStatus !== undefined) {
    return <Label theme={result.httpStatus < 400 ? 'success' : 'danger'}>{result.httpStatus}</Label>
  }
  return <Label theme="danger">ERR</Label>
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
    // Schedule the initial poll outside the effect's synchronous render phase.
    const initialPoll = window.setTimeout(() => { void refresh() }, 0)
    const timer = window.setInterval(() => { void refresh() }, pollingIntervalMs)
    return () => {
      window.clearTimeout(initialPoll)
      window.clearInterval(timer)
      activeRequest.current?.abort()
    }
  }, [open, refresh])

  return (
    <div className="m8-request-console m8-health-console">
      <div className="m8-request-console__header">
        <Text as="h2" variant="header-1">{t('health.title')}</Text>
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
      <div className="m8-request-console__records m8-health-services">
        {monitoredServices.map((service) => (
          <ServiceHealthItem
            key={service.id}
            service={service}
            probes={results[service.id]}
            t={t}
          />
        ))}
      </div>
    </div>
  )
}
