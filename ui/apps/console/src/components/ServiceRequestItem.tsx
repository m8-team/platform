import {memo, useState} from 'react'
import {Button, DefinitionList, Label, Text} from '@gravity-ui/uikit'

import {JsonPreview} from './JsonPreview'
import type {Translate} from '../i18n'
import type {ServiceRequestRecord} from '../platform/http/serviceRequestLog'
import {formatServiceRequestForCopy} from '../platform/http/serviceRequestExport'

interface ServiceRequestItemProps {
  record: ServiceRequestRecord
  open: boolean
  onOpenChange: (recordId: string, open: boolean) => void
  t: Translate
}

export const ServiceRequestItem = memo(function ServiceRequestItem({
  record,
  open,
  onOpenChange,
  t,
}: ServiceRequestItemProps) {
  return (
    <details
      className="m8-request-console__record"
      open={open}
      onToggle={(event) => onOpenChange(record.id, event.currentTarget.open)}
    >
      <summary className="m8-request-console__summary">
        <span className="m8-mono">{record.method}</span>
        <Text ellipsis>{record.url}</Text>
        <RequestStatusLabel record={record} pendingText={t('requestConsole.pending')} />
      </summary>
      {open ? <ServiceRequestDetails record={record} t={t} /> : null}
    </details>
  )
})

function ServiceRequestDetails({record, t}: {record: ServiceRequestRecord; t: Translate}) {
  const [copyStatus, setCopyStatus] = useState<'idle' | 'copied' | 'error'>('idle')

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(formatServiceRequestForCopy(record))
      setCopyStatus('copied')
    } catch {
      setCopyStatus('error')
    }
  }

  return (
    <>
      <div className="m8-request-console__copy-row">
        <Text variant="caption-2" color="secondary">{t('requestConsole.copyNotice')}</Text>
        <Button
          view="outlined"
          size="m"
          disabled={record.pending || Boolean(record.responseBodyPending)}
          onClick={() => void handleCopy()}
          title={record.pending || record.responseBodyPending
            ? t('requestConsole.copyPending') : t('requestConsole.copyFull')}
        >
          {copyStatus === 'copied' ? t('requestConsole.copyFullDone') : t('requestConsole.copyFull')}
        </Button>
      </div>
      {copyStatus === 'error' ? (
        <Text role="alert" color="danger" variant="caption-2" className="m8-request-console__copy-error">
          {t('requestConsole.copyFailed')}
        </Text>
      ) : null}
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
    </>
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

function RequestStatusLabel({record, pendingText}: {record: ServiceRequestRecord; pendingText: string}) {
  const status = getStatusPresentation(record, pendingText)

  return (
    <Label theme={status.theme} loading={record.pending}>
      {status.text}
    </Label>
  )
}

function getStatusPresentation(record: ServiceRequestRecord, pendingText: string) {
  if (record.pending) return {text: pendingText, theme: 'info' as const}
  if (record.status !== undefined) {
    return {
      text: String(record.status),
      theme: record.status < 400 ? ('success' as const) : ('danger' as const),
    }
  }
  return {text: 'ERR', theme: 'danger' as const}
}
