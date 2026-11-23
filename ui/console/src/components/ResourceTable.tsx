import {useCallback, useEffect, useMemo, useState} from 'react'
import type {ReactNode} from 'react'
import {flexRender, useTable} from '@tanstack/react-table'
import type {ColumnVisibilityState, RowData, RowSelectionState, SortingState, Updater} from '@tanstack/react-table'
import {Button, Checkbox, Pagination, TextInput} from '@gravity-ui/uikit'
import {features, selectionColumn} from './tableModel'
import type {TableColumn} from './tableModel'

const selectionColumnId = '_select'
const settingsColumnId = '_settings'

export interface ResourceTableSettings {
  label: string
  moveUp: string
  moveDown: string
  storageKey?: string
  enableSearch?: boolean
  searchPlaceholder?: string
}

export interface ResourceTableClientPagination {
  mode: 'client'
  defaultPageSize?: number
  pageSizeOptions?: number[]
}

export interface ResourceTableServerPagination {
  mode: 'server'
  page: number
  pageSize: number
  total: number
  disabled?: boolean
  pageSizeOptions?: number[]
  onUpdate: (page: number, pageSize: number) => void
}

export type ResourceTablePagination = ResourceTableClientPagination | ResourceTableServerPagination

interface ResourceTableFilteringBase {
  searchPlaceholder?: string
  ariaLabel?: string
}

export interface ResourceTableClientFiltering extends ResourceTableFilteringBase {
  mode: 'client'
}

export interface ResourceTableServerFiltering extends ResourceTableFilteringBase {
  mode: 'server'
  value: string
  onUpdate: (value: string) => void
}

export type ResourceTableFiltering = ResourceTableClientFiltering | ResourceTableServerFiltering

export interface ResourceTableSorting {
  mode: 'server'
  value: SortingState
  onUpdate: (value: SortingState) => void
}

export interface ResourceTableSelectionActions<TData> {
  selectedItems: TData[]
  clearSelection: () => void
}

export interface ResourceTableProps<TData extends RowData> {
  data: TData[]
  columns: TableColumn<TData>[]
  getRowId: (item: TData) => string
  loading?: boolean
  loadingContent: string
  emptyContent: string
  className?: string
  onRowActivate?: (item: TData) => void
  getRowAriaLabel?: (item: TData) => string
  selectPageLabel: string
  selectRowLabel: string
  selectable?: boolean
  onSelectedRowsChange?: (items: TData[]) => void
  settings?: ResourceTableSettings
  renderSelectionActions?: (context: ResourceTableSelectionActions<TData>) => ReactNode
  pagination?: ResourceTablePagination
  filtering?: ResourceTableFiltering
  sortable?: boolean
  sorting?: ResourceTableSorting
}

