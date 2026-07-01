import {useCallback, useState} from 'react'
import {useQuery} from '@tanstack/react-query'
import type {ColumnDef} from '@tanstack/react-table'

import {ActionToolbar} from '../components/ActionToolbar'
import {ChartCard} from '../components/ChartCard'
import {DataTable} from '../components/DataTable'
import {getApprovals} from '../mock/queries'
import type {Approval} from '../mock/types'
import {notifyAction} from '../utils'
import {CommercePage, ErrorState, KpiGrid, LoadingState, StatusCell} from './pageCommon'

const columns: ColumnDef<Approval, unknown>[] = [
  {accessorKey: 'id', header: 'ID'},
  {accessorKey: 'type', header: 'Тип'},
  {accessorKey: 'subject', header: 'SKU / Группа'},
  {accessorKey: 'decision', header: 'Решение'},
  {accessorKey: 'expectedEffect', header: 'Ожидаемый эффект'},
  {accessorKey: 'risk', header: 'Риск', cell: (info) => <StatusCell value={info.row.original.risk} />},
  {accessorKey: 'requestedBy', header: 'Запросил'},
  {accessorKey: 'approver', header: 'Согласующий'},
  {accessorKey: 'status', header: 'Статус', cell: (info) => <StatusCell value={info.row.original.status} />},
  {accessorKey: 'dueAt', header: 'Срок'},
]

export function ApprovalsPage() {
  const [selectedRows, setSelectedRows] = useState<Approval[]>([])
  const query = useQuery({queryKey: ['commerce-intelligence', 'approvals'], queryFn: getApprovals})
  const onSelectionChange = useCallback((rows: Approval[]) => setSelectedRows(rows), [])

  return (
    <CommercePage title="Согласования" subtitle="Очередь решений по ценам, разметке и исключениям guardrails.">
      {query.isLoading ? <LoadingState /> : null}
      {query.isError ? <ErrorState onRetry={() => void query.refetch()} /> : null}
      {query.data ? (
        <>
          <KpiGrid items={query.data.kpis} />
          <ChartCard title="Очередь решений">
            <DataTable
              data={query.data.approvals}
              columns={columns}
              getRowId={(row) => row.id}
              enableRowSelection
              onSelectedRowsChange={onSelectionChange}
              toolbar={
                <ActionToolbar
                  selectedCount={selectedRows.length}
                  onApprove={() => notifyAction('Решения согласованы')}
                  onReject={() => notifyAction('Решения отклонены')}
                  onSchedule={() => notifyAction('Решения запланированы')}
                  onMore={() => notifyAction('Открыты детали согласования')}
                />
              }
            />
          </ChartCard>
        </>
      ) : null}
    </CommercePage>
  )
}
