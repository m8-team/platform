import type {A2UIScreen} from '../types'

export const productsA2UIScreen: A2UIScreen = {
  id: 'commerce-intelligence.products',
  route: '/commerce-intelligence/products',
  title: 'Товары',
  surface: {
    id: 'products.surface',
    component: 'AppShell',
    children: [
      {id: 'products.header', component: 'PageHeader'},
      {id: 'products.kpis', component: 'KpiCard', dataPath: 'kpis'},
      {id: 'products.segments', component: 'FilterSelect', dataPath: 'segments'},
      {id: 'products.filters', component: 'FilterBar', dataPath: 'filters'},
      {id: 'products.table', component: 'DataTable', dataPath: 'products'},
      {id: 'products.drawer', component: 'DetailDrawer', dataPath: 'selectedProduct'},
      {id: 'products.risk', component: 'Heatmap', dataPath: 'riskMap'},
    ],
  },
  dataModel: {filters: {}, selectedSegment: 'Высокая маржа', selectedProduct: null, drawerOpen: false},
  actions: [
    {name: 'selectProduct', type: 'local'},
    {name: 'applyFilters', type: 'local'},
    {name: 'exportTable', type: 'event'},
  ],
}
