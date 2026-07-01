import type {Kpi, Product} from './types'

export const productKpis: Kpi[] = [
  {title: 'Активные SKU', value: '12 842', subtitle: 'в коммерческом портфеле', sparkline: [12210, 12340, 12480, 12610, 12740, 12842]},
  {title: 'SKU с ценовым риском', value: '1 276', delta: '+4.2%', deltaTone: 'negative', subtitle: 'требуют проверки', sparkline: [980, 1040, 1120, 1180, 1230, 1276]},
  {title: 'Сезонные товары', value: '2 341', subtitle: 'активные окна', sparkline: [2100, 2140, 2210, 2290, 2330, 2341]},
  {title: 'SKU с высокой эластичностью', value: '1 103', subtitle: 'чувствительны к цене', sparkline: [940, 980, 1010, 1044, 1080, 1103]},
  {title: 'SKU с низкой конверсией', value: '842', delta: '-3.1%', deltaTone: 'positive', subtitle: 'улучшение за неделю', sparkline: [930, 910, 882, 870, 855, 842]},
  {title: 'Проблемы качества данных', value: '387', delta: '+19', deltaTone: 'negative', subtitle: 'ожидают исправления', sparkline: [290, 310, 336, 350, 368, 387]},
]

export const products: Product[] = [
  {sku: 'SKU10024567', product: 'Air Purifier Pro 3000', category: 'Дом', currentPrice: '$229', marketPrice: '$211', priceIndex: 108.5, stock: '8 240', coverageDays: 64, sellThrough: '38%', sales7d: '420', forecast30d: '1 860', elasticity: '-1.7', lifecycle: 'Зрелый', risk: 'Риск', status: 'Избыточный запас'},
  {sku: 'SKU20013459', product: "Men's Running Tee", category: 'Одежда', currentPrice: '$32', marketPrice: '$29', priceIndex: 110.3, stock: '14 120', coverageDays: 42, sellThrough: '51%', sales7d: '1 840', forecast30d: '7 200', elasticity: '-2.1', lifecycle: 'Сезонный', risk: 'Средняя', status: 'Сезонный'},
  {sku: 'SKU30007890', product: 'Blender X200', category: 'Дом', currentPrice: '$89', marketPrice: '$92', priceIndex: 96.7, stock: '3 780', coverageDays: 58, sellThrough: '34%', sales7d: '310', forecast30d: '1 240', elasticity: '-1.4', lifecycle: 'Новый запуск', risk: 'Риск', status: 'Медленно продается'},
  {sku: 'SKU40055678', product: 'Yoga Mat Premium', category: 'Спорт', currentPrice: '$54', marketPrice: '$51', priceIndex: 105.9, stock: '5 640', coverageDays: 31, sellThrough: '62%', sales7d: '980', forecast30d: '3 880', elasticity: '-1.2', lifecycle: 'Рост', risk: 'Норма', status: 'Здоровый'},
  {sku: 'SKU50011234', product: 'Wireless Headphones', category: 'Электроника', currentPrice: '$149', marketPrice: '$156', priceIndex: 95.5, stock: '2 110', coverageDays: 17, sellThrough: '76%', sales7d: '1 140', forecast30d: '4 920', elasticity: '-0.9', lifecycle: 'Рост', risk: 'Низкая', status: 'Здоровый'},
  {sku: 'SKU60077890', product: 'Stainless Steel Bottle', category: 'Спорт', currentPrice: '$24', marketPrice: '$25', priceIndex: 96.0, stock: '9 800', coverageDays: 49, sellThrough: '44%', sales7d: '1 090', forecast30d: '4 200', elasticity: '-1.1', lifecycle: 'Зрелый', risk: 'Норма', status: 'Здоровый'},
  {sku: 'SKU70022345', product: 'Winter Jacket', category: 'Одежда', currentPrice: '$189', marketPrice: '$171', priceIndex: 110.5, stock: '6 320', coverageDays: 86, sellThrough: '28%', sales7d: '260', forecast30d: '980', elasticity: '-2.4', lifecycle: 'Сезонный', risk: 'Высокая', status: 'Избыточный запас'},
]

export const productSegments = ['Высокая маржа', 'Высокая скорость продаж', 'Кандидаты на разметку', 'Конкурентное давление', 'Новые товары']

export const portfolioDistribution = [
  {category: 'Электроника', sku: 3120, revenue: 42},
  {category: 'Дом', sku: 2840, revenue: 31},
  {category: 'Одежда', sku: 2460, revenue: 28},
  {category: 'Спорт', sku: 1810, revenue: 22},
  {category: 'Красота', sku: 1410, revenue: 18},
]

export const productOpportunities = [
  {sku: 'SKU50011234', product: 'Wireless Headphones', action: 'Повысить цену на $4–$8', effect: '+6–9% выручки'},
  {sku: 'SKU70022345', product: 'Winter Jacket', action: 'Уценка 25–30%', effect: '+18% sell-through'},
  {sku: 'SKU10024567', product: 'Air Purifier Pro 3000', action: 'Снизить цену 12–15%', effect: '-19 дней покрытия'},
]
