import {useEffect, useState} from 'react'
import type {ReactNode} from 'react'
import {ArrowRotateRight, Check, Clock, Magnifier, Speedometer, TriangleExclamation} from '@gravity-ui/icons'
import {Button, Card, Icon, Text, TextInput} from '@gravity-ui/uikit'
import {Link} from '@tanstack/react-router'

import {ConsoleBreadcrumbs} from '../../components/ConsoleBreadcrumbs'
import {useConsoleI18n} from '../../App'
import {
  applyDemoRequest,
  createDemoRequest,
  inspectDemoReservation,
  reviewDemoFinding,
  reviewDemoRequest,
  runDemoScan,
  setDemoScope,
  updateDemoPolicy,
  useQuotaDemo,
  validateQuotaLimit,
} from './demo'
import type {Finding, Quota, QuotaCategory, QuotaMode, QuotaRequest, QuotaScope, Reservation} from './demo'

import './quota-manager.css'

const tabs = [
  {id:'overview',path:'/quota-manager/overview',en:'Overview',ru:'Обзор'},
  {id:'policies',path:'/quota-manager/policies',en:'Quota Policies',ru:'Политики квот'},
  {id:'reservations',path:'/quota-manager/reservations',en:'Reservations',ru:'Резервирования'},
  {id:'requests',path:'/quota-manager/requests',en:'Quota Requests',ru:'Заявки на квоты'},
  {id:'reconciliation',path:'/quota-manager/reconciliation',en:'Reconciliation',ru:'Сверка'},
] as const
type TabId = typeof tabs[number]['id']
type T = (ru: string, en: string) => string
const number = (n: number) => n.toLocaleString('en-US')
const remaining = (q: Quota) => q.limit - q.used - q.reserved
const percent = (numerator: number, denominator: number) => denominator > 0 ? Math.max(0, Math.min(100, numerator / denominator * 100)) : 0

