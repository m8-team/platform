import {useMemo, useState} from 'react'
import {Button, Text} from '@gravity-ui/uikit'
import {useQuery} from '@tanstack/react-query'
import type {ColumnDef} from '@tanstack/react-table'
import {Bar, BarChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis} from 'recharts'

import {ChartCard} from '../components/ChartCard'
import {DataTable} from '../components/DataTable'
import {DetailDrawer} from '../components/DetailDrawer'
import {getProducts} from '../mock/queries'
import type {Product} from '../mock/types'
import {notifyAction} from '../utils'
import {CommercePage, DefinitionGrid, ErrorState, KpiGrid, LoadingState, MiniBars, StatusCell, commonOptions, usePageFilters} from './pageCommon'

const filters = [
  {key: 'category', label: 'Категория', options: commonOptions.category},
  {key: 'brand', label: 'Бренд', options: commonOptions.brand},
  {key: 'lifecycle', label: 'Жизненный цикл', options: [{value: 'all', content: 'Все'}, {value: 'growth', content: 'Рост'}, {value: 'mature', content: 'Зрелый'}, {value: 'seasonal', content: 'Сезонный'}]},
  {key: 'season', label: 'Сезон', options: [{value: 'all', content: 'Все сезоны'}, {value: 'summer', content: 'Лето'}, {value: 'winter', content: 'Зима'}]},
  {key: 'region', label: 'Регион', options: commonOptions.region},
  {key: 'channel', label: 'Канал', options: commonOptions.channel},
  {key: 'warehouse', label: 'Кластер склада', options: [{value: 'all', content: 'Все кластеры'}, {value: 'north', content: 'Север'}, {value: 'south', content: 'Юг'}]},
  {key: 'abc', label: 'ABC/XYZ сегмент', options: [{value: 'all', content: 'Все сегменты'}, {value: 'ax', content: 'AX'}, {value: 'cz', content: 'CZ'}]},
  {key: 'risk', label: 'Складской риск', options: [{value: 'all', content: 'Любой'}, {value: 'risk', content: 'Риск'}, {value: 'normal', content: 'Норма'}]},
]

const columns: ColumnDef<Product, unknown>[] = [
  {accessorKey: 'sku', header: 'SKU'},
  {accessorKey: 'product', header: 'Товар'},
  {accessorKey: 'category', header: 'Категория'},
  {accessorKey: 'currentPrice', header: 'Текущая цена'},
  {accessorKey: 'marketPrice', header: 'Рыночная цена'},
  {accessorKey: 'priceIndex', header: 'Индекс цены'},
  {accessorKey: 'stock', header: 'Остаток'},
  {accessorKey: 'coverageDays', header: 'Дней покрытия'},
  {accessorKey: 'sellThrough', header: 'Sell-through'},
  {accessorKey: 'sales7d', header: 'Продажи 7д'},
  {accessorKey: 'forecast30d', header: 'Прогноз 30д'},
  {accessorKey: 'elasticity', header: 'Эластичность'},
  {accessorKey: 'lifecycle', header: 'Жизненный цикл'},
  {accessorKey: 'risk', header: 'Риск', cell: (info) => <StatusCell value={info.row.original.risk} />},
  {accessorKey: 'status', header: 'Статус', cell: (info) => <StatusCell value={info.row.original.status} />},
]

export function ProductsPage() {
  const {values, setFilter, resetFilters} = usePageFilters(filters)
  const [activeSegment, setActiveSegment] = useState('Высокая маржа')
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null)
  const query = useQuery({queryKey: ['commerce-intelligence', 'products'], queryFn: getProducts})
  const riskBars = useMemo(
    () => (query.data?.products ?? []).map((product) => ({label: product.sku, value: product.coverageDays})),
    [query.data?.products],
  )

  return (
    <CommercePage
      title="Товары"
      subtitle="Цены, спрос и складские риски на уровне SKU."
      filters={filters}
      filterValues={values}
      onFilterChange={setFilter}
      onResetFilters={resetFilters}
      onSaveView={() => notifyAction('Сегмент сохранен')}
    >
      {query.isLoading ? <LoadingState /> : null}
      {query.isError ? <ErrorState onRetry={() => void query.refetch()} /> : null}
      {query.data ? (
        <>
          <KpiGrid items={query.data.kpis} />
          <div className="ci-segments">
            {query.data.segments.map((segment) => (
              <Button key={segment} view={activeSegment === segment ? 'action' : 'outlined'} onClick={() => setActiveSegment(segment)}>
                {segment}
              </Button>
            ))}
          </div>

          <ChartCard title="Инвентарь товаров" subtitle={`Активный сегмент: ${activeSegment}`}>
            <DataTable data={query.data.products} columns={columns} getRowId={(row) => row.sku} onRowClick={setSelectedProduct} />
          </ChartCard>

          <div className="ci-grid ci-grid_1-1-1">
            <ChartCard title="Распределение портфеля по категориям">
              <div className="ci-chart">
                <ResponsiveContainer width="100%" height={220}>
                  <BarChart data={query.data.portfolioDistribution}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="category" />
                    <YAxis />
                    <Tooltip />
                    <Bar name="SKU" dataKey="sku" fill="#2f6fed" />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </ChartCard>
            <ChartCard title="Карта рисков">
              <MiniBars items={riskBars} />
            </ChartCard>
            <ChartCard title="Лучшие ценовые возможности">
              <div className="ci-list">
                {query.data.opportunities.map((item) => (
                  <div className="ci-list__item" key={item.sku}>
                    <Text variant="body-2">{item.product}</Text>
                    <Text variant="caption-2" color="secondary">{item.action} · {item.effect}</Text>
                  </div>
                ))}
              </div>
            </ChartCard>
          </div>

          <DetailDrawer
            open={Boolean(selectedProduct)}
            title={selectedProduct?.product ?? 'Товар'}
            subtitle={selectedProduct?.sku}
            onClose={() => setSelectedProduct(null)}
            actions={<Button view="action" onClick={() => notifyAction('Действие отправлено на проверку')}>Создать рекомендацию</Button>}
            sections={[
              {title: 'История цены 90д', content: <MiniBars items={[{label: '90д', value: 229}, {label: '60д', value: 219}, {label: '30д', value: 209}, {label: 'сейчас', value: 199}]} />},
              {title: 'Прогноз 30д', content: <Text variant="body-2">{selectedProduct?.forecast30d ?? '—'} единиц, эластичность {selectedProduct?.elasticity ?? '—'}</Text>},
              {title: 'Остатки по складам', content: <DefinitionGrid items={[{label: 'Остаток', value: selectedProduct?.stock}, {label: 'Дней покрытия', value: selectedProduct?.coverageDays}]} />},
              {title: 'Лестница конкурентов', content: <Text variant="body-2">M8 находится на индексе {selectedProduct?.priceIndex ?? 100} относительно рынка.</Text>},
              {title: 'Сезонность', content: <Text variant="body-2">Пик сезонности ожидается через 3–5 недель.</Text>},
              {title: 'Последние события', content: <Text variant="body-2">Обновлен прогноз спроса, получены новые наблюдения конкурентов.</Text>},
              {title: 'Рекомендуемое действие', content: <Text variant="body-2">Есть возможность повысить цену на $4–$8. Ожидаемый эффект: +6–9% выручки при минимальном влиянии на объем.</Text>},
            ]}
          />
        </>
      ) : null}
    </CommercePage>
  )
}
