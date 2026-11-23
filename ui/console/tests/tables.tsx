import {useState} from 'react'
import {createRoot} from 'react-dom/client'
import {ThemeProvider} from '@gravity-ui/uikit'
import type {SortingState} from '@tanstack/react-table'
import {ResourceTable} from '../src/components/ResourceTable'
import type {TableColumn} from '../src/components/tableModel'
import {DataTable} from '../src/modules/commerce-intelligence/components/DataTable'

const rows = Array.from({length: 18}, (_, i) => ({id: String(18 - i).padStart(2, '0')}))
const columns: TableColumn<typeof rows[number]>[] = [{accessorKey: 'id', header: 'ID', sortDescFirst: false}]
export function Fixture() {
  const [page, setPage] = useState(1)
  const [sorting, setSorting] = useState<SortingState>([])
  const sorted = sorting.length ? [...rows].sort((a, b) => a.id.localeCompare(b.id) * (sorting[0].desc ? -1 : 1)) : rows
  return <ThemeProvider><section id="client">
    <DataTable data={rows} columns={columns} getRowId={row => row.id} enableRowSelection />
  </section><section id="server">
    <ResourceTable data={sorted.slice((page - 1) * 8, page * 8)} columns={columns} getRowId={row => row.id}
      selectable sortable selectPageLabel="Select page" selectRowLabel="Select row" loadingContent="Loading" emptyContent="Empty"
      sorting={{mode: 'server', value: sorting, onUpdate: setSorting}}
      pagination={{mode: 'server', page, pageSize: 8, total: rows.length, onUpdate: setPage}} />
  </section></ThemeProvider>
}

createRoot(document.getElementById('root')!).render(<Fixture />)
const frame = () => new Promise<void>(resolve => requestAnimationFrame(() => resolve()))
async function check(section: HTMLElement) {
  const body = () => Array.from(section.querySelectorAll<HTMLTableRowElement>('tbody tr'))
  const ids = () => body().map(row => row.cells[1].textContent).join(',')
  const click = async (control: HTMLElement) => { control.click(); await frame(); await frame() }
  const checkbox = (element: Element) => element.querySelector<HTMLInputElement>('input[type=checkbox]')!
  const assert = (ok: boolean, message: string) => { if (!ok) throw Error(section.id + ' ' + message + ': ' + ids()) }
  const sort = section.querySelector<HTMLButtonElement>('thead button')!
  await click(sort)
  assert(ids() === '01,02,03,04,05,06,07,08', 'ascending')
  await click(checkbox(body()[0]))
  assert(checkbox(body()[0]).checked && checkbox(section.querySelector('thead')!).indeterminate, 'single selection')
  await click(sort)
  assert(ids() === '18,17,16,15,14,13,12,11' && !checkbox(body()[0]).checked, 'descending with stable selection')
  const pagination = section.querySelector('.ci-table__pagination, .m8-resource-table-pagination')!
  const buttons = pagination.querySelectorAll('button')
  const next = pagination.querySelector<HTMLButtonElement>('[data-qa="pagination-button-next"]') ?? buttons[buttons.length - 1]
  await click(next)
  assert(ids() === '10,09,08,07,06,05,04,03', 'second page')
  await click(checkbox(section.querySelector('thead')!))
  assert(body().every(row => checkbox(row).checked), 'page selection')
  await click(next)
  assert(ids() === '02,01' && !checkbox(body()[0]).checked && checkbox(body()[1]).checked, 'selection across pages')
  assert(next.disabled, 'last page boundary')
}
async function run() {
  while (!document.querySelector("#server tbody tr")) await frame()
  await check(document.getElementById('client')!)
  await check(document.getElementById('server')!)
}
run().then(() => { document.getElementById('result')!.textContent = 'PASS' })
  .catch(error => { document.getElementById('result')!.textContent = String(error); throw error })