function Tag({children,tone='neutral'}:{children:ReactNode;tone?:'neutral'|'info'|'success'|'warning'|'danger'}) {
  return <span className={`qm-tag qm-tag_${tone}`}>{children}</span>
}
function StatusTag({status}:{status:string}) {
  const tone = status==='Critical'||status==='Rejected'?'danger':status==='Warning'||status==='Reconciling'?'warning':status==='Committed'||status==='Applied'||status==='Healthy'||status==='Resolved'?'success':status==='Approved'?'info':'neutral'
  return <Tag tone={tone}>{status}</Tag>
}
function Panel({title,subtitle,actions,children}:{title:string;subtitle?:string;actions?:ReactNode;children:ReactNode}) {
  return <Card view="outlined" className="qm-panel">
    <div className="qm-panel__head">
      <div><Text as="h2" variant="header-1">{title}</Text>{subtitle?<Text as="p" variant="caption-2" color="secondary">{subtitle}</Text>:null}</div>
      {actions}
    </div>
    {children}
  </Card>
}
function Empty({message}:{message:string}) {
  return <div className="qm-empty"><Text color="secondary">{message}</Text></div>
}
function Stat({label,value,note,tone='default'}:{label:string;value:string|number;note?:string;tone?:string}) {
  return <Card view="outlined" className="qm-stat">
    <Text variant="caption-2" color="secondary">{label}</Text>
    <span className={`qm-stat__value qm-stat__value_${tone}`}>{value}</span>
    {note?<Text variant="caption-2" color="secondary">{note}</Text>:null}
  </Card>
}
function UsageBar({quota}:{quota:Quota}) {
  const used = percent(quota.used,quota.limit)
  const reserved = Math.min(100-used,percent(quota.reserved,quota.limit))
  return <div className="qm-progress" role="meter" aria-label={quota.title} aria-valuenow={quota.used+quota.reserved} aria-valuemin={0} aria-valuemax={Math.max(quota.limit,quota.used+quota.reserved)}>
    <span className="qm-progress__used" style={{width:`${used}%`}}/>
    <span className="qm-progress__reserved" style={{width:`${reserved}%`}}/>
  </div>
}
function Drawer({title,subtitle,onClose,children}:{title:string;subtitle?:string;onClose:()=>void;children:ReactNode}) {
  useEffect(()=>{
    function keydown(event:KeyboardEvent) { if(event.key==='Escape') onClose() }
    window.addEventListener('keydown',keydown)
    return ()=>window.removeEventListener('keydown',keydown)
  },[onClose])
  return <div className="qm-drawer-host">
    <button className="qm-drawer-backdrop" type="button" aria-label="Close details" onClick={onClose}/>
    <aside className="qm-drawer" role="dialog" aria-modal="true" aria-label={title}>
      <header className="qm-drawer__head">
        <div><Text as="h2" variant="header-2">{title}</Text>{subtitle?<Text as="p" variant="caption-2" color="secondary">{subtitle}</Text>:null}</div>
        <Button view="outlined" onClick={onClose} aria-label="Close">✕</Button>
      </header>
      <div className="qm-drawer__body">{children}</div>
    </aside>
  </div>
}
function Detail({label,value}:{label:string;value:ReactNode}) {
  return <div className="qm-detail"><Text variant="caption-2" color="secondary">{label}</Text><div>{value}</div></div>
}
function Filter({label,value,options,onChange}:{label:string;value:string;options:Array<{value:string;label:string}>;onChange:(v:string)=>void}) {
  return <label className="qm-filter"><span>{label}</span><select className="qm-select" value={value} onChange={e=>onChange(e.target.value)}>{options.map(o=><option key={o.value} value={o.value}>{o.label}</option>)}</select></label>
}
function QuotaHeader({tab,t}:{tab:TabId;t:T}) {
  const demo = useQuotaDemo()
  const selected = tabs.find(item=>item.id===tab) ?? tabs[0]
  return <div className="qm-heading">
    <div>
      <ConsoleBreadcrumbs items={[{text:t('Операции платформы','Platform Operations')},{text:'Quota Manager',href:'/quota-manager/overview'},{text:t(selected.ru,selected.en)}]}/>
      <div className="qm-heading__title"><Text as="h1" variant="display-1">Quota Manager</Text><Tag tone="info">DEMO DATA</Tag></div>
      <Text as="p" variant="body-2" color="secondary">{t('Управление квотами, резервами и сверкой ресурсов M8','Control quotas, reservations and provider reconciliation across M8 resources')}</Text>
    </div>
    <div className="qm-heading__scope">
      <Filter label={t('Контекст ресурсов','Resource scope')} value={demo.scope} onChange={value=>setDemoScope(value as QuotaScope)}
        options={[{value:'project',label:'Acme / Production / AI Studio'},{value:'organization',label:'Acme (Organization)'}]}/>
      <div className="qm-provider"><span className="qm-provider__dot"/><span>Cozystack · eu-1</span></div>
    </div>
  </div>
}
function Layout({tab,children}:{tab:TabId;children:(t:T)=>ReactNode}) {
  const {language} = useConsoleI18n()
  const t:T = (ru,en)=>language==='ru'?ru:en
  return <main className="m8-page__body"><section className="m8-page__content qm-page">
    <QuotaHeader tab={tab} t={t}/>
    <nav className="qm-tabs" aria-label="Quota Manager pages">{tabs.map(item=><Link key={item.id} to={item.path} className={`qm-tab ${item.id===tab?'qm-tab_active':''}`} aria-current={item.id===tab?'page':undefined}>{t(item.ru,item.en)}</Link>)}</nav>
    <div className="qm-demo-notice"><Icon data={Speedometer} size={16}/>{t('Демоданные: изменения действуют только до перезагрузки вкладки; Kubernetes, YDB и Temporal не вызываются.','Demo data: changes last until the page reload; no Kubernetes, YDB or Temporal calls are made.')}</div>
    {children(t)}
  </section></main>
}
function QuotaTable({quotas,onOpen,t}:{quotas:Quota[];onOpen:(id:string)=>void;t:T}) {
  return <div className="qm-table-scroll"><table className="qm-table">
    <thead><tr><th>{t('Квота','Quota')}</th><th>{t('Использование','Utilization')}</th><th>{t('Лимит','Limit')}</th><th>{t('Доступно','Available')}</th><th>{t('Источник','Source')}</th><th/></tr></thead>
    <tbody>{quotas.map(q=><tr key={q.id}>
      <td><strong>{q.title}</strong><span className="qm-muted qm-code">{q.id}</span></td>
      <td className="qm-utilization"><div className="qm-utilization__value">{number(q.used)} <span className="qm-muted">+ {number(q.reserved)} {t('резерв','reserved')}</span></div><UsageBar quota={q}/></td>
      <td className="qm-nowrap">{number(q.limit)} <span className="qm-muted">{q.unit}</span></td>
      <td className={remaining(q)<0?'qm-danger':''}>{number(remaining(q))}</td>
      <td><Tag>{q.source}</Tag></td>
      <td><Button size="s" view="outlined" onClick={()=>onOpen(q.id)}>{t('Детали','Details')}</Button></td>
    </tr>)}</tbody></table></div>
}
function MiniLegend({t}:{t:T}) {
  return <div className="qm-legend"><span><i className="qm-legend__used"/>{t('Использовано','Used')}</span><span><i className="qm-legend__reserved"/>{t('Зарезервировано','Reserved')}</span><span><i className="qm-legend__available"/>{t('Свободно','Available')}</span></div>
}

