import {useMemo, useState} from 'react'
import {ArrowRotateRight, Clock, Magnifier, ShieldCheck, TriangleExclamation} from '@gravity-ui/icons'
import {Button, Icon, Text} from '@gravity-ui/uikit'
import {Link, useParams, useRouter} from '@tanstack/react-router'

import {getAuditEvent,listAuditEvents,redactAuditData,resetAuditDemo,setExportDraft,useAuditDemo} from './demo'
import type {AuditCategory,AuditFilters,AuditOutcome,AuditScope,AuditSeverity,ActorType} from './demo'
import {AuditLayout,Empty,Field,Notice,Outcome,Panel,SelectFilter,Severity,Stat,TextField,auditEventPath,dateTimeUTC,fmt,jsonText} from './AuditShell'
import type {T} from './AuditShell'

const categories:AuditCategory[]=['Resource','Gateway','Quota','Identity','Security','System']
const outcomes:AuditOutcome[]=['Success','Failure','Denied']
const severities:AuditSeverity[]=['Info','Warning','Critical']
const scopes:AuditScope[]=['Organization','Project']
const actorTypes:ActorType[]=['User','ServiceAccount','System']
const formatOptions=<T extends string>(values:readonly T[],all:string):Array<{value:string;label:string}>=>[
  {value:'All',label:all},...values.map(value=>({value,label:value})),
]
const pageSize=9

