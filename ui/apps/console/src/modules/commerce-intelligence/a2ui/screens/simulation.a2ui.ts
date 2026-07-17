import type {A2UIScreen} from '../types'

export const simulationA2UIScreen: A2UIScreen = {
  id: 'commerce-intelligence.simulation',
  route: '/commerce-intelligence/simulation',
  title: 'Симуляции',
  surface: {
    id: 'simulation.surface',
    component: 'AppShell',
    children: [
      {id: 'simulation.header', component: 'PageHeader'},
      {id: 'simulation.kpis', component: 'KpiCard', dataPath: 'kpis'},
      {id: 'simulation.filters', component: 'FilterBar', dataPath: 'filters'},
      {id: 'simulation.builder', component: 'ScenarioBuilder', dataPath: 'scenarioComparison'},
      {id: 'simulation.impact', component: 'ChartCard', dataPath: 'priceImpact'},
      {id: 'simulation.guardrails', component: 'GuardrailList', dataPath: 'guardrails'},
      {id: 'simulation.controls', component: 'WhatIfControls', dataPath: 'scenarioSettings'},
      {id: 'simulation.planner', component: 'DataTable', dataPath: 'planner'},
    ],
  },
  dataModel: {filters: {}, scenarioSettings: {markdown: 15, priceIndex: 98, stockDays: 47, risk: 'Средняя'}},
  actions: [
    {name: 'runSimulation', type: 'event'},
    {name: 'saveScenario', type: 'event'},
    {name: 'submitScenarioForApproval', type: 'event'},
  ],
}
