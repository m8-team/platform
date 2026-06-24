import {Icon, Select} from '@gravity-ui/uikit'
import {Globe} from '@gravity-ui/icons'

import type {ConsoleActionBarOption} from './ConsoleActionBar'

interface LanguageSwitcherProps {
  label: string
  value: string
  options: ConsoleActionBarOption[]
  onUpdate: (value: string[]) => void
}

export function LanguageSwitcher({label, value, options, onUpdate}: LanguageSwitcherProps) {
  return (
    <div className="m8-language-switcher">
      <Icon data={Globe} size={14} />
      <div className="m8-field m8-switcher">
        <Select aria-label={label} value={[value]} options={options} width="max" onUpdate={onUpdate} />
      </div>
    </div>
  )
}
