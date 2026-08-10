import {Avatar, Button, Icon, Text, TextInput} from '@gravity-ui/uikit'
import {BellDot, CircleQuestion, Gear, Magnifier} from '@gravity-ui/icons'

export function TopBar() {
  return (
    <header className="ci-topbar">
      <div className="ci-topbar__brand">
        <div className="ci-topbar__mark">M8</div>
        <div>
          <Text variant="body-2">M8 Commerce Intelligence</Text>
          <Text variant="caption-2" color="secondary">
            Цены, спрос, конкуренты
          </Text>
        </div>
      </div>

      <div className="ci-topbar__search">
        <TextInput startContent={<Icon data={Magnifier} size={14} />} placeholder="Поиск (⌘ + K)" size="m" />
      </div>

      <div className="ci-topbar__actions">
        <Button view="flat" size="m" title="Уведомления">
          <Icon data={BellDot} size={16} />
        </Button>
        <Button view="flat" size="m" title="Помощь">
          <Icon data={CircleQuestion} size={16} />
        </Button>
        <Button view="flat" size="m" title="Настройки">
          <Icon data={Gear} size={16} />
        </Button>
        <Avatar text="SS" size="s" theme="brand" />
      </div>
    </header>
  )
}
