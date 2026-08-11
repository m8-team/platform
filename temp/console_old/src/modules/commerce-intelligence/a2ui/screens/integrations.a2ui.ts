import type {A2UIScreen} from '../types'

export const integrationsA2UIScreen: A2UIScreen = {
  id: 'commerce-intelligence.integrations',
  route: '/commerce-intelligence/integrations',
  title: 'Интеграции',
  surface: {
    id: 'integrations.surface',
    component: 'AppShell',
    children: [
      {id: 'integrations.header', component: 'PageHeader'},
      {id: 'integrations.providers', component: 'DataTable', dataPath: 'integrations'},
    ],
  },
  dataModel: {selectedProvider: null},
  actions: [
    {name: 'configureIntegration', type: 'event'},
    {name: 'testConnection', type: 'event'},
    {name: 'runSync', type: 'event'},
    {name: 'openLogs', type: 'event'},
  ],
}
