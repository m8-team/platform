import {Checkbox} from '@gravity-ui/uikit'
import {
  columnOrderingFeature, columnSizingFeature, columnVisibilityFeature, columnFilteringFeature,
  createFilteredRowModel, createPaginatedRowModel, createSortedRowModel, filterFns,
  globalFilteringFeature, rowPaginationFeature, rowSelectionFeature, rowSortingFeature,
  sortFns, tableFeatures,
} from '@tanstack/react-table'
import type {ColumnDef, RowData} from '@tanstack/react-table'

export const features = tableFeatures({
  columnOrderingFeature, columnSizingFeature, columnVisibilityFeature, columnFilteringFeature,
  globalFilteringFeature, rowPaginationFeature, rowSelectionFeature, rowSortingFeature,
  filteredRowModel: createFilteredRowModel(), paginatedRowModel: createPaginatedRowModel(),
  sortedRowModel: createSortedRowModel(), filterFns, sortFns,
})

export type TableColumn<T extends RowData> = ColumnDef<typeof features, T>

export function selectionColumn<T extends RowData>(id: string, pageLabel: string, rowLabel: string): TableColumn<T> {
  return {
    id, size: 42, enableSorting: false, enableHiding: false,
    header: ({table}) => <Checkbox aria-label={pageLabel} checked={table.getIsAllPageRowsSelected()}
      indeterminate={table.getIsSomePageRowsSelected()} onUpdate={checked => table.toggleAllPageRowsSelected(checked)} />,
    cell: ({row}) => <span onClick={event => event.stopPropagation()}>
      <Checkbox aria-label={rowLabel} checked={row.getIsSelected()} onUpdate={checked => row.toggleSelected(checked)} />
    </span>,
  }
}
