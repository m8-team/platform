import type {CompetitorMatch, HeatmapCell, Kpi} from './types'

export const competitorKpis: Kpi[] = [
  {title: 'Отслеживаемые конкуренты', value: '18', subtitle: 'активных источников', sparkline: [12, 13, 14, 16, 18, 18]},
  {title: 'Сопоставлено товаров', value: '42 318', delta: '+2 140', deltaTone: 'positive', subtitle: 'за 7 дней', sparkline: [36100, 37420, 38900, 40120, 41400, 42318]},
  {title: 'Уверенность сопоставления', value: '86.7%', delta: '+1.2 п.п.', deltaTone: 'positive', subtitle: 'средняя по рынку', sparkline: [82, 83, 84, 85, 86, 86.7]},
  {title: 'Изменений цен сегодня', value: '1 243', subtitle: 'по всем конкурентам', sparkline: [420, 760, 880, 1030, 1180, 1243]},
  {title: 'События отсутствия товара', value: '576', delta: '+9.4%', deltaTone: 'negative', subtitle: 'за сутки', sparkline: [420, 450, 470, 520, 544, 576]},
  {title: 'Свежесть обхода', value: '96.3%', delta: '+0.8 п.п.', deltaTone: 'positive', subtitle: 'за последние 6 часов', sparkline: [92, 93, 94, 95, 96, 96.3]},
]

export const priceIndexTrend = [
  {date: '12 мая', M8: 100, Amazon: 98, Walmart: 101, Target: 99, 'Best Buy': 102, Newegg: 97},
  {date: '18 мая', M8: 101, Amazon: 99, Walmart: 100, Target: 98, 'Best Buy': 101, Newegg: 96},
  {date: '24 мая', M8: 99, Amazon: 98, Walmart: 101, Target: 100, 'Best Buy': 103, Newegg: 97},
  {date: '30 мая', M8: 98, Amazon: 97, Walmart: 100, Target: 99, 'Best Buy': 102, Newegg: 96},
  {date: '5 июн', M8: 99, Amazon: 98, Walmart: 101, Target: 98, 'Best Buy': 101, Newegg: 97},
  {date: '10 июн', M8: 98.6, Amazon: 97.4, Walmart: 100.6, Target: 98.8, 'Best Buy': 101.2, Newegg: 96.8},
]

export const competitorHeatmap: HeatmapCell[] = ['Amazon', 'Walmart', 'Target', 'Best Buy', 'Newegg', 'eBay'].map((row, index) => ({
  row,
  values: ['Электроника', 'Дом', 'Одежда', 'Красота', 'Спорт', 'Игрушки', 'Продукты', 'Итого'].map((column, columnIndex) => ({
    column,
    value: 92 + ((index * 7 + columnIndex * 3) % 18),
  })),
}))

export const marketLadder = [
  {name: 'Newegg', price: 229},
  {name: 'M8', price: 239},
  {name: 'Amazon', price: 244},
  {name: 'Best Buy', price: 249},
  {name: 'Walmart', price: 252},
]

export const competitorMatches: CompetitorMatch[] = [
  {id: 'cm-1', sku: 'SKU10024567', competitor: 'Amazon', competitorProduct: 'Air Purifier Pro 3000 White', ourPrice: '$229', competitorPrice: '$219', delivery: '2 дня', availability: 'В наличии', seller: 'Amazon', matchConfidence: '94%', lastSeen: '12 мин назад', differencePct: '+4.6%', alert: 'Выше рынка'},
  {id: 'cm-2', sku: 'SKU50011234', competitor: 'Best Buy', competitorProduct: 'Wireless Headphones Black', ourPrice: '$149', competitorPrice: '$159', delivery: '1 день', availability: 'В наличии', seller: 'Best Buy', matchConfidence: '91%', lastSeen: '18 мин назад', differencePct: '-6.3%', alert: 'Ниже рынка'},
  {id: 'cm-3', sku: 'SKU30007890', competitor: 'Walmart', competitorProduct: 'Blender X200 Kitchen Set', ourPrice: '$89', competitorPrice: '$84', delivery: '3 дня', availability: 'Мало', seller: 'Walmart', matchConfidence: '87%', lastSeen: '34 мин назад', differencePct: '+6.0%', alert: 'Давление'},
  {id: 'cm-4', sku: 'SKU40055678', competitor: 'Target', competitorProduct: 'Premium Yoga Mat 6mm', ourPrice: '$54', competitorPrice: '$52', delivery: '2 дня', availability: 'В наличии', seller: 'Target', matchConfidence: '89%', lastSeen: '41 мин назад', differencePct: '+3.8%', alert: 'Норма'},
  {id: 'cm-5', sku: 'SKU70022345', competitor: 'eBay', competitorProduct: 'Winter Jacket Insulated', ourPrice: '$189', competitorPrice: '$169', delivery: '5 дней', availability: 'В наличии', seller: 'Marketplace', matchConfidence: '82%', lastSeen: '1 ч назад', differencePct: '+11.8%', alert: 'Высокий разрыв'},
]
