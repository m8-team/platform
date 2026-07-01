import {Button, Icon, Text} from '@gravity-ui/uikit'
import {useRouter, useRouterState} from '@tanstack/react-router'

import {commerceNavItems} from '../../navigation'

export function SidebarNav() {
  const router = useRouter()
  const pathname = useRouterState({select: (state) => state.location.pathname})

  return (
    <aside className="ci-sidebar">
      <div className="ci-sidebar__section">
        <Text variant="caption-2" color="secondary">
          Навигация
        </Text>
        <nav className="ci-sidebar__nav">
          {commerceNavItems.map((item) => {
            const active = pathname === item.path && !item.muted

            return (
              <Button
                key={`${item.title}-${item.path}`}
                className={`ci-sidebar__item${active ? ' ci-sidebar__item_active' : ''}${item.muted ? ' ci-sidebar__item_muted' : ''}`}
                view="flat"
                width="max"
                onClick={() => {
                  void router.navigate({to: item.path})
                }}
              >
                <Icon data={item.icon} size={16} />
                <span>{item.title}</span>
              </Button>
            )
          })}
        </nav>
      </div>

      <div className="ci-sidebar__status">
        <Text variant="caption-2" color="secondary">
          Качество данных
        </Text>
        <Text variant="body-2">Хорошее</Text>
        <div className="ci-sidebar__quality" aria-hidden="true">
          <span style={{width: '92%'}} />
        </div>
      </div>
    </aside>
  )
}
