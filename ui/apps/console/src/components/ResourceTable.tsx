import {useCallback, useEffect, useMemo, useRef, useState} from 'react'
import type {ReactNode} from 'react'
import {Gear} from '@gravity-ui/icons'
import {
  Button,
  Checkbox,
  Icon,
  Pagination,
  Table,
  TableColumnSetup,
  TextInput,
} from '@gravity-ui/uikit'
import type {
  TableColumnConfig,
  TableColumnSetupItem,
  TableDataItem,
} from '@gravity-ui/uikit'

const selectionColumnId = '_selection'
const settingsColumnId = '_settings'

type ColumnVisibility = Record<string, boolean>

export interface ResourceTableSettings {
  storageKey?: string
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

export interface ResourceTableSortItem {
  column: string
  order: 'asc' | 'desc'
}

export type ResourceTableSortingState = ResourceTableSortItem[]

export interface ResourceTableSorting {
  mode: 'server'
  value: ResourceTableSortingState
  onUpdate: (value: ResourceTableSortingState) => void
}

export interface ResourceTableSelectionActions<TData> {
  selectedItems: TData[]
  clearSelection: () => void
}

export type ResourceTableColumn<TData> = TableColumnConfig<TData>

export interface ResourceTableProps<TData extends TableDataItem> {
  data: TData[]
  columns: ResourceTableColumn<TData>[]
  getRowId: (item: TData, index: number) => string
  loading?: boolean
  loadingContent: string
  emptyContent: string
  className?: string
  onRowActivate?: (item: TData) => void
  getRowClassNames?: (item: TData) => string[]
  selectable?: boolean
  onSelectedRowsChange?: (items: TData[]) => void
  settings?: ResourceTableSettings
  renderSelectionActions?: (context: ResourceTableSelectionActions<TData>) => ReactNode
  pagination?: ResourceTablePagination
  filtering?: ResourceTableFiltering
  sortable?: boolean
  sorting?: ResourceTableSorting
}

export function ResourceTable<TData extends TableDataItem>({
  data,
  columns,
  getRowId,
  loading = false,
  loadingContent,
  emptyContent,
  className,
  onRowActivate,
  getRowClassNames,
  selectable = false,
  onSelectedRowsChange,
  settings,
  renderSelectionActions,
  pagination,
  filtering,
  sortable = false,
  sorting: controlledSorting,
}: ResourceTableProps<TData>) {
  const [selectedIds, setSelectedIds] = useState<string[]>([])
  const [internalPage, setInternalPage] = useState(1)
  const [internalPageSize, setInternalPageSize] = useState(
    pagination?.mode === 'client' ? pagination.defaultPageSize ?? 20 : 20,
  )
  const [internalSorting, setInternalSorting] = useState<ResourceTableSortingState>([])
  const [internalFilter, setInternalFilter] = useState('')
  const initialSettings = useMemo(
    () => readTableSettings(settings?.storageKey),
    [settings?.storageKey],
  )
  const [columnVisibility, setColumnVisibility] = useState<ColumnVisibility>(
    initialSettings.columnVisibility,
  )
  const [columnOrder, setColumnOrder] = useState<string[]>(initialSettings.columnOrder)

  const serverPagination = pagination?.mode === 'server'
  const page = serverPagination ? pagination.page : internalPage
  const pageSize = serverPagination ? pagination.pageSize : internalPageSize
  const serverFiltering = filtering?.mode === 'server'
  const globalFilter = serverFiltering ? filtering.value : internalFilter
  const serverSorting = controlledSorting?.mode === 'server'
  const sorting = controlledSorting?.value ?? internalSorting

  const orderedColumns = useMemo(
    () => orderColumns(columns, columnOrder),
    [columnOrder, columns],
  )
  const visibleColumns = useMemo(
    () => orderedColumns.filter((column) => columnVisibility[column.id] !== false),
    [columnVisibility, orderedColumns],
  )
  const filteredData = useMemo(
    () =>
      filtering && !serverFiltering
        ? filterRows(data, columns, globalFilter)
        : data,
    [columns, data, filtering, globalFilter, serverFiltering],
  )
  const sortedData = useMemo(
    () =>
      sortable && !serverSorting
        ? sortRows(filteredData, columns, sorting)
        : filteredData,
    [columns, filteredData, serverSorting, sortable, sorting],
  )
  const filteredRowCount = serverPagination ? pagination.total : sortedData.length
  const lastPage = Math.max(1, Math.ceil(filteredRowCount / pageSize))
  const effectivePage = Math.min(page, lastPage)
  const displayedData = useMemo(
    () =>
      serverPagination
        ? sortedData
        : sortedData.slice((effectivePage - 1) * pageSize, effectivePage * pageSize),
    [effectivePage, pageSize, serverPagination, sortedData],
  )

  const selectedItems = useMemo(() => {
    const selected = new Set(selectedIds)
    return data.filter((item, index) => selected.has(getRowId(item, index)))
  }, [data, getRowId, selectedIds])
  const lastNotifiedSelectedIds = useRef(selectedIds)

  useEffect(() => {
    if (haveSameIds(lastNotifiedSelectedIds.current, selectedIds)) return

    lastNotifiedSelectedIds.current = selectedIds
    onSelectedRowsChange?.(selectedItems)
  }, [onSelectedRowsChange, selectedIds, selectedItems])

  const updateSorting = useCallback(
    (column: string) => {
      const current = sorting.find((item) => item.column === column)
      const next: ResourceTableSortingState = current
        ? current.order === 'asc'
          ? [{column, order: 'desc'}]
          : []
        : [{column, order: 'asc'}]

      if (serverSorting) controlledSorting.onUpdate(next)
      else {
        setInternalSorting(next)
        setInternalPage(1)
      }
    },
    [controlledSorting, serverSorting, sorting],
  )

  const renderedColumns = useMemo(() => {
    const contentColumns = visibleColumns.map((column) =>
      sortable && column.meta?.sortable !== false
        ? withSortingHeader(column, sorting, updateSorting)
        : column,
    )
    const result: ResourceTableColumn<TData>[] = []

    if (selectable) {
      result.push(
        createSelectionColumn(displayedData, selectedIds, setSelectedIds, getRowId),
      )
    }
    result.push(...contentColumns)

    if (settings) {
      result.push({
        id: settingsColumnId,
        width: 52,
        align: 'center',
        name: () => (
          <TableColumnSetup
            items={createSettingsItems(columns, columnOrder, columnVisibility)}
            sortable
            showStatus
            popupPlacement={['bottom-end', 'bottom', 'top-end', 'top']}
            renderSwitcher={({onClick, onKeyDown}) => (
              <Button
                view="flat-secondary"
                size="m"
                aria-label="Настройки таблицы"
                onClick={onClick}
                onKeyDown={onKeyDown}
              >
                <Icon data={Gear} size={16} />
              </Button>
            )}
            onUpdate={(items) => {
              const nextOrder = items.map((item) => item.id)
              const nextVisibility = Object.fromEntries(
                items.map((item) => [item.id, item.selected !== false]),
              )
              setColumnOrder(nextOrder)
              setColumnVisibility(nextVisibility)
              persistTableSettings(settings.storageKey, nextVisibility, nextOrder)
            }}
          />
        ),
        template: () => null,
        meta: {sortable: false},
      })
    }

    return result
  }, [
    columnOrder,
    columnVisibility,
    columns,
    displayedData,
    getRowId,
    selectable,
    selectedIds,
    settings,
    sortable,
    sorting,
    updateSorting,
    visibleColumns,
  ])

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
        <Table
          data={displayedData}
          columns={renderedColumns}
          width="max"
          className={className}
          emptyMessage={loading ? loadingContent : emptyContent}
          getRowDescriptor={(item, index) => {
            const id = getRowId(item, index)
            return {
              id,
              interactive: Boolean(onRowActivate),
              classNames: [
                ...(getRowClassNames?.(item) ?? []),
                ...(selectedIds.includes(id) ? ['m8-resource-table-row_selected'] : []),
              ],
            }
          }}
          onRowClick={
            onRowActivate
              ? (item, _index, event) => {
                  if (!isInteractiveTarget(event.target)) onRowActivate(item)
                }
              : undefined
          }
        />
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
                setInternalPage(nextPageSize === pageSize ? nextPage : 1)
                setInternalPageSize(nextPageSize)
              }
            }}
            showPages={!serverPagination}
            showInput={false}
            className={
              serverPagination && pagination.disabled
                ? 'm8-resource-table-pagination_disabled'
                : undefined
            }
          />
        </div>
      ) : null}
      {selectedItems.length > 0 && renderSelectionActions ? (
        <div className="m8-selection-actions-footer">
          {renderSelectionActions({
            selectedItems,
            clearSelection: () => setSelectedIds([]),
          })}
        </div>
      ) : null}
    </div>
  )
}

