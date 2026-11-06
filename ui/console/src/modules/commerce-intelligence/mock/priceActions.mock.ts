import {overviewRecommendations} from './overview.mock'
import type {Kpi, PriceAction} from './types'

export const priceActionKpis: Kpi[] = [
  {title: 'Черновики рекомендаций', value: '1 248', subtitle: 'готовятся к проверке', sparkline: [820, 910, 1020, 1110, 1180, 1248]},
  {title: 'Готово к согласованию', value: '236', delta: '+14', deltaTone: 'positive', subtitle: 'за сегодня', sparkline: [180, 188, 206, 218, 230, 236]},
  {title: 'Запланированные изменения', value: '412', subtitle: 'в ближайшие 7 дней', sparkline: [320, 344, 370, 390, 404, 412]},
  {title: 'Оценка прироста выручки', value: '$2.17M', delta: '+8.3%', deltaTone: 'positive', subtitle: 'после применения', sparkline: [1.4, 1.6, 1.9, 2.0, 2.1, 2.17]},
  {title: 'Оценка влияния на маржу', value: '$742K', delta: '+4.8%', deltaTone: 'positive', subtitle: 'ожидаемая маржа', sparkline: [480, 520, 610, 680, 710, 742]},
  {title: 'Исключения guardrails', value: '23', delta: '-3', deltaTone: 'positive', subtitle: 'нужна ручная проверка', sparkline: [31, 29, 27, 25, 26, 23]},
]

export const priceActions: PriceAction[] = [
  ...overviewRecommendations,
  {id: 'pa-20013459', sku: 'SKU20013459', product: "Men's Running Tee", category: 'Одежда', currentPrice: '$32', recommendedPrice: '$26', deltaPct: '-18.8%', reason: 'Сезонная распродажа', expectedRevenue: '+$48K', expectedMargin: '+$6K', confidence: '89%', guardrailStatus: 'Пройдено', approver: 'Анна Р.', status: 'На проверке', risk: 'Средняя'},
  {id: 'pa-40055678', sku: 'SKU40055678', product: 'Yoga Mat Premium', category: 'Спорт', currentPrice: '$54', recommendedPrice: '$49', deltaPct: '-9.3%', reason: 'Сравнять с конкурентом', expectedRevenue: '+$19K', expectedMargin: '+$4K', confidence: '78%', guardrailStatus: 'Ошибка', approver: 'Дмитрий К.', status: 'Отклонено', risk: 'Высокая'},
  {id: 'pa-60077890', sku: 'SKU60077890', product: 'Stainless Steel Bottle', category: 'Спорт', currentPrice: '$24', recommendedPrice: '$24', deltaPct: '0%', reason: 'Поддерживать — стабильно', expectedRevenue: '+$7K', expectedMargin: '+$2K', confidence: '93%', guardrailStatus: 'Пройдено', approver: 'Мария И.', status: 'Применено', risk: 'Низкая'},
  {id: 'pa-70022345', sku: 'SKU70022345', product: 'Winter Jacket', category: 'Одежда', currentPrice: '$189', recommendedPrice: '$139', deltaPct: '-26.5%', reason: 'Разметка — низкий спрос', expectedRevenue: '+$116K', expectedMargin: '-$18K', confidence: '86%', guardrailStatus: 'Исключение', approver: 'Иван С.', status: 'Согласовано', risk: 'Высокая'},
]

export const appliedVsActual = [
  {name: 'Прирост выручки', expected: 1240, actual: 1185},
  {name: 'Влияние на маржу', expected: 420, actual: 448},
  {name: 'Запросы', expected: 780, actual: 805},
  {name: 'В срок', expected: 92, actual: 89},
]
