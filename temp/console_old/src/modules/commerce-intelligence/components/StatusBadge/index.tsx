import {Label} from '@gravity-ui/uikit'

import type {StatusTone} from '../../mock/types'

const themeByTone: Record<StatusTone, 'success' | 'warning' | 'danger' | 'info' | 'normal' | 'utility'> = {
  success: 'success',
  warning: 'warning',
  danger: 'danger',
  info: 'info',
  neutral: 'normal',
  utility: 'utility',
}

export function StatusBadge({children, tone = 'neutral'}: {children: string; tone?: StatusTone}) {
  return (
    <Label className={`ci-status ci-status_${tone}`} theme={themeByTone[tone]}>
      {children}
    </Label>
  )
}
