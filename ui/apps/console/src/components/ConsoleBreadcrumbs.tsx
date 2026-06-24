import {Breadcrumbs} from '@gravity-ui/uikit'

interface ConsoleBreadcrumbItem {
  text: string
  href?: string
}

interface ConsoleBreadcrumbsProps {
  items: ConsoleBreadcrumbItem[]
}

export function ConsoleBreadcrumbs({items}: ConsoleBreadcrumbsProps) {
  return (
    <Breadcrumbs aria-label="Breadcrumbs" maxItems={4}>
      {items.map((item) => (
        <Breadcrumbs.Item key={`${item.href ?? 'current'}-${item.text}`} href={item.href}>
          {item.text}
        </Breadcrumbs.Item>
      ))}
    </Breadcrumbs>
  )
}
