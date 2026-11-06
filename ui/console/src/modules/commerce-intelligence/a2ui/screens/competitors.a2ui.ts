import type {A2UIScreen} from '../types'

export const competitorsA2UIScreen: A2UIScreen = {
  id: 'commerce-intelligence.competitors',
  route: '/commerce-intelligence/competitors',
  title: 'Конкуренты',
  surface: {
    id: 'competitors.surface',
    component: 'AppShell',
    children: [
      {id: 'competitors.header', component: 'PageHeader'},
      {id: 'competitors.kpis', component: 'KpiCard', dataPath: 'kpis'},
      {id: 'competitors.filters', component: 'FilterBar', dataPath: 'filters'},
      {id: 'competitors.trend', component: 'ChartCard', dataPath: 'trend'},
      {id: 'competitors.heatmap', component: 'Heatmap', dataPath: 'heatmap'},
      {id: 'competitors.ladder', component: 'ChartCard', dataPath: 'ladder'},
      {id: 'competitors.table', component: 'DataTable', dataPath: 'matches'},
      {id: 'competitors.drawer', component: 'DetailDrawer', dataPath: 'selectedMatch'},
    ],
  },
  dataModel: {filters: {}, selectedMatch: null, drawerOpen: false},
  actions: [
    {name: 'openCompetitorMatch', type: 'local'},
    {name: 'applyFilters', type: 'local'},
    {name: 'exportTable', type: 'event'},
  ],
}