export function ResourceTable<TData extends RowData>({
  data,
  columns,
  getRowId,
  loading = false,
  loadingContent,
  emptyContent,
  className,
  onRowActivate,
  getRowAriaLabel,
  selectable = false,
  selectPageLabel,
  selectRowLabel,
  onSelectedRowsChange,
  settings,
  renderSelectionActions,
  pagination,
  filtering,
  sortable = false,
  sorting: controlledSorting,
}: ResourceTableProps<TData>) {
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({})
  const [internalPage, setInternalPage] = useState(1)
  const [internalPageSize, setInternalPageSize] = useState(
    pagination?.mode === 'client' ? pagination.defaultPageSize ?? 20 : 20,
  )
  const serverPagination = pagination?.mode === 'server'
  const page = serverPagination ? pagination.page : internalPage
  const pageSize = serverPagination ? pagination.pageSize : internalPageSize
  const [internalSorting, setInternalSorting] = useState<SortingState>([])
  const sorting = controlledSorting?.value ?? internalSorting
  const serverSorting = controlledSorting?.mode === 'server'
  const [internalFilter, setInternalFilter] = useState('')
  const serverFiltering = filtering?.mode === 'server'
  const globalFilter = serverFiltering ? filtering.value : internalFilter
  const [columnVisibility, setColumnVisibility] = useState<ColumnVisibilityState>(() =>
    readTableSettings(settings?.storageKey).columnVisibility,
  )
  const [columnOrder, setColumnOrder] = useState<string[]>(() => {
    const storedOrder = readTableSettings(settings?.storageKey).columnOrder
    return storedOrder.length === 0 ? [] : normalizeColumnOrder(storedOrder, selectable, Boolean(settings))
  })
  const [columnSearch, setColumnSearch] = useState('')
  const tableColumns = useMemo<TableColumn<TData>[]>(() => [
    ...(selectable ? [selectionColumn<TData>(selectionColumnId, selectPageLabel, selectRowLabel)] : []),
    ...columns,
    ...(settings ? [{id: settingsColumnId, size: 44, enableSorting: false, enableHiding: false}] : []),
  ], [columns, selectable, selectPageLabel, selectRowLabel, settings])
  const updateColumnVisibility = useCallback(
    (updater: Updater<ColumnVisibilityState>) => {
      setColumnVisibility((current) => {
        const next = applyUpdater(updater, current)
        persistTableSettings(settings?.storageKey, next, columnOrder)
        return next
      })
    },
    [columnOrder, settings?.storageKey],
  )
  const updateColumnOrder = useCallback(
    (updater: Updater<string[]>) => {
      setColumnOrder((current) => {
        const requested = applyUpdater(updater, current)
        const next = normalizeColumnOrder(requested, selectable, Boolean(settings))
        persistTableSettings(settings?.storageKey, columnVisibility, next)
        return next
      })
    },
    [columnVisibility, selectable, settings],
  )
  const table = useTable({
    features,
    columns: tableColumns,
    data,
    getRowId,
    enableRowSelection: selectable,
    enableMultiRowSelection: selectable,
    enableSorting: sortable,
    onRowSelectionChange: setRowSelection,
    onColumnOrderChange: updateColumnOrder,
    onColumnVisibilityChange: updateColumnVisibility,
    onSortingChange: (updater) => {
      const next = applyUpdater(updater, sorting)
      if (serverSorting) controlledSorting.onUpdate(next)
      else setInternalSorting(next)
    },
    onGlobalFilterChange: (updater) => {
      const next = applyUpdater(updater, globalFilter)
      if (serverFiltering) filtering.onUpdate(next)
      else setInternalFilter(next)
    },
    manualPagination: !pagination || serverPagination,
    manualFiltering: !filtering || serverFiltering,
    manualSorting: !sortable || serverSorting,
    state: {
      rowSelection,
      columnOrder,
      columnVisibility,
      sorting,
      globalFilter,
      pagination: {pageIndex: page - 1, pageSize},
    },
  })
  const filteredRowCount = serverPagination
    ? pagination.total
    : filtering
      ? table.getFilteredRowModel().rows.length
      : data.length
  const lastPage = Math.max(1, Math.ceil(filteredRowCount / pageSize))
  const effectivePage = Math.min(page, lastPage)

  const selectedItems = useMemo(
    () => data.filter((item) => rowSelection[getRowId(item)]),
    [data, getRowId, rowSelection],
  )

  useEffect(() => {
    onSelectedRowsChange?.(selectedItems)
  }, [onSelectedRowsChange, selectedItems])

  const activate = (item: TData) => onRowActivate?.(item)

  return (
    <div className="m8-resource-table">
      {filtering ? (
        <div className="m8-resource-table-filters">
          <TextInput
            value={globalFilter}
            placeholder={filtering.searchPlaceholder}
            aria-label={filtering.ariaLabel ?? filtering.searchPlaceholder}
            hasClear
            onUpdate={(value) => {
              if (serverFiltering) filtering.onUpdate(value)
              else setInternalFilter(value)
              if (!serverPagination) setInternalPage(1)
            }}
          />
        </div>
      ) : null}
      <div className="m8-resource-table-shell" aria-busy={loading}>
        <table className={className}>
          <thead>{table.getHeaderGroups().map(group => <tr key={group.id}>
            {group.headers.map(header => {
              const column = header.column
              const direction = column.getIsSorted()
              const label = flexRender(column.columnDef.header, header.getContext())
              return <th key={header.id} style={{width: header.getSize()}}
                aria-sort={direction === 'asc' ? 'ascending' : direction === 'desc' ? 'descending' : undefined}>
                {column.id === settingsColumnId && settings ? (
                  <details className="m8-table-settings">
                    <summary aria-label={settings.label}>⚙</summary>
                    <div className="m8-table-settings-panel">
                      {settings.enableSearch && <TextInput value={columnSearch} onUpdate={setColumnSearch}
                        aria-label={settings.searchPlaceholder} placeholder={settings.searchPlaceholder} />}
                      {table.getAllLeafColumns().filter(item => item.getCanHide()).map((item, index, items) => {
                        const title = String(item.columnDef.header ?? item.id)
                        if (!title.toLocaleLowerCase().includes(columnSearch.toLocaleLowerCase())) return null
                        const move = (step: number) => {
                          const order = items.map(entry => entry.id)
                          order.splice(index + step, 0, ...order.splice(index, 1))
                          table.setColumnOrder(order)
                        }
                        return <div key={item.id}>
                          <Checkbox checked={item.getIsVisible()} onUpdate={checked => item.toggleVisibility(checked)}>{title}</Checkbox>
                          <Button aria-label={settings.moveUp} disabled={index === 0} onClick={() => move(-1)}>↑</Button>
                          <Button aria-label={settings.moveDown} disabled={index === items.length - 1} onClick={() => move(1)}>↓</Button>
                        </div>
                      })}
                    </div>
                  </details>
                ) : column.getCanSort() ? <Button view="flat" onClick={column.getToggleSortingHandler()}>
                  {label}{direction === 'asc' ? ' ↑' : direction === 'desc' ? ' ↓' : ''}
                </Button> : label}
              </th>
            })}
          </tr>)}</thead>
          <tbody>{table.getRowModel().rows.map(row => <tr key={row.id}
            tabIndex={onRowActivate ? 0 : undefined} aria-label={getRowAriaLabel?.(row.original)}
            onClick={event => { if (!isInteractiveTarget(event.target)) activate(row.original) }}
            onKeyDown={event => {
              if (onRowActivate && event.target === event.currentTarget && (event.key === 'Enter' || event.key === ' ')) {
                event.preventDefault()
                activate(row.original)
              }
            }}>
            {row.getVisibleCells().map(cell => <td key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</td>)}
          </tr>)}
            {!table.getRowModel().rows.length && <tr><td colSpan={table.getVisibleLeafColumns().length}>
              {loading ? loadingContent : emptyContent}
            </td></tr>}
          </tbody>
        </table>
      </div>
      {pagination && filteredRowCount > 0 ? (
        <div
          className="m8-resource-table-pagination"
          aria-disabled={serverPagination && pagination.disabled ? true : undefined}
        >
          <Pagination
            page={effectivePage}
            pageSize={pageSize}
            total={filteredRowCount}
            pageSizeOptions={pagination.pageSizeOptions ?? [10, 20, 50, 100]}
            onUpdate={(nextPage, nextPageSize) => {
              if (serverPagination) {
                if (!pagination.disabled) pagination.onUpdate(nextPage, nextPageSize)
              } else {
                setInternalPage(nextPage)
                setInternalPageSize(nextPageSize)
              }
            }}
            showPages={!serverPagination}
            showInput={false}
            className={serverPagination && pagination.disabled ? 'm8-resource-table-pagination_disabled' : undefined}
          />
        </div>
      ) : null}
      {selectedItems.length > 0 && renderSelectionActions
        ? (
            <div className="m8-selection-actions-footer">
              {renderSelectionActions({
                selectedItems,
                clearSelection: () => setRowSelection({}),
              })}
            </div>
          )
        : null}
    </div>
  )
}

