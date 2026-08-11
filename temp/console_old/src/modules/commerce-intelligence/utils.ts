import {toaster} from '@gravity-ui/uikit/toaster-singleton'

export function notifyAction(title: string, content?: string) {
  toaster.add({
    name: `${Date.now()}-${Math.random()}`,
    title,
    content,
    theme: 'success',
    autoHiding: 3500,
  })
}

export function statusTone(status: string) {
  if (['Согласовано', 'Запланировано', 'Применено', 'Пройдено', 'Норма', 'Здоровый', 'Активно'].includes(status)) {
    return 'success' as const
  }

  if (['На проверке', 'Предупреждение', 'Средняя', 'Риск', 'Медленно продается', 'Сезонный', 'Просрочено'].includes(status)) {
    return 'warning' as const
  }

  if (['Ошибка', 'Отклонено', 'Высокая', 'Избыточный запас'].includes(status)) {
    return 'danger' as const
  }

  if (['Черновик', 'Новый запуск', 'Ожидает'].includes(status)) {
    return 'info' as const
  }

  if (['Исключение'].includes(status)) {
    return 'utility' as const
  }

  return 'neutral' as const
}