export function QuotaOverviewPage() {
  const demo=useQuotaDemo()
  const [category,setCategory]=useState('All')
  const [opened,setOpened]=useState<string|null>(null)
  const quotas=demo.quotas.filter(q=>q.scope===demo.scope)
  const visible=quotas.filter(q=>category==='All'||q.category===category)
  const alerts=quotas.filter(q=>q.used+q.reserved>=q.limit*.8)
  return <Layout tab="overview">{t=><>
    <div className="qm-section-heading"><div><Text as="h2" variant="header-2">{t('Обзор квот','Quota Overview')}</Text><Text color="secondary" variant="body-2">{t('Эффективные лимиты и потребление выбранного scope','Effective limits and allocations in the selected scope')}</Text></div><Link to="/quota-manager/requests" className="qm-primary-link">{t('Запросить увеличение','Request increase')}</Link></div>
    <div className="qm-kpis">
      <Stat label={t('Активные квоты','Tracked quotas')} value={quotas.length} note={t('По всем категориям','Across all categories')}/>
      <Stat label={t('Приближаются к лимиту','Near limit')} value={alerts.length} note={t('Использование ≥ 80%','Usage ≥ 80%')} tone={alerts.length?'warning':'default'}/>
      <Stat label={t('Открытые заявки','Open requests')} value={demo.requests.filter(r=>r.scope===demo.scope&&r.status==='Pending').length} note={t('Ожидают согласования','Pending approval')}/>
      <Stat label={t('Найденные расхождения','Drift findings')} value={demo.findings.filter(f=>f.status==='Critical'||f.status==='Warning').length} note={t('Требуют внимания','Need attention')} tone="warning"/>
    </div>
    <Panel title={t('Использование ресурсов','Resource utilization')} subtitle={t('Used + Reserved / Effective limit','Used + Reserved / Effective limit')} actions={<Filter label={t('Категория','Category')} value={category} onChange={setCategory} options={[{value:'All',label:t('Все','All')},...(['Compute','Storage','Network','AI'] as QuotaCategory[]).map(x=>({value:x,label:x}))]}/>}>
      <div className="qm-bars">{visible.map(q=><div key={q.id} className="qm-bars__row"><div className="qm-bars__meta"><div><strong>{q.title}</strong><span className="qm-muted qm-code">{q.id}</span></div><span className="qm-bars__count">{number(q.used+q.reserved)} / {number(q.limit)} {q.unit}</span></div><UsageBar quota={q}/><div className="qm-bars__meta qm-bars__footer"><span>{t('Резерв','Reserved')}: {number(q.reserved)}</span><span className={remaining(q)<0?'qm-danger':''}>{t('Доступно','Available')}: {number(remaining(q))}</span></div></div>)}</div>
      <MiniLegend t={t}/>
    </Panel>
    <div className="qm-columns">
      <Panel title={t('Распределение квот','Quota allocation')} subtitle={t('Пример распределения между дочерними scope','Example distribution across child scopes')}>
        <div className="qm-allocation">
          {(demo.scope==='organization'?[{name:'Production / AI Studio',part:40},{name:'Development',part:25},{name:'Shared / unallocated',part:35}]:[{name:'AI Gateway',part:45},{name:'Inference workloads',part:35},{name:'Other services / free',part:20}]).map((x,i)=><div key={x.name} className="qm-allocation__row"><span className={`qm-allocation__dot qm-allocation__dot_${i}`}/><span>{x.name}</span><div className="qm-allocation__track"><span style={{width:`${x.part}%`}}/></div><strong>{x.part}%</strong></div>)}
        </div>
      </Panel>
      <Panel title={t('Последние события','Recent activity')} subtitle={t('Действия только в демо-сеансе','Demo-session activity only')}>
        <ul className="qm-activity">{demo.activity.slice(0,5).map((item,i)=><li key={i}><span className="qm-activity__dot"/>{item}</li>)}</ul>
      </Panel>
    </div>
    <Panel title={t('Все квоты','All quotas')} subtitle={t('Детализация effective policy и резерва','Effective policy and reservation breakdown')}><QuotaTable quotas={quotas} onOpen={setOpened} t={t}/></Panel>
    {opened&&quotas.find(q=>q.id===opened)?<Drawer title={quotas.find(q=>q.id===opened)!.title} subtitle={opened} onClose={()=>setOpened(null)}>
      {(()=>{const q=quotas.find(q=>q.id===opened)!;return <div className="qm-details"><Detail label={t('Область','Scope')} value={demo.scope}/><Detail label={t('Лимит','Limit')} value={`${number(q.limit)} ${q.unit}`}/><Detail label={t('Использовано','Used')} value={number(q.used)}/><Detail label={t('Зарезервировано','Reserved')} value={number(q.reserved)}/><Detail label={t('Доступно','Available')} value={number(remaining(q))}/><Detail label={t('Режим','Mode')} value={q.mode}/><Detail label={t('Источник','Source')} value={q.source}/><Text color="secondary">{q.description}</Text></div>})()}
    </Drawer>:null}
  </>}</Layout>
}