function isInteractiveTarget(target: EventTarget | null) {
  return target instanceof Element && Boolean(target.closest('button, input, a, [role="checkbox"]'))
}

interface StoredTableSettings {
  columnVisibility: ColumnVisibilityState
  columnOrder: string[]
}

function readTableSettings(storageKey: string | undefined): StoredTableSettings {
  if (!storageKey || typeof window === 'undefined') return {columnVisibility: {}, columnOrder: []}
  try {
    const value = JSON.parse(window.localStorage.getItem(storageKey) ?? '{}') as Partial<StoredTableSettings>
    return {
      columnVisibility: value.columnVisibility ?? {},
      columnOrder: Array.isArray(value.columnOrder) ? value.columnOrder : [],
    }
  } catch {
    return {columnVisibility: {}, columnOrder: []}
  }
}

function persistTableSettings(
  storageKey: string | undefined,
  columnVisibility: ColumnVisibilityState,
  columnOrder: string[],
) {
  if (!storageKey || typeof window === 'undefined') return
  try {
    window.localStorage.setItem(storageKey, JSON.stringify({columnVisibility, columnOrder}))
  } catch {
    // Table settings are an optional enhancement; storage restrictions must not break the table.
  }
}

function normalizeColumnOrder(order: string[], selectable: boolean, withSettings: boolean) {
  const contentColumns = order.filter((id) => id !== selectionColumnId && id !== settingsColumnId)
  return [
    ...(selectable ? [selectionColumnId] : []),
    ...contentColumns,
    ...(withSettings ? [settingsColumnId] : []),
  ]
}

function applyUpdater<T>(updater: Updater<T>, current: T): T {
  return typeof updater === 'function' ? (updater as (old: T) => T)(current) : updater
}
