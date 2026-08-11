import type {A2UIScreen} from '../types'

export const priceActionsA2UIScreen: A2UIScreen = {
  id: 'commerce-intelligence.price-actions',
  route: '/commerce-intelligence/price-actions',
  title: 'Ценовые действия',
  surface: {
    id: 'price-actions.surface',
    component: 'AppShell',
    children: [
      {id: 'price-actions.header', component: 'PageHeader'},
      {id: 'price-actions.kpis', component: 'KpiCard', dataPath: 'kpis'},
      {id: 'price-actions.filters', component: 'FilterBar', dataPath: 'filters'},
      {id: 'price-actions.toolbar', component: 'ActionToolbar'},
      {id: 'price-actions.table', component: 'DataTable', dataPath: 'actions'},
      {id: 'price-actions.drawer', component: 'DetailDrawer', dataPath: 'selectedAction'},
      {id: 'price-actions.applied', component: 'ChartCard', dataPath: 'appliedVsActual'},
    ],
  },
  dataModel: {filters: {}, activeTab: 'Черновики', selectedRows: [], selectedAction: null, drawerOpen: false},
  actions: [
    {name: 'approvePriceAction', type: 'event'},
    {name: 'rejectPriceAction', type: 'event'},
    {name: 'schedulePriceAction', type: 'event'},
    {name: 'openPriceAction', type: 'local'},
    {name: 'exportTable', type: 'event'},
    {name: 'customizeColumns', type: 'local'},
  ],
}