function PolicyEditor({quota,onClose,t}:{quota:Quota;onClose:()=>void;t:T}) {
  const [limit,setLimit]=useState(String(quota.limit))
  const [mode,setMode]=useState<QuotaMode>(quota.mode)
  const [error,setError]=useState('')
  const value=Number(limit)
  const demo=useQuotaDemo()
  const previewError=validateQuotaLimit(quota.scope,quota.id,value,mode)
  const parent=demo.quotas.find(q=>q.scope==='organization'&&q.id===quota.id)
  const overQuota=Number.isFinite(value)&&value<quota.used+quota.reserved
  return <Drawer title={t('Редактирование политики','Edit Quota Policy')} subtitle={quota.id} onClose={onClose}>
    <div className="qm-form">
      <Detail label={t('Область применения','Scope')} value={quota.scope==='project'?'Acme / Production / AI Studio':'Acme (Organization)'}/>
      <Detail label={t('Quota dimension','Quota dimension')} value={quota.id}/>
      <label className="qm-field"><span>{t('Режим распределения','Allocation mode')}</span><select className="qm-select" value={mode} onChange={e=>setMode(e.target.value as QuotaMode)}><option value="Dedicated">Dedicated</option><option value="Shared">Shared</option></select></label>
      <label className="qm-field"><span>{t('Новый лимит','New quota limit')} ({quota.unit})</span><TextInput type="number" value={limit} onUpdate={setLimit}/></label>
      <div className="qm-impact"><strong>{t('Предварительная проверка','Impact preview')}</strong>
        <Detail label={t('Текущий лимит','Current limit')} value={number(quota.limit)}/>
        <Detail label={t('Изменение','Delta')} value={Number.isFinite(value)?`${value-quota.limit>=0?'+':''}${number(value-quota.limit)} ${quota.unit}`:'—'}/>
        {quota.scope==='project'&&parent?<Detail label={t('Бюджет родителя','Parent budget')} value={`${number(parent.limit)} ${quota.unit}`}/>:null}
        {overQuota?<p className="qm-alert qm-alert_warning">{t('Уменьшение ниже текущего использования создаст состояние Over quota. Ресурсы не удаляются.','Reducing below current usage creates Over quota. No resources are deleted.')}</p>:null}
        {previewError?<p className="qm-alert qm-alert_danger" role="alert">{previewError}</p>:null}
      </div>
      {error?<p className="qm-alert qm-alert_danger" role="alert">{error}</p>:null}
      <div className="qm-form__actions"><Button view="outlined" onClick={onClose}>{t('Отмена','Cancel')}</Button><Button view="action" disabled={!!previewError} onClick={()=>{const issue=updateDemoPolicy(quota.scope,quota.id,value,mode);if(issue)setError(issue);else onClose()}}>{t('Сохранить в демо','Save demo policy')}</Button></div>
      <Text variant="caption-2" color="secondary">{t('Изменяется только локальная модель, без запросов к Quota Manager API.','This updates local demo state only, not the Quota Manager API.')}</Text>
    </div>
  </Drawer>
}
export function QuotaPoliciesPage() {
  const demo=useQuotaDemo()
  const [search,setSearch]=useState('')
  const [category,setCategory]=useState('All')
  const [edit,setEdit]=useState<string|null>(null)
  const quotas=demo.quotas.filter(q=>q.scope===demo.scope)
  const visible=quotas.filter(q=>(category==='All'||q.category===category)&&`${q.title} ${q.id}`.toLowerCase().includes(search.toLowerCase()))
  const quota=quotas.find(q=>q.id===edit)
  return <Layout tab="policies">{t=><>
    <div className="qm-section-heading"><div><Text as="h2" variant="header-2">{t('Политики квот','Quota Policies')}</Text><Text color="secondary" variant="body-2">{t('Лимиты, наследование и выделенные бюджеты','Limits, inheritance and allocated budgets')}</Text></div><Tag tone="info">{t('Demo policy engine','Demo policy engine')}</Tag></div>
    <div className="qm-kpis"><Stat label={t('Прямые политики','Direct policies')} value={quotas.filter(q=>q.source==='Direct').length}/><Stat label={t('Наследуемые политики','Inherited policies')} value={quotas.filter(q=>q.source==='Inherited').length}/><Stat label={t('Выделенные квоты','Dedicated quotas')} value={quotas.filter(q=>q.mode==='Dedicated').length}/><Stat label={t('Общие квоты','Shared quotas')} value={quotas.filter(q=>q.mode==='Shared').length}/></div>
    <Panel title={t('Каталог политик','Policy catalog')} subtitle={t('Нажмите Изменить, чтобы открыть редактор с проверкой родительского бюджета','Select Edit for impact preview and parent budget validation')}
      actions={<div className="qm-toolbar"><TextInput startContent={<Icon data={Magnifier}/>} value={search} onUpdate={setSearch} placeholder={t('Поиск квоты','Search quota')}/><Filter label={t('Категория','Category')} value={category} onChange={setCategory} options={[{value:'All',label:t('Все','All')},...(['Compute','Storage','Network','AI'] as QuotaCategory[]).map(x=>({value:x,label:x}))]}/></div>}>
      <div className="qm-table-scroll"><table className="qm-table">
        <thead><tr><th>{t('Квота','Quota')}</th><th>{t('Режим','Mode')}</th><th>{t('Источник','Source')}</th><th>{t('Использовано','Used')}</th><th>{t('Лимит','Limit')}</th><th>{t('Статус','Status')}</th><th/></tr></thead>
        <tbody>{visible.map(q=><tr key={q.id}>
          <td><strong>{q.title}</strong><span className="qm-muted qm-code">{q.id}</span></td><td>{q.mode}</td><td><Tag>{q.source}</Tag></td><td>{number(q.used+q.reserved)}</td>
          <td>{number(q.limit)} <span className="qm-muted">{q.unit}</span></td>
          <td><StatusTag status={remaining(q)<0?'Warning':'Healthy'}/></td>
          <td><Button size="s" view="outlined" onClick={()=>setEdit(q.id)}>{t('Изменить','Edit')}</Button></td>
        </tr>)}</tbody>
      </table></div>
      {!visible.length?<Empty message={t('Политики не найдены','No policies match the filters')}/>:null}
    </Panel>
    <Panel title={t('Правила наследования','Inheritance rules')}><div className="qm-guidance"><div><strong>Dedicated</strong><p>{t('Гарантированная доля бюджета родителя. Сумма выделений не может превышать лимит родителя.','Reserved share of the parent budget. Child allocations cannot exceed the parent limit.')}</p></div><div><strong>Shared</strong><p>{t('Расходует общий пул родителя; лимиты проверяются при admission.','Consumes the shared parent pool and is checked at admission.')}</p></div></div></Panel>
    {quota?<PolicyEditor key={`${quota.scope}:${quota.id}`} quota={quota} t={t} onClose={()=>setEdit(null)}/>:null}
  </>}</Layout>
}

