import {useRef, useState} from 'react'
import {Button, Card, Text, TextInput} from '@gravity-ui/uikit'
import {useMutation, useQueryClient} from '@tanstack/react-query'
import {useRouter} from '@tanstack/react-router'
import type {AppLanguage} from '../../../i18n'
import {createOrganization} from '../api/organizationMutations'
import '../pages/organization-editor.css'

const uuidPattern = /^[a-f\d]{8}-[a-f\d]{4}-[1-8][a-f\d]{3}-[89ab][a-f\d]{3}-[a-f\d]{12}$/i

interface Props {language: AppLanguage; onCancel: () => void}
export function OrganizationCreatePanel({language, onCancel}: Props) {
  const ru = language === 'ru'
  const router = useRouter()
  const queryClient = useQueryClient()
  const [tenantId, setTenantId] = useState('')
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const key = useRef(crypto.randomUUID())
  const mutation = useMutation({
    mutationFn: () => createOrganization({tenantId, name, description}, key.current),
    onSuccess: async (id) => {
      await queryClient.invalidateQueries({queryKey: ['resource-manager', 'organizations']})
      await router.navigate({to: '/resource-manager/organizations/$organizationId', params: {organizationId: id}})
    },
  })
  const update = (field: 'tenant' | 'name' | 'description', value: string) => {
    key.current = crypto.randomUUID()
    mutation.reset()
    if (field === 'tenant') setTenantId(value)
    else if (field === 'name') setName(value)
    else setDescription(value)
  }
  const valid = uuidPattern.test(tenantId.trim()) && name.trim().length > 0 && name.trim().length <= 256
    && description.length <= 1024
  return (
    <Card view="outlined" className="m8-org-editor">
      <Text as="h2" variant="header-2">{ru ? 'Создать организацию' : 'Create organization'}</Text>
      <Text variant="body-2" color="secondary">
        {ru
          ? 'Новая организация принадлежит выбранному Tenant. Backend одновременно создаёт пустую структуру подразделений (ревизия 1).'
          : 'The new organization belongs to a Tenant. The backend creates an empty revision-1 hierarchy atomically.'}
      </Text>
      <form className="m8-org-editor__form" onSubmit={(event) => {event.preventDefault();if(valid&&!mutation.isPending)mutation.mutate()}}>
        <label className="m8-org-editor__field">
          <Text variant="caption-2">{ru ? 'Tenant ID (UUID)' : 'Tenant ID (UUID)'}</Text>
          <TextInput value={tenantId} placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
            onUpdate={(value) => update('tenant', value)} disabled={mutation.isPending}/>
        </label>
        <label className="m8-org-editor__field">
          <Text variant="caption-2">{ru ? 'Название *' : 'Name *'}</Text>
          <TextInput value={name} placeholder={ru ? 'Например, Альфа Ритейл' : 'For example, Alpha Retail'}
            onUpdate={(value) => update('name', value)} disabled={mutation.isPending}/>
        </label>
        <label className="m8-org-editor__field">
          <Text variant="caption-2">{ru ? 'Описание' : 'Description'}</Text>
          <textarea className="m8-org-editor__textarea" value={description} maxLength={1024} rows={3}
            onChange={(event) => update('description', event.target.value)}
            disabled={mutation.isPending}/>
        </label>
        {mutation.isError ? <Text color="danger" variant="body-2" role="alert">{mutation.error.message}</Text> : null}
        <Text variant="caption-2" color="secondary">
          {ru
            ? 'Изменение полей создаёт новый ключ операции; повтор после сетевой ошибки использует тот же ключ.'
            : 'Editing fields generates a new operation key; retrying a failed request preserves the key.'}
        </Text>
        <div className="m8-org-editor__actions">
          <Button view="outlined" type="button" onClick={onCancel} disabled={mutation.isPending}>
            {ru ? 'Отмена' : 'Cancel'}
          </Button>
          <Button view="action" type="submit" disabled={!valid || mutation.isPending} loading={mutation.isPending}>
            {ru ? 'Создать' : 'Create'}
          </Button>
        </div>
      </form>
    </Card>
  )
}
