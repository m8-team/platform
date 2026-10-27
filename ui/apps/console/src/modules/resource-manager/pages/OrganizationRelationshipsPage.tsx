import {useState} from 'react'
import {Button, Card, Text, TextInput} from '@gravity-ui/uikit'

import {ConsoleBreadcrumbs} from '../../../components/ConsoleBreadcrumbs'
import type {Translate} from '../../../i18n'
import {OrganizationRelationshipsPanel} from '../components/OrganizationRelationshipsPanel'
import '../pages/organization-groups.css'

const uuid = /^[a-f\d]{8}-[a-f\d]{4}-[1-8][a-f\d]{3}-[89ab][a-f\d]{3}-[a-f\d]{12}$/i
const key = 'm8.resource-manager.organizations.tenant-id'
function initialTenant() {
  let saved = ''
  try {saved = sessionStorage.getItem(key) ?? ''} catch { /* optional */ }
  const configured = (import.meta.env.VITE_RESOURCE_MANAGER_TENANT_ID ?? '').trim()
  return uuid.test(saved) ? saved : uuid.test(configured) ? configured : ''
}
export function OrganizationRelationshipsPage({t}: {t: Translate}) {
  const [tenantId, setTenantId] = useState(initialTenant)
  const [input, setInput] = useState(initialTenant)
  return <main className="m8-page__body"><section className="m8-page__content">
    <div className="m8-page__heading"><div>
      <ConsoleBreadcrumbs items={[
        {text: t('breadcrumb.resourceManager'), href: '/resource-manager'},
        {text: t('relations.title')},
      ]}/>
      <Text as="h1" variant="display-1">{t('relations.title')}</Text>
      <Text as="p" variant="body-2" color="secondary">{t('relations.intro')}</Text>
    </div></div>
    <Card view="outlined" className="m8-groups__scope">
      <Text variant="subheader-2">{t('relations.scope')}</Text>
      <form className="m8-groups__row" onSubmit={(e) => {
        e.preventDefault()
        const next = input.trim()
        if (!uuid.test(next)) return
        try {sessionStorage.setItem(key, next)} catch { /* optional */ }
        setTenantId(next)
      }}>
        <TextInput value={input} onUpdate={setInput} className="m8-groups__input"
          placeholder={t('groups.tenantLabel')} aria-label={t('groups.tenantLabel')}/>
        <Button view="action" type="submit" disabled={!uuid.test(input.trim())}>{t('relations.tenantSelect')}</Button>
      </form>
    </Card>
    {tenantId ? <OrganizationRelationshipsPanel t={t} tenantId={tenantId}/> :
      <Text color="secondary">{t('relations.tenantMissing')}</Text>}
  </section></main>
}
