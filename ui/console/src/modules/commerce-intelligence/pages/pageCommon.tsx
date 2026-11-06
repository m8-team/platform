/* eslint-disable react-refresh/only-export-components */
import type {ReactNode} from 'react'
import {useMemo, useState} from 'react'
import {Button, Card, Text} from '@gravity-ui/uikit'

import {AppShell} from '../components/AppShell'
import {FilterBar} from '../components/FilterBar'
import type {FilterConfig, FilterValues} from '../components/FilterBar'
import {KpiCard} from '../components/KpiCard'
import {PageHeader} from '../components/PageHeader'
import {StatusBadge} from '../components/StatusBadge'
import type {Kpi} from '../mock/types'
import {notifyAction, statusTone} from '../utils'

export const commonOptions = {
  region: [
    {value: 'all', content: 'Все регионы'},
    {value: 'vienna', content: 'Вена'},
    {value: 'eu', content: 'Европа'},
    {value: 'us', content: 'США'},
  ],
  channel: [
    {value: 'all', content: 'Все каналы'},
    {value: 'online', content: 'Онлайн'},
    {value: 'retail', content: 'Розница'},
    {value: 'marketplace', content: 'Маркетплейс'},
  ],
  category: [
    {value: 'all', content: 'Все категории'},
    {value: 'electronics', content: 'Электроника'},
    {value: 'home', content: 'Дом'},
    {value: 'apparel', content: 'Одежда'},
    {value: 'sport', content: 'Спорт'},
  ],
  brand: [
    {value: 'all', content: 'Все бренды'},
    {value: 'm8', content: 'M8 Select'},
    {value: 'premium', content: 'Premium Brand'},
    {value: 'seasonal', content: 'Seasonal Line'},
  ],
  period: [{value: '12 мая — 10 июня 2025', content: '12 мая — 10 июня 2025'}],
}

export function createInitialFilters(filters: FilterConfig[]) {
  return filters.reduce<FilterValues>((values, filter) => {
    values[filter.key] = filter.options?.[0]?.value ?? ''
    return values
  }, {})
}

export function usePageFilters(filters: FilterConfig[]) {
  const initialFilters = useMemo(() => createInitialFilters(filters), [filters])
  const [values, setValues] = useState<FilterValues>(initialFilters)

  return {
    values,
    setFilter: (key: string, value: string) => setValues((current) => ({...current, [key]: value})),
    resetFilters: () => {
      setValues(initialFilters)
      notifyAction('Фильтры сброшены')
    },
  }
}

export function CommercePage({
  title,
  subtitle,
  filters,
  filterValues,
  onFilterChange,
  onResetFilters,
  onSaveView,
  primaryFilterAction,
  children,
  actions,
}: {
  title: string
  subtitle: string
  filters?: FilterConfig[]
  filterValues?: FilterValues
  onFilterChange?: (key: string, value: string) => void
  onResetFilters?: () => void
  onSaveView?: () => void
  primaryFilterAction?: {label: string; onClick: () => void}
  children: ReactNode
  actions?: ReactNode
}) {
  return (
    <AppShell>
      <div className="ci-page">
        <PageHeader title={title} subtitle={subtitle} actions={actions} />
        {filters && filterValues && onFilterChange && onResetFilters ? (
          <FilterBar
            filters={filters}
            values={filterValues}
            onChange={onFilterChange}
            onReset={onResetFilters}
            onSaveView={onSaveView}
            primaryAction={primaryFilterAction}
          />
        ) : null}
        {children}
      </div>
    </AppShell>
  )
}

export function KpiGrid({items}: {items: Kpi[]}) {
  return (
    <div className="ci-kpi-grid">
      {items.map((item) => (
        <KpiCard key={item.title} {...item} />
      ))}
    </div>
  )
}

export function LoadingState() {
  return (
    <div className="ci-loading-grid">
      {Array.from({length: 8}, (_, index) => (
        <Card key={index} view="outlined" type="container" className="ci-skeleton-card">
          <div className="ci-skeleton ci-skeleton_short" />
          <div className="ci-skeleton ci-skeleton_tall" />
          <div className="ci-skeleton" />
        </Card>
      ))}
    </div>
  )
}

export function ErrorState({onRetry}: {onRetry: () => void}) {
  return (
    <Card view="outlined" type="container" className="ci-error">
      <Text variant="header-1">Не удалось загрузить данные</Text>
      <Text variant="body-2" color="secondary">
        Повторите запрос или проверьте состояние mock API.
      </Text>
      <Button view="action" onClick={onRetry}>
        Повторить
      </Button>
    </Card>
  )
}

export function StatusCell({value}: {value: string}) {
  return <StatusBadge tone={statusTone(value)}>{value}</StatusBadge>
}

export function DefinitionGrid({items}: {items: Array<{label: string; value: ReactNode}>}) {
  return (
    <dl className="ci-definition-grid">
      {items.map((item) => (
        <div key={item.label}>
          <dt>{item.label}</dt>
          <dd>{item.value}</dd>
        </div>
      ))}
    </dl>
  )
}

export function MiniBars({items}: {items: Array<{label: string; value: number}>}) {
  const max = Math.max(...items.map((item) => item.value), 1)

  return (
    <div className="ci-mini-bars">
      {items.map((item) => (
        <div className="ci-mini-bars__row" key={item.label}>
          <Text variant="caption-2" color="secondary">
            {item.label}
          </Text>
          <span>
            <i style={{width: `${(item.value / max) * 100}%`}} />
          </span>
          <Text variant="caption-2">{item.value}</Text>
        </div>
      ))}
    </div>
  )
}
