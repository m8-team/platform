import {ActionsPanel, Icon} from '@gravity-ui/uikit'
import {ArrowRight, Copy} from '@gravity-ui/icons'
import {toaster} from '@gravity-ui/uikit/toaster-singleton'

import type {Translate} from '../../../i18n'
import type {Organization} from '../api/organizations'

export interface OrganizationActionsPanelProps {
  organizations: Organization[]
  onClear: () => void
  onOpen: (organization: Organization) => void
  t: Translate
}

export function OrganizationActionsPanel({organizations, onClear, onOpen, t}: OrganizationActionsPanelProps) {
  const openSelected = () => {
    if (organizations.length === 1) onOpen(organizations[0])
  }
  const copySelectedIds = async () => {
    try {
      await navigator.clipboard.writeText(organizations.map((organization) => organization.id).join('\n'))
      toaster.add({
        name: 'organization-ids-copied',
        title: t('organizations.actions.idsCopied'),
        theme: 'success',
        autoHiding: 3000,
      })
    } catch {
      toaster.add({
        name: 'organization-ids-copy-failed',
        title: t('organizations.actions.copyFailed'),
        theme: 'danger',
        autoHiding: 5000,
      })
    }
  }

  return (
    <ActionsPanel
      className="m8-selection-actions-panel"
      renderNote={() => `${t('organizations.actions.selected')}: ${organizations.length}`}
      onClose={onClear}
      actions={[
        {
          id: 'open',
          button: {
            props: {
              children: [<Icon key="icon" data={ArrowRight} size={16} />, t('organizations.actions.open')],
              disabled: organizations.length !== 1,
              onClick: openSelected,
            },
          },
          dropdown: {
            item: {
              text: t('organizations.actions.open'),
              disabled: organizations.length !== 1,
              action: openSelected,
              iconStart: <Icon data={ArrowRight} size={16} />,
            },
          },
        },
        {
          id: 'copy-ids',
          button: {
            props: {
              children: [<Icon key="icon" data={Copy} size={16} />, t('organizations.actions.copyIds')],
              onClick: () => void copySelectedIds(),
            },
          },
          dropdown: {
            item: {
              text: t('organizations.actions.copyIds'),
              action: () => void copySelectedIds(),
              iconStart: <Icon data={Copy} size={16} />,
            },
          },
        },
      ]}
    />
  )
}
