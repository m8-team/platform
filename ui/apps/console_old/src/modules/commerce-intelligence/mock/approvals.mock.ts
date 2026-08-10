import type {Approval, Kpi} from './types'

export const approvalKpis: Kpi[] = [
  {title: 'Ожидают', value: '23', subtitle: 'в очереди решений', sparkline: [18, 20, 21, 24, 23]},
  {title: 'Просрочены', value: '4', delta: '+1', deltaTone: 'negative', subtitle: 'нужна эскалация', sparkline: [2, 3, 3, 4, 4]},
  {title: 'Высокое влияние', value: '7', subtitle: 'выручка или маржа', sparkline: [5, 6, 6, 7, 7]},
  {title: 'Согласовано за 7 дней', value: '184', delta: '+16', deltaTone: 'positive', subtitle: 'решений', sparkline: [122, 138, 151, 169, 184]},
  {title: 'Среднее время согласования', value: '3.2 ч', delta: '-0.4 ч', deltaTone: 'positive', subtitle: 'по всем типам', sparkline: [4.1, 3.8, 3.6, 3.4, 3.2]},
]

export const approvals: Approval[] = [
  {id: 'APR-1024', type: 'Цена', subject: 'SKU50011234', decision: 'Повысить цену до $159', expectedEffect: '+$124K выручки', risk: 'Средняя', requestedBy: 'AI pricing', approver: 'Иван С.', status: 'Ожидает', dueAt: 'Сегодня 17:00'},
  {id: 'APR-1025', type: 'Разметка', subject: 'SKU70022345', decision: 'Уценка 30%', expectedEffect: '+24% sell-through', risk: 'Высокая', requestedBy: 'Markdown Center', approver: 'Анна Р.', status: 'Ожидает', dueAt: 'Завтра 12:00'},
  {id: 'APR-1026', type: 'Исключение', subject: 'Категория Дом', decision: 'Выйти за индекс 105', expectedEffect: '+$82K маржи', risk: 'Высокая', requestedBy: 'Симуляции', approver: 'Мария И.', status: 'Просрочено', dueAt: 'Вчера 18:00'},
  {id: 'APR-1027', type: 'Новая позиция', subject: 'SKU30007890', decision: 'Стартовая цена $89', expectedEffect: '+$34K выручки', risk: 'Низкая', requestedBy: 'Каталог', approver: 'Ольга П.', status: 'На проверке', dueAt: '12 июн 14:00'},
]
