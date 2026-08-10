import type {ForecastPoint, HeatmapCell, Kpi, PriceAction} from './types'

export const overviewKpis: Kpi[] = [
  {title: 'Возможность по выручке', value: '$8.42M', delta: '+12.6%', deltaTone: 'positive', subtitle: 'за последние 30 дней', sparkline: [32, 38, 41, 39, 45, 52, 61, 66]},
  {title: 'Возможность по марже', value: '$2.17M', delta: '+8.3%', deltaTone: 'positive', subtitle: 'за последние 30 дней', sparkline: [18, 22, 25, 28, 30, 31, 36, 39]},
  {title: 'Индекс цен vs рынок', value: '98.6', delta: '-1.4%', deltaTone: 'negative', subtitle: 'за последние 30 дней', sparkline: [101, 100, 99, 99, 98, 99, 98, 98]},
  {title: 'Риск избыточных запасов', value: '412 SKU', delta: '+7.1%', deltaTone: 'negative', subtitle: 'за последние 30 дней', sparkline: [240, 260, 291, 315, 330, 360, 389, 412]},
  {title: 'Риск перебоев в поставках', value: '178 SKU', delta: '+5.4%', deltaTone: 'negative', subtitle: 'за последние 30 дней', sparkline: [120, 132, 126, 148, 151, 160, 171, 178]},
  {title: 'Ожидающие согласования', value: '23', delta: 'Посмотреть очередь', deltaTone: 'neutral', subtitle: '18 цен, 3 разметки, 2 новых позиции', sparkline: [18, 21, 20, 24, 23, 22, 25, 23]},
]

export const overviewForecast: ForecastPoint[] = [
  {date: '12 мая', forecast: 940, actual: 910, upper: 1040, lower: 820},
  {date: '16 мая', forecast: 980, actual: 1005, upper: 1100, lower: 850, promo: 1},
  {date: '20 мая', forecast: 1030, actual: 1018, upper: 1160, lower: 900},
  {date: '24 мая', forecast: 1100, actual: 1072, upper: 1230, lower: 960},
  {date: '28 мая', forecast: 1185, actual: 1196, upper: 1310, lower: 1010, promo: 1},
  {date: '1 июн', forecast: 1220, actual: 1244, upper: 1370, lower: 1080},
  {date: '5 июн', forecast: 1190, actual: 1175, upper: 1320, lower: 1040},
  {date: '10 июн', forecast: 1260, actual: 1235, upper: 1410, lower: 1110},
]

export const marketHeatmap: HeatmapCell[] = [
  {row: 'Электроника', values: [{column: 'Вы', value: 104}, {column: 'Конкурент A', value: 99}, {column: 'Конкурент B', value: 96}, {column: 'Конкурент C', value: 101}, {column: 'Среднее по рынку', value: 100}]},
  {row: 'Дом', values: [{column: 'Вы', value: 92}, {column: 'Конкурент A', value: 97}, {column: 'Конкурент B', value: 101}, {column: 'Конкурент C', value: 99}, {column: 'Среднее по рынку', value: 100}]},
  {row: 'Одежда', values: [{column: 'Вы', value: 95}, {column: 'Конкурент A', value: 93}, {column: 'Конкурент B', value: 98}, {column: 'Конкурент C', value: 102}, {column: 'Среднее по рынку', value: 100}]},
  {row: 'Красота', values: [{column: 'Вы', value: 101}, {column: 'Конкурент A', value: 100}, {column: 'Конкурент B', value: 97}, {column: 'Конкурент C', value: 103}, {column: 'Среднее по рынку', value: 100}]},
  {row: 'Спорт', values: [{column: 'Вы', value: 89}, {column: 'Конкурент A', value: 96}, {column: 'Конкурент B', value: 102}, {column: 'Конкурент C', value: 99}, {column: 'Среднее по рынку', value: 100}]},
  {row: 'Игрушки', values: [{column: 'Вы', value: 97}, {column: 'Конкурент A', value: 103}, {column: 'Конкурент B', value: 101}, {column: 'Конкурент C', value: 94}, {column: 'Среднее по рынку', value: 100}]},
  {row: 'Продукты', values: [{column: 'Вы', value: 100}, {column: 'Конкурент A', value: 99}, {column: 'Конкурент B', value: 98}, {column: 'Конкурент C', value: 101}, {column: 'Среднее по рынку', value: 100}]},
]

export const overviewInsights = [
  {title: 'Спрос снизится в категории «Дом»', text: 'Прогнозируемое снижение спроса на 12% на следующей неделе. Проверьте тренды поисковых запросов и сезонные факторы.', tone: 'warning' as const},
  {title: 'Электроника выше рынка', text: 'Вы выше рынка по 272 SKU в категории «Электроника». Цены на 10% выше рынка в среднем.', tone: 'info' as const},
  {title: 'Риск избыточных запасов', text: 'У 412 SKU высокий риск избыточных запасов из-за низкого прогноза спроса и высокого уровня остатков.', tone: 'danger' as const},
  {title: 'Окно разметки открыто', text: 'Окно для разметки открыто для 89 сезонных позиций. Нужно проверить товары с низкой скоростью продаж.', tone: 'warning' as const},
  {title: 'Влияние защитных правил', text: '23 рекомендации скорректированы для сохранения маржи и целевого индекса цен.', tone: 'success' as const},
]

export const markdownAlerts = [
  {sku: 'SKU245667', product: 'Air Purifier Pro 3000', reason: 'Сезонный товар', action: 'Снизить цену 15%'},
  {sku: 'SKU136450', product: "Men's Running Tee", reason: 'Низкая скорость продаж', action: 'Снизить цену 20%'},
  {sku: 'SKU078900', product: 'Blender X200', reason: 'Новый товар', action: 'Снизить цену 10%'},
]

export const overviewRecommendations: PriceAction[] = [
  {id: 'pa-10024567', sku: 'SKU10024567', product: 'Air Purifier Pro 3000', category: 'Дом', currentPrice: '$229', recommendedPrice: '$199', deltaPct: '-13.1%', reason: 'Разметка — низкий спрос', expectedRevenue: '+$82K', expectedMargin: '+$21K', confidence: '91%', guardrailStatus: 'Пройдено', approver: 'Мария И.', status: 'На проверке', risk: 'Средняя'},
  {id: 'pa-50011234', sku: 'SKU50011234', product: 'Wireless Headphones', category: 'Электроника', currentPrice: '$149', recommendedPrice: '$159', deltaPct: '+6.7%', reason: 'Повышение — высокий спрос', expectedRevenue: '+$124K', expectedMargin: '+$46K', confidence: '88%', guardrailStatus: 'Исключение', approver: 'Иван С.', status: 'Черновик', risk: 'Высокая'},
  {id: 'pa-30007890', sku: 'SKU30007890', product: 'Blender X200', category: 'Дом', currentPrice: '$89', recommendedPrice: '$79', deltaPct: '-11.2%', reason: 'Избыточный запас — высокий риск', expectedRevenue: '+$34K', expectedMargin: '-$8K', confidence: '84%', guardrailStatus: 'Пройдено', approver: 'Ольга П.', status: 'Запланировано', risk: 'Средняя'},
]

export const approvalSummary = [
  {label: 'Ожидают', value: '23'},
  {label: 'Действия по ценам', value: '18'},
  {label: 'По разметке', value: '3'},
  {label: 'Новые позиции', value: '2'},
]
