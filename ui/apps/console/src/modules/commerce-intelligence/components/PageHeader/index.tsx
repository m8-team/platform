import type {ReactNode} from 'react'
import {Text} from '@gravity-ui/uikit'

export function PageHeader({
  title,
  subtitle,
  actions,
}: {
  title: string
  subtitle: string
  actions?: ReactNode
}) {
  return (
    <div className="ci-page-header">
      <div>
        <Text className="ci-breadcrumb" variant="caption-2" color="secondary">
          M8 / Commerce Intelligence
        </Text>
        <Text as="h1" variant="display-1">
          {title}
        </Text>
        <Text as="p" variant="body-2" color="secondary">
          {subtitle}
        </Text>
      </div>
      {actions ? <div className="ci-page-header__actions">{actions}</div> : null}
    </div>
  )
}
