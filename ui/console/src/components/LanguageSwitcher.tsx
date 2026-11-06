import {Globe} from '@gravity-ui/icons'
import {Button, DropdownMenu, Icon} from '@gravity-ui/uikit'
import type {DropdownMenuItem} from '@gravity-ui/uikit'

import type {ConsoleActionBarOption} from './ConsoleActionBar'

interface LanguageSwitcherProps {
  label: string
  value: string
  options: ConsoleActionBarOption[]
  onUpdate: (value: string[]) => void
}

export function LanguageSwitcher({label, value, options, onUpdate}: LanguageSwitcherProps) {
  const menuItems: DropdownMenuItem[] = options.map((option) => ({
    text: option.content,
    selected: option.value === value,
    action: () => onUpdate([option.value]),
  }))

  return (
      <DropdownMenu
        items={menuItems}
        renderSwitcher={({onClick, onKeyDown}) => (
          <Button aria-label={label} view="flat" onClick={onClick} onKeyDown={onKeyDown}>
            <Icon data={Globe} size={14} />
          </Button>
        )}
      />
  )
}
