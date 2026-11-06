import type {A2UIScreen} from '../types'

export const forecastsA2UIScreen: A2UIScreen = {
  id: 'commerce-intelligence.forecasts',
  route: '/commerce-intelligence/forecasts',
  title: 'Прогнозы',
  surface: {
    id: 'forecasts.surface',
    component: 'AppShell',
    children: [
      {id: 'forecasts.header', component: 'PageHeader'},
      {id: 'forecasts.kpis', component: 'KpiCard', dataPath: 'kpis'},
      {id: 'forecasts.filters', component: 'FilterBar', dataPath: 'filters'},
      {id: 'forecasts.forecast', component: 'ChartCard', dataPath: 'forecastVsActual'},
      {id: 'forecasts.accuracy', component: 'ChartCard', dataPath: 'categoryAccuracy'},
      {id: 'forecasts.risk', component: 'Heatmap', dataPath: 'inventoryRiskMatrix'},
      {id: 'forecasts.insights', component: 'InsightPanel', dataPath: 'insights'},
      {id: 'forecasts.table', component: 'DataTable', dataPath: 'atRiskSkus'},
    ],
  },
  dataModel: {filters: {}, selectedSku: null},
  actions: [
    {name: 'applyFilters', type: 'local'},
    {name: 'exportTable', type: 'event'},
  ],
}
