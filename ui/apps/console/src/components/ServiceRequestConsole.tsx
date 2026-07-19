import {useSyncExternalStore} from 'react'
import {TrashBin, Xmark} from '@gravity-ui/icons'
import {Button, DefinitionList, Icon, Label, Text} from '@gravity-ui/uikit'

import {JsonPreview} from './JsonPreview'
import {serviceRequestLog} from '../platform/http/serviceRequestLog'
import type {ServiceRequestRecord} from '../platform/http/serviceRequestLog'
import type {Translate} from '../i18n'

export function ServiceRequestConsole({t, onClose}: {t: Translate; onClose: () => void}) {
  const records = useSyncExternalStore(
    serviceRequestLog.subscribe,
    serviceRequestLog.getSnapshot,
    serviceRequestLog.getSnapshot,
  )

  return (
    <div className="m8-request-console">
      <div className="m8-request-console__header">
        <Text as="h2" variant="header-1">{t('requestConsole.title')}</Text>
        <div className="m8-request-console__header-actions">
          <Button
            view="flat"
            size="m"
            disabled={records.length === 0}
            aria-label={t('requestConsole.clear')}
            title={t('requestConsole.clear')}
            onClick={serviceRequestLog.clear}
          >
            <Icon data={TrashBin} size={18} />
          </Button>
          <Button
            view="flat"
            size="m"
            aria-label={t('requestConsole.close')}
            title={t('requestConsole.close')}
            onClick={onClose}
          >
            <Icon data={Xmark} size={18} />
          </Button>
        </div>
      </div>
      <div className="m8-request-console__records">
        {records.length === 0 ? (
          <Text variant="body-2" color="secondary">{t('requestConsole.empty')}</Text>
        ) : records.map((record) => (
          <details className="m8-request-console__record" key={record.id}>
            <summary className="m8-request-console__summary">
              <span className="m8-mono">{record.method}</span>
              <Text ellipsis>{record.url}</Text>
              <RequestStatusLabel record={record} pendingText={t('requestConsole.pending')} />
            </summary>
                <DefinitionList
                  className="m8-request-console__definitions"
                  direction="horizontal"
                  nameMaxWidth={145}
                  aria-label={`${record.method} ${record.url}`}
                >
                  <DefinitionList.Item name={t('requestConsole.service')}>{record.service}</DefinitionList.Item>
                  <DefinitionList.Item name={t('requestConsole.parameters')}>
                    <RequestValue value={record.parameters} t={t} />
                  </DefinitionList.Item>
                  <DefinitionList.Item name={t('requestConsole.requestHeaders')}>
                    <RequestValue value={record.requestHeaders} t={t} />
                  </DefinitionList.Item>
                  <DefinitionList.Item name={t('requestConsole.requestBody')}>
                    <RequestValue value={record.requestBody} t={t} />
                  </DefinitionList.Item>
                  <DefinitionList.Item name={t('requestConsole.startedAt')}>
                    {new Date(record.startedAt).toLocaleTimeString()}
                  </DefinitionList.Item>
                  <DefinitionList.Item name={t('requestConsole.duration')}>
                    {record.durationMs === undefined ? '—' : `${record.durationMs} ms`}
                  </DefinitionList.Item>
                  <DefinitionList.Item name={t('requestConsole.responseStatus')}>
                    <RequestStatusLabel record={record} pendingText={t('requestConsole.pending')} />
                  </DefinitionList.Item>
                  <DefinitionList.Item name={t('requestConsole.responseHeaders')}>
                    <RequestValue value={record.responseHeaders} t={t} />
                  </DefinitionList.Item>
                  <DefinitionList.Item name={t('requestConsole.responseBody')}>
                    <RequestValue
                      value={record.responseBody}
                      pending={record.responseBodyPending}
                      pendingText={t('requestConsole.responseBodyPending')}
                      t={t}
                    />
                  </DefinitionList.Item>
                  {record.error ? (
                    <DefinitionList.Item name={t('requestConsole.error')}>{record.error}</DefinitionList.Item>
                  ) : null}
                </DefinitionList>
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
  t,
}: {
  value: unknown
  pending?: boolean
  pendingText?: string
  t: Translate
}) {
  if (pending) return <Text variant="caption-2" color="secondary">{pendingText ?? '…'}</Text>
  if (value === undefined) return <Text variant="caption-2" color="secondary">—</Text>
  return (
    <JsonPreview
      value={value}
      copyText={t('resource.copy')}
      copiedText={t('resource.copied')}
      openText={t('requestConsole.openJson')}
      overlayTitle={t('requestConsole.jsonPreview')}
      closeText={t('requestConsole.closeJson')}
    />
  )
}

function statusText(record: ServiceRequestRecord) {
  return record.status === undefined ? 'ERR' : String(record.status)
}

function RequestStatusLabel({record, pendingText}: {record: ServiceRequestRecord; pendingText: string}) {
  return (
    <Label theme={statusTheme(record)} loading={record.pending}>
      {record.pending ? pendingText : statusText(record)}
    </Label>
  )
}

function statusTheme(record: ServiceRequestRecord) {
  if (record.pending) return 'info' as const
  return record.status !== undefined && record.status < 400 ? 'success' as const : 'danger' as const
}
