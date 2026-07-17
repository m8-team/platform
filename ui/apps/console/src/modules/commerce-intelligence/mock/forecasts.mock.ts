import {overviewForecast} from './overview.mock'
import type {ForecastRiskSku, HeatmapCell, Kpi} from './types'

export const forecastKpis: Kpi[] = [
  {title: 'Точность прогноза WAPE', value: '18.6%', delta: '-1.1 п.п.', deltaTone: 'positive', subtitle: 'лучше прошлой версии', sparkline: [22, 21, 20, 19.7, 19.1, 18.6]},
  {title: 'Прогнозируемые продажи 30д', value: '2.84M', delta: '+5.8%', deltaTone: 'positive', subtitle: 'единиц', sparkline: [2.3, 2.4, 2.55, 2.68, 2.76, 2.84]},
  {title: 'Экспозиция избыточных запасов', value: '$2.17M', delta: '+8.1%', deltaTone: 'negative', subtitle: 'в зоне риска', sparkline: [1.7, 1.8, 1.92, 2.04, 2.11, 2.17]},
  {title: 'Экспозиция отсутствия товара', value: '$1.42M', delta: '+3.4%', deltaTone: 'negative', subtitle: 'потенциальный дефицит', sparkline: [1.1, 1.18, 1.22, 1.31, 1.37, 1.42]},
  {title: 'Риск срока поставки', value: 'Высокий', subtitle: 'поставщики одежды и дома', sparkline: [42, 46, 48, 51, 57, 62]},
  {title: 'Сезонная волатильность', value: '0.48', subtitle: 'портфельный индекс', sparkline: [0.38, 0.39, 0.41, 0.44, 0.46, 0.48]},
]

export const forecastVsActual = overviewForecast

export const categoryAccuracy = [
  {category: 'Электроника', wape: 14.2},
  {category: 'Дом', wape: 21.4},
  {category: 'Одежда', wape: 24.8},
  {category: 'Красота', wape: 16.9},
  {category: 'Спорт', wape: 18.1},
  {category: 'Игрушки', wape: 20.2},
]

export const inventoryRiskMatrix: HeatmapCell[] = [
  {row: 'Низкий запас', values: [{column: 'Низкий спрос', value: 26}, {column: 'Средний спрос', value: 58}, {column: 'Высокий спрос', value: 91}]},
  {row: 'Средний запас', values: [{column: 'Низкий спрос', value: 41}, {column: 'Средний спрос', value: 55}, {column: 'Высокий спрос', value: 77}]},
  {row: 'Высокий запас', values: [{column: 'Низкий спрос', value: 94}, {column: 'Средний спрос', value: 68}, {column: 'Высокий спрос', value: 48}]},
]

export const forecastInsights = [
  {title: 'Электроника +12%', text: 'Ожидается рост спроса из-за сезонного роста, поисковых трендов и промо.', tone: 'success' as const},
  {title: 'Дом -6%', text: 'Ожидается снижение из-за окончания весеннего сезона и роста запасов.', tone: 'warning' as const},
  {title: 'Одежда +5%', text: 'Рост связан с летними коллекциями и увеличением web-трафика.', tone: 'info' as const},
]

export const atRiskSkus: ForecastRiskSku[] = [
  {sku: 'SKU10024567', product: 'Air Purifier Pro 3000', category: 'Дом', forecast30d: '1 860', stock: '8 240', coverageDays: 64, overstockScore: 88, outOfStockScore: 12, leadTime: '21 день', confidence: '91%', suggestedAction: 'Скорректировать заказ / разметка'},
  {sku: 'SKU50011234', product: 'Wireless Headphones', category: 'Электроника', forecast30d: '4 920', stock: '2 110', coverageDays: 17, overstockScore: 18, outOfStockScore: 82, leadTime: '14 дней', confidence: '86%', suggestedAction: 'Пополнить 6 000 ед.'},
  {sku: 'SKU70022345', product: 'Winter Jacket', category: 'Одежда', forecast30d: '980', stock: '6 320', coverageDays: 86, overstockScore: 94, outOfStockScore: 6, leadTime: '30 дней', confidence: '89%', suggestedAction: 'Снизить заказ на 25%'},
  {sku: 'SKU40055678', product: 'Yoga Mat Premium', category: 'Спорт', forecast30d: '3 880', stock: '5 640', coverageDays: 31, overstockScore: 42, outOfStockScore: 35, leadTime: '12 дней', confidence: '78%', suggestedAction: 'Снизить заказ на 15%'},
  {sku: 'SKU20013459', product: "Men's Running Tee", category: 'Одежда', forecast30d: '7 200', stock: '14 120', coverageDays: 42, overstockScore: 63, outOfStockScore: 21, leadTime: '18 дней', confidence: '84%', suggestedAction: 'Ускорить пополнение'},
]

export const modelDrivers = [
  {name: 'Цена', value: 24},
  {name: 'Сезонность', value: 21},
  {name: 'Давление конкурентов', value: 18},
  {name: 'Промо', value: 15},
  {name: 'Уровень запасов', value: 12},
  {name: 'Тренды поиска и web', value: 10},
]
