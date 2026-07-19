import {useCallback, useEffect, useMemo, useState} from 'react'
import type {ReactNode} from 'react'
import {getSettingsColumn, selectionColumn, Table, useTable} from '@gravity-ui/table'
import {
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
} from '@gravity-ui/table/tanstack'
import type {
  ColumnDef,
  RowSelectionState,
  SortingState,
  Updater,
  VisibilityState,
} from '@gravity-ui/table/tanstack'
import {Pagination, TextInput} from '@gravity-ui/uikit'

const selectionColumnId = '_select'
const settingsColumnId = '_settings'

export interface ResourceTableSettings {
  storageKey?: string
  enableSearch?: boolean
  searchPlaceholder?: string
}

export interface ResourceTablePagination {
  page?: number
  pageSize?: number
  total?: number
  pageSizeOptions?: number[]
  onUpdate?: (page: number, pageSize: number) => void
}

export interface ResourceTableFiltering {
  searchPlaceholder?: string
}

export interface ResourceTableSelectionActions<TData> {
  selectedItems: TData[]
  clearSelection: () => void
}

export interface ResourceTableProps<TData> {
  data: TData[]
  columns: ColumnDef<TData>[]
  getRowId: (item: TData) => string
  loading?: boolean
  loadingContent: string
  emptyContent: string
  className?: string
  onRowActivate?: (item: TData) => void
  selectable?: boolean
  onSelectedRowsChange?: (items: TData[]) => void
  settings?: ResourceTableSettings
  renderSelectionActions?: (context: ResourceTableSelectionActions<TData>) => ReactNode
  pagination?: ResourceTablePagination
  filtering?: ResourceTableFiltering
  sortable?: boolean
}

export function ResourceTable<TData>({
  data,
  columns,
  getRowId,
  loading = false,
  loadingContent,
  emptyContent,
  className,
  onRowActivate,
  selectable = false,
  onSelectedRowsChange,
  settings,
  renderSelectionActions,
  pagination,
  filtering,
  sortable = false,
}: ResourceTableProps<TData>) {
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({})
  const [internalPage, setInternalPage] = useState(1)
  const [internalPageSize, setInternalPageSize] = useState(pagination?.pageSize ?? 20)
  const page = pagination?.page ?? internalPage
  const pageSize = pagination?.pageSize ?? internalPageSize
  const serverPagination = Boolean(pagination?.onUpdate)
  const [sorting, setSorting] = useState<SortingState>([])
  const [globalFilter, setGlobalFilter] = useState('')
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>(() =>
    readTableSettings(settings?.storageKey).columnVisibility,
  )
  const [columnOrder, setColumnOrder] = useState<string[]>(() => {
    const storedOrder = readTableSettings(settings?.storageKey).columnOrder
    return storedOrder.length === 0 ? [] : normalizeColumnOrder(storedOrder, selectable, Boolean(settings))
  })
  const tableColumns = useMemo(
    () => [
      ...(selectable ? [selectionColumn as ColumnDef<TData>] : []),
      ...columns,
      ...(settings
        ? [
            getSettingsColumn<TData>(settingsColumnId, {
              sortable: true,
              filterable: true,
              enableSearch: settings.enableSearch,
              searchPlaceholder: settings.searchPlaceholder,
            }),
          ]
        : []),
    ],
    [columns, selectable, settings],
  )
  const updateColumnVisibility = useCallback(
    (updater: Updater<VisibilityState>) => {
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
    columns: tableColumns,
    data,
    getRowId,
    enableRowSelection: selectable,
    enableMultiRowSelection: selectable,
    enableSorting: sortable,
    onRowSelectionChange: setRowSelection,
    onColumnOrderChange: updateColumnOrder,
    onColumnVisibilityChange: updateColumnVisibility,
    onSortingChange: setSorting,
    onGlobalFilterChange: setGlobalFilter,
    getFilteredRowModel: filtering ? getFilteredRowModel() : undefined,
    getSortedRowModel: sortable ? getSortedRowModel() : undefined,
    getPaginationRowModel: pagination && !serverPagination ? getPaginationRowModel() : undefined,
    manualPagination: serverPagination,
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
    ? pagination?.total ?? data.length
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
            hasClear
            onUpdate={(value) => {
              setGlobalFilter(value)
              if (pagination?.onUpdate) pagination.onUpdate(1, pageSize)
              else setInternalPage(1)
            }}
          />
        </div>
      ) : null}
      <div className="m8-resource-table-shell" aria-busy={loading}>
        <Table
          table={table}
          size="m"
          className={className}
          emptyContent={loading ? loadingContent : emptyContent}
          onRowClick={
            onRowActivate
              ? (row, event) => {
                  if (!isInteractiveTarget(event.target)) activate(row.original)
                }
              : undefined
          }
          rowAttributes={
            onRowActivate
              ? (row) => ({
                  tabIndex: 0,
                  onKeyDown: (event) => {
                    if (isInteractiveTarget(event.target)) return
                    if (event.key === 'Enter' || event.key === ' ') {
                      event.preventDefault()
                      activate(row.original)
                    }
                  },
                })
              : undefined
          }
        />
      </div>
      {pagination && data.length > 0 ? (
        <div className="m8-resource-table-pagination">
          <Pagination
            page={effectivePage}
            pageSize={pageSize}
            total={filteredRowCount}
            pageSizeOptions={pagination.pageSizeOptions ?? [10, 20, 50, 100]}
            onUpdate={(nextPage, nextPageSize) => {
              if (pagination.onUpdate) pagination.onUpdate(nextPage, nextPageSize)
              else {
                setInternalPage(nextPage)
                setInternalPageSize(nextPageSize)
              }
            }}
            showPages={!serverPagination}
            showInput={false}
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
  columnVisibility: VisibilityState
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
  columnVisibility: VisibilityState,
  columnOrder: string[],
) {
  if (!storageKey || typeof window === 'undefined') return
  window.localStorage.setItem(storageKey, JSON.stringify({columnVisibility, columnOrder}))
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
