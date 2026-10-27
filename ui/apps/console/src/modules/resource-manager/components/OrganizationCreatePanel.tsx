import {useRef, useState} from 'react'
import {Button, Card, Text, TextInput} from '@gravity-ui/uikit'
import {useMutation, useQueryClient} from '@tanstack/react-query'
import {useRouter} from '@tanstack/react-router'
import type {AppLanguage} from '../../../i18n'
import {
  createOrganization, createTenant, TenantNotFoundError, verifyTenant,
} from '../api/organizationMutations'
import '../pages/organization-editor.css'

const uuidPattern = /^[a-f\d]{8}-[a-f\d]{4}-[1-8][a-f\d]{3}-[89ab][a-f\d]{3}-[a-f\d]{12}$/i
type TenantMode = 'existing' | 'new'
interface Props {language: AppLanguage; onCancel: () => void}

export function OrganizationCreatePanel({language, onCancel}: Props) {
  const ru = language === 'ru'
  const router = useRouter()
  const queryClient = useQueryClient()
  const [tenantMode, setTenantMode] = useState<TenantMode>('existing')
  const [tenantId, setTenantId] = useState('')
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
    ? ru
      ? 'Tenant не найден в текущем Resource Manager. Проверьте его ID и адрес API. UUID из другого модуля не создаёт Tenant автоматически.'
      : 'The Tenant does not exist in this Resource Manager. Check the ID and API environment; an ID from another module does not provision a Tenant.'
    : mutation.error?.message

  return (
    <Card view="outlined" className="m8-org-editor">
      <Text as="h2" variant="header-2">{ru ? 'Создать организацию' : 'Create organization'}</Text>
      <Text variant="body-2" color="secondary">
        {ru
          ? 'Организация всегда должна принадлежать существующему Tenant в этом Resource Manager. Для новой инсталляции администратор сначала создаёт Tenant.'
          : 'Every Organization belongs to an existing Tenant in this Resource Manager. On a new installation, an administrator provisions the Tenant first.'}
      </Text>
      <fieldset className="m8-org-editor__tenant-mode" disabled={mutation.isPending}>
        <legend>{ru ? 'Родительский Tenant' : 'Parent Tenant'}</legend>
        <label>
          <input type="radio" name="tenant-mode" checked={tenantMode === 'existing'}
            onChange={() => switchMode('existing')}/>
          {ru ? 'Использовать существующий' : 'Use an existing Tenant'}
        </label>
        <label>
          <input type="radio" name="tenant-mode" checked={tenantMode === 'new'}
            onChange={() => switchMode('new')}/>
          {ru ? 'Создать новый (администратор платформы)' : 'Create new (platform administrator)'}
        </label>
      </fieldset>
      <form className="m8-org-editor__form" onSubmit={(event) => {
        event.preventDefault()
        if (valid && !mutation.isPending) mutation.mutate()
      }}>
        {tenantMode === 'existing' ? (
          <label className="m8-org-editor__field">
            <Text variant="caption-2">Tenant ID (UUID)</Text>
            <TextInput value={tenantId} placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
              onUpdate={(value) => {setTenantId(value);resetOperation()}} disabled={mutation.isPending}/>
            <Text variant="caption-2" color="secondary">
              {ru
                ? 'При создании будет проверено, что Tenant существует и активен. При отсутствии прав сервер вернёт 401/403.'
                : 'Before creating an Organization, the API checks that this Tenant exists and is active. Access errors remain 401/403.'}
            </Text>
          </label>
        ) : (
          <label className="m8-org-editor__field">
            <Text variant="caption-2">{ru ? 'Название нового Tenant *' : 'New Tenant name *'}</Text>
            <TextInput value={tenantName}
              placeholder={ru ? 'Например, Группа компаний Альфа' : 'For example, Alpha Group'}
              onUpdate={(value) => {setTenantName(value);resetOperation(true)}} disabled={mutation.isPending}/>
            <Text variant="caption-2" color="secondary">
              {ru
                ? 'Создание Tenant — отдельная привилегированная операция. Без разрешения M8 Access сервер отклонит запрос.'
                : 'Tenant creation is a separate privileged operation. The server rejects it without M8 Access permission.'}
            </Text>
          </label>
        )}
        <label className="m8-org-editor__field">
          <Text variant="caption-2">{ru ? 'Название организации *' : 'Organization name *'}</Text>
          <TextInput value={name} placeholder={ru ? 'Например, Альфа Ритейл' : 'For example, Alpha Retail'}
            onUpdate={(value) => {setName(value);resetOperation()}} disabled={mutation.isPending}/>
        </label>
        <label className="m8-org-editor__field">
          <Text variant="caption-2">{ru ? 'Описание' : 'Description'}</Text>
          <textarea className="m8-org-editor__textarea" value={description} maxLength={1024} rows={3}
            onChange={(event) => {setDescription(event.target.value);resetOperation()}}
            disabled={mutation.isPending}/>
        </label>
        {createdTenantId && tenantMode === 'new' ? (
          <div className="m8-org-editor__created-tenant" role="status">
            <Text variant="body-2">{ru ? 'Tenant уже создан' : 'Tenant was created'}</Text>
            <Text variant="caption-2" className="m8-mono">{createdTenantId}</Text>
            <Text variant="caption-2" color="secondary">
              {ru
                ? 'При повторной отправке будет использован этот же Tenant, новый не создаётся.'
                : 'Retry will reuse this Tenant instead of creating another one.'}
            </Text>
          </div>
        ) : null}
        {mutation.isError ? (
          <div className="m8-org-editor__error" role="alert">
            <Text color="danger" variant="body-2">{errorMessage}</Text>
            {missingTenant ? <Text variant="caption-2" color="secondary">
              {ru
                ? 'Для новой среды переключитесь на «Создать новый Tenant», если у вас есть права администратора платформы.'
                : 'For a new environment, switch to Create new Tenant if you have platform administrator permission.'}
            </Text> : null}
          </div>
        ) : null}
        <Text variant="caption-2" color="secondary">
          {ru
            ? 'Повтор после сетевой ошибки сохраняет ключ операции. Изменение полей организации создаёт новый ключ.'
            : 'A retry after network failure preserves the operation key. Editing Organization fields creates a new key.'}
        </Text>
        <div className="m8-org-editor__actions">
          <Button view="outlined" type="button" onClick={onCancel} disabled={mutation.isPending}>
            {ru ? 'Отмена' : 'Cancel'}
          </Button>
          <Button view="action" type="submit" disabled={!valid || mutation.isPending} loading={mutation.isPending}>
            {ru ? 'Создать организацию' : 'Create organization'}
          </Button>
        </div>
      </form>
    </Card>
  )
}
