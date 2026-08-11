import {Check} from '@gravity-ui/icons'
import {Card, Icon, Label, Text} from '@gravity-ui/uikit'

import {ConsoleBreadcrumbs} from '../../../components/ConsoleBreadcrumbs'
import {Metric} from '../../../components/Metric'
import {useConsoleI18n} from '../../../console/ConsoleContext'
import type {Translate, TranslationKey} from '../../../i18n'
import {resourceManagerRoutes} from '../routes'

type OverviewTone = 'brand' | 'info' | 'positive' | 'warning' | 'danger'

const resourceOverviewMix = [
  {labelKey: 'overview.resource.database', value: 42, tone: 'brand'},
  {labelKey: 'overview.resource.kafka', value: 31, tone: 'info'},
  {labelKey: 'overview.resource.cache', value: 18, tone: 'positive'},
  {labelKey: 'overview.resource.storage', value: 9, tone: 'warning'},
] satisfies Array<{labelKey: TranslationKey; value: number; tone: OverviewTone}>

const resourceOverviewLifecycle = [
  {labelKey: 'status.Active', value: 78, tone: 'positive'},
  {labelKey: 'status.Provisioning', value: 12, tone: 'info'},
  {labelKey: 'status.Suspended', value: 6, tone: 'warning'},
  {labelKey: 'status.Failed', value: 4, tone: 'danger'},
] satisfies Array<{labelKey: TranslationKey; value: number; tone: OverviewTone}>

const resourceOverviewOperations = [
  {
    titleKey: 'overview.operation.projectCreate',
    descriptionKey: 'overview.operation.projectCreateDescription',
    theme: 'info',
  },
  {
    titleKey: 'overview.operation.workspaceSuspend',
    descriptionKey: 'overview.operation.workspaceSuspendDescription',
    theme: 'warning',
  },
  {
    titleKey: 'overview.operation.resourceReconcile',
    descriptionKey: 'overview.operation.resourceReconcileDescription',
    theme: 'success',
  },
] satisfies Array<{
  titleKey: TranslationKey
  descriptionKey: TranslationKey
  theme: 'success' | 'warning' | 'danger' | 'info' | 'normal'
}>

const resourceOverviewSignals = [
  {
    titleKey: 'overview.signal.sourceOfTruth',
    descriptionKey: 'overview.signal.sourceOfTruthDescription',
  },
  {
    titleKey: 'overview.signal.safeMutations',
    descriptionKey: 'overview.signal.safeMutationsDescription',
  },
  {
    titleKey: 'overview.signal.auditReady',
    descriptionKey: 'overview.signal.auditReadyDescription',
  },
] satisfies Array<{titleKey: TranslationKey; descriptionKey: TranslationKey}>

