import {Globe} from '@gravity-ui/icons'
import {Button, Icon, Menu, MenuItem} from '@gravity-ui/uikit'

import type {ConsoleActionBarOption} from './ConsoleActionBar'

interface LanguageSwitcherProps {
  label: string
  value: string
  options: ConsoleActionBarOption[]
  onUpdate: (value: string[]) => void
}

export function LanguageSwitcher({label, value, options, onUpdate}: LanguageSwitcherProps) {
  return (
    <Menu
      trigger={(
        <Button aria-label={label} view="flat">
          <Icon data={Globe} size={14} />
        </Button>
      )}
    >
      {options.map((option) => (
        <MenuItem
          key={option.value}
          selected={option.value === value}
          onClick={() => onUpdate([option.value])}
        >
          {option.content}
        </MenuItem>
      ))}
    </Menu>
  )
}
