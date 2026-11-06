import type {A2UIScreen} from '../types'

export const rulesA2UIScreen: A2UIScreen = {
  id: 'commerce-intelligence.rules',
  route: '/commerce-intelligence/rules',
  title: 'Правила',
  surface: {
    id: 'rules.surface',
    component: 'AppShell',
    children: [
      {id: 'rules.header', component: 'PageHeader'},
      {id: 'rules.groups', component: 'FilterSelect', dataPath: 'groups'},
      {id: 'rules.table', component: 'DataTable', dataPath: 'rules'},
    ],
  },
  dataModel: {activeGroup: 'Активные правила', selectedRule: null},
  actions: [
    {name: 'createRule', type: 'event'},
    {name: 'editRule', type: 'event'},
    {name: 'viewRuleImpact', type: 'event'},
  ],
}
