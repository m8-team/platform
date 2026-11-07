import {useState} from 'react'
import {Check, Code, Magnifier, ShieldCheck, Speedometer} from '@gravity-ui/icons'
import {Button, Icon, Text} from '@gravity-ui/uikit'
import {Link} from '@tanstack/react-router'

import {
  saveDemoPolicy,saveDemoLimit,toggleDemoPolicy,toggleDemoLimit,useGatewayDemo,
} from './demo'
import type {DemoGateway,GatewayPolicy,GatewayRateLimit,GatewayPolicyKind,RateTargetType,RatePeriod} from './demo'
import {gatewayManifest,apiServiceDescriptor} from './manifest'
import {GatewayPage,Panel,Metric,Status,Pill,Drawer,Selector,TextField,DescriptionField,Detail,Notice,Empty,UsageBar,formatNumber} from './GatewayShell'
import type {T} from './GatewayShell'

const policyKinds:GatewayPolicyKind[]=['JWT','API Key','mTLS','CORS','WAF']
const targetTypes=['Gateway','Service','Route'] as const
const rateTargets:RateTargetType[]=['Gateway','Service','Route','Consumer']
const periods:RatePeriod[]=['minute','hour','day']
const targetOptions=(type:RateTargetType,demo:Pick<DemoGateway,'services'|'routes'|'consumers'>):Array<{value:string;label:string}>=>{
  if(type==='Gateway')return[{value:'gateway-production',label:'gateway-production'}]
  if(type==='Service')return demo.services.map(s=>({value:s.id,label:s.name}))
  if(type==='Route')return demo.routes.map(r=>({value:r.id,label:`${r.name} (${r.kind})`}))
  return demo.consumers.map(c=>({value:c.id,label:c.name}))
}
const targetName=(type:RateTargetType,id:string,demo:Pick<DemoGateway,'services'|'routes'|'consumers'>):string=>
  targetOptions(type,demo).find(o=>o.value===id)?.label??id

