import {useMemo, useState} from 'react'
import {Button, ClipboardButton, Label, Text, TextInput} from '@gravity-ui/uikit'
import type {AppLanguage} from '../../../i18n'
import type {OrganizationResource} from '../api/organizationDetail'

type Props = {resources: OrganizationResource[]; organizationId: string; language: AppLanguage}

export function ResourceOwnershipCards({resources, organizationId, language}: Props) {
  const ru = language === 'ru'
  const [search, setSearch] = useState('')
  const [selectedId, setSelectedId] = useState('')
  const items = useMemo(() => resources.filter((r) =>
    [r.name, r.kind, r.id].some((value) => value.toLocaleLowerCase().includes(search.trim().toLocaleLowerCase())),
  ), [resources, search])
  const selected = items.find((r) => r.id === selectedId) ?? items[0]
  return (
    <section className="m8-ownership">
      <div className="m8-ownership__header">
        <div>
          <Text as="h2" variant="header-2">{ru ? 'Ответственность за ресурсы' : 'Resource accountability'}</Text>
          <Text variant="body-2" color="secondary">
            {ru
              ? 'Структурная принадлежность берётся из Resource Manager. Бизнес-владелец, технический владелец и исполнитель должны назначаться отдельно — эти API ещё не подключены.'
              : 'Structural scope comes from Resource Manager. Business owner, technical owner and operating team are separate assignments; their APIs are not connected yet.'}
          </Text>
        </div>
        <Label theme="warning">{ru ? 'Нет данных об ответственных' : 'No accountable owners API'}</Label>
      </div>
      <div className="m8-ownership__workspace">
        <div className="m8-ownership__list">
          <TextInput value={search} onUpdate={setSearch}
            placeholder={ru ? 'Найти ресурс' : 'Search resources'}/>
          {items.length === 0 ? <Text color="secondary">
            {ru ? 'Ресурсы не найдены' : 'No matching resources'}
          </Text> : null}
          {items.map((r) => (
            <button type="button" key={`${r.kind}:${r.id}`}
              className={`m8-ownership__resource ${r.id === selected?.id ? 'm8-ownership__resource_active' : ''}`}
              aria-pressed={r.id === selected?.id} onClick={() => setSelectedId(r.id)}>
              <Text variant="body-2">{r.name}</Text>
              <Text variant="caption-2" color="secondary">{r.kind} · {r.state}</Text>
            </button>
          ))}
        </div>
        <div className="m8-ownership__inspector">
          {selected ? (
            <>
              <div className="m8-ownership__title">
                <Text as="h3" variant="header-2">{selected.name}</Text>
                <ClipboardButton text={selected.id} size="s"/>
              </div>
              <div className="m8-ownership__summary">
                <div className="m8-ownership__role">
                  <Text variant="caption-2" color="secondary">{ru ? 'Организация-владелец' : 'Owning organization'}</Text>
                  <Text variant="body-2" className="m8-mono">{organizationId}</Text>
                  <Label theme="info">{ru ? 'Структурная принадлежность' : 'Structural ownership'}</Label>
                </div>
                <div className="m8-ownership__role">
                  <Text variant="caption-2" color="secondary">{ru ? 'Непосредственный контейнер' : 'Immediate container'}</Text>
                  <Text variant="body-2">{selected.structuralOwner?.kind ?? '—'}</Text>
                  <Text variant="caption-2" className="m8-mono" color="secondary">
                    {selected.structuralOwner?.id ?? '—'}
                  </Text>
                </div>
              </div>
              <Text as="h4" variant="header-1">{ru ? 'Ответственные' : 'Accountable roles'}</Text>
              <div className="m8-ownership__summary">
                {([
                  [ru ? 'Бизнес-владелец' : 'Business owner', 'business'],
                  [ru ? 'Технический владелец' : 'Technical owner', 'technical'],
                  [ru ? 'Эксплуатационная команда' : 'Operations team', 'operations'],
                  [ru ? 'Ответственное подразделение' : 'Accountable OrgUnit', 'unit'],
                ] as const).map(([label, key]) => (
                  <div className="m8-ownership__role" key={key}>
                    <Text variant="caption-2" color="secondary">{label}</Text>
                    <Text variant="body-2">{ru ? 'Данные недоступны' : 'Data unavailable'}</Text>
                    <Label theme="warning">API pending</Label>
                  </div>
                ))}
              </div>
              <Text variant="body-2" color="secondary">
                {ru
                  ? 'Недоступность API не означает, что владелец отсутствует. Изменение ответственности и назначение людей будут включены только после интеграции M8 Directory и Access.'
                  : 'API unavailable does not mean no owner is assigned. Editing responsible people requires M8 Directory and Access integration.'}
              </Text>
              <Button view="outlined" disabled>
                {ru ? 'Назначить владельца — ожидается API' : 'Assign owner — API pending'}
              </Button>
            </>
          ) : <Text color="secondary">{ru ? 'Выберите ресурс' : 'Select a resource'}</Text>}
        </div>
      </div>
    </section>
  )
}