function haveSameIds(left: string[], right: string[]) {
  if (left.length !== right.length) return false

  const rightIds = new Set(right)
  return left.every((id) => rightIds.has(id))
}

function createSelectionColumn<TData extends TableDataItem>(
  displayedData: TData[],
  selectedIds: string[],
  setSelectedIds: (ids: string[]) => void,
  getRowId: (item: TData, index: number) => string,
): ResourceTableColumn<TData> {
  const displayedIds = displayedData.map(getRowId)
  const selected = new Set(selectedIds)
  const selectedOnPage = displayedIds.filter((id) => selected.has(id)).length

  return {
    id: selectionColumnId,
    width: 48,
    align: 'center',
    name: () => (
      <Checkbox
        size="l"
        checked={displayedIds.length > 0 && selectedOnPage === displayedIds.length}
        indeterminate={selectedOnPage > 0 && selectedOnPage < displayedIds.length}
        disabled={displayedIds.length === 0}
        aria-label="Выбрать все строки"
        onUpdate={(checked) => {
          if (checked) {
            setSelectedIds([...new Set([...selectedIds, ...displayedIds])])
          } else {
            const pageIds = new Set(displayedIds)
            setSelectedIds(selectedIds.filter((id) => !pageIds.has(id)))
          }
        }}
      />
    ),
    template: (item, index) => {
      const id = getRowId(item, index)
      return (
        <span onClick={(event) => event.stopPropagation()}>
          <Checkbox
            size="l"
            checked={selected.has(id)}
            aria-label={`Выбрать строку ${id}`}
            onUpdate={(checked) =>
              setSelectedIds(
                checked
                  ? [...new Set([...selectedIds, id])]
                  : selectedIds.filter((selectedId) => selectedId !== id),
              )
            }
          />
        </span>
      )
    },
    meta: {sortable: false},
  }
}