function PolicyEditor({policy,close,t}:{policy?:GatewayPolicy;close:()=>void;t:T}){
  const demo=useGatewayDemo()
  const [name,setName]=useState(policy?.name??'')
  const [kind,setKind]=useState<GatewayPolicyKind>(policy?.kind??'JWT')
  const [targetType,setTargetType]=useState<GatewayPolicy['targetType']>(policy?.targetType??'Gateway')
  const [targetId,setTargetId]=useState(policy?.targetId??'gateway-production')
  const [mode,setMode]=useState<GatewayPolicy['mode']>(policy?.mode??'Enforce')
  const [description,setDescription]=useState(policy?.description??'')
  const [error,setError]=useState('')
  return <Drawer title={policy?t('Изменить политику','Edit Policy'):t('Создать политику','Create Policy')} subtitle={t('Демонстрационная policy-модель','Demo policy metadata')} close={close}>
    <form className="gw-form" onSubmit={event=>{event.preventDefault();const issue=saveDemoPolicy({name,kind,targetType,targetId,mode,description},policy?.id);if(issue)setError(issue);else close()}}>
      <TextField label={t('Имя политики','Policy name')} value={name} onChange={setName} placeholder="Gateway JWT policy"/>
      <Selector label={t('Тип политики','Policy type')} value={kind} onChange={v=>setKind(v as GatewayPolicyKind)} options={policyKinds.map(v=>({value:v,label:v}))}/>
      <div className="gw-form__two">
        <Selector label={t('Уровень','Scope')} value={targetType} onChange={v=>{const type=v as GatewayPolicy['targetType'];setTargetType(type);setTargetId(targetOptions(type,demo)[0]?.value??'')}} options={targetTypes.map(v=>({value:v,label:v}))}/>
        <Selector label={t('Объект применения','Target')} value={targetId} onChange={setTargetId} options={targetOptions(targetType,demo)}/>
      </div>
      <Selector label={t('Режим','Mode')} value={mode} onChange={v=>setMode(v as GatewayPolicy['mode'])} options={[{value:'Enforce',label:'Enforce'},{value:'Monitor',label:'Monitor'}]}/>
      <DescriptionField label={t('Описание','Description')} value={description} onChange={setDescription}/>
      <Notice kind="info">{t('Правила JWT, mTLS, CORS и WAF отображаются как метаданные. Настройки провайдера и enforcement здесь не реализованы.','JWT, mTLS, CORS and WAF are metadata only. Provider configuration and enforcement are not implemented.')}</Notice>
      {error?<Notice kind="danger">{error}</Notice>:null}
      <div className="gw-form__actions"><Button view="outlined" type="button" onClick={close}>{t('Отмена','Cancel')}</Button><Button view="action" type="submit">{t('Сохранить','Save')}</Button></div>
    </form>
  </Drawer>
}
export function GatewayPoliciesPage(){
  const demo=useGatewayDemo()
  const [search,setSearch]=useState('')
  const [kind,setKind]=useState('All')
  const [opened,setOpened]=useState<string|null>(null)
  const [editing,setEditing]=useState<string|null>(null)
  const [creating,setCreating]=useState(false)
  const visible=demo.policies.filter(p=>(kind==='All'||p.kind===kind)&&`${p.name} ${p.description} ${p.targetId}`.toLowerCase().includes(search.toLowerCase()))
  const selected=demo.policies.find(p=>p.id===opened)
  const edited=demo.policies.find(p=>p.id===editing)
  return <GatewayPage active="policies" title="Policies" description="Authentication, authorization, CORS and WAF policy inventory">
    {t=><>
      <div className="gw-section-title"><div><Text as="h2" variant="header-2">{t('Политики Gateway','Gateway Policies')}</Text><Text variant="body-2" color="secondary">{t('Применение на уровне Gateway, API-сервиса или маршрута','Gateway-, service- and route-scoped policies')}</Text></div><Button view="action" onClick={()=>setCreating(true)}>+ {t('Новая политика','New policy')}</Button></div>
      <div className="gw-kpis"><Metric label={t('Политик','Total policies')} value={demo.policies.length}/><Metric label="Enabled" value={demo.policies.filter(p=>p.status==='Enabled').length} tone="success"/><Metric label="Monitor" value={demo.policies.filter(p=>p.mode==='Monitor').length}/><Metric label={t('Защита доступа','Auth policies')} value={demo.policies.filter(p=>['JWT','API Key','mTLS'].includes(p.kind)).length}/></div>
      <Panel title={t('Каталог политик','Policy catalog')} subtitle={t('Декларативные привязки, не подтверждённые провайдером','Desired policy bindings, not provider-verified state')}
        actions={<div className="gw-toolbar"><div className="gw-search"><Icon data={Magnifier} size={16}/><input aria-label="Search policies" placeholder={t('Имя или объект','Name or target')} value={search} onChange={e=>setSearch(e.target.value)}/></div><Selector label="Type" value={kind} onChange={setKind} options={['All',...policyKinds].map(v=>({value:v,label:v}))}/></div>}>
        <div className="gw-table-scroll"><table className="gw-table"><thead><tr><th>{t('Политика','Policy')}</th><th>Kind</th><th>{t('Применяется к','Target')}</th><th>{t('Режим','Mode')}</th><th>{t('Статус','Status')}</th><th/></tr></thead><tbody>{visible.map(p=><tr key={p.id}><td><strong>{p.name}</strong><span className="gw-muted">{p.description}</span></td><td><Pill>{p.kind}</Pill></td><td>{p.targetType} · {targetName(p.targetType,p.targetId,demo)}</td><td><Pill tone={p.mode==='Monitor'?'warning':'neutral'}>{p.mode}</Pill></td><td><Status status={p.status}/></td><td><Button view="outlined" size="s" onClick={()=>setOpened(p.id)}>{t('Детали','Details')}</Button></td></tr>)}</tbody></table></div>
        {!visible.length?<Empty>{t('Политики не найдены','No policies found')}</Empty>:null}
      </Panel>
      <Panel title={t('Модель применения политик','Policy precedence')} subtitle={t('Возможности control plane, требующие реализации provider adapters','Control-plane model pending provider adapters')}>
        <div className="gw-feature-list gw-feature-list_columns"><div><Icon data={ShieldCheck} size={20}/><span><strong>Gateway</strong><br/>{t('Общие правила на слушателе, маршруты наследуют общую политику.','Cross-cutting rules applied to gateway traffic.')}</span></div><div><Icon data={ShieldCheck} size={20}/><span><strong>Route</strong><br/>{t('Специфичные условия аутентификации, CORS и WAF.','Per-route auth, CORS or WAF configuration.')}</span></div><div><Icon data={ShieldCheck} size={20}/><span><strong>Service</strong><br/>{t('Защита конкретного backend-сервиса и его клиентов.','Backend-specific security requirements.')}</span></div></div>
        <Notice kind="warning">{t('Показанная иерархия не означает автоматическое наследование правил при реальном provisioning. Конфликты должны разрешаться сервером и провайдером.','Displayed layers are a proposed model, not guaranteed provider enforcement. Conflict resolution belongs to the server and provider.')}</Notice>
      </Panel>
      {selected?<Drawer title={selected.name} subtitle={selected.id} close={()=>setOpened(null)}>
        <div className="gw-details"><Detail label="Kind" value={selected.kind}/><Detail label="Target" value={`${selected.targetType}: ${targetName(selected.targetType,selected.targetId,demo)}`}/><Detail label="Mode" value={selected.mode}/><Detail label="Status" value={<Status status={selected.status}/>}/></div>
        <Text variant="body-2" color="secondary">{selected.description}</Text>
        <Notice kind="info">{t('В этом прототипе включение/отключение изменяет только метаданные.','Enable/disable changes only demo metadata, not actual enforcement.')}</Notice>
        <div className="gw-form__actions"><Button view="outlined" onClick={()=>{setEditing(selected.id);setOpened(null)}}>{t('Редактировать','Edit')}</Button><Button view={selected.status==='Enabled'?'outlined-danger':'action'} onClick={()=>toggleDemoPolicy(selected.id)}>{selected.status==='Enabled'?t('Отключить (демо)','Disable (demo)'):t('Включить (демо)','Enable (demo)')}</Button></div>
      </Drawer>:null}
      {creating||edited?<PolicyEditor key={edited?.id??'new'} policy={edited} t={t} close={()=>{setCreating(false);setEditing(null)}}/>:null}
    </>}
  </GatewayPage>
}

