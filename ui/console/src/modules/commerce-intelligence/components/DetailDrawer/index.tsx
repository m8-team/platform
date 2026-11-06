import type {ReactNode} from 'react'
import {Button, Drawer, Text} from '@gravity-ui/uikit'

export type DetailSection = {
  title: string
  content: ReactNode
}

export function DetailDrawer({
  open,
  title,
  subtitle,
  sections,
  actions,
  onClose,
}: {
  open: boolean
  title: string
  subtitle?: string
  sections: DetailSection[]
  actions?: ReactNode
  onClose: () => void
}) {
  return (
    <Drawer open={open} onOpenChange={(nextOpen) => !nextOpen && onClose()} placement="right" size={480} hideVeil>
      <div className="ci-drawer">
        <div className="ci-drawer__header">
          <div>
            <Text as="h2" variant="header-2">
              {title}
            </Text>
            {subtitle ? (
              <Text variant="caption-2" color="secondary">
                {subtitle}
              </Text>
            ) : null}
          </div>
          <Button view="flat" onClick={onClose}>
            Закрыть
          </Button>
        </div>
        {actions ? <div className="ci-drawer__actions">{actions}</div> : null}
        <div className="ci-drawer__sections">
          {sections.map((section) => (
            <section className="ci-drawer__section" key={section.title}>
              <Text as="h3" variant="subheader-2">
                {section.title}
              </Text>
              <div>{section.content}</div>
            </section>
          ))}
        </div>
      </div>
    </Drawer>
  )
}
