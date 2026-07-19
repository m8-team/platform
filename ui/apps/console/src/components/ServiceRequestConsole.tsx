import {useCallback, useEffect, useRef, useState, useSyncExternalStore} from 'react'
import {TrashBin, Xmark} from '@gravity-ui/icons'
import {Button, Icon, Text} from '@gravity-ui/uikit'

import {ServiceRequestItem} from './ServiceRequestItem'
import {serviceRequestLog} from '../platform/http/serviceRequestLog'
import type {Translate} from '../i18n'

interface ServiceRequestConsoleProps {
  t: Translate
  onClose: () => void
}

export function ServiceRequestConsole({t, onClose}: ServiceRequestConsoleProps) {
  const records = useSyncExternalStore(
    serviceRequestLog.subscribe,
    serviceRequestLog.getSnapshot,
    serviceRequestLog.getSnapshot,
  )
  const [openRecordIds, setOpenRecordIds] = useState<ReadonlySet<string>>(() => new Set())
  const visibleRecordIdsRef = useRef<ReadonlySet<string>>(new Set())

  useEffect(() => {
    visibleRecordIdsRef.current = new Set(records.map((record) => record.id))
  }, [records])

  const handleRecordToggle = useCallback((recordId: string, open: boolean) => {
    setOpenRecordIds((current) => {
      const next = new Set([...current].filter((id) => visibleRecordIdsRef.current.has(id)))
      if (open) next.add(recordId)
      else next.delete(recordId)
      return next
    })
  }, [])

  const handleClear = () => {
    serviceRequestLog.clear()
    setOpenRecordIds(new Set())
  }

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
            onClick={handleClear}
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
          <ServiceRequestItem
            key={record.id}
            record={record}
            open={openRecordIds.has(record.id)}
            onOpenChange={handleRecordToggle}
            t={t}
          />
        ))}
      </div>
    </div>
  )
}
