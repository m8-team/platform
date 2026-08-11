import type {A2UIScreen} from '../types'

export const approvalsA2UIScreen: A2UIScreen = {
  id: 'commerce-intelligence.approvals',
  route: '/commerce-intelligence/approvals',
  title: 'Согласования',
  surface: {
    id: 'approvals.surface',
    component: 'AppShell',
    children: [
      {id: 'approvals.header', component: 'PageHeader'},
      {id: 'approvals.kpis', component: 'KpiCard', dataPath: 'kpis'},
      {id: 'approvals.toolbar', component: 'ActionToolbar'},
      {id: 'approvals.table', component: 'DataTable', dataPath: 'approvals'},
    ],
  },
  dataModel: {selectedRows: [], selectedApproval: null},
  actions: [
    {name: 'approvePriceAction', type: 'event'},
    {name: 'rejectPriceAction', type: 'event'},
    {name: 'schedulePriceAction', type: 'event'},
    {name: 'openApprovalQueue', type: 'local'},
  ],
}
