import type {A2UIScreen} from '../types'

export const markdownA2UIScreen: A2UIScreen = {
  id: 'commerce-intelligence.markdown',
  route: '/commerce-intelligence/markdown',
  title: 'Центр разметки',
  surface: {
    id: 'markdown.surface',
    component: 'AppShell',
    children: [
      {id: 'markdown.header', component: 'PageHeader'},
      {id: 'markdown.kpis', component: 'KpiCard', dataPath: 'kpis'},
      {id: 'markdown.filters', component: 'FilterBar', dataPath: 'filters'},
      {id: 'markdown.table', component: 'DataTable', dataPath: 'candidates'},
      {id: 'markdown.drawer', component: 'DetailDrawer', dataPath: 'selectedCandidate'},
      {id: 'markdown.guardrails', component: 'GuardrailList', dataPath: 'guardrails'},
    ],
  },
  dataModel: {filters: {}, selectedCandidate: null, drawerOpen: false},
  actions: [
    {name: 'applyFilters', type: 'local'},
    {name: 'approvePriceAction', type: 'event'},
    {name: 'rejectPriceAction', type: 'event'},
    {name: 'exportTable', type: 'event'},
  ],
}
