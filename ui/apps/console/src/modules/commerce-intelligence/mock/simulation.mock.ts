import type {Kpi, SimulationGuardrail, SimulationPlannerRow} from './types'

export const simulationKpis: Kpi[] = [
  {title: 'Выручка сценария', value: '$9.42M', delta: '+6.8%', deltaTone: 'positive', subtitle: 'к базовому сценарию', sparkline: [8.5, 8.7, 8.9, 9.1, 9.3, 9.42]},
  {title: 'Маржа сценария', value: '$2.31M', delta: '+0.6 п.п.', deltaTone: 'positive', subtitle: 'валовая маржа', sparkline: [2.0, 2.1, 2.18, 2.22, 2.28, 2.31]},
  {title: 'Сокращение дней запасов', value: '-14 дней', subtitle: 'после разметки', sparkline: [0, -3, -6, -9, -12, -14]},
  {title: 'Рост sell-through', value: '+9.6%', delta: '+2.4 п.п.', deltaTone: 'positive', subtitle: 'по сезонным SKU', sparkline: [3, 4.4, 5.8, 7.1, 8.4, 9.6]},
  {title: 'Нарушения guardrails', value: '1', delta: '-2', deltaTone: 'positive', subtitle: 'требует проверки', sparkline: [4, 3, 3, 2, 1, 1]},
  {title: 'Рекомендуемый бюджет разметки', value: '$1.28M', subtitle: 'сценарий A', sparkline: [0.8, 0.92, 1.05, 1.16, 1.22, 1.28]},
]

export const scenarioComparison = [
  {metric: 'Выручка', base: '$8.82M', scenarioA: '$9.42M', scenarioB: '$9.11M'},
  {metric: 'Валовая маржа', base: '$2.18M', scenarioA: '$2.31M', scenarioB: '$2.06M'},
  {metric: 'Валовая маржа %', base: '24.7%', scenarioA: '25.3%', scenarioB: '22.6%'},
  {metric: 'Продано единиц', base: '418K', scenarioA: '452K', scenarioB: '486K'},
  {metric: 'Sell-through %', base: '54%', scenarioA: '63%', scenarioB: '71%'},
  {metric: 'Дней запасов', base: '61', scenarioA: '47', scenarioB: '39'},
  {metric: 'Бюджет разметки', base: '$0.72M', scenarioA: '$1.28M', scenarioB: '$1.96M'},
]

export const priceImpact = [
  {markdown: '0%', revenue: 8.82, margin: 2.18, units: 418},
  {markdown: '5%', revenue: 9.01, margin: 2.24, units: 432},
  {markdown: '10%', revenue: 9.28, margin: 2.30, units: 446},
  {markdown: '15%', revenue: 9.42, margin: 2.31, units: 452},
  {markdown: '20%', revenue: 9.36, margin: 2.22, units: 468},
  {markdown: '25%', revenue: 9.11, margin: 2.06, units: 486},
]

export const simulationGuardrails: SimulationGuardrail[] = [
  {rule: 'Мин. валовая маржа %', limit: '24%', scenarioA: '25.3%', scenarioB: '22.6%', status: 'Предупреждение'},
  {rule: 'Макс. уценка %', limit: '40%', scenarioA: '30%', scenarioB: '38%', status: 'Пройдено'},
  {rule: 'Индекс цен конкурентов', limit: '95–105', scenarioA: '98.4', scenarioB: '93.1', status: 'Ошибка'},
  {rule: 'Частота изменения цены', limit: '1 раз / 7 дней', scenarioA: 'Пройдено', scenarioB: 'Пройдено', status: 'Пройдено'},
  {rule: 'Ручное согласование', limit: '>30%', scenarioA: 'Не требуется', scenarioB: 'Требуется', status: 'Предупреждение'},
  {rule: 'Макс. дней запасов', limit: '55', scenarioA: '47', scenarioB: '39', status: 'Пройдено'},
]

export const markdownPlanner: SimulationPlannerRow[] = [
  {sku: 'SKU70022345', currentPrice: '$189', markdown: '30%', sellThroughLift: '+24%', marginImpact: '-$18K', seasonEndStock: '3 420', confidence: '86%', status: 'На проверке'},
  {sku: 'SKU20013459', currentPrice: '$32', markdown: '18%', sellThroughLift: '+16%', marginImpact: '+$6K', seasonEndStock: '4 900', confidence: '89%', status: 'Черновик'},
  {sku: 'SKU10024567', currentPrice: '$229', markdown: '12%', sellThroughLift: '+13%', marginImpact: '+$21K', seasonEndStock: '2 120', confidence: '91%', status: 'Согласовано'},
]

export const scenarioSummary = [
  {metric: 'Выручка', base: '$8.82M', scenarioA: '$9.42M', scenarioB: '$9.11M', deltaA: '+6.8%', deltaB: '+3.3%'},
  {metric: 'Маржа', base: '$2.18M', scenarioA: '$2.31M', scenarioB: '$2.06M', deltaA: '+6.0%', deltaB: '-5.5%'},
  {metric: 'Дней запасов', base: '61', scenarioA: '47', scenarioB: '39', deltaA: '-14', deltaB: '-22'},
  {metric: 'Нарушения правил', base: '3', scenarioA: '1', scenarioB: '3', deltaA: '-2', deltaB: '0'},
]
