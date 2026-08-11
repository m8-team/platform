import {useQuery} from '@tanstack/react-query'

import {ApprovalQueue} from '../components/ApprovalQueue'
import {BusinessChart} from '../components/BusinessChart'
import {ChartCard} from '../components/ChartCard'
import {DataTable, type DataTableColumn} from '../components/DataTable'
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

const recommendationColumns: DataTableColumn<PriceAction>[] = [
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
              <BusinessChart
                data={query.data.forecast}
                xKey="date"
                ariaLabel="Прогноз спроса, фактические продажи и промо-события"
                series={[
                  {dataKey: 'forecast', name: 'Прогноз', emphasis: true},
                  {dataKey: 'actual', name: 'Факт. продажи', emphasis: true},
                  {dataKey: 'upper', name: 'Верхняя граница', dashed: true},
                  {dataKey: 'lower', name: 'Нижняя граница', dashed: true},
                  {kind: 'bar', dataKey: 'promo', name: 'Промо-события'},
                ]}
              />
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
