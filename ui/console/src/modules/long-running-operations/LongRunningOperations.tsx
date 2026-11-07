import {useMemo, useState} from 'react'
import type {ReactNode} from 'react'
import {ArrowRotateRight, BarsPlay, Check, Clock, Code, Magnifier, OctagonXmark, TriangleExclamation} from '@gravity-ui/icons'
import {Button, Card, Icon, Text, TextInput} from '@gravity-ui/uikit'
import {Link, useParams, useRouter} from '@tanstack/react-router'

import {ConsoleBreadcrumbs} from '../../components/ConsoleBreadcrumbs'
import {useConsoleI18n} from '../../App'
import {
  advanceDemoOperation,
  cancelDemoOperation,
  resetOperationsDemo,
  retryDemoOperation,
  useOperationsDemo,
} from './demo'
import type {LongRunningOperation, OperationLog, OperationScope, OperationStatus, StepStatus} from './demo'
import './long-running-operations.css'

type T = (ru:string,en:string)=>string
const fmt = (n:number)=>n.toLocaleString('en-US')
const activeStatuses:OperationStatus[]=['Queued','Running','Waiting','CancelRequested']
const finishedStatuses:OperationStatus[]=['Succeeded','Failed','Cancelled']
const statusRu:Record<OperationStatus,string>={
  Queued:'В очереди',Running:'Выполняется',Waiting:'Ожидание',CancelRequested:'Отмена запрошена',
  Succeeded:'Успешно',Failed:'Ошибка',Cancelled:'Отменено',
}
const stepRu:Record<StepStatus,string>={Pending:'Ожидает',Running:'Выполняется',Succeeded:'Завершено',Failed:'Ошибка'}
const statusTone:Record<OperationStatus,string>={
  Queued:'neutral',Running:'info',Waiting:'warning',CancelRequested:'warning',Succeeded:'success',Failed:'danger',Cancelled:'neutral',
}
const pageRoute='/platform-operations/long-running'

function Badge({children,tone='neutral'}:{children:ReactNode;tone?:string}){
  return <span className={`lro-badge lro-badge_${tone}`}>{children}</span>
}
function Status({status,t}:{status:OperationStatus;t:T}){
  return <Badge tone={statusTone[status]}>{t(statusRu[status],status)}</Badge>
}
function Progress({operation,label}:{operation:LongRunningOperation;label?:string}){
  const value=operation.progress
  return <div className="lro-progress-wrap">
    <div className="lro-progress" role="progressbar" aria-valuenow={value} aria-valuemin={0} aria-valuemax={100} aria-label={label??operation.title}>
      <div className={`lro-progress__fill lro-progress__fill_${statusTone[operation.status]}`} style={{width:`${Math.min(100,Math.max(0,value))}%`}}/>
    </div><span className="lro-progress__value">{value}%</span>
  </div>
}
function Tile({label,value,caption,icon,tone='default'}:{label:string;value:number;caption:string;icon:typeof Clock;tone?:string}){
  return <Card view="outlined" className="lro-tile"><div className="lro-tile__top"><Text variant="caption-2" color="secondary">{label}</Text><Icon data={icon} size={17}/></div><span className={`lro-tile__count lro-tile__count_${tone}`}>{fmt(value)}</span><Text variant="caption-2" color="secondary">{caption}</Text></Card>
}
function Panel({title,subtitle,action,children}:{title:string;subtitle?:string;action?:ReactNode;children:ReactNode}){
  return <Card view="outlined" className="lro-panel"><div className="lro-panel__header"><div><Text variant="header-1" as="h2">{title}</Text>{subtitle?<Text variant="caption-2" color="secondary" as="p">{subtitle}</Text>:null}</div>{action}</div>{children}</Card>
}
function OptionFilter({label,value,onChange,options}:{label:string;value:string;onChange:(value:string)=>void;options:Array<{value:string;label:string}>}){
  return <label className="lro-filter"><span>{label}</span><select value={value} onChange={e=>onChange(e.target.value)} className="lro-select">{options.map(o=><option key={o.value} value={o.value}>{o.label}</option>)}</select></label>
}
function Header({detail,t}:{detail?:LongRunningOperation;t:T}){
  return <div className="lro-heading">
    <div>
      <ConsoleBreadcrumbs items={[
        {text:t('Операции платформы','Platform Operations')},
        ...(detail?[{text:t('Длительные операции','Long-running Operations'),href:pageRoute}]:[]),
        {text:detail?.id??t('Длительные операции','Long-running Operations')},
      ]}/>
      <div className="lro-heading__title"><Text as="h1" variant="display-1">{detail?.title??t('Длительные операции','Long-running Operations')}</Text><Badge tone="info">DEMO DATA</Badge></div>
      <Text variant="body-2" color="secondary" as="p">
        {detail?detail.description:t('Центральный журнал длительных операций Resource Manager, Quota Manager и провайдеров','Central view of long-running Resource Manager, Quota Manager and provider operations')}
      </Text>
    </div>
  </div>
}
function Layout({children,detail}:{children:(t:T)=>ReactNode;detail?:LongRunningOperation}){
  const {language}=useConsoleI18n()
  const t:T=(ru,en)=>language==='ru'?ru:en
  return <main className="m8-page__body"><section className="m8-page__content lro-page">
    <Header detail={detail} t={t}/>
    <div className="lro-demo-notice"><Icon data={BarsPlay} size={18}/><span>{t('Демонстрационный интерфейс: операции, прогресс и действия моделируются в памяти вкладки. Temporal и реальные ресурсы не изменяются.','Demo only: operations, progress and actions are simulated in memory. No real Temporal workflow or infrastructure changes.')}</span></div>
    {children(t)}
  </section></main>
}
function Property({label,value}:{label:string;value:ReactNode}){
  return <div className="lro-property"><dt>{label}</dt><dd>{value}</dd></div>
}
function FormatTime({value}:{value:string}){return <span className="lro-mono">{value}</span>}