function withSortingHeader<TData>(
  column: ResourceTableColumn<TData>,
  sorting: ResourceTableSortingState,
  onUpdate: (column: string) => void,
): ResourceTableColumn<TData> {
  const current = sorting.find((item) => item.column === column.id)
  const title = resolveColumnTitle(column)

  return {
    ...column,
    name: () => (
      <button
        type="button"
        className="m8-resource-table-sort"
        onClick={(event) => {
          event.stopPropagation()
          onUpdate(column.id)
        }}
      >
        <span>{title}</span>
        <span aria-hidden>{current?.order === 'asc' ? '↑' : current?.order === 'desc' ? '↓' : ''}</span>
      </button>
    ),
  }
}

function filterRows<TData>(
  data: TData[],
  columns: ResourceTableColumn<TData>[],
  rawFilter: string,
) {
  const filter = rawFilter.trim().toLocaleLowerCase()
  if (!filter) return data

  return data.filter((item) =>
    columns.some((column) =>
      String(readColumnValue(item, column.id) ?? '')
        .toLocaleLowerCase()
        .includes(filter),
    ),
  )
}

function sortRows<TData>(
  data: TData[],
  columns: ResourceTableColumn<TData>[],
  sorting: ResourceTableSortingState,
) {
  const selected = sorting[0]
  if (!selected) return data
  const column = columns.find((item) => item.id === selected.column)
  if (!column || column.meta?.sortable === false) return data

  const compare = column.meta?.compare as
    | ((left: TData, right: TData) => number)
    | undefined
  return [...data].sort((left, right) => {
    const result = compare
      ? compare(left, right)
      : compareValues(
          readColumnValue(left, selected.column),
          readColumnValue(right, selected.column),
        )
    return selected.order === 'desc' ? -result : result
  })
}

function compareValues(left: unknown, right: unknown) {
  if (left === right) return 0
  if (left === undefined || left === null) return -1
  if (right === undefined || right === null) return 1
  if (typeof left === 'number' && typeof right === 'number') return left - right
  return String(left).localeCompare(String(right))
}

function readColumnValue<TData>(item: TData, columnId: string) {
  return (item as Record<string, unknown>)[columnId]
}

function createSettingsItems<TData>(
  columns: ResourceTableColumn<TData>[],
  order: string[],
  visibility: ColumnVisibility,
): TableColumnSetupItem[] {
  return orderColumns(columns, order).map((column) => ({
    id: column.id,
    title: resolveColumnTitle(column),
    selected: visibility[column.id] !== false,
  }))
}

function resolveColumnTitle<TData>(column: ResourceTableColumn<TData>): ReactNode {
  if (typeof column.name === 'function') return column.name()
  return column.name ?? column.id
}

function orderColumns<TData>(columns: ResourceTableColumn<TData>[], order: string[]) {
  const byId = new Map(columns.map((column) => [column.id, column]))
  return [
    ...order.map((id) => byId.get(id)).filter((column): column is ResourceTableColumn<TData> => Boolean(column)),
    ...columns.filter((column) => !order.includes(column.id)),
  ]
}

function isInteractiveTarget(target: EventTarget | null) {
  return target instanceof Element && Boolean(target.closest('button, input, a, [role="checkbox"]'))
}

interface StoredTableSettings {
  columnVisibility: ColumnVisibility
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
  columnVisibility: ColumnVisibility,
  columnOrder: string[],
) {
  if (!storageKey || typeof window === 'undefined') return
  try {
    window.localStorage.setItem(storageKey, JSON.stringify({columnVisibility, columnOrder}))
  } catch {
    // Настройки таблицы необязательны и не должны ломать основной интерфейс.
  }
}
