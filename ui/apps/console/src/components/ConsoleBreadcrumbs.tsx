import {Breadcrumbs, BreadcrumbsItem, type BreadcrumbsItemProps, type BreadcrumbsProps} from '@gravity-ui/uikit'
import {createLink} from '@tanstack/react-router';

const RouterLink = createLink(BreadcrumbsItem);

type BreadcrumbsItem = Omit<BreadcrumbsItemProps, 'children'> & {
  text: string
}

export function ConsoleBreadcrumbs({
                                     items,
                                     ...breadcrumbsProps
                                   }: Omit<BreadcrumbsProps, 'children' | 'itemComponent'> & {
  items: BreadcrumbsItem[]
}) {
  return (
    <Breadcrumbs
      {...breadcrumbsProps}
      itemComponent={RouterLink}
    >
      {items.map(({text, href, ...item}) => (
        <RouterLink key={`${href ?? 'current'}-${text}`} {...item} to={href}>
          {text}
        </RouterLink>
      ))}
    </Breadcrumbs>
  )
}
