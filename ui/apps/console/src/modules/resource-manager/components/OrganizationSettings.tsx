import {useEffect, useRef, useState} from 'react'
import {Button, Label, Text, TextInput} from '@gravity-ui/uikit'
import {useMutation, useQueryClient} from '@tanstack/react-query'
import type {AppLanguage} from '../../../i18n'
import type {OrganizationDetail} from '../api/organizationDetail'
import {changeOrganizationState, updateOrganization} from '../api/organizationMutations'

export function OrganizationSettings({
  organization, language,
}: {organization: OrganizationDetail; language: AppLanguage}) {
  const ru = language === 'ru'
  const queryClient = useQueryClient()
  const [name, setName] = useState(organization.name)
  const [description, setDescription] = useState(organization.description ?? '')
  const [confirmSuspend, setConfirmSuspend] = useState(false)
  const saveKey = useRef(crypto.randomUUID())
  const stateKey = useRef(crypto.randomUUID())
  useEffect(() => {
    setName(organization.name)
    setDescription(organization.description ?? '')
    setConfirmSuspend(false)
  }, [organization.id, organization.name, organization.description])
  const refresh = async () => {
    await queryClient.invalidateQueries({queryKey: ['resource-manager', 'organization', organization.id]})
    await queryClient.invalidateQueries({queryKey: ['resource-manager', 'organizations']})
  }
  const save = useMutation({
    mutationFn: () => updateOrganization(organization, {name, description}, saveKey.current),
    onSuccess: refresh,
  })
  const changeState = useMutation({
    mutationFn: (state: 'SUSPENDED' | 'ACTIVE') => changeOrganizationState(organization, state, stateKey.current),
    onSuccess: async () => {setConfirmSuspend(false); await refresh()},
  })
  const dirty = name.trim() !== organization.name || description.trim() !== (organization.description ?? '')
  const valid = name.trim().length > 0 && name.length <= 256 && description.length <= 1024
  const canSuspend = organization.state === 'ACTIVE'
  const canResume = organization.state === 'SUSPENDED'
  return (
    <div className="m8-org-editor m8-org-editor_settings">
      <Text as="h2" variant="header-2">{ru ? 'Настройки организации' : 'Organization settings'}</Text>
      <Text variant="body-2" color="secondary">
        {ru
          ? 'ID и принадлежность Tenant неизменяемы. Изменения используют ожидаемую ревизию для защиты от конкурентных операций.'
          : 'Resource ID and Tenant ownership are immutable. Updates use optimistic revision checks.'}
      </Text>
      <form className="m8-org-editor__form" onSubmit={(e) => {e.preventDefault();if(valid&&dirty)save.mutate()}}>
        <label className="m8-org-editor__field">
          <Text variant="caption-2">{ru ? 'Название организации' : 'Organization name'}</Text>
          <TextInput value={name} onUpdate={(v) => {setName(v);saveKey.current=crypto.randomUUID();save.reset()}}
            disabled={save.isPending || changeState.isPending}/>
        </label>
        <label className="m8-org-editor__field">
          <Text variant="caption-2">{ru ? 'Описание' : 'Description'}</Text>
          <textarea className="m8-org-editor__textarea" value={description} maxLength={1024} rows={3}
            onChange={(e) => {setDescription(e.target.value);saveKey.current=crypto.randomUUID();save.reset()}}
            disabled={save.isPending || changeState.isPending}/>
        </label>
        <div className="m8-org-editor__readonly">
          <Text variant="caption-2" color="secondary">ID</Text>
          <Text variant="body-2" className="m8-mono">{organization.id}</Text>
          <Text variant="caption-2" color="secondary">Tenant ID</Text>
          <Text variant="body-2" className="m8-mono">{organization.tenantId}</Text>
          <Text variant="caption-2" color="secondary">{ru ? 'Версия' : 'Revision'}</Text>
          <Text variant="body-2">{organization.revision ?? '—'}</Text>
        </div>
        {save.isError ? <Text color="danger" role="alert">{save.error.message}</Text> : null}
        {save.isSuccess ? <Text color="positive">{ru ? 'Настройки сохранены' : 'Settings saved'}</Text> : null}
        <div className="m8-org-editor__actions">
          <Button type="button" view="outlined" disabled={!dirty || save.isPending}
            onClick={() => {setName(organization.name);setDescription(organization.description ?? '');save.reset()}}>
            {ru ? 'Сбросить' : 'Reset'}
          </Button>
          <Button type="submit" view="action" disabled={!valid || !dirty || save.isPending || !organization.revision}
            loading={save.isPending}>{ru ? 'Сохранить' : 'Save changes'}</Button>
        </div>
      </form>
      <hr className="m8-org-editor__rule"/>
      <Text as="h3" variant="header-1">{ru ? 'Состояние организации' : 'Organization lifecycle'}</Text>
      <Label theme={canSuspend ? 'success' : canResume ? 'warning' : 'normal'}>{organization.state}</Label>
      <Text variant="body-2" color="secondary">
        {ru ? 'Приостановка влияет на создание дочерних ресурсов. Изменения доступа требуют отдельной интеграции M8 Access.'
          : 'Suspending affects child resource provisioning. Access reconciliation is managed separately by M8 Access.'}
      </Text>
      {canSuspend ? (
        confirmSuspend ? (
          <div className="m8-org-editor__danger">
            <Text variant="body-2">{ru ? 'Подтвердите приостановку организации.' : 'Confirm suspending the Organization.'}</Text>
            <div className="m8-org-editor__actions">
              <Button view="outlined" onClick={() => setConfirmSuspend(false)}>{ru ? 'Отмена' : 'Cancel'}</Button>
              <Button view="outlined-danger" loading={changeState.isPending} disabled={changeState.isPending}
                onClick={() => changeState.mutate('SUSPENDED')}>
                {ru ? 'Подтвердить приостановку' : 'Confirm suspension'}
              </Button>
            </div>
          </div>
        ) : (
          <div><Button view="outlined-danger" onClick={() => {stateKey.current=crypto.randomUUID();setConfirmSuspend(true)}}>
            {ru ? 'Приостановить' : 'Suspend'}</Button></div>
        )
      ) : null}
      {canResume ? <div><Button view="outlined" loading={changeState.isPending}
        onClick={() => {stateKey.current=crypto.randomUUID();changeState.mutate('ACTIVE')}}>
        {ru ? 'Возобновить' : 'Resume'}</Button></div> : null}
      {changeState.isError ? <Text role="alert" color="danger">{changeState.error.message}</Text> : null}
      {organization.companyProfileRef ? (
        <Text color="secondary">Company Profile: {organization.companyProfileRef}</Text>
      ) : null}
    </div>
  )
}
