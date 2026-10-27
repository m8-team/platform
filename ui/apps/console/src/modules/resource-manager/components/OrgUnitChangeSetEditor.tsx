import {useMemo, useState} from 'react'
import {Button, Label, Text, TextInput} from '@gravity-ui/uikit'
import {useMutation, useQueryClient} from '@tanstack/react-query'
import type {AppLanguage} from '../../../i18n'
import type {OrganizationStructure, OrgUnit} from '../api/organizationDetail'
import {
  applyChangeSet, createDraft, diffStructure, validateChangeSet,
} from '../api/organizationMutations'
import type {ChangeSetDraft, MutationKind, UnitMutation} from '../api/organizationMutations'

type Props = {organizationId: string; structure: OrganizationStructure; language: AppLanguage}
const kinds: MutationKind[] = ['CREATE', 'RENAME', 'MOVE', 'CLOSE']
const labels: Record<AppLanguage, Record<MutationKind, string>> = {
  ru: {CREATE: 'Создать', RENAME: 'Переименовать', MOVE: 'Переместить', CLOSE: 'Закрыть'},
  en: {CREATE: 'Create', RENAME: 'Rename', MOVE: 'Move', CLOSE: 'Close'},
}
const empty = (kind: MutationKind): UnitMutation => ({
  kind, orgUnitId: '', parentOrgUnitId: '', displayName: '', authoritativeSourceId: '',
})