export function ResourceManagerOverviewPage() {
  const {t} = useConsoleI18n()

  return (
    <main className="m8-page__body">
      <section className="m8-page__content">
        <div className="m8-page__heading">
          <div>
            <ConsoleBreadcrumbs
              items={[
                {text: t('breadcrumb.resourceManager'), href: resourceManagerRoutes.overview},
                {text: t('menu.resources.overview')},
              ]}
            />
            <Text as="h1" variant="display-1">
              {t('page.resourceManager.title')}
            </Text>
            <Text as="p" variant="body-2" color="secondary">
              {t('page.resourceManager.description')}
            </Text>
          </div>

          <div className="m8-summary m8-summary_overview">
            <Metric
              label={t('overview.metric.organizations')}
              value="2"
              description={t('overview.metric.organizationsDescription')}
            />
            <Metric
              label={t('overview.metric.workspaces')}
              value="3"
              description={t('overview.metric.workspacesDescription')}
            />
            <Metric
              label={t('overview.metric.projects')}
              value="147"
              description={t('overview.metric.projectsDescription')}
            />
            <Metric
              label={t('overview.metric.operations')}
              value="9"
              description={t('overview.metric.operationsDescription')}
              tone="warning"
            />
          </div>
        </div>

        <div className="m8-overview-grid">
          <Card view="outlined" type="container" className="m8-overview-card">
            <OverviewCardHeader title={t('overview.hierarchy.title')} description={t('overview.hierarchy.description')} />
            <div className="m8-overview-hierarchy">
              <OverviewHierarchyNode label={t('overview.hierarchy.organizations')} value="2" />
              <OverviewHierarchyNode label={t('overview.hierarchy.workspaces')} value="3" />
              <OverviewHierarchyNode label={t('overview.hierarchy.projects')} value="147" />
              <OverviewHierarchyNode label={t('overview.hierarchy.resources')} value="612" />
            </div>
          </Card>

          <Card view="outlined" type="container" className="m8-overview-card">
            <OverviewCardHeader title={t('overview.resourceMix.title')} description={t('overview.resourceMix.description')} />
            <OverviewBarChart items={resourceOverviewMix} t={t} />
          </Card>

          <Card view="outlined" type="container" className="m8-overview-card">
            <OverviewCardHeader title={t('overview.lifecycle.title')} description={t('overview.lifecycle.description')} />
            <OverviewStackChart items={resourceOverviewLifecycle} t={t} />
          </Card>

          <Card view="outlined" type="container" className="m8-overview-card">
            <OverviewCardHeader title={t('overview.operations.title')} description={t('overview.operations.description')} />
            <div className="m8-overview-list">
              {resourceOverviewOperations.map((operation) => (
                <div className="m8-overview-list__item" key={operation.titleKey}>
                  <div>
                    <Text variant="body-2">{t(operation.titleKey)}</Text>
                    <Text variant="caption-2" color="secondary">
                      {t(operation.descriptionKey)}
                    </Text>
                  </div>
                  <Label theme={operation.theme}>{t('overview.operation.running')}</Label>
                </div>
              ))}
            </div>
          </Card>

          <Card view="outlined" type="container" className="m8-overview-card m8-overview-card_full">
            <OverviewCardHeader title={t('overview.info.title')} description={t('overview.info.description')} />
            <div className="m8-overview-signals">
              {resourceOverviewSignals.map((signal) => (
                <div className="m8-overview-signal" key={signal.titleKey}>
                  <Icon data={Check} size={16} />
                  <div>
                    <Text variant="body-2">{t(signal.titleKey)}</Text>
                    <Text variant="caption-2" color="secondary">
                      {t(signal.descriptionKey)}
                    </Text>
                  </div>
                </div>
              ))}
            </div>
          </Card>
        </div>
      </section>
    </main>
  )
}

function OverviewCardHeader({title, description}: {title: string; description: string}) {
  return (
    <div>
      <Text as="h2" variant="header-1">
        {title}
      </Text>
      <Text variant="caption-2" color="secondary">
        {description}
      </Text>
    </div>
  )
}

function OverviewHierarchyNode({label, value}: {label: string; value: string}) {
  return (
    <div className="m8-overview-hierarchy__node">
      <Text variant="caption-2" color="secondary">
        {label}
      </Text>
      <Text variant="header-2">{value}</Text>
    </div>
  )
}

function OverviewBarChart({
  items,
  t,
}: {
  items: Array<{labelKey: TranslationKey; value: number; tone: OverviewTone}>
  t: Translate
}) {
  return (
    <div className="m8-bar-chart">
      {items.map((item) => (
        <div className="m8-bar-chart__row" key={item.labelKey}>
          <Text variant="caption-2" color="secondary">
            {t(item.labelKey)}
          </Text>
          <div className="m8-bar-chart__track" aria-hidden="true">
            <div className={`m8-bar-chart__bar m8-overview-tone_${item.tone}`} style={{width: `${item.value}%`}} />
          </div>
          <Text variant="caption-2">{item.value}%</Text>
        </div>
      ))}
    </div>
  )
}

function OverviewStackChart({
  items,
  t,
}: {
  items: Array<{labelKey: TranslationKey; value: number; tone: OverviewTone}>
  t: Translate
}) {
  return (
    <div className="m8-stack-chart-shell">
      <div className="m8-stack-chart" aria-hidden="true">
        {items.map((item) => (
          <span
            className={`m8-stack-chart__segment m8-overview-tone_${item.tone}`}
            key={item.labelKey}
            style={{width: `${item.value}%`}}
          />
        ))}
      </div>
      <div className="m8-stack-chart__legend">
        {items.map((item) => (
          <div className="m8-stack-chart__legend-item" key={item.labelKey}>
            <span className={`m8-stack-chart__dot m8-overview-tone_${item.tone}`} />
            <Text variant="caption-2" color="secondary">
              {t(item.labelKey)}
            </Text>
            <Text variant="caption-2">{item.value}%</Text>
          </div>
        ))}
      </div>
    </div>
  )
}
