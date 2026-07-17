import {useState} from 'react'
import {Button, Text} from '@gravity-ui/uikit'
import {useQuery} from '@tanstack/react-query'
import type {ColumnDef} from '@tanstack/react-table'
import {CartesianGrid, Legend, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis} from 'recharts'

import {ChartCard} from '../components/ChartCard'
import {DataTable} from '../components/DataTable'
import {GuardrailList} from '../components/GuardrailList'
import {ScenarioBuilder} from '../components/ScenarioBuilder'
import {WhatIfControls} from '../components/WhatIfControls'
import type {ScenarioSettings} from '../components/WhatIfControls'
import {getSimulation} from '../mock/queries'
import type {SimulationPlannerRow} from '../mock/types'
import {notifyAction} from '../utils'
import {CommercePage, ErrorState, KpiGrid, LoadingState, StatusCell, commonOptions, usePageFilters} from './pageCommon'

const filters = [
  {key: 'scenario', label: 'Сценарий', options: [{value: 'a', content: 'Сценарий A'}, {value: 'b', content: 'Сценарий B'}, {value: 'base', content: 'Базовый'}]},
  {key: 'region', label: 'Регион', options: commonOptions.region},
  {key: 'category', label: 'Категория', options: commonOptions.category},
  {key: 'brand', label: 'Бренд', options: commonOptions.brand},
  {key: 'horizon', label: 'Горизонт', options: [{value: '30', content: '30 дней'}, {value: '60', content: '60 дней'}]},
  {key: 'goal', label: 'Цель стратегии', options: [{value: 'margin', content: 'Максимум маржи'}, {value: 'stock', content: 'Очистка запасов'}, {value: 'balanced', content: 'Баланс'}]},
  {key: 'period', label: 'Период', type: 'date' as const, options: commonOptions.period},
]

const plannerColumns: ColumnDef<SimulationPlannerRow, unknown>[] = [
  {accessorKey: 'sku', header: 'SKU'},
  {accessorKey: 'currentPrice', header: 'Текущая цена'},
  {accessorKey: 'markdown', header: 'Предлагаемая уценка'},
  {accessorKey: 'sellThroughLift', header: 'Рост sell-through'},
  {accessorKey: 'marginImpact', header: 'Влияние на маржу'},
  {accessorKey: 'seasonEndStock', header: 'Остаток к концу сезона'},
  {accessorKey: 'confidence', header: 'Уверенность'},
  {accessorKey: 'status', header: 'Статус', cell: (info) => <StatusCell value={info.row.original.status} />},
]

export function SimulationPage() {
  const {values, setFilter, resetFilters} = usePageFilters(filters)
  const [settings, setSettings] = useState<ScenarioSettings>({markdown: 15, priceIndex: 98, stockDays: 47, risk: 'Средняя'})
  const query = useQuery({queryKey: ['commerce-intelligence', 'simulation'], queryFn: getSimulation})

  return (
    <CommercePage
      title="Симуляции"
      subtitle="Планирование сценариев, оптимизация разметки и проверка guardrails."
      filters={filters}
      filterValues={values}
      onFilterChange={setFilter}
      onResetFilters={resetFilters}
      primaryFilterAction={{label: 'Запустить симуляцию', onClick: () => notifyAction('Симуляция запущена', `Уценка ${settings.markdown}%, индекс ${settings.priceIndex}`)}}
    >
      {query.isLoading ? <LoadingState /> : null}
      {query.isError ? <ErrorState onRetry={() => void query.refetch()} /> : null}
      {query.data ? (
        <>
          <KpiGrid items={query.data.kpis} />
          <div className="ci-grid ci-grid_2-1">
            <ScenarioBuilder items={query.data.scenarioComparison} />
            <WhatIfControls value={settings} onChange={setSettings} onRun={() => notifyAction('Симуляция пересчитана')} />
          </div>

          <div className="ci-grid ci-grid_1-1">
            <ChartCard title="Влияние изменения цены">
              <div className="ci-chart">
                <ResponsiveContainer width="100%" height={280}>
                  <LineChart data={query.data.priceImpact}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="markdown" label={{value: 'Средняя уценка %', position: 'insideBottom', offset: -4}} />
                    <YAxis />
                    <Tooltip />
                    <Legend />
                    <Line name="Выручка" dataKey="revenue" stroke="#2f6fed" strokeWidth={2} />
                    <Line name="Маржа" dataKey="margin" stroke="#1a9b68" strokeWidth={2} />
                    <Line name="Продано единиц" dataKey="units" stroke="#f2994a" strokeWidth={2} />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </ChartCard>
            <GuardrailList title="Guardrails сценариев" items={query.data.guardrails} />
          </div>

          <ChartCard title="Планировщик разметки">
            <DataTable data={query.data.planner} columns={plannerColumns} getRowId={(row) => row.sku} />
          </ChartCard>

          <div className="ci-grid ci-grid_1-1">
            <ChartCard
              title="Сравнение сценариев"
              actions={
                <>
                  <Button view="outlined" onClick={() => notifyAction('Сценарий сохранен')}>Сохранить сценарий</Button>
                  <Button view="action" onClick={() => notifyAction('Сценарий отправлен на согласование')}>Отправить на согласование</Button>
                </>
              }
            >
              <DataTable
                data={query.data.scenarioSummary}
                columns={[
                  {accessorKey: 'metric', header: 'Метрика'},
                  {accessorKey: 'base', header: 'Базовый'},
                  {accessorKey: 'scenarioA', header: 'Сценарий A'},
                  {accessorKey: 'scenarioB', header: 'Сценарий B'},
                  {accessorKey: 'deltaA', header: 'Δ A к базовому'},
                  {accessorKey: 'deltaB', header: 'Δ B к базовому'},
                ]}
                getRowId={(row) => row.metric}
              />
            </ChartCard>
            <ChartCard title="Почему рекомендован сценарий A">
              <div className="ci-explanation">
                <Text variant="body-2">
                  Сценарий A балансирует рост маржи и ускорение sell-through, оставаясь в пределах guardrails. Он сокращает дни запасов на 14, повышает валовую маржу на 0.6 п.п. и требует умеренного бюджета разметки.
                </Text>
                {[
                  'Увеличивает валовую маржу',
                  'Сохраняет ценовую дисциплину относительно конкурентов',
                  'Улучшает sell-through',
                  'Остается в рамках обязательных guardrails',
                  'Имеет низкий риск выполнения',
                ].map((item) => (
                  <div className="ci-explanation__benefit" key={item}>{item}</div>
                ))}
              </div>
            </ChartCard>
          </div>
        </>
      ) : null}
    </CommercePage>
  )
}
