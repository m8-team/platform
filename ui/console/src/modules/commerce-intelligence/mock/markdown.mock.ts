import type {Kpi, MarkdownCandidate} from './types'

export const markdownKpis: Kpi[] = [
  {title: 'Товары в окне разметки', value: '1 248', subtitle: 'активных кандидатов', sparkline: [920, 1010, 1120, 1180, 1220, 1248]},
  {title: 'Рекомендуемый объем разметки', value: '$2.83M', delta: '+12.4%', deltaTone: 'neutral', subtitle: 'планируемый бюджет', sparkline: [1.9, 2.1, 2.3, 2.55, 2.7, 2.83]},
  {title: 'SKU с сезонным риском', value: '412', delta: '+7.1%', deltaTone: 'negative', subtitle: 'до конца сезона', sparkline: [290, 318, 340, 366, 390, 412]},
  {title: 'Медленно продающиеся SKU', value: '786', subtitle: 'ниже целевой скорости', sparkline: [720, 744, 760, 772, 780, 786]},
  {title: 'Ожидаемый рост sell-through', value: '+18.7%', delta: '+2.1 п.п.', deltaTone: 'positive', subtitle: 'после уценки', sparkline: [10, 12, 14, 16, 17, 18.7]},
  {title: 'Ожидают согласования разметки', value: '23', subtitle: 'в очереди решений', sparkline: [18, 20, 23, 21, 22, 23]},
]

export const markdownCandidates: MarkdownCandidate[] = [
  {sku: 'SKU70022345', product: 'Winter Jacket', currentPrice: '$189', markdown: '30%', recommendedPrice: '$132', reason: 'Сезонный риск', seasonEndStock: '3 420', sellThroughLift: '+24%', marginImpact: '-$18K', confidence: '86%', status: 'На проверке'},
  {sku: 'SKU20013459', product: "Men's Running Tee", currentPrice: '$32', markdown: '20%', recommendedPrice: '$26', reason: 'Медленные продажи', seasonEndStock: '4 900', sellThroughLift: '+18%', marginImpact: '+$6K', confidence: '89%', status: 'На проверке'},
  {sku: 'SKU10024567', product: 'Air Purifier Pro 3000', currentPrice: '$229', markdown: '15%', recommendedPrice: '$195', reason: 'Избыточный запас', seasonEndStock: '2 120', sellThroughLift: '+15%', marginImpact: '+$21K', confidence: '91%', status: 'Согласовано'},
  {sku: 'SKU30007890', product: 'Blender X200', currentPrice: '$89', markdown: '10%', recommendedPrice: '$80', reason: 'Медленные продажи', seasonEndStock: '1 060', sellThroughLift: '+11%', marginImpact: '-$8K', confidence: '84%', status: 'Запланировано'},
  {sku: 'SKU40055678', product: 'Yoga Mat Premium', currentPrice: '$54', markdown: '8%', recommendedPrice: '$50', reason: 'Избыточный запас', seasonEndStock: '920', sellThroughLift: '+7%', marginImpact: '+$4K', confidence: '78%', status: 'Черновик'},
]

export const markdownWindows = [
  {name: 'Летний спорт', period: '15 июн — 10 июл', sku: '284 SKU', status: 'Запланировано'},
  {name: 'Дом: окончание сезона', period: '20 июн — 18 июл', sku: '412 SKU', status: 'На проверке'},
  {name: 'Одежда: капсула весна', period: '1 июл — 22 июл', sku: '552 SKU', status: 'Черновик'},
]

export const markdownGuardrails = [
  {name: 'Минимальная маржа', value: '20%', status: 'Пройдено' as const},
  {name: 'Максимальная уценка', value: '40%', status: 'Пройдено' as const},
  {name: 'Согласование требуется', value: '>30%', status: 'Предупреждение' as const},
  {name: 'Ограничения брендов', value: '5', status: 'Пройдено' as const},
]
