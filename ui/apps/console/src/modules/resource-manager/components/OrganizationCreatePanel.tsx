import {useRef, useState} from 'react'
import {Button, Card, Text, TextInput} from '@gravity-ui/uikit'
import {useMutation, useQueryClient} from '@tanstack/react-query'
import {useRouter} from '@tanstack/react-router'

import type {Translate} from '../../../i18n'
import {
  createOrganization, createTenant, TenantNotFoundError, verifyTenant,
} from '../api/organizationMutations'
import '../pages/organization-editor.css'

const uuidPattern = /^[a-f\d]{8}-[a-f\d]{4}-[1-8][a-f\d]{3}-[89ab][a-f\d]{3}-[a-f\d]{12}$/i
type TenantMode = 'existing' | 'new'
interface Props {t: Translate; initialTenantId?: string; onCancel: () => void}

export function OrganizationCreatePanel({t, initialTenantId, onCancel}: Props) {
  const router = useRouter()
  const queryClient = useQueryClient()
  const [tenantMode, setTenantMode] = useState<TenantMode>('existing')
  const [tenantId, setTenantId] = useState(initialTenantId ?? '')
  const [tenantName, setTenantName] = useState('')
  const [createdTenantId, setCreatedTenantId] = useState('')
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  // Preserve both keys on a failed/retried submission. Tenant creation and
  // Organization creation are two distinct idempotent API transactions.
  const keys = useRef({tenant: crypto.randomUUID(), organization: crypto.randomUUID()})
  const mutation = useMutation({
    mutationFn: async () => {
      let parentId: string
      if (tenantMode === 'new') {
        parentId = createdTenantId || await createTenant(tenantName, keys.current.tenant)
        if (!createdTenantId) setCreatedTenantId(parentId)
      } else {
        parentId = tenantId.trim()
        await verifyTenant(parentId)
      }
      const id = await createOrganization({tenantId: parentId, name, description}, keys.current.organization)
      return {id, parentId}
    },
    onSuccess: async ({id, parentId}) => {
      sessionStorage.setItem('m8.resource-manager.organizations.tenant-id', parentId)
      await queryClient.invalidateQueries({queryKey: ['resource-manager', 'organizations']})
      await router.navigate({to: '/resource-manager/organizations/$organizationId', params: {organizationId: id}})
    },
  })

  const resetOperation = (resetTenant = false) => {
    keys.current.organization = crypto.randomUUID()
    if (resetTenant) {
      keys.current.tenant = crypto.randomUUID()
      setCreatedTenantId('')
    }
    mutation.reset()
  }
  const switchMode = (mode: TenantMode) => {
    if (mutation.isPending || mode === tenantMode) return
    setTenantMode(mode)
    resetOperation(true)
  }
  const validParent = tenantMode === 'existing'
    ? uuidPattern.test(tenantId.trim())
    : tenantName.trim().length > 0 && tenantName.trim().length <= 256
  const valid = validParent && name.trim().length > 0 && name.trim().length <= 256
    && description.length <= 1024
  const missingTenant = mutation.error instanceof TenantNotFoundError
    || mutation.error?.message.includes('Parent Tenant does not exist')
  const errorMessage = missingTenant
    ? t('organizations.create.tenantNotFound')
    : mutation.error?.message

  return (
    <Card view="outlined" className="m8-org-editor">
      <Text as="h2" variant="header-2">{t('organizations.create.title')}</Text>
      <Text variant="body-2" color="secondary">
        {t('organizations.create.intro')}
      </Text>
      <fieldset className="m8-org-editor__tenant-mode" disabled={mutation.isPending}>
        <legend>{t('organizations.create.parentTenant')}</legend>
        <label>
          <input type="radio" name="tenant-mode" checked={tenantMode === 'existing'}
            onChange={() => switchMode('existing')}/>
          {t('organizations.create.existingTenant')}
        </label>
        <label>
          <input type="radio" name="tenant-mode" checked={tenantMode === 'new'}
            onChange={() => switchMode('new')}/>
          {t('organizations.create.newTenant')}
        </label>
      </fieldset>
      <form className="m8-org-editor__form" onSubmit={(event) => {
        event.preventDefault()
        if (valid && !mutation.isPending) mutation.mutate()
      }}>
        {tenantMode === 'existing' ? (
          <label className="m8-org-editor__field">
            <Text variant="caption-2">{t('organizations.tenant.inputLabel')}</Text>
            <TextInput value={tenantId} placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
              onUpdate={(value) => {setTenantId(value);resetOperation()}} disabled={mutation.isPending}/>
            <Text variant="caption-2" color="secondary">
              {t('organizations.create.tenantHelp')}
            </Text>
          </label>
        ) : (
          <label className="m8-org-editor__field">
            <Text variant="caption-2">{t('organizations.create.newTenantName')}</Text>
            <TextInput value={tenantName}
              placeholder={t('organizations.create.newTenantPlaceholder')}
              onUpdate={(value) => {setTenantName(value);resetOperation(true)}} disabled={mutation.isPending}/>
            <Text variant="caption-2" color="secondary">
              {t('organizations.create.newTenantHelp')}
            </Text>
          </label>
        )}
        <label className="m8-org-editor__field">
          <Text variant="caption-2">{t('organizations.create.organizationName')}</Text>
          <TextInput value={name} placeholder={t('organizations.create.organizationPlaceholder')}
            onUpdate={(value) => {setName(value);resetOperation()}} disabled={mutation.isPending}/>
        </label>
        <label className="m8-org-editor__field">
          <Text variant="caption-2">{t('organizations.create.description')}</Text>
          <textarea className="m8-org-editor__textarea" value={description} maxLength={1024} rows={3}
            onChange={(event) => {setDescription(event.target.value);resetOperation()}}
            disabled={mutation.isPending}/>
        </label>
        {createdTenantId && tenantMode === 'new' ? (
          <div className="m8-org-editor__created-tenant" role="status">
            <Text variant="body-2">{t('organizations.create.tenantCreated')}</Text>
            <Text variant="caption-2" className="m8-mono">{createdTenantId}</Text>
            <Text variant="caption-2" color="secondary">
              {t('organizations.create.tenantReuse')}
            </Text>
          </div>
        ) : null}
        {mutation.isError ? (
          <div className="m8-org-editor__error" role="alert">
            <Text color="danger" variant="body-2">{errorMessage}</Text>
            {missingTenant ? <Text variant="caption-2" color="secondary">
              {t('organizations.create.tenantNotFoundHelp')}
            </Text> : null}
          </div>
        ) : null}
        <Text variant="caption-2" color="secondary">
          {t('organizations.create.retryHelp')}
        </Text>
        <div className="m8-org-editor__actions">
          <Button view="outlined" type="button" onClick={onCancel} disabled={mutation.isPending}>
            {t('organizations.create.cancel')}
          </Button>
          <Button view="action" type="submit" disabled={!valid || mutation.isPending} loading={mutation.isPending}>
            {t('organizations.create.title')}
          </Button>
        </div>
      </form>
    </Card>
  )
}