export function AuditEventsPage(){
  const demo=useAuditDemo()
  const router=useRouter()
  const [filters,setFilters]=useState<AuditFilters>({})
  const [page,setPage]=useState(1)
  const update=(key:keyof AuditFilters,value:string)=>{
    setFilters(previous=>({...previous,[key]:value==='All'?'':value}))
    setPage(1)
  }
  const invalidRange=Boolean(filters.fromDate&&filters.toDate&&filters.fromDate>filters.toDate)
  const filtered=useMemo(()=>invalidRange?[]:listAuditEvents(filters,demo.events),[demo.events,filters,invalidRange])
  const lastPage=Math.max(1,Math.ceil(filtered.length/pageSize))
  const rows=filtered.slice((Math.min(page,lastPage)-1)*pageSize,Math.min(page,lastPage)*pageSize)
  const clear=()=>{setFilters({});setPage(1)}
  const createExport=()=>{
    if(invalidRange)return
    setExportDraft(filters)
    void router.navigate({to:'/audit/exports'})
  }
  return <AuditLayout page="events">{t=><>
    <div className="au-section-title">
      <div><Text as="h2" variant="header-2">{t('Журнал событий','Event Journal')}</Text><Text color="secondary" variant="body-2">{t('События в хронологическом порядке с привязкой к ресурсам, субъектам и операциям','Chronological events with actor, resource and operation correlations')}</Text></div>
      <div className="au-actions"><Button view="outlined" onClick={()=>{resetAuditDemo();clear()}}><Icon data={ArrowRotateRight} size={15}/>{t('Сбросить демо','Reset demo')}</Button><Button view="action" disabled={invalidRange} onClick={createExport}>{t('Экспортировать выборку','Export filtered events')}</Button></div>
    </div>
    <div className="au-kpis">
      <Stat label={t('Событий','Events')} value={fmt(filtered.length)} caption={t('С учётом фильтров','Matching filters')}/>
      <Stat label={t('Успешные','Successful')} value={fmt(filtered.filter(e=>e.outcome==='Success').length)}/>
      <Stat label={t('Отказы и ошибки','Denied / Failed')} value={fmt(filtered.filter(e=>e.outcome!=='Success').length)} tone="warning"/>
      <Stat label={t('Критичные','Critical')} value={fmt(filtered.filter(e=>e.severity==='Critical').length)} tone="danger"/>
    </div>
    <Panel title={t('Поиск и фильтрация','Search and filters')} subtitle={t('Все даты в UTC. Период включительный. Демо-запрос выполняется в памяти.','All dates UTC, inclusive; query is evaluated in memory.')}>
      <div className="au-filters">
        <div className="au-filter au-filter_query"><span>{t('Поиск','Search')}</span><div className="au-search"><Icon data={Magnifier} size={16}/><input value={filters.query??''} onChange={e=>update('query',e.target.value)} placeholder={t('Actor, action, resource, ID, correlation','Actor, action, resource, ID, correlation')}/></div></div>
        <SelectFilter label={t('Категория','Category')} value={filters.category||'All'} onChange={v=>update('category',v)} options={formatOptions(categories,t('Все','All'))}/>
        <SelectFilter label={t('Результат','Outcome')} value={filters.outcome||'All'} onChange={v=>update('outcome',v)} options={formatOptions(outcomes,t('Все','All'))}/>
        <SelectFilter label={t('Важность','Severity')} value={filters.severity||'All'} onChange={v=>update('severity',v)} options={formatOptions(severities,t('Все','All'))}/>
        <SelectFilter label="Scope" value={filters.scope||'All'} onChange={v=>update('scope',v)} options={formatOptions(scopes,t('Все','All'))}/>
        <SelectFilter label={t('Субъект','Actor type')} value={filters.actorType||'All'} onChange={v=>update('actorType',v)} options={formatOptions(actorTypes,t('Все','All'))}/>
        <TextField label={t('С даты (UTC)','From date (UTC)')} value={filters.fromDate??''} onChange={v=>update('fromDate',v)} type="date"/>
        <TextField label={t('По дату (UTC)','To date (UTC)')} value={filters.toDate??''} onChange={v=>update('toDate',v)} type="date"/>
      </div>
      <div className="au-filter-footer"><Text variant="caption-2" color="secondary">{t('Найдено','Matched')}: {fmt(filtered.length)} / {fmt(demo.events.length)}</Text><Button view="flat" onClick={clear}>{t('Сбросить фильтры','Clear filters')}</Button></div>
      {invalidRange?<Notice type="danger">{t('Дата начала не может быть позднее даты окончания.','Start date cannot be later than end date.')}</Notice>:null}
    </Panel>
    <Panel title={t('События','Events')} subtitle={t('Демоданные, без подтверждения целостности журнала','Illustrative data; audit-log integrity has NOT been verified')}>
      <div className="au-table-scroll"><table className="au-table">
        <thead><tr><th>{t('Время UTC','Time UTC')}</th><th>{t('Действие','Action')}</th><th>{t('Субъект','Actor')}</th><th>{t('Ресурс','Resource')}</th><th>{t('Результат','Outcome')}</th><th>{t('Важность','Severity')}</th><th/></tr></thead>
        <tbody>{rows.map(e=><tr key={e.id}>
          <td><time className="au-nowrap au-mono" dateTime={e.timestamp}>{dateTimeUTC(e.timestamp)}</time><span className="au-muted au-mono">{e.id}</span></td>
          <td><Link className="au-link" to="/audit/events/$eventId" params={{eventId:e.id}}>{e.action}</Link><span className="au-muted">{e.category} · {e.source}</span></td>
          <td><strong>{e.actor.displayName}</strong><span className="au-muted">{e.actor.type}</span></td>
          <td><strong>{e.resourceName}</strong><span className="au-muted">{e.resourceType} · {e.scope}</span></td>
          <td><Outcome outcome={e.outcome}/></td><td><Severity severity={e.severity}/></td>
          <td><Link className="au-link" to="/audit/events/$eventId" params={{eventId:e.id}}>{t('Детали','Details')}</Link></td>
        </tr>)}</tbody>
      </table></div>
      {!rows.length?<Empty>{t('Нет событий, соответствующих фильтрам','No events match the selected filters')}</Empty>:null}
      <div className="au-pagination">
        <span>{t('Страница','Page')} {Math.min(page,lastPage)} / {lastPage}</span>
        <div><Button size="s" view="outlined" disabled={page<=1} onClick={()=>setPage(p=>Math.max(1,p-1))}>{t('Назад','Previous')}</Button><Button size="s" view="outlined" disabled={page>=lastPage} onClick={()=>setPage(p=>Math.min(lastPage,p+1))}>{t('Далее','Next')}</Button></div>
      </div>
    </Panel>
    <div className="au-columns">
      <Panel title={t('Источники аудита','Audit sources')} subtitle={t('Покрытие демонстрационного журнала','Coverage represented in the demo')}>
        <div className="au-sources">{categories.map(category=><div className="au-source" key={category}><span>{category}</span><div className="au-source__track"><i style={{width:`${demo.events.length?demo.events.filter(e=>e.category===category).length/demo.events.length*100:0}%`}}/></div><strong>{demo.events.filter(e=>e.category===category).length}</strong></div>)}</div>
      </Panel>
      <Panel title={t('Политика аудита','Audit policy')} subtitle={t('Требования к производственной версии','Production design requirements')}>
        <div className="au-guidance"><div><Icon data={ShieldCheck} size={19}/><span>{t('Append-only ingest; доступ на изменение истории не предусмотрен','Append-only ingest; no event editing or deletion through the Console')}</span></div><div><Icon data={Clock} size={19}/><span>{t('Пример: 365 дней хранения, фактическая настройка не применяется','Sample: 365-day retention, not enforced in the demo')}</span></div><div><Icon data={TriangleExclamation} size={19}/><span>{t('Целостность и цепочка доказательств НЕ проверялись','Integrity signatures and chain of custody are NOT verified')}</span></div></div>
      </Panel>
    </div>
  </>}</AuditLayout>
}
function EventTarget({eventId,resourceType,operationId,t}:{eventId:string;resourceType:string;operationId?:string;t:T}){
  const links=[
    ...(operationId?[{to:'/platform-operations/long-running/$operationId' as const,id:operationId,label:`${t('Операция','Operation')} ${operationId}` }]:[]),
  ]
  return <div className="au-correlation-links">
    {links.map(link=><Link key={link.id} className="au-link" to={link.to} params={{operationId:link.id}}>{link.label}</Link>)}
    {resourceType==='AIGatewayRoute'||resourceType==='GatewayPolicy'||resourceType==='Consumer'?<Link className="au-link" to="/gateway/routes">{t('Gateway Console','Gateway Console')}</Link>:null}
    {resourceType==='QuotaReservation'||resourceType==='QuotaPolicy'||resourceType==='QuotaChangeRequest'?<Link className="au-link" to="/quota-manager/overview">{t('Quota Manager','Quota Manager')}</Link>:null}
    {!links.length&&resourceType!=='AIGatewayRoute'&&resourceType!=='GatewayPolicy'&&resourceType!=='Consumer'&&resourceType!=='QuotaReservation'&&resourceType!=='QuotaPolicy'&&resourceType!=='QuotaChangeRequest'?<span className="au-muted">{t('Нет доступной ссылки на ресурс в демоданных','No linked Console resource available in demo')}</span>:null}
    <span className="au-muted au-mono">{eventId}</span>
  </div>
}
export function AuditEventDetailsPage(){
  const {eventId}=useParams({from:'/audit/events/$eventId'})
  useAuditDemo()
  const event=getAuditEvent(eventId)
  return <AuditLayout page="events" details={eventId}>{t=>
    !event?<Panel title={t('Событие не найдено','Event not found')}><Empty><Link className="au-link" to={auditEventPath}>{t('Вернуться к событиям','Back to events')}</Link></Empty></Panel>:
    <>
      <div className="au-detail-heading">
        <div><Link className="au-link" to={auditEventPath}>‹ {t('Все события','All events')}</Link><Text as="h2" variant="header-2">{event.action}</Text><Text variant="body-2" color="secondary">{event.description}</Text></div>
        <div className="au-actions"><Outcome outcome={event.outcome}/><Severity severity={event.severity}/></div>
      </div>
      <div className="au-detail-grid">
        <Panel title={t('Событие','Event details')}>
          <dl className="au-properties">
            <Field label={t('Идентификатор события','Event ID')}><span className="au-mono">{event.id}</span></Field>
            <Field label={t('Время (UTC)','Timestamp (UTC)')}><time dateTime={event.timestamp}>{dateTimeUTC(event.timestamp)}</time></Field>
            <Field label={t('Категория','Category')}>{event.category}</Field>
            <Field label={t('Источник','Source')}>{event.source}</Field>
            <Field label="Scope">{event.scopeName}</Field>
            <Field label={t('Тип ресурса','Resource type')}>{event.resourceType}</Field>
            <Field label={t('Ресурс','Resource')}><span className="au-mono">{event.resourceId}</span> · {event.resourceName}</Field>
          </dl>
        </Panel>
        <Panel title={t('Субъект и корреляция','Actor and correlation')}>
          <dl className="au-properties">
            <Field label={t('Субъект','Actor')}>{event.actor.displayName}</Field>
            <Field label={t('Тип субъекта','Actor type')}>{event.actor.type}</Field>
            <Field label="Actor ID"><span className="au-mono">{event.actor.id}</span></Field>
            <Field label="Request ID"><span className="au-mono">{event.requestId}</span></Field>
            <Field label="Correlation ID"><span className="au-mono">{event.correlationId}</span></Field>
            <Field label="Trace ID"><span className="au-mono">{event.traceId??'—'}</span></Field>
            <Field label={t('Исходный адрес','Source IP')}>{event.ipAddress??'—'}</Field>
          </dl>
        </Panel>
      </div>
      <div className="au-detail-grid">
        <Panel title={t('Изменения','Changes')} subtitle={t('Before/After JSON; потенциально чувствительные поля скрыты','Before/after JSON; sensitive fields are redacted')}>
          {event.changes?<div className="au-diff">
            <div><span>Before</span><pre>{jsonText(redactAuditData(event.changes.before??{}))}</pre></div>
            <div><span>After</span><pre>{jsonText(redactAuditData(event.changes.after??{}))}</pre></div>
          </div>:<Empty>{t('Для этого события нет пары Before/After','No before/after snapshot was captured')}</Empty>}
        </Panel>
        <Panel title={t('Контекст выполнения','Execution context')}>
          <EventTarget eventId={event.id} resourceType={event.resourceType} operationId={event.operationId} t={t}/>
          <Text variant="caption-2" color="secondary">Metadata (redacted)</Text>
          <pre className="au-json">{jsonText(redactAuditData(event.metadata??{}))}</pre>
          <Notice type="warning">{t('Демо-событие не имеет подтверждённой цифровой подписи или криптографического доказательства целостности.','This demo event has no validated signature, cryptographic integrity proof or chain-of-custody attestation.')}</Notice>
        </Panel>
      </div>
    </>
  }</AuditLayout>
}