function ReservationDetails({reservation,onClose,t}:{reservation:Reservation;onClose:()=>void;t:T}) {
  return <Drawer title={reservation.id} subtitle={reservation.resource} onClose={onClose}>
    <div className="qm-details"><Detail label={t('Статус','Status')} value={<StatusTag status={reservation.status}/>}/><Detail label={t('Ресурс','Resource')} value={`${reservation.kind} / ${reservation.resource}`}/><Detail label={t('Объём резерва','Requested quota')} value={reservation.quantity}/><Detail label="Temporal Operation" value={reservation.operation}/><Detail label="Provider" value={reservation.provider}/><Detail label={t('Возраст','Age')} value={reservation.age}/></div>
    <div className="qm-timeline"><Text as="h3" variant="header-1">{t('Жизненный цикл','Lifecycle')}</Text><ol>{reservation.timeline.map((event,i)=><li key={i}>{event}</li>)}</ol></div>
    {reservation.status==='Reconciling'?<div className="qm-form__actions"><Button view="action" onClick={()=>inspectDemoReservation(reservation.id)}><Icon data={ArrowRotateRight}/> {t('Проверить повторно (демо)','Retry inspection (demo)')}</Button></div>:null}
    <p className="qm-alert qm-alert_warning">{t('Принудительное освобождение резерва не предусмотрено: результат provisioning должен быть подтверждён.','Force release is disabled: the provisioning result must be confirmed first.')}</p>
  </Drawer>
}
export function QuotaReservationsPage() {
  const demo=useQuotaDemo()
  const [filter,setFilter]=useState('All')
  const [search,setSearch]=useState('')
  const [opened,setOpened]=useState<string|null>(null)
  const filtered=demo.reservations.filter(r=>(filter==='All'||r.status===filter)&&`${r.id} ${r.resource} ${r.operation}`.toLowerCase().includes(search.toLowerCase()))
  const selected=demo.reservations.find(r=>r.id===opened)
  return <Layout tab="reservations">{t=><>
    <div className="qm-section-heading"><div><Text as="h2" variant="header-2">{t('Резервирования квот','Quota Reservations')}</Text><Text color="secondary" variant="body-2">{t('Provisioning, идемпотентность и зависшие резервы','Provisioning, idempotency and uncertain reservations')}</Text></div></div>
    <div className="qm-kpis"><Stat label="Reserved" value={demo.reservations.filter(r=>r.status==='Reserved').length}/><Stat label="Reconciling" value={demo.reservations.filter(r=>r.status==='Reconciling').length} tone="warning"/><Stat label="Committed" value={demo.reservations.filter(r=>r.status==='Committed').length}/><Stat label="Released" value={demo.reservations.filter(r=>r.status==='Released').length}/></div>
    <Panel title={t('Операции и резервирования','Operations and reservations')} subtitle={t('Статусы демонстрационные. События Temporal не загружаются.','Statuses are simulated; Temporal events are not queried.')} actions={<div className="qm-toolbar"><TextInput startContent={<Icon data={Magnifier}/>} value={search} onUpdate={setSearch} placeholder={t('Ресурс или operation ID','Resource or operation ID')}/><Filter label={t('Статус','Status')} value={filter} onChange={setFilter} options={['All','Reserved','Reconciling','Committed','Released'].map(s=>({value:s,label:s}))}/></div>}>
      <div className="qm-table-scroll"><table className="qm-table"><thead><tr><th>ID</th><th>{t('Ресурс','Resource')}</th><th>{t('Квота','Quota')}</th><th>{t('Статус','Status')}</th><th>{t('Возраст','Age')}</th><th/></tr></thead><tbody>{filtered.map(r=><tr key={r.id}><td className="qm-code">{r.id}</td><td><strong>{r.resource}</strong><span className="qm-muted">{r.kind}</span></td><td>{r.quantity}</td><td><StatusTag status={r.status}/></td><td>{r.age}</td><td><Button view="outlined" size="s" onClick={()=>setOpened(r.id)}>{t('Детали','Details')}</Button></td></tr>)}</tbody></table></div>
      {!filtered.length?<Empty message={t('Резервирования не найдены','No reservations found')}/>:null}
    </Panel>
    <div className="qm-demo-notice"><Icon data={Clock} size={16}/>{t('Истечение TTL само по себе не снимает резерв. Для неизвестного исхода сначала выполняется reconciliation.','A reservation TTL does not automatically release quota. Unknown outcomes must be reconciled first.')}</div>
    {selected?<ReservationDetails reservation={selected} t={t} onClose={()=>setOpened(null)}/>:null}
  </>}</Layout>
}

