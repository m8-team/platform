import {useEffect, useMemo, useState} from 'react'
import type {ReactNode} from 'react'
import {Button, Text} from '@gravity-ui/uikit'
import type {TableDataItem} from '@gravity-ui/uikit'

import {
  ResourceTable,
  type ResourceTableColumn,
} from '../../../../components/ResourceTable'

export interface DataTableCellContext<T> {
  row: {original: T}
}

export interface DataTableColumn<T> {
  accessorKey: Extract<keyof T, string>
  header: ReactNode
  cell?: (context: DataTableCellContext<T>) => ReactNode
  size?: number
  enableSorting?: boolean
}

export interface DataTableProps<T extends TableDataItem> {
  title?: string
  data: T[]
  columns: DataTableColumn<T>[]
  getRowId: (row: T, index: number) => string
  onRowClick?: (row: T) => void
  enableRowSelection?: boolean
  onSelectedRowsChange?: (rows: T[]) => void
  toolbar?: ReactNode
  emptyTitle?: string
  emptyDescription?: string
  density?: 'compact' | 'normal'
}

export function DataTable<T extends TableDataItem>({
  title,
  data,
  columns,
  getRowId,
  onRowClick,
  enableRowSelection,
  onSelectedRowsChange,
  toolbar,
  emptyTitle = 'Нет данных',
  emptyDescription = 'Измените фильтры или повторите загрузку.',
  density: initialDensity = 'compact',
}: DataTableProps<T>) {
  const [selectedRows, setSelectedRows] = useState<T[]>([])
  const [density, setDensity] = useState(initialDensity)
  const tableColumns = useMemo<ResourceTableColumn<T>[]>(
    () =>
      columns.map((column) => ({
        id: column.accessorKey,
        name: typeof column.header === 'string' ? column.header : () => column.header,
        width: column.size,
        template: column.cell
          ? (item) => column.cell?.({row: {original: item}})
          : (item) => renderValue(item[column.accessorKey]),
        meta: {sortable: column.enableSorting !== false},
      })),
    [columns],
  )

  useEffect(() => {
    onSelectedRowsChange?.(selectedRows)
  }, [onSelectedRowsChange, selectedRows])

  if (data.length === 0) {
    return (
      <div className="ci-table-empty">
        <Text variant="body-2">{emptyTitle}</Text>
        <Text variant="caption-2" color="secondary">
          {emptyDescription}
        </Text>
      </div>
    )
  }

  return (
    <div className={`ci-table ci-table_${density}`}>
      <div className="ci-table__toolbar">
        <div>
          {title ? (
            <Text as="h2" variant="header-1">
              {title}
            </Text>
          ) : null}
          <Text variant="caption-2" color="secondary">
            {selectedRows.length > 0 ? `${selectedRows.length} выбрано` : `${data.length} строк`}
          </Text>
        </div>
        <div className="ci-table__toolbar-actions">
          {toolbar}
          <Button
            view="outlined"
            onClick={() => setDensity(density === 'compact' ? 'normal' : 'compact')}
          >
            Плотность
          </Button>
          <Button view="outlined">Экспорт</Button>
        </div>
      </div>

      <ResourceTable
        data={data}
        columns={tableColumns}
        getRowId={getRowId}
        loadingContent="Загрузка…"
        emptyContent={emptyTitle}
        className="ci-table__uikit-table"
        onRowActivate={onRowClick}
        selectable={enableRowSelection}
        onSelectedRowsChange={setSelectedRows}
        sortable
        settings={{}}
        pagination={{
          mode: 'client',
          defaultPageSize: 8,
          pageSizeOptions: [8, 16, 32],
        }}
      />
    </div>
  )
}

function renderValue(value: unknown) {
  if (value === undefined || value === null || value === '') return '—'
  if (typeof value === 'string' || typeof value === 'number') return value
  if (typeof value === 'boolean') return value ? 'Да' : 'Нет'
  return String(value)
}
