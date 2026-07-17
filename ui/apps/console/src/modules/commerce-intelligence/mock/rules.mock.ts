import type {Rule} from './types'

export const rules: Rule[] = [
  {name: 'Минимальная маржа электроники', type: 'Минимальная маржа', scope: 'Электроника', limit: '18%', priority: 'Высокий', status: 'Активно', updatedAt: '10 июн 2025', author: 'Мария И.'},
  {name: 'Максимальная уценка сезонной одежды', type: 'Максимальная уценка', scope: 'Одежда / сезон', limit: '40%', priority: 'Высокий', status: 'Активно', updatedAt: '9 июн 2025', author: 'Анна Р.'},
  {name: 'Индекс цены конкурентов', type: 'Индекс цены конкурентов', scope: 'Все категории', limit: '95–105', priority: 'Средний', status: 'Активно', updatedAt: '8 июн 2025', author: 'Иван С.'},
  {name: 'Частота изменения цены', type: 'Частота изменения цены', scope: 'Все SKU', limit: '1 раз / 7 дней', priority: 'Средний', status: 'Активно', updatedAt: '7 июн 2025', author: 'Дмитрий К.'},
  {name: 'MAP / РРЦ для бренда Premium', type: 'MAP / РРЦ', scope: 'Premium Brand', limit: 'не ниже РРЦ', priority: 'Критичный', status: 'Активно', updatedAt: '6 июн 2025', author: 'Ольга П.'},
]

export const ruleGroups = ['Активные правила', 'Группы правил', 'Правила по категориям', 'Правила по брендам', 'Исключения', 'История изменений']