export function OrgUnitChangeSetEditor({organizationId, structure, language}: Props) {
  const ru = language === 'ru'
  const client = useQueryClient()
  const [kind, setKind] = useState<MutationKind>('CREATE')
  const [value, setValue] = useState<UnitMutation>(() => empty('CREATE'))
  const [changes, setChanges] = useState<UnitMutation[]>([])
  const [draftId, setDraftId] = useState(() => crypto.randomUUID())
  const [validatedFingerprint, setValidatedFingerprint] = useState('')
  const [confirmed, setConfirmed] = useState(false)
  const [feedback, setFeedback] = useState('')
  const draft: ChangeSetDraft = useMemo(
    () => ({id: draftId, organizationId, expectedStructureRevision: structure.revision, mutations: changes}),
    [draftId, organizationId, structure.revision, changes],
  )
  const fingerprint = JSON.stringify(draft)
  const validation = useMutation({mutationFn: () => validateChangeSet(draft), onSuccess: () => {
    setValidatedFingerprint(fingerprint)
    setConfirmed(false)
    setFeedback('')
  }})
  const apply = useMutation({
    mutationFn: () => applyChangeSet(draft),
    onSuccess: async ({structureRevision, replayed}) => {
      setFeedback(ru
        ? `ChangeSet ${replayed ? 'повторно подтверждён' : 'применён'}. Ревизия: ${structureRevision}`
        : `ChangeSet ${replayed ? 'replayed' : 'applied'}. Revision: ${structureRevision}`)
      setChanges([])
      setDraftId(crypto.randomUUID())
      setConfirmed(false)
      setValidatedFingerprint('')
      validation.reset()
      await client.invalidateQueries({queryKey: ['resource-manager', 'organization', organizationId, 'structure']})
      await client.invalidateQueries({queryKey: ['resource-manager', 'organization', organizationId]})
    },
  })
  const available = structure.units.filter((u) => u.state === 'ACTIVE' && u.managementMode !== 'EXTERNAL')
  const selected = structure.units.find((u) => u.id === value.orgUnitId)
  const serverOwned = selected?.managementMode === 'EXTERNAL'
  const rootIds = new Set(structure.units.map((u) => u.id))
  const targetParent = value.parentOrgUnitId
  const matchesParent = targetParent === '' || rootIds.has(targetParent)
  const candidateIds = changes.map((c) => c.orgUnitId)
  const duplicate = candidateIds.includes(value.orgUnitId)
  const isCycle = (() => {
    if (kind !== 'MOVE' || !selected || !targetParent) return false
    const index = new Map(structure.units.map((u) => [u.id, u]))
    let cursor: OrgUnit | undefined = index.get(targetParent)
    const seen = new Set<string>()
    while (cursor && !seen.has(cursor.id)) {
      if (cursor.id === selected.id) return true
      seen.add(cursor.id)
      cursor = index.get(cursor.parentId)
    }
    return false
  })()
  const valid = kind === 'CREATE'
    ? Boolean(value.displayName.trim() && matchesParent)
    : Boolean(selected && !serverOwned && !duplicate &&
      (kind === 'RENAME' ? value.displayName.trim() && value.displayName.trim() !== selected.name
        : kind === 'MOVE' ? matchesParent && !isCycle && selected.parentId !== targetParent
          : true))
  const busy = validation.isPending || apply.isPending
  const add = () => {
    if (!valid || busy) return
    const unitId = kind === 'CREATE' ? crypto.randomUUID() : value.orgUnitId
    const mutation: UnitMutation = {
      kind, orgUnitId: unitId,
      parentOrgUnitId: (kind === 'CREATE' || kind === 'MOVE') ? value.parentOrgUnitId : '',
      displayName: (kind === 'CREATE' || kind === 'RENAME') ? value.displayName.trim() : '',
      authoritativeSourceId: '',
    }
    setChanges((existing) => [...existing, mutation])
    setDraftId(crypto.randomUUID())
    setValidatedFingerprint('')
    setConfirmed(false)
    setFeedback('')
    validation.reset()
    apply.reset()
    setValue(empty(kind))
  }
  const remove = (index: number) => {
    setChanges((items) => items.filter((_, i) => i !== index))
    setDraftId(crypto.randomUUID())
    setValidatedFingerprint('')
    setConfirmed(false)
    validation.reset()
    apply.reset()
  }
  const diffs = diffStructure(structure, changes)
  const isValidated = Boolean(validation.data?.valid && validatedFingerprint === fingerprint &&
    validation.data.blockers.length === 0)
  return (
    <section className="m8-changeset">
      <div className="m8-changeset__header">
        <div>
          <Text as="h3" variant="header-1">{ru ? 'Редактор изменений оргструктуры' : 'Org structure change editor'}</Text>
          <Text variant="body-2" color="secondary">
            {ru
              ? 'Изменения сначала собираются в пакет, затем проходят серверную проверку и только после подтверждения применяются.'
              : 'Stage mutations first, validate on the server, and apply only after explicit confirmation.'}
          </Text>
        </div>
        <Label theme="info">{ru ? 'Базовая ревизия' : 'Base revision'}: {structure.revision}</Label>
      </div>
      <div className="m8-changeset__editor">
        <label className="m8-org-editor__field">
          <Text variant="caption-2">{ru ? 'Операция' : 'Operation'}</Text>
          <select className="m8-org-editor__select" value={kind} disabled={busy}
            onChange={(e) => {const next=e.target.value as MutationKind;setKind(next);setValue(empty(next))}}>
            {kinds.map((item) => <option key={item} value={item}>{labels[language][item]}</option>)}
          </select>
        </label>
        {kind !== 'CREATE' ? (
          <label className="m8-org-editor__field">
            <Text variant="caption-2">{ru ? 'Подразделение' : 'Org unit'}</Text>
            <select className="m8-org-editor__select" value={value.orgUnitId} disabled={busy}
              onChange={(e) => setValue((m) => ({...m, orgUnitId: e.target.value}))}>
              <option value="">{ru ? 'Выбрать подразделение' : 'Select an org unit'}</option>
              {structure.units.filter((u) => u.state === 'ACTIVE').map((u) => (
                <option value={u.id} key={u.id}>{u.name}{u.managementMode === 'EXTERNAL' ? ' · external' : ''}</option>
              ))}
            </select>
          </label>
        ) : null}
        {(kind === 'CREATE' || kind === 'RENAME') ? (
          <label className="m8-org-editor__field">
            <Text variant="caption-2">{ru ? 'Название' : 'Name'}</Text>
            <TextInput value={value.displayName} disabled={busy}
              onUpdate={(v) => setValue((m) => ({...m, displayName: v}))}/>
          </label>
        ) : null}
        {(kind === 'CREATE' || kind === 'MOVE') ? (
          <label className="m8-org-editor__field">
            <Text variant="caption-2">{ru ? 'Родительское подразделение' : 'Parent org unit'}</Text>
            <select className="m8-org-editor__select" value={targetParent} disabled={busy}
              onChange={(e) => setValue((m) => ({...m, parentOrgUnitId: e.target.value}))}>
              <option value="">{ru ? 'Корневой уровень' : 'Root level'}</option>
              {available.map((u) => (
                <option key={u.id} value={u.id}>{u.name}</option>
              ))}
            </select>
          </label>
        ) : null}
        {serverOwned ? <Text color="warning" variant="body-2">
          {ru ? 'Подразделение управляется внешним источником. Редактирование через UI запрещено.'
            : 'This org unit is externally managed and cannot be changed here.'}
        </Text> : null}
        {isCycle ? <Text color="danger">{ru ? 'Недопустимо перемещать подразделение в собственного потомка.' : 'Cannot move a unit under its descendant.'}</Text> : null}
        <div className="m8-org-editor__actions">
          <Button view="outlined" disabled={!valid || busy} onClick={add}>
            {ru ? 'Добавить в ChangeSet' : 'Stage mutation'}
          </Button>
        </div>
      </div>

      <div className="m8-changeset__preview">
        <Text as="h4" variant="header-1">{ru ? 'Сравнение до / после' : 'Before / after diff'}</Text>
        {changes.length === 0 ? <Text color="secondary">
          {ru ? 'Изменений пока нет. Выберите операцию и добавьте её в пакет.' : 'No staged changes yet.'}
        </Text> : (
          <div className="m8-org-detail__table-wrap">
            <table className="m8-org-detail__table">
              <thead><tr>
                <th>{ru ? 'Действие' : 'Operation'}</th><th>{ru ? 'Объект' : 'Unit ID'}</th>
                <th>{ru ? 'Название: до → после' : 'Name: before → after'}</th>
                <th>{ru ? 'Родитель: до → после' : 'Parent: before → after'}</th>
                <th>{ru ? 'Состояние' : 'State'}</th><th/>
              </tr></thead>
              <tbody>{diffs.map((row, index) => (
                <tr key={row.unitId}>
                  <td><Label theme="info">{labels[language][row.kind]}</Label></td>
                  <td className="m8-mono">{row.unitId}</td>
                  <td>{row.beforeName} → {row.afterName}</td>
                  <td className="m8-mono">{row.beforeParent} → {row.afterParent}</td>
                  <td>{row.beforeState} → {row.afterState}</td>
                  <td><Button view="flat-danger" size="s" disabled={busy} onClick={() => remove(index)}>
                    {ru ? 'Убрать' : 'Remove'}</Button></td>
                </tr>
              ))}</tbody>
            </table>
          </div>
        )}
        <Text variant="caption-2" color="secondary">
          {ru
            ? 'Предпросмотр отражает изменения структуры, но не рассчитывает затронутые права, сотрудников или владельцев. Серверная оценка воздействия обязательна.'
            : 'Diff shows hierarchy changes, not downstream access, employees or ownership impact. Server-side impact validation is mandatory.'}
        </Text>
      </div>

      {changes.length > 0 ? (
        <div className="m8-changeset__approval">
          <div className="m8-org-editor__readonly">
            <Text variant="caption-2" color="secondary">ChangeSet ID</Text>
            <Text className="m8-mono">{draftId}</Text>
          </div>
          <div className="m8-org-editor__actions">
            <Button view="outlined" disabled={busy} loading={validation.isPending}
              onClick={() => validation.mutate()}>
              {ru ? 'Проверить на сервере' : 'Validate on server'}
            </Button>
            {isValidated ? <Label theme="success">
              {ru ? 'Серверная проверка пройдена' : 'Server validated'}
            </Label> : null}
          </div>
          {validation.isError ? <Text color="danger" role="alert">{validation.error.message}</Text> : null}
          {validation.data?.blockers?.length ? (
            <ul>{validation.data.blockers.map((blocker) => <li key={blocker}>{blocker}</li>)}</ul>
          ) : null}
          {isValidated ? (
            <>
              <label className="m8-changeset__confirm">
                <input type="checkbox" checked={confirmed} disabled={busy} onChange={(e) => setConfirmed(e.target.checked)}/>
                <Text variant="body-2">{ru
                  ? 'Подтверждаю применение к текущей структуре и понимаю, что возврат потребует нового ChangeSet.'
                  : 'I confirm applying these changes; reversal requires another ChangeSet.'}</Text>
              </label>
              <Button view="action" loading={apply.isPending} disabled={!confirmed || busy}
                onClick={() => apply.mutate()}>{ru ? 'Применить ChangeSet' : 'Apply ChangeSet'}</Button>
            </>
          ) : null}
          {apply.isError ? <Text color="danger" role="alert">{apply.error.message}</Text> : null}
        </div>
      ) : null}
      {feedback ? <Text color="positive" role="status">{feedback}</Text> : null}
    </section>
  )
}
