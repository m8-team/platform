import type {TranslationKey} from '../../../i18n'
import type {ProjectStatus} from '../model/project'

export const workspaceOptionConfigs = [
  {value: 'ws_prod-eu1', titleKey: 'workspace.platform'},
  {value: 'ws_shared-eu1', titleKey: 'workspace.sharedServices'},
  {value: 'ws_legacy-eu1', titleKey: 'workspace.legacy'},
] satisfies Array<{value: string; titleKey: TranslationKey}>

export const statusOptionConfigs = [
  {value: 'all', titleKey: 'status.all'},
  {value: 'Active', titleKey: 'status.Active'},
  {value: 'Suspended', titleKey: 'status.Suspended'},
  {value: 'Failed', titleKey: 'status.Failed'},
  {value: 'Provisioning', titleKey: 'status.Provisioning'},
  {value: 'Deleting', titleKey: 'status.Deleting'},
] satisfies Array<{value: ProjectStatus | 'all'; titleKey: TranslationKey}>

export const ownerOptionConfigs = [
  {value: 'all', titleKey: 'owner.all'},
  {value: 'usr_19bd4027_sre', content: 'usr_19bd4027_sre'},
  {value: 'usr_2f0c81aa_sec', content: 'usr_2f0c81aa_sec'},
] satisfies Array<{value: string; content?: string; titleKey?: TranslationKey}>
