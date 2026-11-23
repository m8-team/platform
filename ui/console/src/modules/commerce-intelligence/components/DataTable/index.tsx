import {useEffect, useMemo, useState} from 'react'
import type {ReactNode} from 'react'
import {Button, Checkbox, Text} from '@gravity-ui/uikit'
import {flexRender, useTable} from '@tanstack/react-table'
import type {RowData} from '@tanstack/react-table'
import {features, selectionColumn} from '../../../../components/tableModel'
import type {TableColumn} from '../../../../components/tableModel'

export type DataTableProps<T extends RowData> = {
  title?: string
  data: T[]
  columns: TableColumn<T>[]
  getRowId?: (row: T, index: number) => string
  onRowClick?: (row: T) => void
  enableRowSelection?: boolean
  onSelectedRowsChange?: (rows: T[]) => void
  toolbar?: ReactNode
  emptyTitle?: string
  emptyDescription?: string
  density?: 'compact' | 'normal'
}

export function DataTable<T extends RowData>({
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
  const [density, setDensity] = useState(initialDensity)

  const tableColumns = useMemo(() => {
    if (!enableRowSelection) {
      return columns
    }

    return [selectionColumn<T>('__select', 'Выбрать страницу', 'Выбрать строку'), ...columns]
  }, [columns, enableRowSelection])

  const table = useTable({
    features,
    data,
    columns: tableColumns,
    enableRowSelection,
    getRowId,
    initialState: {pagination: {pageIndex: 0, pageSize: 8}},
  })

  const selectedModel = table.getSelectedRowModel()
  const selectedRows = useMemo(() => selectedModel.rows.map(row => row.original), [selectedModel])

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
          <Button view="outlined" onClick={() => setDensity(density === 'compact' ? 'normal' : 'compact')}>
            Плотность
          </Button>
          <details className="ci-column-menu">
            <summary>Колонки</summary>
            <div className="ci-column-menu__content">
              {table
                .getAllLeafColumns()
                .filter((column) => column.id !== '__select')
                .map((column) => (
                  <Checkbox key={column.id} checked={column.getIsVisible()} onUpdate={(checked) => column.toggleVisibility(checked)}>
                    {String(column.columnDef.header ?? column.id)}
                  </Checkbox>
                ))}
            </div>
          </details>
          <Button view="outlined">Экспорт</Button>
        </div>
      </div>

      <div className="ci-table__scroll">
        <table>
          <thead>
            {table.getHeaderGroups().map((headerGroup) => (
              <tr key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <th key={header.id} style={{width: header.getSize()}}>
                    {header.isPlaceholder ? null : !header.column.getCanSort() ? flexRender(header.column.columnDef.header, header.getContext()) : (
                      <button
                        className="ci-table__sort"
                        type="button"
                        onClick={header.column.getToggleSortingHandler()}
                        disabled={!header.column.getCanSort()}
                      >
                        {flexRender(header.column.columnDef.header, header.getContext())}
                        {header.column.getIsSorted() === 'asc' ? ' ↑' : header.column.getIsSorted() === 'desc' ? ' ↓' : ''}
                      </button>
                    )}
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody>
            {table.getRowModel().rows.map((row) => (
              <tr key={row.id} className={row.getIsSelected() ? 'ci-table__row_selected' : undefined} onClick={() => onRowClick?.(row.original)}>
                {row.getVisibleCells().map((cell) => (
                  <td key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="ci-table__pagination">
        <Button view="outlined" disabled={!table.getCanPreviousPage()} onClick={() => table.previousPage()}>
          Назад
        </Button>
        <Text variant="caption-2" color="secondary">
          Страница {table.state.pagination.pageIndex + 1} из {table.getPageCount()}
        </Text>
        <Button view="outlined" disabled={!table.getCanNextPage()} onClick={() => table.nextPage()}>
          Вперед
        </Button>
      </div>
    </div>
  )
}
