import {useState} from 'react'
import {Button, Text} from '@gravity-ui/uikit'
import {useQuery} from '@tanstack/react-query'
import type {ColumnDef} from '@tanstack/react-table'
import {Scatter, ScatterChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis} from 'recharts'

import {ChartCard} from '../components/ChartCard'
import {DataTable} from '../components/DataTable'
import {DetailDrawer} from '../components/DetailDrawer'
import {GuardrailList} from '../components/GuardrailList'
import {getMarkdownCandidates} from '../mock/queries'
import type {MarkdownCandidate} from '../mock/types'
import {notifyAction} from '../utils'
import {CommercePage, DefinitionGrid, ErrorState, KpiGrid, LoadingState, MiniBars, StatusCell, commonOptions, usePageFilters} from './pageCommon'

const filters = [
  {key: 'season', label: 'Сезон', options: [{value: 'all', content: 'Все сезоны'}, {value: 'summer', content: 'Лето'}, {value: 'winter', content: 'Зима'}]},
  {key: 'category', label: 'Категория', options: commonOptions.category},
  {key: 'brand', label: 'Бренд', options: commonOptions.brand},
  {key: 'region', label: 'Регион', options: commonOptions.region},
  {key: 'channel', label: 'Канал', options: commonOptions.channel},
  {key: 'storeCluster', label: 'Кластер магазина', options: [{value: 'all', content: 'Все кластеры'}, {value: 'flagship', content: 'Флагман'}, {value: 'outlet', content: 'Outlet'}]},
  {key: 'horizon', label: 'Горизонт', options: [{value: '30', content: '30 дней'}, {value: '60', content: '60 дней'}]},
  {key: 'period', label: 'Период', type: 'date' as const, options: commonOptions.period},
]

const columns: ColumnDef<MarkdownCandidate, unknown>[] = [
  {accessorKey: 'sku', header: 'SKU'},
  {accessorKey: 'product', header: 'Товар'},
  {accessorKey: 'currentPrice', header: 'Текущая цена'},
  {accessorKey: 'markdown', header: 'Предлагаемая уценка'},
  {accessorKey: 'recommendedPrice', header: 'Рекомендуемая цена'},
  {accessorKey: 'reason', header: 'Причина'},
  {accessorKey: 'seasonEndStock', header: 'Остаток к концу сезона'},
  {accessorKey: 'sellThroughLift', header: 'Рост sell-through'},
  {accessorKey: 'marginImpact', header: 'Влияние на маржу'},
  {accessorKey: 'confidence', header: 'Уверенность'},
  {accessorKey: 'status', header: 'Статус', cell: (info) => <StatusCell value={info.row.original.status} />},
]

export function MarkdownCenterPage() {
  const {values, setFilter, resetFilters} = usePageFilters(filters)
  const [selectedCandidate, setSelectedCandidate] = useState<MarkdownCandidate | null>(null)
  const query = useQuery({queryKey: ['commerce-intelligence', 'markdown'], queryFn: getMarkdownCandidates})
  const riskPoints = (query.data?.candidates ?? []).map((item, index) => ({risk: 20 + index * 16, stock: 32 + index * 11, name: item.sku}))

  return (
    <CommercePage
      title="Центр разметки"
      subtitle="Оптимизация сезонной распродажи, медленных товаров и sell-through."
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
          <ChartCard title="Кандидаты на разметку">
            <DataTable data={query.data.candidates} columns={columns} getRowId={(row) => row.sku} onRowClick={setSelectedCandidate} />
          </ChartCard>

          <div className="ci-grid ci-grid_1-1-1">
            <ChartCard title="Предстоящие окна разметки">
              <div className="ci-list">
                {query.data.windows.map((windowItem) => (
                  <div className="ci-list__item" key={windowItem.name}>
                    <Text variant="body-2">{windowItem.name}</Text>
                    <Text variant="caption-2" color="secondary">{windowItem.period} · {windowItem.sku} · {windowItem.status}</Text>
                  </div>
                ))}
              </div>
            </ChartCard>
            <ChartCard title="Сезонный риск vs позиция запасов">
              <div className="ci-chart">
                <ResponsiveContainer width="100%" height={220}>
                  <ScatterChart>
                    <CartesianGrid />
                    <XAxis dataKey="risk" name="Сезонный риск" />
                    <YAxis dataKey="stock" name="Запасы" />
                    <Tooltip cursor={{strokeDasharray: '3 3'}} />
                    <Scatter name="SKU" data={riskPoints} fill="#2f6fed" />
                  </ScatterChart>
                </ResponsiveContainer>
              </div>
            </ChartCard>
            <GuardrailList title="Политики и guardrails разметки" items={query.data.guardrails} />
          </div>

          <ChartCard title="Пакеты разметки">
            <MiniBars items={[{label: 'Сезонный риск', value: 412}, {label: 'Медленные продажи', value: 786}, {label: 'Избыточный запас', value: 338}]} />
          </ChartCard>

          <DetailDrawer
            open={Boolean(selectedCandidate)}
            title={selectedCandidate?.product ?? "Women's Hybrid Insulated Jacket"}
            subtitle={selectedCandidate?.sku}
            onClose={() => setSelectedCandidate(null)}
            actions={
              <>
                <Button view="action" onClick={() => notifyAction('Уценка согласована')}>Согласовать</Button>
                <Button view="outlined" onClick={() => notifyAction('Уценка отправлена на доработку')}>Вернуть</Button>
              </>
            }
            sections={[
              {title: 'Кривая сезонности', content: <MiniBars items={[{label: 'Неделя 1', value: 82}, {label: 'Неделя 2', value: 64}, {label: 'Неделя 3', value: 38}, {label: 'Неделя 4', value: 22}]} />},
              {title: 'Остатки по локациям', content: <DefinitionGrid items={[{label: 'Склад Север', value: '1 240'}, {label: 'Склад Юг', value: '860'}, {label: 'Магазины', value: '1 320'}]} />},
              {title: 'Ценообразование', content: <DefinitionGrid items={[{label: 'Текущая цена', value: selectedCandidate?.currentPrice}, {label: 'Уценка', value: selectedCandidate?.markdown}, {label: 'Рекомендуемая цена', value: selectedCandidate?.recommendedPrice}]} />},
              {title: 'Контекст конкурентов', content: <Text variant="body-2">Средняя цена рынка ниже на 8–12%, есть давление по двум конкурентам.</Text>},
              {title: 'Ожидаемая дата распродажи', content: <Text variant="body-2">Через 18 дней после применения уценки.</Text>},
              {title: 'Уверенность', content: <Text variant="body-2">{selectedCandidate?.confidence}</Text>},
            ]}
          />
        </>
      ) : null}
    </CommercePage>
  )
}
