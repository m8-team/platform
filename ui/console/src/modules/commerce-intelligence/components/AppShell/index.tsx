import type {ReactNode} from 'react'
import {Text} from '@gravity-ui/uikit'

import {SidebarNav} from '../SidebarNav'
import {TopBar} from '../TopBar'

export function AppShell({children}: {children: ReactNode}) {
  return (
    <div className="ci-app">
      <TopBar />
      <div className="ci-app__body">
        <SidebarNav />
        <main className="ci-main">{children}</main>
      </div>
      <footer className="ci-footer">
        <Text variant="caption-2">Данные обновлены 5 мин назад</Text>
        <Text variant="caption-2">Все время указано по Europe/Vienna</Text>
        <Text variant="caption-2">Качество данных: Хорошее</Text>
        <Text variant="caption-2">Статус API: Норма</Text>
        <Text variant="caption-2">Статус операций: В норме</Text>
      </footer>
    </div>
  )
}
