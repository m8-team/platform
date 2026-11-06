import {useCallback, useMemo, useState} from 'react'
import {Button, Text} from '@gravity-ui/uikit'
import {useQuery} from '@tanstack/react-query'
import type {ColumnDef} from '@tanstack/react-table'
import {Bar, BarChart, CartesianGrid, Legend, ResponsiveContainer, Tooltip, XAxis, YAxis} from 'recharts'

import {ActionToolbar} from '../components/ActionToolbar'
import {ChartCard} from '../components/ChartCard'
import {DataTable} from '../components/DataTable'
import {DetailDrawer} from '../components/DetailDrawer'
import {getPriceActions} from '../mock/queries'
import type {PriceAction} from '../mock/types'
import {notifyAction} from '../utils'
import {CommercePage, DefinitionGrid, ErrorState, KpiGrid, LoadingState, StatusCell, commonOptions, usePageFilters} from './pageCommon'

const filters = [
  {key: 'region', label: 'Регион', options: commonOptions.region},
  {key: 'channel', label: 'Канал', options: commonOptions.channel},
  {key: 'category', label: 'Категория', options: commonOptions.category},
  {key: 'brand', label: 'Бренд', options: commonOptions.brand},
  {key: 'type', label: 'Тип рекомендации', options: [{value: 'all', content: 'Все типы'}, {value: 'raise', content: 'Повышение'}, {value: 'markdown', content: 'Уценка'}, {value: 'match', content: 'Сравнять с конкурентом'}]},
  {key: 'status', label: 'Статус', options: [{value: 'all', content: 'Все статусы'}, {value: 'draft', content: 'Черновик'}, {value: 'review', content: 'На проверке'}]},
  {key: 'confidence', label: 'Уверенность', options: [{value: 'all', content: 'Любая'}, {value: 'high', content: 'Высокая'}, {value: 'medium', content: 'Средняя'}]},
  {key: 'period', label: 'Период', type: 'date' as const, options: commonOptions.period},
]

const tabs = ['Черновики', 'На проверке', 'Согласовано', 'Запланировано', 'Применено', 'Отклонено']

