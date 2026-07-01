import type {A2UIScreen} from '../types'

export const overviewA2UIScreen: A2UIScreen = {
  id: 'commerce-intelligence.overview',
  route: '/commerce-intelligence/overview',
  title: 'Обзор',
  surface: {
    id: 'overview.surface',
    component: 'AppShell',
    children: [
      {id: 'overview.header', component: 'PageHeader'},
      {id: 'overview.filters', component: 'FilterBar', dataPath: 'filters'},
      {id: 'overview.kpis', component: 'KpiCard', dataPath: 'kpis'},
      {id: 'overview.forecast', component: 'ChartCard', dataPath: 'forecast'},
      {id: 'overview.heatmap', component: 'Heatmap', dataPath: 'heatmap'},
      {id: 'overview.insights', component: 'InsightPanel', dataPath: 'insights'},
      {id: 'overview.recommendations', component: 'DataTable', dataPath: 'recommendations'},
      {id: 'overview.approvals', component: 'ApprovalQueue', dataPath: 'approvalSummary'},
    ],
  },
  dataModel: {filters: {}, selectedRecommendation: null, drawerOpen: false},
  actions: [
    {name: 'applyFilters', type: 'local'},
    {name: 'resetFilters', type: 'local'},
    {name: 'saveView', type: 'event'},
    {name: 'openApprovalQueue', type: 'event'},
  ],
}