function LimitEditor({limit,close,t}:{limit?:GatewayRateLimit;close:()=>void;t:T}){
  const demo=useGatewayDemo()
  const [name,setName]=useState(limit?.name??'')
  const [targetType,setTargetType]=useState<RateTargetType>(limit?.targetType??'Route')
  const [targetId,setTargetId]=useState(limit?.targetId??demo.routes[0]?.id??'')
  const [requests,setRequests]=useState(String(limit?.requests??1000))
  const [period,setPeriod]=useState<RatePeriod>(limit?.period??'minute')
  const [burst,setBurst]=useState(String(limit?.burst??100))
  const [mode,setMode]=useState<GatewayRateLimit['mode']>(limit?.mode??'Enforce')
  const [error,setError]=useState('')
  return <Drawer title={limit?t('Редактировать rate limit','Edit Rate Limit'):t('Новый rate limit','New Rate Limit')} subtitle={t('Политика ограничения трафика','Traffic budget policy')} close={close}>
    <form className="gw-form" onSubmit={event=>{event.preventDefault();const issue=saveDemoLimit({name,targetType,targetId,requests:Number(requests),period,burst:Number(burst),mode},limit?.id);if(issue)setError(issue);else close()}}>
      <TextField label={t('Название','Name')} value={name} onChange={setName} placeholder="Chat completions"/>
      <div className="gw-form__two"><Selector label={t('Уровень','Scope')} value={targetType} onChange={v=>{const type=v as RateTargetType;setTargetType(type);setTargetId(targetOptions(type,demo)[0]?.value??'')}} options={rateTargets.map(v=>({value:v,label:v}))}/><Selector label={t('Объект','Target')} value={targetId} onChange={setTargetId} options={targetOptions(targetType,demo)}/></div>
      <div className="gw-form__two"><TextField label={t('Запросов','Request count')} type="number" value={requests} onChange={setRequests}/><Selector label={t('Период','Period')} value={period} onChange={v=>setPeriod(v as RatePeriod)} options={periods.map(v=>({value:v,label:v}))}/></div>
      <div className="gw-form__two"><TextField label={t('Дополнительный burst','Burst allowance')} value={burst} type="number" onChange={setBurst}/><Selector label={t('Режим','Mode')} value={mode} onChange={v=>setMode(v as GatewayRateLimit['mode'])} options={[{value:'Enforce',label:'Enforce'},{value:'Monitor',label:'Monitor'}]}/></div>
      <Notice kind="info">{t('Requests — окно rate limit, не квота CPU/памяти. Burst показан отдельно и не суммируется в usage.','Requests is a traffic rate window, not a CPU/memory quota. Burst is tracked separately from usage.')}</Notice>
      {error?<Notice kind="danger">{error}</Notice>:null}
      <div className="gw-form__actions"><Button view="outlined" type="button" onClick={close}>{t('Отмена','Cancel')}</Button><Button view="action" type="submit">{t('Сохранить','Save')}</Button></div>
    </form>
  </Drawer>
}

