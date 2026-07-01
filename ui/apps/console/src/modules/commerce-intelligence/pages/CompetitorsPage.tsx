import {useState} from 'react'
import {Text} from '@gravity-ui/uikit'
import {useQuery} from '@tanstack/react-query'
import type {ColumnDef} from '@tanstack/react-table'
import {Bar, BarChart, CartesianGrid, Legend, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis} from 'recharts'

import {ChartCard} from '../components/ChartCard'
import {DataTable} from '../components/DataTable'
import {DetailDrawer} from '../components/DetailDrawer'
import {Heatmap} from '../components/Heatmap'
import {getCompetitors} from '../mock/queries'
import type {CompetitorMatch} from '../mock/types'
import {notifyAction} from '../utils'
import {CommercePage, DefinitionGrid, ErrorState, KpiGrid, LoadingState, StatusCell, commonOptions, usePageFilters} from './pageCommon'

const filters = [
  {key: 'competitor', label: 'Конкурент', options: [{value: 'all', content: 'Все конкуренты'}, {value: 'amazon', content: 'Amazon'}, {value: 'walmart', content: 'Walmart'}, {value: 'target', content: 'Target'}]},
  {key: 'market', label: 'Рынок', options: [{value: 'all', content: 'Все рынки'}, {value: 'eu', content: 'Европа'}, {value: 'us', content: 'США'}]},
  {key: 'category', label: 'Категория', options: commonOptions.category},
  {key: 'brand', label: 'Бренд', options: commonOptions.brand},
  {key: 'match', label: 'Качество сопоставления', options: [{value: 'all', content: 'Любое'}, {value: 'high', content: 'Высокое'}, {value: 'medium', content: 'Среднее'}]},
  {key: 'availability', label: 'Наличие', options: [{value: 'all', content: 'Любое'}, {value: 'in_stock', content: 'В наличии'}, {value: 'low', content: 'Мало'}]},
  {key: 'period', label: 'Период', type: 'date' as const, options: commonOptions.period},
]

const columns: ColumnDef<CompetitorMatch, unknown>[] = [
  {accessorKey: 'sku', header: 'Наш SKU'},
  {accessorKey: 'competitor', header: 'Конкурент'},
  {accessorKey: 'competitorProduct', header: 'Товар конкурента'},
  {accessorKey: 'ourPrice', header: 'Наша цена'},
  {accessorKey: 'competitorPrice', header: 'Цена конкурента'},
  {accessorKey: 'delivery', header: 'Доставка'},
  {accessorKey: 'availability', header: 'Наличие'},
  {accessorKey: 'seller', header: 'Продавец'},
  {accessorKey: 'matchConfidence', header: 'Уверенность сопоставления'},
  {accessorKey: 'lastSeen', header: 'Последнее наблюдение'},
  {accessorKey: 'differencePct', header: 'Разница %'},
  {accessorKey: 'alert', header: 'Алерт', cell: (info) => <StatusCell value={info.row.original.alert === 'Норма' ? 'Норма' : 'Риск'} />},
]

export function CompetitorsPage() {
  const {values, setFilter, resetFilters} = usePageFilters(filters)
  const [selectedMatch, setSelectedMatch] = useState<CompetitorMatch | null>(null)
  const query = useQuery({queryKey: ['commerce-intelligence', 'competitors'], queryFn: getCompetitors})

  return (
    <CommercePage
      title="Конкуренты"
      subtitle="Мониторинг рынка, сопоставление товаров и анализ индекса цен."
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
            <ChartCard title="Динамика индекса цен vs M8">
              <div className="ci-chart">
                <ResponsiveContainer width="100%" height={280}>
                  <LineChart data={query.data.trend}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="date" />
                    <YAxis />
                    <Tooltip />
                    <Legend />
                    <Line dataKey="M8" stroke="#2f6fed" strokeWidth={2} dot={false} />
                    <Line dataKey="Amazon" stroke="#f2994a" dot={false} />
                    <Line dataKey="Walmart" stroke="#1a9b68" dot={false} />
                    <Line dataKey="Target" stroke="#9b51e0" dot={false} />
                    <Line dataKey="Best Buy" stroke="#eb5757" dot={false} />
                    <Line dataKey="Newegg" stroke="#56ccf2" dot={false} />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </ChartCard>
            <ChartCard title="Лестница рынка — AirPods Pro 2">
              <div className="ci-chart">
                <ResponsiveContainer width="100%" height={280}>
                  <BarChart data={query.data.ladder} layout="vertical">
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis type="number" />
                    <YAxis dataKey="name" type="category" />
                    <Tooltip />
                    <Bar dataKey="price" fill="#2f6fed" />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </ChartCard>
          </div>

          <ChartCard title="Индекс цен по категориям">
            <Heatmap rows={query.data.heatmap} />
          </ChartCard>

          <ChartCard title="Конкурентная ценовая разведка">
            <DataTable data={query.data.matches} columns={columns} getRowId={(row) => row.id} onRowClick={setSelectedMatch} />
          </ChartCard>

          <div className="ci-grid ci-grid_1-1-1">
            {['Исключения сопоставления товаров', 'Оповещения по конкурентам', 'Сводка изменений цен'].map((title) => (
              <ChartCard key={title} title={title}>
                <Text variant="body-2">Данные готовы к подключению к реальному источнику конкурентного мониторинга.</Text>
              </ChartCard>
            ))}
          </div>

          <DetailDrawer
            open={Boolean(selectedMatch)}
            title={selectedMatch?.competitorProduct ?? 'Товар конкурента'}
            subtitle={selectedMatch ? `${selectedMatch.competitor} · ${selectedMatch.sku}` : undefined}
            onClose={() => setSelectedMatch(null)}
            sections={[
              {title: 'Цена конкурента', content: <DefinitionGrid items={[{label: 'Наша цена', value: selectedMatch?.ourPrice}, {label: 'Цена конкурента', value: selectedMatch?.competitorPrice}, {label: 'Разница', value: selectedMatch?.differencePct}]} />},
              {title: 'Наличие', content: <Text variant="body-2">{selectedMatch?.availability}</Text>},
              {title: 'Доставка', content: <Text variant="body-2">{selectedMatch?.delivery}</Text>},
              {title: 'Продавец', content: <Text variant="body-2">{selectedMatch?.seller}</Text>},
              {title: 'Детали сопоставления', content: <Text variant="body-2">Уверенность сопоставления {selectedMatch?.matchConfidence}. Совпали бренд, модель, объем и цвет.</Text>},
              {title: 'Источник и последнее наблюдение', content: <Text variant="body-2">{selectedMatch?.lastSeen}</Text>},
              {title: 'Заметки по сопоставлению', content: <Text variant="body-2">Нужна ручная проверка только при изменении продавца или комплектации.</Text>},
              {title: 'Сырые данные', content: <Text variant="body-2">JSON наблюдения будет доступен после подключения провайдера.</Text>},
            ]}
          />
        </>
      ) : null}
    </CommercePage>
  )
}
