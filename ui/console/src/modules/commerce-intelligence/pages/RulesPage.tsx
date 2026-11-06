import {useState} from 'react'
import {Button, Text, TextInput} from '@gravity-ui/uikit'
import {useQuery} from '@tanstack/react-query'
import type {ColumnDef} from '@tanstack/react-table'

import {ChartCard} from '../components/ChartCard'
import {DataTable} from '../components/DataTable'
import {getRules} from '../mock/queries'
import type {Rule} from '../mock/types'
import {notifyAction} from '../utils'
import {CommercePage, ErrorState, LoadingState, StatusCell} from './pageCommon'

const columns: ColumnDef<Rule, unknown>[] = [
  {accessorKey: 'name', header: 'Название'},
  {accessorKey: 'type', header: 'Тип'},
  {accessorKey: 'scope', header: 'Область действия'},
  {accessorKey: 'limit', header: 'Лимит'},
  {accessorKey: 'priority', header: 'Приоритет'},
  {accessorKey: 'status', header: 'Статус', cell: (info) => <StatusCell value={info.row.original.status} />},
  {accessorKey: 'updatedAt', header: 'Последнее изменение'},
  {accessorKey: 'author', header: 'Автор'},
]

export function RulesPage() {
  const [activeGroup, setActiveGroup] = useState('Активные правила')
  const query = useQuery({queryKey: ['commerce-intelligence', 'rules'], queryFn: getRules})

  return (
    <CommercePage
      title="Правила"
      subtitle="Ценовые ограничения, guardrails и политики безопасного изменения цен."
      actions={<Button view="action" onClick={() => notifyAction('Форма создания правила открыта')}>Создать правило</Button>}
    >
      {query.isLoading ? <LoadingState /> : null}
      {query.isError ? <ErrorState onRetry={() => void query.refetch()} /> : null}
      {query.data ? (
        <>
          <div className="ci-rule-layout">
            <ChartCard title="Разделы правил">
              <div className="ci-segments ci-segments_vertical">
                {query.data.groups.map((group) => (
                  <Button key={group} view={activeGroup === group ? 'action' : 'outlined'} width="max" onClick={() => setActiveGroup(group)}>
                    {group}
                  </Button>
                ))}
              </div>
            </ChartCard>

            <ChartCard
              title={activeGroup}
              subtitle="Каркас готов для подключения редактора условий, приоритетов и проверки влияния."
              actions={
                <>
                  <Button view="outlined" onClick={() => notifyAction('Редактор открыт')}>Редактировать</Button>
                  <Button view="outlined" onClick={() => notifyAction('Правило отключено')}>Отключить</Button>
                  <Button view="outlined" onClick={() => notifyAction('Расчет влияния запущен')}>Посмотреть влияние</Button>
                </>
              }
            >
              <div className="ci-form-grid">
                <label className="ci-filter">
                  <Text variant="caption-2" color="secondary">Название</Text>
                  <TextInput placeholder="Например: минимальная маржа категории" />
                </label>
                <label className="ci-filter">
                  <Text variant="caption-2" color="secondary">Лимит</Text>
                  <TextInput placeholder="20%" />
                </label>
                <label className="ci-filter">
                  <Text variant="caption-2" color="secondary">Область действия</Text>
                  <TextInput placeholder="Категория, бренд или SKU" />
                </label>
              </div>
              <DataTable data={query.data.rules} columns={columns} getRowId={(row) => row.name} />
            </ChartCard>
          </div>
        </>
      ) : null}
    </CommercePage>
  )
}
