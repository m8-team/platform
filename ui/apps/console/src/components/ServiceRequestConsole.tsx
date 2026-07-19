import {useSyncExternalStore} from 'react'
import {Button, Label, Text} from '@gravity-ui/uikit'

import {serviceRequestLog} from '../platform/http/serviceRequestLog'
import type {ServiceRequestRecord} from '../platform/http/serviceRequestLog'
import type {Translate} from '../i18n'

export function ServiceRequestConsole({t}: {t: Translate}) {
  const records = useSyncExternalStore(
    serviceRequestLog.subscribe,
    serviceRequestLog.getSnapshot,
    serviceRequestLog.getSnapshot,
  )

  return (
    <div className="m8-request-console">
      <div className="m8-request-console__header">
        <div>
          <Text as="h2" variant="header-1">{t('requestConsole.title')}</Text>
          <Text variant="caption-2" color="secondary">{t('requestConsole.description')}</Text>
        </div>
        <Button view="outlined" disabled={records.length === 0} onClick={serviceRequestLog.clear}>
          {t('requestConsole.clear')}
        </Button>
      </div>
      <div className="m8-request-console__records">
        {records.length === 0 ? (
          <Text variant="body-2" color="secondary">{t('requestConsole.empty')}</Text>
        ) : records.map((record) => (
          <details className="m8-request-console__record" key={record.id}>
            <summary>
              <span className="m8-mono">{record.method}</span>
              <Text ellipsis>{record.url}</Text>
              <Label theme={statusTheme(record)}>{statusText(record)}</Label>
            </summary>
            <dl>
              <dt>{t('requestConsole.service')}</dt><dd>{record.service}</dd>
              <dt>{t('requestConsole.parameters')}</dt><dd><RequestValue value={record.parameters} /></dd>
              <dt>{t('requestConsole.requestHeaders')}</dt><dd><RequestValue value={record.requestHeaders} /></dd>
              <dt>{t('requestConsole.requestBody')}</dt><dd><RequestValue value={record.requestBody} /></dd>
              <dt>{t('requestConsole.startedAt')}</dt><dd>{new Date(record.startedAt).toLocaleTimeString()}</dd>
              <dt>{t('requestConsole.duration')}</dt><dd>{record.durationMs === undefined ? '—' : `${record.durationMs} ms`}</dd>
              <dt>{t('requestConsole.responseStatus')}</dt><dd>{statusText(record)}</dd>
              <dt>{t('requestConsole.responseHeaders')}</dt><dd><RequestValue value={record.responseHeaders} /></dd>
              <dt>{t('requestConsole.responseBody')}</dt>
              <dd>
                <RequestValue
                  value={record.responseBody}
                  pending={record.responseBodyPending}
                  pendingText={t('requestConsole.responseBodyPending')}
                />
              </dd>
              {record.error ? <><dt>{t('requestConsole.error')}</dt><dd>{record.error}</dd></> : null}
            </dl>
          </details>
        ))}
      </div>
    </div>
  )
}

function RequestValue({
  value,
  pending = false,
  pendingText,
}: {
  value: unknown
  pending?: boolean
  pendingText?: string
}) {
  if (pending) return <Text variant="caption-2" color="secondary">{pendingText ?? '…'}</Text>
  if (value === undefined) return <Text variant="caption-2" color="secondary">—</Text>
  return <pre>{formatRequestValue(value)}</pre>
}

function formatRequestValue(value: unknown) {
  if (typeof value === 'string') return value
  try {
    return JSON.stringify(value, null, 2)
  } catch {
    return '[VALUE_UNAVAILABLE]'
  }
}

function statusText(record: ServiceRequestRecord) {
  if (record.pending) return '…'
  return record.status === undefined ? 'ERR' : String(record.status)
}

function statusTheme(record: ServiceRequestRecord) {
  if (record.pending) return 'info' as const
  return record.status !== undefined && record.status < 400 ? 'success' as const : 'danger' as const
}