export function GatewayRateLimitsPage(){
  const demo=useGatewayDemo()
  const [scope,setScope]=useState('All')
  const [search,setSearch]=useState('')
  const [opened,setOpened]=useState<string|null>(null)
  const [editing,setEditing]=useState<string|null>(null)
  const [creating,setCreating]=useState(false)
  const visible=demo.limits.filter(l=>(scope==='All'||l.targetType===scope)&&`${l.name} ${l.targetId}`.toLowerCase().includes(search.toLowerCase()))
  const selected=demo.limits.find(l=>l.id===opened)
  const edited=demo.limits.find(l=>l.id===editing)
  return <GatewayPage active="limits" title="Rate Limits" description="Request budgets, distributed throttling policies and traffic safeguards">
    {t=><>
      <div className="gw-section-title"><div><Text as="h2" variant="header-2">{t('Лимиты запросов','Rate Limits')}</Text><Text variant="body-2" color="secondary">{t('Ограничения трафика Gateway, маршрутов, сервисов и клиентов','Gateway, route, service and consumer traffic controls')}</Text></div><Button view="action" onClick={()=>setCreating(true)}>+ {t('Новый лимит','New rate limit')}</Button></div>
      <div className="gw-kpis"><Metric label={t('Политик лимитов','Rate limit policies')} value={demo.limits.length}/><Metric label="Enforce" value={demo.limits.filter(l=>l.mode==='Enforce'&&l.status==='Enabled').length}/><Metric label="Monitor" value={demo.limits.filter(l=>l.mode==='Monitor'&&l.status==='Enabled').length}/><Metric label={t('Высокая загрузка','High utilization')} value={demo.limits.filter(l=>l.requests>0&&l.used/l.requests>=.75).length} tone="warning"/></div>
      <Panel title={t('Политики лимитов','Rate limit policies')} subtitle={t('Использование — синтетические значения выбранного окна','Usage represents synthetic values for the respective time window')}
        actions={<div className="gw-toolbar"><div className="gw-search"><Icon data={Magnifier} size={16}/><input aria-label="Search limits" value={search} onChange={e=>setSearch(e.target.value)} placeholder={t('Название или цель','Name or target')}/></div><Selector label={t('Уровень','Scope')} value={scope} onChange={setScope} options={['All',...rateTargets].map(v=>({value:v,label:v}))}/></div>}>
        <div className="gw-table-scroll"><table className="gw-table"><thead><tr><th>{t('Политика','Policy')}</th><th>{t('Цель','Target')}</th><th>{t('Лимит','Limit')}</th><th>{t('Использование окна','Window usage')}</th><th>{t('Режим','Mode')}</th><th>Status</th><th/></tr></thead><tbody>{visible.map(l=><tr key={l.id}><td><strong>{l.name}</strong><span className="gw-muted">{l.id}</span></td><td>{l.targetType}: {targetName(l.targetType,l.targetId,demo)}</td><td className="gw-nowrap">{formatNumber(l.requests)} / {l.period}<span className="gw-muted">burst {formatNumber(l.burst)}</span></td><td><div className="gw-meter-cell"><span>{formatNumber(l.used)} · {Math.round(l.used/l.requests*100)}%</span><UsageBar value={l.used} total={l.requests}/></div></td><td><Pill tone={l.mode==='Monitor'?'warning':'neutral'}>{l.mode}</Pill></td><td><Status status={l.status}/></td><td><Button size="s" view="outlined" onClick={()=>setOpened(l.id)}>{t('Детали','Details')}</Button></td></tr>)}</tbody></table></div>
        {!visible.length?<Empty>{t('Лимиты не найдены','No rate limits found')}</Empty>:null}
      </Panel>
      <Panel title={t('Как взаимодействует с Quota Manager','Relationship to Quota Manager')}>
        <div className="gw-feature-list gw-feature-list_columns"><div><Icon data={Speedometer} size={20}/><span><strong>Rate Limit</strong><br/>{t('Ограничивает поток запросов по окну и клиенту.','Restricts request throughput by time window and target.')}</span></div><div><Icon data={ShieldCheck} size={20}/><span><strong>Quota Manager</strong><br/>{t('Ограничивает выделение облачных ресурсов и потребление за период.','Limits allocated platform resources and period consumption.')}</span></div></div>
        <Notice kind="info">{t('Данные счётчиков в этом UI не связаны с реальным Envoy rate limit service, Redis или OpenMeter.','Counters are not connected to Envoy rate limit service, Redis or OpenMeter.')}</Notice>
      </Panel>
      {selected?<Drawer title={selected.name} subtitle={selected.id} close={()=>setOpened(null)}>
        <div className="gw-details"><Detail label="Status" value={<Status status={selected.status}/>}/><Detail label="Target" value={`${selected.targetType}: ${targetName(selected.targetType,selected.targetId,demo)}`}/><Detail label="Limit" value={`${formatNumber(selected.requests)} / ${selected.period}`}/><Detail label="Burst" value={formatNumber(selected.burst)}/><Detail label="Mode" value={selected.mode}/><Detail label="Used in window" value={formatNumber(selected.used)}/></div>
        <UsageBar value={selected.used} total={selected.requests}/>
        <Notice kind="warning">{t('Отключение применится только к локальному демореестру.','Disabling updates only the demo registry.')}</Notice>
        <div className="gw-form__actions"><Button view="outlined" onClick={()=>{setEditing(selected.id);setOpened(null)}}>{t('Редактировать','Edit')}</Button><Button view={selected.status==='Enabled'?'outlined-danger':'action'} onClick={()=>toggleDemoLimit(selected.id)}>{selected.status==='Enabled'?t('Отключить','Disable'):t('Включить','Enable')}</Button></div>
      </Drawer>:null}
      {creating||edited?<LimitEditor key={edited?.id??'new'} limit={edited} close={()=>{setCreating(false);setEditing(null)}} t={t}/>:null}
    </>}
  </GatewayPage>
}

