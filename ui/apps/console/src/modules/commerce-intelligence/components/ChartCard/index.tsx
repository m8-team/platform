import type {ReactNode} from 'react'
import {Card, Text} from '@gravity-ui/uikit'

export function ChartCard({
  title,
  subtitle,
  children,
  actions,
  className,
}: {
  title: string
  subtitle?: string
  children: ReactNode
  actions?: ReactNode
  className?: string
}) {
  return (
    <Card view="outlined" type="container" className={`ci-card ${className ?? ''}`}>
      <div className="ci-card__header">
        <div>
          <Text as="h2" variant="header-1">
            {title}
          </Text>
          {subtitle ? (
            <Text variant="caption-2" color="secondary">
              {subtitle}
            </Text>
          ) : null}
        </div>
        {actions}
      </div>
      <div className="ci-card__body">{children}</div>
    </Card>
  )
}