export function LongRunningOperationsPage(){
  const demo=useOperationsDemo()
  const {language}=useConsoleI18n()
  const t:T=(ru,en)=>language==='ru'?ru:en
  const [search,setSearch]=useState('')
  const [status,setStatus]=useState('All')
  const [service,setService]=useState('All')
  const [scope,setScope]=useState('All')
  const [sort,setSort]=useState('Newest')
  const services=useMemo(()=>[...new Set(demo.operations.map(o=>o.service))].sort(),[demo.operations])
  const rows=useMemo(()=>{
    const filtered=demo.operations.filter(o=>{
      const text=`${o.id} ${o.title} ${o.resource} ${o.service} ${o.workflowId} ${o.correlationId}`.toLowerCase()
      return text.includes(search.toLowerCase().trim())&&(status==='All'||o.status===status||(status==='Active'&&activeStatuses.includes(o.status))||(status==='Finished'&&finishedStatuses.includes(o.status)))&&(service==='All'||o.service===service)&&(scope==='All'||o.scope===scope)
    })
    return sort==='Oldest'?[...filtered].reverse():sort==='Progress'?[...filtered].sort((a,b)=>b.progress-a.progress):filtered
  },[demo.operations,scope,search,service,sort,status])
  const active=demo.operations.filter(o=>activeStatuses.includes(o.status))
  const failed=demo.operations.filter(o=>o.status==='Failed')
  const succeeded=demo.operations.filter(o=>o.status==='Succeeded')
  return <Layout>{t=><>
    <div className="lro-section-title"><div><Text as="h2" variant="header-2">{t('Мониторинг операций','Operation monitoring')}</Text><Text color="secondary" variant="body-2">{t('Статусы, прогресс, ошибки и история выполнения','Statuses, progress, failures and execution history')}</Text></div><Button view="outlined" onClick={resetOperationsDemo}><Icon data={ArrowRotateRight} size={15}/>{t('Сбросить демо','Reset demo')}</Button></div>
    <div className="lro-stats">
      <Tile label={t('Активные','Active')} value={active.length} caption={t('Очередь, выполнение, ожидание','Queued, running, waiting')} icon={BarsPlay}/>
      <Tile label={t('Завершено успешно','Succeeded')} value={succeeded.length} caption={t('Результат подтверждён','Result confirmed')} icon={Check}/>
      <Tile label={t('Ошибка','Failed')} value={failed.length} caption={t('Требуют анализа','Need attention')} icon={TriangleExclamation} tone="danger"/>
      <Tile label={t('Ожидают провайдера','Waiting')} value={demo.operations.filter(o=>o.status==='Waiting'||o.status==='CancelRequested').length} caption={t('Ожидание и отмена','Provider or cancellation')} icon={Clock} tone="warning"/>
    </div>
    <Panel title={t('Операции','Operations')} subtitle={t('Выберите операцию для просмотра шагов, журналов и ошибок','Open an operation to inspect steps, logs and errors')}
      action={<span className="lro-count">{rows.length} / {demo.operations.length}</span>}>
      <div className="lro-toolbar">
        <label className="lro-search"><span>{t('Поиск','Search')}</span><TextInput value={search} onUpdate={setSearch} startContent={<Icon data={Magnifier} size={16}/>} placeholder={t('Имя, ресурс, ID, workflow','Name, resource, ID, workflow')}/></label>
        <OptionFilter label={t('Статус','Status')} value={status} onChange={setStatus} options={[{value:'All',label:t('Все','All')},{value:'Active',label:t('Активные','Active')},{value:'Finished',label:t('Завершённые','Finished')},...(['Queued','Running','Waiting','CancelRequested','Succeeded','Failed','Cancelled'] as OperationStatus[]).map(s=>({value:s,label:t(statusRu[s],s)}))]}/>
        <OptionFilter label={t('Сервис','Service')} value={service} onChange={setService} options={[{value:'All',label:t('Все сервисы','All services')},...services.map(s=>({value:s,label:s}))]}/>
        <OptionFilter label={t('Scope','Scope')} value={scope} onChange={setScope} options={[{value:'All',label:t('Все','All')},{value:'project',label:'Project'},{value:'organization',label:'Organization'}]}/>
        <OptionFilter label={t('Сортировка','Sort')} value={sort} onChange={setSort} options={[{value:'Newest',label:t('Новые сначала','Newest first')},{value:'Oldest',label:t('Старые сначала','Oldest first')},{value:'Progress',label:t('По прогрессу','By progress')}]}/>
      </div>
      <div className="lro-table-scroll"><table className="lro-table">
        <thead><tr><th>{t('Операция','Operation')}</th><th>{t('Ресурс','Resource')}</th><th>{t('Статус','Status')}</th><th>{t('Прогресс','Progress')}</th><th>{t('Время','Duration')}</th><th>{t('Владелец','Owner')}</th><th/></tr></thead>
        <tbody>{rows.map(o=><tr key={o.id}>
          <td><Link to="/platform-operations/long-running/$operationId" params={{operationId:o.id}} className="lro-operation-link">{o.title}</Link><span className="lro-table__sub lro-mono">{o.id} · {o.service}</span></td>
          <td><strong>{o.resource}</strong><span className="lro-table__sub">{o.resourceType}</span></td>
          <td><Status status={o.status} t={t}/></td>
          <td className="lro-progress-cell"><Progress operation={o}/></td>
          <td><span className="lro-nowrap">{o.duration}</span><span className="lro-table__sub">{o.updatedAt.slice(11)}</span></td>
          <td>{o.owner}</td>
          <td><Link to="/platform-operations/long-running/$operationId" params={{operationId:o.id}} className="lro-view-link">{t('Детали','Details')}</Link></td>
        </tr>)}</tbody>
      </table></div>
      {!rows.length?<div className="lro-empty">{t('Операции не найдены. Измените фильтры.','No operations match the selected filters.')}</div>:null}
    </Panel>
    <div className="lro-bottom-grid">
      <Panel title={t('Сервисы-исполнители','Services')} subtitle={t('Количество операций в демо','Number of demo operations')}>
        <div className="lro-service-list">{services.map(s=>{
          const count=demo.operations.filter(o=>o.service===s).length
          return <div key={s}><span>{s}</span><div className="lro-service-bar"><i style={{width:`${count/demo.operations.length*100}%`}}/></div><strong>{count}</strong></div>
        })}</div>
      </Panel>
      <Panel title={t('Операционные принципы','Lifecycle guarantees')}>
        <ul className="lro-guidance">
          <li>{t('Cancel — запрос на отмену, а не мгновенное завершение.','Cancel requests termination; completion is asynchronous.')}</li>
          <li>{t('Retry создаёт новую операцию и сохраняет исходную ошибку.','Retry creates a new operation and retains the original failure.')}</li>
          <li>{t('Внешнее состояние подтверждается через reconciler.','Provider state is verified through reconciliation.')}</li>
          <li>{t('Неопределённый исход не освобождает резерв квоты.','Unknown outcomes never release quota automatically.')}</li>
        </ul>
      </Panel>
    </div>
  </>}</Layout>
}