export function GatewayYamlPage(){
  const demo=useGatewayDemo()
  const [resource,setResource]=useState<'Route'|'Service'>('Route')
  const [id,setId]=useState(demo.routes[0]?.id??'')
  const [copyStatus,setCopyStatus]=useState('')
  const route=demo.routes.find(r=>r.id===id)
  const service=demo.services.find(s=>s.id===id)
  const backend=route?demo.services.find(s=>s.id===route.serviceId):undefined
  const manifest=resource==='Service'&&service?apiServiceDescriptor(service):route&&backend?gatewayManifest(route,backend):'# Select a resource\n'
  const deployable=resource==='Route'&&route?.kind==='HTTPRoute'
  const isConceptual=!deployable
  const copy=async()=>{
    try{await navigator.clipboard.writeText(manifest);setCopyStatus('Copied')}
    catch{setCopyStatus('Clipboard unavailable')}
  }
  return <GatewayPage active="yaml" title="Gateway YAML" description="Inspect generated manifests and proposed M8 resource definitions">
    {t=><>
      <div className="gw-section-title"><div><Text as="h2" variant="header-2">{t('Просмотр конфигурации','Configuration Preview')}</Text><Text variant="body-2" color="secondary">{t('Генерация read-only YAML по выбранным деморесурсам','Read-only YAML generated from demo resources')}</Text></div><div className="gw-actions"><Button view="outlined" onClick={()=>void copy()}><Icon data={Code} size={16}/>{t('Скопировать','Copy YAML')}</Button><Link to="/gateway/routes" className="gw-link">{t('Маршруты','Routes')}</Link></div></div>
      <div className="gw-yaml-layout">
        <Panel title={t('Выбор ресурса','Resource selection')} subtitle={t('Элементы из локального реестра Gateway','Entries from local Gateway demo store')}>
          <Selector label={t('Категория','Category')} value={resource} onChange={v=>{const type=v as 'Route'|'Service';setResource(type);setId(type==='Route'?demo.routes[0]?.id??'':demo.services[0]?.id??'');setCopyStatus('')}} options={[{value:'Route',label:'Route / HTTPRoute / AIGatewayRoute'},{value:'Service',label:'ApiService (M8 conceptual)'}]}/>
          <Selector label={t('Ресурс','Resource')} value={id} onChange={v=>{setId(v);setCopyStatus('')}} options={resource==='Route'?demo.routes.map(r=>({value:r.id,label:`${r.name} · ${r.kind}`})):demo.services.map(s=>({value:s.id,label:s.name}))}/>
          <div className="gw-details">
            <Detail label="Kind" value={resource==='Route'?route?.kind??'—':'ApiService'}/>
            <Detail label="Provider" value={resource==='Route'?route?.provider??'—':'M8 Resource Type Registry'}/>
            <Detail label={t('Состояние','Status')} value={route?<Status status={route.status}/>:service?<Status status={service.health}/>:null}/>
          </div>
          <Notice kind={isConceptual?'warning':'info'}>
            {isConceptual?t('Это концептуальный M8 descriptor, НЕ применимый Kubernetes CRD. Схему Agent Router нужно подтвердить по провайдеру.','This is a conceptual M8 descriptor, NOT a deployable Kubernetes CRD. Confirm Agent Router schema with the provider.'):
              t('Это пример HTTPRoute Gateway API. Проверьте namespace, parentRefs, разрешения backend и listeners перед применением.','Illustrative Gateway API HTTPRoute. Verify namespace, parentRefs, backend permissions and listeners before applying.')}
          </Notice>
          {copyStatus?<Notice kind={copyStatus==='Copied'?'success':'warning'}>{copyStatus==='Copied'?t('YAML скопирован','YAML copied'):t('Clipboard недоступен','Clipboard unavailable')}</Notice>:null}
        </Panel>
        <Panel title={t('Манифест','Manifest')} subtitle={t('Только чтение · изменения исходных деморесурсов автоматически отражаются здесь','Read only · demo resource edits are reflected here')}>
          <div className="gw-code-head"><span>{deployable?'gateway.networking.k8s.io/v1 · HTTPRoute':'M8 conceptual resource descriptor'}</span><Pill tone={isConceptual?'warning':'info'}>{isConceptual?'NOT DEPLOYABLE':'EXAMPLE'}</Pill></div>
          <pre className="gw-yaml-code"><code>{manifest}</code></pre>
        </Panel>
      </div>
      <Panel title={t('Что генерируется в production','Production manifest workflow')}>
        <div className="gw-feature-list gw-feature-list_columns"><div><Icon data={Check} size={18}/><span>{t('Provider API выбирает версионированную schema/CRD на основе Resource Type Registry.','Provider API selects a versioned schema/CRD from Resource Type Registry.')}</span></div><div><Icon data={Check} size={18}/><span>{t('Quota admission и Temporal выполняются до/во время provisioning.','Quota admission and Temporal coordinate provisioning.')}</span></div><div><Icon data={Check} size={18}/><span>{t('Kubernetes reconciler сообщает observed state и условия готовности.','Kubernetes reconciler reports observed state and readiness conditions.')}</span></div></div>
        <Notice kind="warning">{t('В этой версии нет Apply YAML: никакие запросы Kubernetes API не выполняются.','Apply YAML is deliberately unavailable: no Kubernetes API calls are made.')}</Notice>
      </Panel>
    </>}
  </GatewayPage>
}
