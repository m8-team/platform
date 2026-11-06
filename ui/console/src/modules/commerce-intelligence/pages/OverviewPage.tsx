import {useQuery} from '@tanstack/react-query'
import type {ColumnDef} from '@tanstack/react-table'
import {Bar, CartesianGrid, Legend, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis} from 'recharts'

import {ApprovalQueue} from '../components/ApprovalQueue'
import {ChartCard} from '../components/ChartCard'
import {DataTable} from '../components/DataTable'
import {Heatmap} from '../components/Heatmap'
import {InsightPanel} from '../components/InsightPanel'
import {getOverviewDashboard} from '../mock/queries'
import type {PriceAction} from '../mock/types'
import {notifyAction} from '../utils'
import {CommercePage, ErrorState, KpiGrid, LoadingState, StatusCell, commonOptions, usePageFilters} from './pageCommon'

const filters = [
  {key: 'region', label: 'Регион', options: commonOptions.region},
  {key: 'channel', label: 'Канал', options: commonOptions.channel},
  {key: 'category', label: 'Категория', options: commonOptions.category},
  {key: 'brand', label: 'Бренд', options: commonOptions.brand},
  {key: 'period', label: 'Период', type: 'date' as const, options: commonOptions.period},
  {key: 'scenario', label: 'Сценарий', options: [{value: 'base', content: 'Базовый'}, {value: 'margin', content: 'Максимум маржи'}]},
]

const recommendationColumns: ColumnDef<PriceAction, unknown>[] = [
  {accessorKey: 'sku', header: 'SKU'},
  {accessorKey: 'product', header: 'Товар'},
  {accessorKey: 'currentPrice', header: 'Текущая'},
  {accessorKey: 'recommendedPrice', header: 'Рекомендуемая'},
  {accessorKey: 'reason', header: 'Причина'},
  {accessorKey: 'expectedRevenue', header: 'Ожидаемое влияние'},
  {accessorKey: 'confidence', header: 'Уверенность'},
  {accessorKey: 'status', header: 'Статус', cell: (info) => <StatusCell value={info.row.original.status} />},
]

export function OverviewPage() {
  const {values, setFilter, resetFilters} = usePageFilters(filters)
  const query = useQuery({queryKey: ['commerce-intelligence', 'overview'], queryFn: getOverviewDashboard})

  return (
    <CommercePage
      title="Обзор"
      subtitle="Данные о ценах, спросе и запасах в реальном времени по вашему портфелю."
      filters={filters}
      filterValues={values}
      onFilterChange={setFilter}
      onResetFilters={resetFilters}
      onSaveView={() => notifyAction('Вид сохранен', 'Сохраненные виды появятся в следующем релизе.')}
    >
      {query.isLoading ? <LoadingState /> : null}
      {query.isError ? <ErrorState onRetry={() => void query.refetch()} /> : null}
      {query.data ? (
        <>
          <KpiGrid items={query.data.kpis} />

          <div className="ci-grid ci-grid_2-1">
            <ChartCard title="Прогноз спроса vs фактические продажи" subtitle="Прогноз, факт, границы и промо-события">
              <div className="ci-chart">
                <ResponsiveContainer width="100%" height={280}>
                  <LineChart data={query.data.forecast}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="date" />
                    <YAxis />
                    <Tooltip />
                    <Legend />
                    <Line name="Прогноз" type="monotone" dataKey="forecast" stroke="#2f6fed" strokeWidth={2} dot={false} />
                    <Line name="Факт. продажи" type="monotone" dataKey="actual" stroke="#1a9b68" strokeWidth={2} dot={false} />
                    <Line name="Верхняя граница" type="monotone" dataKey="upper" stroke="#9fb7ff" strokeDasharray="4 4" dot={false} />
                    <Line name="Нижняя граница" type="monotone" dataKey="lower" stroke="#9fb7ff" strokeDasharray="4 4" dot={false} />
                    <Bar name="Промо-события" dataKey="promo" fill="#f2c94c" />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </ChartCard>

            <InsightPanel insights={query.data.insights} />
          </div>

          <div className="ci-grid ci-grid_1-1">
            <ChartCard title="Позиция на рынке">
              <Heatmap rows={query.data.heatmap} />
            </ChartCard>

            <ApprovalQueue items={query.data.approvalSummary} />
          </div>

          <div className="ci-grid ci-grid_1-1">
            <ChartCard title="Оповещения по разметке">
              <DataTable
                data={query.data.markdownAlerts}
                columns={[
                  {accessorKey: 'sku', header: 'SKU'},
                  {accessorKey: 'product', header: 'Товар'},
                  {accessorKey: 'reason', header: 'Причина'},
                  {accessorKey: 'action', header: 'Рекомендуемое действие'},
                ]}
                getRowId={(row) => row.sku}
              />
            </ChartCard>

            <ChartCard title="Рекомендации по ценам">
              <DataTable data={query.data.recommendations} columns={recommendationColumns} getRowId={(row) => row.id} />
            </ChartCard>
          </div>
        </>
      ) : null}
    </CommercePage>
  )
}