function RequestForm({onClose,t}:{onClose:()=>void;t:T}) {
  const demo=useQuotaDemo()
  const [quotaId,setQuotaId]=useState('compute.vcpu.allocated')
  const current=demo.quotas.find(q=>q.scope===demo.scope&&q.id===quotaId)
  const [limit,setLimit]=useState('')
  const [reason,setReason]=useState('')
  const [error,setError]=useState('')
  return <Drawer title={t('Запросить увеличение','Request quota increase')} subtitle={t('Демо-заявка, без отправки на сервер','Demo request: not sent to the server')} onClose={onClose}>
    <form className="qm-form" onSubmit={e=>{e.preventDefault();const issue=createDemoRequest(demo.scope,quotaId,Number(limit),reason);if(issue)setError(issue);else onClose()}}>
      <Detail label={t('Область','Scope')} value={demo.scope==='project'?'Acme / Production / AI Studio':'Acme'}/>
      <label className="qm-field"><span>{t('Измерение квоты','Quota dimension')}</span><select className="qm-select" value={quotaId} onChange={e=>{setQuotaId(e.target.value);setLimit('');setError('')}}>{demo.quotas.filter(q=>q.scope===demo.scope).map(q=><option key={q.id} value={q.id}>{q.title} — {q.id}</option>)}</select></label>
      <Detail label={t('Текущий лимит','Current limit')} value={current?`${number(current.limit)} ${current.unit}`:'—'}/>
      <label className="qm-field"><span>{t('Запрашиваемый лимит','Requested limit')} ({current?.unit})</span><TextInput type="number" value={limit} onUpdate={setLimit} placeholder={current?String(current.limit+10):'64'}/></label>
      <label className="qm-field"><span>{t('Обоснование','Business justification')}</span><textarea className="qm-textarea" rows={5} value={reason} onChange={e=>setReason(e.target.value)} placeholder={t('Опишите планируемую нагрузку и необходимость квоты','Describe expected workload and business need')}/></label>
      {error?<div className="qm-alert qm-alert_danger" role="alert">{error}</div>:null}
      <div className="qm-form__actions"><Button view="outlined" onClick={onClose}>{t('Отмена','Cancel')}</Button><Button type="submit" view="action">{t('Создать демо-заявку','Create demo request')}</Button></div>
    </form>
  </Drawer>
}
function RequestDetails({request,onClose,t}:{request:QuotaRequest;onClose:()=>void;t:T}) {
  const demo=useQuotaDemo()
  const [error,setError]=useState('')
  const quota=demo.quotas.find(q=>q.scope===request.scope&&q.id===request.quotaId)
  return <Drawer title={request.id} subtitle={request.quotaId} onClose={onClose}>
    <div className="qm-details"><Detail label={t('Статус','Status')} value={<StatusTag status={request.status}/>}/><Detail label={t('Область','Scope')} value={request.scope}/><Detail label={t('Запросил','Requested by')} value={request.requester}/><Detail label={t('Дата','Date')} value={request.created}/><Detail label={t('Текущий лимит','Current limit')} value={quota?`${number(quota.limit)} ${quota.unit}`:'—'}/><Detail label={t('Запрошенный лимит','Requested limit')} value={number(request.requestedLimit)}/><Detail label={t('Обоснование','Justification')} value={request.reason}/></div>
    {request.status==='Pending'?<div className="qm-form__actions"><Button view="outlined" onClick={()=>reviewDemoRequest(request.id,false)}>{t('Отклонить (демо)','Reject (demo)')}</Button><Button view="action" onClick={()=>reviewDemoRequest(request.id,true)}>{t('Согласовать (демо)','Approve (demo)')}</Button></div>:null}
    {request.status==='Approved'?<div className="qm-form__actions"><Button view="action" onClick={()=>setError(applyDemoRequest(request.id)??'')}>{t('Применить политику (демо)','Apply policy (demo)')}</Button></div>:null}
    {error?<p className="qm-alert qm-alert_danger" role="alert">{error}</p>:null}
    <p className="qm-alert qm-alert_warning">{t('Approved не означает Applied. Перед применением лимит и бюджет родителя проверяются заново.','Approved is not Applied. The quota and parent budget are revalidated before applying.')}</p>
  </Drawer>
}
export function QuotaRequestsPage() {
  const demo=useQuotaDemo()
  const [filter,setFilter]=useState('All')
  const [opened,setOpened]=useState<string|null>(null)
  const [creating,setCreating]=useState(false)
  const filtered=demo.requests.filter(r=>(filter==='All'||r.status===filter)&&(r.scope===demo.scope))
  const selected=demo.requests.find(r=>r.id===opened)
  return <Layout tab="requests">{t=><>
    <div className="qm-section-heading"><div><Text as="h2" variant="header-2">{t('Заявки на квоты','Quota Requests')}</Text><Text color="secondary" variant="body-2">{t('Создание, согласование и применение новых лимитов','Create, approve and apply quota changes')}</Text></div><Button view="action" onClick={()=>setCreating(true)}>+ {t('Новая заявка','New request')}</Button></div>
    <div className="qm-kpis"><Stat label="Pending" value={demo.requests.filter(r=>r.scope===demo.scope&&r.status==='Pending').length}/><Stat label="Approved" value={demo.requests.filter(r=>r.scope===demo.scope&&r.status==='Approved').length}/><Stat label="Applied" value={demo.requests.filter(r=>r.scope===demo.scope&&r.status==='Applied').length}/><Stat label="Rejected" value={demo.requests.filter(r=>r.scope===demo.scope&&r.status==='Rejected').length}/></div>
    <Panel title={t('Очередь заявок','Request queue')} actions={<Filter label={t('Статус','Status')} value={filter} onChange={setFilter} options={['All','Pending','Approved','Applied','Rejected'].map(s=>({value:s,label:s}))}/>}>
      <div className="qm-table-scroll"><table className="qm-table"><thead><tr><th>{t('Заявка','Request')}</th><th>{t('Квота','Quota')}</th><th>{t('Изменение','Change')}</th><th>{t('Инициатор','Requester')}</th><th>{t('Статус','Status')}</th><th/></tr></thead><tbody>{filtered.map(r=>{const q=demo.quotas.find(x=>x.scope===r.scope&&x.id===r.quotaId);return <tr key={r.id}><td><strong>{r.id}</strong><span className="qm-muted">{r.created}</span></td><td>{q?.title??r.quotaId}</td><td className="qm-nowrap">{q?number(q.limit):'—'} → {number(r.requestedLimit)}</td><td>{r.requester}</td><td><StatusTag status={r.status}/></td><td><Button size="s" view="outlined" onClick={()=>setOpened(r.id)}>{t('Открыть','Open')}</Button></td></tr>})}</tbody></table></div>
      {!filtered.length?<Empty message={t('Нет заявок в выбранном фильтре','No requests match this filter')}/>:null}
    </Panel>
    {creating?<RequestForm t={t} onClose={()=>setCreating(false)}/>:null}
    {selected?<RequestDetails request={selected} t={t} onClose={()=>setOpened(null)}/>:null}
  </>}</Layout>
}