function StepList({operation,t}:{operation:LongRunningOperation;t:T}){
  return <ol className="lro-steps">{operation.steps.map((step,index)=><li className={`lro-step lro-step_${step.status.toLowerCase()}`} key={step.id}>
    <span className="lro-step__marker">{step.status==='Succeeded'?<Icon data={Check} size={14}/>:index+1}</span>
    <div className="lro-step__text"><strong>{step.name}</strong>{step.detail?<p>{step.detail}</p>:null}</div>
    <span className="lro-step__duration">{step.duration??t(stepRu[step.status],step.status)}</span>
  </li>)}</ol>
}
function EventLog({logs,t}:{logs:OperationLog[];t:T}){
  const [levels,setLevels]=useState('All')
  const filtered=logs.filter(item=>levels==='All'||levels===item.level)
  return <>
    <div className="lro-log-toolbar"><OptionFilter label={t('Уровень журнала','Log level')} value={levels} onChange={setLevels} options={['All','INFO','WARN','ERROR'].map(s=>({value:s,label:s}))}/><span className="lro-count">{filtered.length} {t('событий','events')}</span></div>
    <div className="lro-logs" role="log">{filtered.map((entry,index)=><div key={index} className="lro-log-row"><time>{entry.at}</time><Badge tone={entry.level==='ERROR'?'danger':entry.level==='WARN'?'warning':'neutral'}>{entry.level}</Badge><span>{entry.message}</span></div>)}</div>
  </>
}
function RawDetails({operation}:{operation:LongRunningOperation}){
  const snapshot={name:`operations/${operation.id}`,metadata:{service:operation.service,action:operation.action,resource:operation.resource,workflowId:operation.workflowId,requestId:operation.requestId,correlationId:operation.correlationId,attempt:operation.attempt,parentOperationId:operation.parentOperationId},done:finishedStatuses.includes(operation.status),status:operation.status,progress:operation.progress,response:operation.result?{summary:operation.result}:undefined,error:operation.error}
  return <pre className="lro-json">{JSON.stringify(snapshot,null,2)}</pre>
}
export function LongRunningOperationDetailsPage(){
  const {operationId}=useParams({from:'/platform-operations/long-running/$operationId'})
  const demo=useOperationsDemo()
  const router=useRouter()
  const operation=demo.operations.find(o=>o.id===operationId)
  const [tab,setTab]=useState<'steps'|'events'|'json'>('steps')
  const [notice,setNotice]=useState('')
  const [confirmCancel,setConfirmCancel]=useState(false)
  const [showRawError,setShowRawError]=useState(false)
  if(!operation)return <Layout>{t=><div className="lro-empty lro-empty_full"><Text variant="header-1">{t('Операция не найдена','Operation not found')}</Text><p>{t('Проверьте идентификатор операции. Демоданные сбрасываются при перезагрузке.','Check the operation ID. Demo state resets on reload.')}</p><Link to={pageRoute} className="lro-view-link">{t('Вернуться к операциям','Back to operations')}</Link></div>}</Layout>
  const canCancel=operation.cancellable&&(['Queued','Running','Waiting'] as OperationStatus[]).includes(operation.status)
  const canAdvance=activeStatuses.includes(operation.status)
  const canRetry=operation.status==='Failed'&&operation.retryable&&operation.error?.retryable
  return <Layout detail={operation}>{t=><>
    <div className="lro-detail-top">
      <Link to={pageRoute} className="lro-back-link">‹ {t('Все операции','All operations')}</Link>
      <div className="lro-detail-actions">
        <Status status={operation.status} t={t}/>
        {canAdvance?<Button view="outlined" onClick={()=>{if(advanceDemoOperation(operation.id))setNotice(t('Демо-операция обновлена','Demo operation advanced'))}}><Icon data={BarsPlay} size={15}/>{t('Следующий шаг (демо)','Advance demo')}</Button>:null}
        {canRetry?<Button view="action" onClick={()=>{const id=retryDemoOperation(operation.id);if(id)void router.navigate({to:'/platform-operations/long-running/$operationId',params:{operationId:id}})}}><Icon data={ArrowRotateRight} size={16}/>{t('Создать повтор (демо)','Retry as new operation')}</Button>:null}
        {canCancel?<Button view="outlined-danger" onClick={()=>setConfirmCancel(true)}><Icon data={OctagonXmark} size={16}/>{t('Запросить отмену','Request cancellation')}</Button>:null}
      </div>
    </div>
    {notice?<div className="lro-notice" role="status">{notice}</div>:null}
    <div className="lro-detail-grid">
      <Panel title={t('Текущий статус','Current state')} subtitle={`${t('Обновлено','Updated')}: ${operation.updatedAt}`}>
        <div className="lro-state-summary"><div className="lro-state-summary__meta"><strong>{t('Выполнение','Execution')}</strong><span>{operation.progress}%</span></div><Progress operation={operation}/></div>
        <dl className="lro-properties">
          <Property label={t('Операция','Operation ID')} value={<FormatTime value={operation.id}/>}/>
          <Property label={t('Сервис','Service')} value={operation.service}/>
          <Property label={t('Действие','Action')} value={operation.action}/>
          <Property label={t('Ресурс','Resource')} value={`${operation.resourceType} / ${operation.resource}`}/>
          <Property label="Provider" value={operation.provider}/>
          <Property label={t('Область','Scope')} value={operation.scope}/>
          <Property label={t('Создано','Submitted')} value={operation.submittedAt}/>
          <Property label={t('Длительность','Duration')} value={operation.duration}/>
        </dl>
      </Panel>
      <Panel title={t('Контекст выполнения','Execution context')} subtitle={t('Технические идентификаторы и связанные ресурсы','Technical correlation and related resources')}>
        <dl className="lro-properties">
          <Property label="Temporal Workflow ID" value={<FormatTime value={operation.workflowId}/>}/>
          <Property label="Request ID" value={<FormatTime value={operation.requestId}/>}/>
          <Property label="Correlation ID" value={<FormatTime value={operation.correlationId}/>}/>
          <Property label={t('Попытка','Attempt')} value={operation.attempt}/>
          <Property label={t('Инициатор','Owner')} value={operation.owner}/>
          {operation.parentOperationId?<Property label={t('Исходная операция','Retry of')} value={<Link to="/platform-operations/long-running/$operationId" params={{operationId:operation.parentOperationId}}>{operation.parentOperationId}</Link>}/>:null}
          {operation.reservationId?<Property label={t('Резерв квоты','Quota reservation')} value={<Link to="/quota-manager/reservations">{operation.reservationId}</Link>}/>:null}
        </dl>
        {operation.error?<div className="lro-error"><div className="lro-error__header"><Icon data={TriangleExclamation} size={18}/><strong>{operation.error.code}</strong></div><p>{operation.error.message}</p><Button view="flat" onClick={()=>setShowRawError(!showRawError)}><Icon data={Code} size={14}/>{showRawError?t('Скрыть ошибку','Hide error'):t('Показать JSON ошибки','View error JSON')}</Button>{showRawError?<pre>{JSON.stringify(operation.error,null,2)}</pre>:null}</div>:null}
        {operation.result?<div className="lro-success"><Icon data={Check} size={16}/>{operation.result}</div>:null}
      </Panel>
    </div>
    <Panel title={t('Выполнение операции','Operation execution')} subtitle={t('Этапы выполнения, журнал событий и Operation JSON','Execution stages, event log and operation JSON')}>
      <div className="lro-inner-tabs" role="tablist">
        <button type="button" role="tab" aria-selected={tab==='steps'} className={tab==='steps'?'lro-inner-tab_active':''} onClick={()=>setTab('steps')}>{t('Этапы','Steps')} ({operation.steps.length})</button>
        <button type="button" role="tab" aria-selected={tab==='events'} className={tab==='events'?'lro-inner-tab_active':''} onClick={()=>setTab('events')}>{t('События','Events')} ({operation.logs.length})</button>
        <button type="button" role="tab" aria-selected={tab==='json'} className={tab==='json'?'lro-inner-tab_active':''} onClick={()=>setTab('json')}>JSON</button>
      </div>
      {tab==='steps'?<StepList operation={operation} t={t}/>:null}
      {tab==='events'?<EventLog logs={operation.logs} t={t}/>:null}
      {tab==='json'?<RawDetails operation={operation}/>:null}
    </Panel>
    <div className="lro-demo-notice"><Icon data={Clock} size={16}/>{t('Состояние провайдера проверяется отдельно от состояния Workflow. Для операций с неопределённым исходом квоты не освобождаются автоматически.','Provider state is verified separately from workflow state. Quota is not automatically released for uncertain outcomes.')}</div>
    {confirmCancel?<div className="lro-modal-overlay" role="presentation"><div className="lro-modal" role="alertdialog" aria-modal="true" aria-labelledby="lro-cancel-title" aria-describedby="lro-cancel-description">
      <Text as="h2" id="lro-cancel-title" variant="header-2">{t('Запросить отмену операции?','Request cancellation?')}</Text>
      <Text as="p" id="lro-cancel-description" color="secondary">{t('Отмена асинхронная и не гарантирует мгновенное прекращение provisioning. Подтверждение выполняется в демонстрационной модели.','Cancellation is asynchronous and does not guarantee immediate termination of provisioning. This updates demo state only.')}</Text>
      <div className="lro-modal__actions"><Button view="outlined" onClick={()=>setConfirmCancel(false)}>{t('Назад','Back')}</Button><Button view="outlined-danger" onClick={()=>{if(cancelDemoOperation(operation.id))setNotice(t('Отмена запрошена — ожидается подтверждение workflow','Cancellation requested; awaiting workflow confirmation'));setConfirmCancel(false)}}>{t('Запросить отмену (демо)','Request cancellation (demo)')}</Button></div>
    </div></div>:null}
  </>}</Layout>
}