export function PriceActionsPage() {
  const {values, setFilter, resetFilters} = usePageFilters(filters)
  const [activeTab, setActiveTab] = useState('Черновики')
  const [selectedRows, setSelectedRows] = useState<PriceAction[]>([])
  const [selectedAction, setSelectedAction] = useState<PriceAction | null>(null)
  const query = useQuery({queryKey: ['commerce-intelligence', 'price-actions'], queryFn: getPriceActions})
  const onSelectionChange = useCallback((rows: PriceAction[]) => setSelectedRows(rows), [])

  const columns = useMemo<ColumnDef<PriceAction, unknown>[]>(
    () => [
      {accessorKey: 'sku', header: 'SKU'},
      {accessorKey: 'product', header: 'Товар'},
      {accessorKey: 'currentPrice', header: 'Текущая цена'},
      {accessorKey: 'recommendedPrice', header: 'Рекомендуемая цена'},
      {accessorKey: 'deltaPct', header: 'Δ %'},
      {accessorKey: 'reason', header: 'Причина'},
      {accessorKey: 'expectedRevenue', header: 'Ожидаемая выручка'},
      {accessorKey: 'expectedMargin', header: 'Ожидаемая маржа'},
      {accessorKey: 'confidence', header: 'Уверенность'},
      {accessorKey: 'guardrailStatus', header: 'Статус правил', cell: (info) => <StatusCell value={info.row.original.guardrailStatus} />},
      {accessorKey: 'approver', header: 'Согласующий'},
      {accessorKey: 'status', header: 'Статус', cell: (info) => <StatusCell value={info.row.original.status} />},
    ],
    [],
  )

  const filteredActions = useMemo(() => {
    const actions = query.data?.actions ?? []
    return activeTab === 'Черновики' ? actions : actions.filter((action) => action.status === activeTab)
  }, [activeTab, query.data?.actions])

  const runAction = (label: string) => notifyAction(label, selectedRows.length > 0 ? `${selectedRows.length} действий обновлено` : 'Выберите строки для массового действия.')

  return (
    <CommercePage
      title="Ценовые действия"
      subtitle="AI-рекомендации по изменению цен, готовые к проверке и выполнению."
      filters={filters}
      filterValues={values}
      onFilterChange={setFilter}
      onResetFilters={resetFilters}
      onSaveView={() => notifyAction('Вид сохранен')}
    >
      {query.isLoading ? <LoadingState /> : null}
      {query.isError ? <ErrorState onRetry={() => void query.refetch()} /> : null}
      {query.data ? (
        <>
          <KpiGrid items={query.data.kpis} />
          <div className="ci-tabs">
            {tabs.map((tab) => (
              <Button key={tab} view={activeTab === tab ? 'action' : 'outlined'} onClick={() => setActiveTab(tab)}>
                {tab}
              </Button>
            ))}
          </div>

          <ChartCard title="Рекомендации к изменению цен">
            <DataTable
              data={filteredActions}
              columns={columns}
              getRowId={(row) => row.id}
              enableRowSelection
              onSelectedRowsChange={onSelectionChange}
              onRowClick={setSelectedAction}
              toolbar={
                <ActionToolbar
                  selectedCount={selectedRows.length}
                  onApprove={() => runAction('Рекомендации согласованы')}
                  onReject={() => runAction('Рекомендации отклонены')}
                  onSchedule={() => runAction('Изменения запланированы')}
                  onMore={() => notifyAction('Открыто меню дополнительных действий')}
                />
              }
            />
          </ChartCard>

          <ChartCard title="Ожидание vs факт по примененным изменениям">
            <div className="ci-chart">
              <ResponsiveContainer width="100%" height={240}>
                <BarChart data={query.data.appliedVsActual}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="name" />
                  <YAxis />
                  <Tooltip />
                  <Legend />
                  <Bar name="Ожидание" dataKey="expected" fill="#2f6fed" />
                  <Bar name="Факт" dataKey="actual" fill="#1a9b68" />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </ChartCard>

          <DetailDrawer
            open={Boolean(selectedAction)}
            title={selectedAction?.product ?? 'Детали рекомендации'}
            subtitle={selectedAction ? selectedAction.sku : undefined}
            onClose={() => setSelectedAction(null)}
            actions={
              <>
                <Button view="action" onClick={() => notifyAction('Рекомендация согласована')}>Согласовать</Button>
                <Button view="outlined" onClick={() => notifyAction('Рекомендация отклонена')}>Отклонить</Button>
                <Button view="outlined" onClick={() => notifyAction('Рекомендация запланирована')}>Запланировать</Button>
              </>
            }
            sections={[
              {
                title: 'AI-объяснение',
                content: <Text variant="body-2">Рекомендация основана на эластичности, конкурентном индексе и прогнозе спроса на 30 дней.</Text>,
              },
              {
                title: 'Снимок конкурентов',
                content: selectedAction ? (
                  <DefinitionGrid
                    items={[
                      {label: 'Текущая цена', value: selectedAction.currentPrice},
                      {label: 'Рекомендуемая', value: selectedAction.recommendedPrice},
                      {label: 'Причина', value: selectedAction.reason},
                    ]}
                  />
                ) : null,
              },
              {title: 'Оценка эластичности', content: <Text variant="body-2">Эластичность в диапазоне -1.4…-2.1, ожидаемый объем устойчив к изменению цены.</Text>},
              {title: 'Влияние на прогноз спроса', content: <Text variant="body-2">Ожидается рост спроса на 8–14% после применения цены.</Text>},
              {title: 'Складской риск', content: <StatusCell value={selectedAction?.risk ?? 'Средняя'} />},
              {title: 'Проверки guardrails', content: <StatusCell value={selectedAction?.guardrailStatus ?? 'Пройдено'} />},
              {title: 'История согласования', content: <Text variant="body-2">Создано AI pricing, назначен согласующий, ожидается решение владельца категории.</Text>},
            ]}
          />
        </>
      ) : null}
    </CommercePage>
  )
}
