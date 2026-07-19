import {useMemo, useState} from 'react'
import type {ReactNode} from 'react'
import {ArrowsExpand, Xmark} from '@gravity-ui/icons'
import {Button, ClipboardButton, Icon, Modal, Text} from '@gravity-ui/uikit'

const maximumPreviewLength = 32 * 1024

export interface JsonPreviewProps {
  value: unknown
  copyText: string
  copiedText: string
  openText: string
  overlayTitle: string
  closeText: string
}

export function JsonPreview({
  value,
  copyText,
  copiedText,
  openText,
  overlayTitle,
  closeText,
}: JsonPreviewProps) {
  const [overlayOpen, setOverlayOpen] = useState(false)
  const preview = useMemo(() => formatPreview(value), [value])
  const compact = preview.text.length > 600 || preview.text.split('\n').length > 8

  return (
    <>
      <JsonCodeBlock
        preview={preview}
        compact={compact}
        copyText={copyText}
        copiedText={copiedText}
        openText={openText}
        onOpen={() => setOverlayOpen(true)}
      />
      <Modal
        open={overlayOpen}
        onOpenChange={setOverlayOpen}
        contentClassName="m8-json-overlay"
        aria-label={overlayTitle}
      >
        <div className="m8-json-overlay__header">
          <Text as="h2" variant="header-1">{overlayTitle}</Text>
          <Button view="flat" size="m" aria-label={closeText} onClick={() => setOverlayOpen(false)}>
            <Icon data={Xmark} size={18} />
          </Button>
        </div>
        <JsonCodeBlock preview={preview} copyText={copyText} copiedText={copiedText} expanded />
      </Modal>
    </>
  )
}

interface FormattedPreview {
  text: string
  json: boolean
}

interface JsonCodeBlockProps {
  preview: FormattedPreview
  copyText: string
  copiedText: string
  compact?: boolean
  expanded?: boolean
  openText?: string
  onOpen?: () => void
}

function JsonCodeBlock({
  preview,
  copyText,
  copiedText,
  compact = false,
  expanded = false,
  openText,
  onOpen,
}: JsonCodeBlockProps) {
  const className = [
    'm8-json-preview',
    compact ? 'm8-json-preview_compact' : '',
    expanded ? 'm8-json-preview_expanded' : '',
  ].filter(Boolean).join(' ')

  return (
    <div className={className}>
      <div className="m8-json-preview__actions">
        {onOpen && openText ? (
          <Button view="flat-secondary" size="s" aria-label={openText} onClick={onOpen}>
            <Icon data={ArrowsExpand} size={16} />
          </Button>
        ) : null}
        <ClipboardButton
          text={preview.text}
          view="flat-secondary"
          size="s"
          tooltipInitialText={copyText}
          tooltipSuccessText={copiedText}
        />
      </div>
      <pre tabIndex={0}>
        <code>{preview.json ? highlightJSON(preview.text) : preview.text}</code>
      </pre>
    </div>
  )
}

function formatPreview(value: unknown): FormattedPreview {
  if (typeof value === 'string') {
    if (value.length > maximumPreviewLength) return {text: limitPreview(value), json: false}
    const parsed = parseJSONString(value)
    if (parsed !== undefined) return {text: limitPreview(JSON.stringify(parsed, null, 2)), json: true}
    return {text: value, json: false}
  }

  try {
    const serialized = JSON.stringify(value, null, 2)
    return typeof serialized === 'string'
      ? {text: limitPreview(serialized), json: true}
      : {text: '[VALUE_UNAVAILABLE]', json: false}
  } catch {
    return {text: '[VALUE_UNAVAILABLE]', json: false}
  }
}

function limitPreview(value: string) {
  return value.length > maximumPreviewLength
    ? `${value.slice(0, maximumPreviewLength)}\n[TRUNCATED: preview exceeds 32 KiB]`
    : value
}

function parseJSONString(value: string) {
  const trimmed = value.trim()
  if (!trimmed.startsWith('{') && !trimmed.startsWith('[')) return undefined
  try {
    return JSON.parse(trimmed) as unknown
  } catch {
    return undefined
  }
}

function highlightJSON(value: string) {
  const tokenPattern = /("(?:\\(?:["\\/bfnrt]|u[0-9a-fA-F]{4})|[^"\\])*"(?=\s*:))|("(?:\\(?:["\\/bfnrt]|u[0-9a-fA-F]{4})|[^"\\])*")|(-?\b\d+(?:\.\d+)?(?:[eE][+-]?\d+)?\b)|(true|false)|(null)/g
  const content: ReactNode[] = []
  let offset = 0

  for (const match of value.matchAll(tokenPattern)) {
    const index = match.index
    if (index > offset) content.push(value.slice(offset, index))
    const token = match[0]
    const tokenType = match[1]
      ? 'key'
      : match[2]
        ? 'string'
        : match[3]
          ? 'number'
          : match[4]
            ? 'boolean'
            : 'null'
    content.push(
      <span className={`m8-json-preview__token_${tokenType}`} key={`${index}:${tokenType}`}>
        {token}
      </span>,
    )
    offset = index + token.length
  }

  if (offset < value.length) content.push(value.slice(offset))
  return content
}
