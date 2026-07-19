import {useSyncExternalStore} from 'react'
import {Accordion, Button, DefinitionList, Label, Text} from '@gravity-ui/uikit'

import {JsonPreview} from './JsonPreview'
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
        <Text as="h2" variant="header-1">{t('requestConsole.title')}</Text>
        <Button view="outlined" disabled={records.length === 0} onClick={serviceRequestLog.clear}>
          {t('requestConsole.clear')}
        </Button>
      </div>
      <div className="m8-request-console__records">
        {records.length === 0 ? (
          <Text variant="body-2" color="secondary">{t('requestConsole.empty')}</Text>
        ) : (
          <Accordion
            className="m8-request-console__accordion"
            arrowPosition="end"
            ariaLevel={3}
            ariaLabel={t('requestConsole.title')}
          >
            {records.map((record) => (
              <Accordion.Item
                value={record.id}
                key={record.id}
                keepMounted={false}
                summary={(
                  <div className="m8-request-console__summary">
                    <span className="m8-mono">{record.method}</span>
                    <Text ellipsis>{record.url}</Text>
                    <Label theme={statusTheme(record)}>{statusText(record)}</Label>
                  </div>
                )}
              >
                <DefinitionList
                  className="m8-request-console__definitions"
                  direction="horizontal"
                  responsive
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
                    {statusText(record)}
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
              </Accordion.Item>
            ))}
          </Accordion>
        )}
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
  if (record.pending) return '…'
  return record.status === undefined ? 'ERR' : String(record.status)
}

function statusTheme(record: ServiceRequestRecord) {
  if (record.pending) return 'info' as const
  return record.status !== undefined && record.status < 400 ? 'success' as const : 'danger' as const
}
