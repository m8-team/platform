import {Card, Text} from '@gravity-ui/uikit'

import {ConsoleBreadcrumbs} from '../../../components/ConsoleBreadcrumbs'
import {useConsoleI18n} from '../../../console/ConsoleContext'
import {resourceManagerRoutes} from '../routes'

export function ResourceOrganizationDetailsPage() {
  const {t} = useConsoleI18n()

  return (
    <ResourcePlaceholderPage
      current={t('menu.resources.organizations')}
      title={t('page.organizationDetails.title')}
      description={t('page.organizationDetails.description')}
    />
  )
}

export function ResourceWorkspaceDetailsPage() {
  const {t} = useConsoleI18n()

  return (
    <ResourcePlaceholderPage
      current={t('menu.resources.workspaces')}
      title={t('page.workspaceDetails.title')}
      description={t('page.workspaceDetails.description')}
    />
  )
}

function ResourcePlaceholderPage({
  current,
  title,
  description,
}: {
  current: string
  title: string
  description: string
}) {
  const {t} = useConsoleI18n()

  return (
    <main className="m8-page__body">
      <section className="m8-page__content">
        <div className="m8-page__heading">
          <div>
            <ConsoleBreadcrumbs
              items={[
                {text: t('breadcrumb.resourceManager'), href: resourceManagerRoutes.overview},
                {text: current},
              ]}
            />
            <Text as="h1" variant="display-1">
              {title}
            </Text>
            <Text as="p" variant="body-2" color="secondary">
              {description}
            </Text>
          </div>
        </div>

        <Card view="outlined" type="container" className="m8-placeholder-card">
          <Text as="h2" variant="header-1">
            {t('page.placeholder.title')}
          </Text>
          <Text variant="body-2" color="secondary">
            {t('page.placeholder.description')}
          </Text>
        </Card>
      </section>
    </main>
  )
}