function FindingDetails({finding,onClose,t}:{finding:Finding;onClose:()=>void;t:T}) {
  return <Drawer title={finding.id} subtitle={finding.resource} onClose={onClose}>
    <div className="qm-details"><Detail label={t('Состояние','Status')} value={<StatusTag status={finding.status}/>}/><Detail label={t('Тип ресурса','Resource kind')} value={finding.kind}/><Detail label="Provider" value={finding.provider}/><Detail label={t('Причина','Analysis')} value={finding.detail}/></div>
    <div className="qm-diff"><div><span>{t('M8 desired / recorded','M8 desired / recorded')}</span><strong>{finding.expected}</strong></div><div><span>{t('Provider observed','Provider observed')}</span><strong>{finding.observed}</strong></div></div>
    {finding.status==='Warning'||finding.status==='Critical'?<div className="qm-form__actions"><Button view="action" onClick={()=>reviewDemoFinding(finding.id)}><Icon data={Check}/>{t('Отметить проверенным (демо)','Mark reviewed (demo)')}</Button></div>:null}
    <div className="qm-alert qm-alert_warning">{t('Изменения в кластере не выполняются. Реальный repair потребует dry-run, RBAC и проверки версии провайдерского ресурса.','No cluster changes are performed. Real repairs require dry-run, RBAC and provider resource-version checks.')}</div>
  </Drawer>
}
export function QuotaReconciliationPage() {
  const demo=useQuotaDemo()
  const [filter,setFilter]=useState('All')
  const [opened,setOpened]=useState<string|null>(null)
  const selected=demo.findings.find(f=>f.id===opened)
  const filtered=demo.findings.filter(f=>filter==='All'||f.status===filter)
  return <Layout tab="reconciliation">{t=><>
    <div className="qm-section-heading"><div><Text as="h2" variant="header-2">{t('Сверка ресурсов','Resource Reconciliation')}</Text><Text color="secondary" variant="body-2">{t('Сравнение состояния M8 и Kubernetes / Cozystack','Compare M8 accounting with Kubernetes / Cozystack observed state')}</Text></div><Button view="action" onClick={runDemoScan}><Icon data={ArrowRotateRight}/>{t('Запустить демо-сверку','Run demo scan')}</Button></div>
    <div className="qm-kpis"><Stat label="Healthy" value={demo.findings.filter(f=>f.status==='Healthy').length}/><Stat label="Warning" value={demo.findings.filter(f=>f.status==='Warning').length} tone="warning"/><Stat label="Critical" value={demo.findings.filter(f=>f.status==='Critical').length} tone="danger"/><Stat label={t('Сверок в демо','Demo scans')} value={demo.scanCount}/></div>
    <Panel title={t('Обнаруженные расхождения','Reconciliation findings')} subtitle={t('Раздельно сравниваются desired, observed и статус резерва','Desired state, observed state and reservation status are compared separately')} actions={<Filter label={t('Статус','Status')} value={filter} onChange={setFilter} options={['All','Critical','Warning','Healthy','Resolved'].map(s=>({value:s,label:s}))}/>}>
      <div className="qm-table-scroll"><table className="qm-table"><thead><tr><th>{t('Ресурс','Resource')}</th><th>M8</th><th>Provider</th><th>{t('Приоритет','Severity')}</th><th/></tr></thead><tbody>{filtered.map(f=><tr key={f.id}><td><strong>{f.resource}</strong><span className="qm-muted">{f.id} · {f.kind}</span></td><td>{f.expected}</td><td>{f.observed}</td><td><StatusTag status={f.status}/></td><td><Button size="s" view="outlined" onClick={()=>setOpened(f.id)}>{t('Разобрать','Inspect')}</Button></td></tr>)}</tbody></table></div>
      {!filtered.length?<Empty message={t('Расхождения не найдены','No reconciliation findings')}/>:null}
    </Panel>
    <Panel title={t('Проверки и безопасность','Reconciliation safeguards')}>
      <div className="qm-guidance"><div><Icon data={TriangleExclamation} size={20}/><strong>{t('Нет автоматического удаления','No automatic deletion')}</strong><p>{t('Неизвестный результат provisioning не освобождает резерв автоматически.','An unknown provisioning outcome never automatically releases a reservation.')}</p></div><div><Icon data={Check} size={20}/><strong>{t('Контролируемый repair','Controlled repair')}</strong><p>{t('Сравниваются UID, ownership, desired revision и resourceVersion до применения изменений.','Verify UID, ownership, desired revision and resourceVersion before changes.')}</p></div></div>
    </Panel>
    {selected?<FindingDetails finding={selected} onClose={()=>setOpened(null)} t={t}/>:null}
  </>}</Layout>
}
