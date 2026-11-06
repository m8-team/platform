import {ClipboardButton, Label} from '@gravity-ui/uikit'

import type {Translate} from '../../../i18n'
import type {Organization} from '../api/organizations'

export function CopyableOrganizationID({id, t}: {id: string; t: Translate}) {
  return (
    <div className="m8-copyable-cell">
      <span className="m8-mono">{id}</span>
      <ClipboardButton
        text={id}
        view="flat-secondary"
        size="s"
        tooltipInitialText={t('resource.copy')}
        tooltipSuccessText={t('resource.copied')}
      />
    </div>
  )
}

export function OrganizationStateLabel({state, t}: {state: Organization['state']; t: Translate}) {
  const theme =
    state === 'ACTIVE'
      ? 'success'
      : state === 'FAILED'
        ? 'danger'
        : state === 'SUSPENDED' || state === 'DELETING'
          ? 'warning'
          : 'normal'
  return <Label theme={theme}>{t(`organizations.state.${state}`)}</Label>
}
