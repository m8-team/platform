import {useQuery} from '@tanstack/react-query'
import type {ColumnDef} from '@tanstack/react-table'
import {Bar, BarChart, CartesianGrid, Legend, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis} from 'recharts'

import {ChartCard} from '../components/ChartCard'
import {DataTable} from '../components/DataTable'
import {Heatmap} from '../components/Heatmap'
import {InsightPanel} from '../components/InsightPanel'
import {getForecasts} from '../mock/queries'
import type {ForecastRiskSku} from '../mock/types'
import {notifyAction} from '../utils'
import {CommercePage, ErrorState, KpiGrid, LoadingState, MiniBars, commonOptions, usePageFilters} from './pageCommon'

const filters = [
  {key: 'region', label: 'Регион', options: commonOptions.region},
  {key: 'channel', label: 'Канал', options: commonOptions.channel},
  {key: 'category', label: 'Категория', options: commonOptions.category},
  {key: 'brand', label: 'Бренд', options: commonOptions.brand},
  {key: 'horizon', label: 'Горизонт прогноза', options: [{value: '30', content: '30 дней'}, {value: '60', content: '60 дней'}, {value: '90', content: '90 дней'}]},
  {key: 'model', label: 'Версия модели', options: [{value: 'v4', content: 'Demand AI v4'}, {value: 'v3', content: 'Demand AI v3'}]},
  {key: 'warehouse', label: 'Кластер склада', options: [{value: 'all', content: 'Все кластеры'}, {value: 'north', content: 'Север'}, {value: 'south', content: 'Юг'}]},
  {key: 'period', label: 'Период', type: 'date' as const, options: commonOptions.period},
]

const columns: ColumnDef<ForecastRiskSku, unknown>[] = [
  {accessorKey: 'sku', header: 'SKU'},
  {accessorKey: 'product', header: 'Товар'},
  {accessorKey: 'category', header: 'Категория'},
  {accessorKey: 'forecast30d', header: 'Прогноз 30д'},
  {accessorKey: 'stock', header: 'Текущий остаток'},
  {accessorKey: 'coverageDays', header: 'Дней покрытия'},
  {accessorKey: 'overstockScore', header: 'Скор избыточных запасов'},
  {accessorKey: 'outOfStockScore', header: 'Скор отсутствия товара'},
  {accessorKey: 'leadTime', header: 'Срок поставки'},
  {accessorKey: 'confidence', header: 'Уверенность'},
  {accessorKey: 'suggestedAction', header: 'Предлагаемое действие'},
]

export function ForecastsPage() {
  const {values, setFilter, resetFilters} = usePageFilters(filters)
  const query = useQuery({queryKey: ['commerce-intelligence', 'forecasts'], queryFn: getForecasts})

  return (
    <CommercePage
      title="Прогнозы"
      subtitle="AI-прогноз спроса, покрытия запасов и операционных рисков."
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
          <div className="ci-grid ci-grid_2-1">
            <ChartCard title="Прогноз vs факт">
              <div className="ci-chart">
                <ResponsiveContainer width="100%" height={280}>
                  <LineChart data={query.data.forecastVsActual}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="date" />
                    <YAxis />
                    <Tooltip />
                    <Legend />
                    <Line name="Прогноз" dataKey="forecast" stroke="#2f6fed" strokeWidth={2} dot={false} />
                    <Line name="Факт" dataKey="actual" stroke="#1a9b68" strokeWidth={2} dot={false} />
                    <Line name="Верхняя граница" dataKey="upper" stroke="#9fb7ff" strokeDasharray="4 4" dot={false} />
                    <Line name="Нижняя граница" dataKey="lower" stroke="#9fb7ff" strokeDasharray="4 4" dot={false} />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </ChartCard>
            <InsightPanel title="Почему изменится спрос" insights={query.data.insights} />
          </div>

          <div className="ci-grid ci-grid_1-1">
            <ChartCard title="Точность прогноза по категориям">
              <div className="ci-chart">
                <ResponsiveContainer width="100%" height={240}>
                  <BarChart data={query.data.categoryAccuracy}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="category" />
                    <YAxis />
                    <Tooltip />
                    <Bar name="WAPE" dataKey="wape" fill="#2f6fed" />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </ChartCard>
            <ChartCard title="Матрица складского риска">
              <Heatmap rows={query.data.inventoryRiskMatrix} />
            </ChartCard>
          </div>

          <ChartCard title="SKU в зоне риска">
            <DataTable data={query.data.atRiskSkus} columns={columns} getRowId={(row) => row.sku} />
          </ChartCard>

          <div className="ci-grid ci-grid_1-1-1">
            <ChartCard title="Драйверы модели">
              <MiniBars items={query.data.modelDrivers.map((driver) => ({label: driver.name, value: driver.value}))} />
            </ChartCard>
            <ChartCard title="Календарь событий">
              <MiniBars items={[{label: 'Промо', value: 8}, {label: 'Сезонные пики', value: 5}, {label: 'Запуски товаров', value: 3}]} />
            </ChartCard>
            <ChartCard title="Важность факторов">
              <MiniBars items={query.data.modelDrivers.slice(0, 4).map((driver) => ({label: driver.name, value: driver.value}))} />
            </ChartCard>
          </div>
        </>
      ) : null}
    </CommercePage>
  )
}
