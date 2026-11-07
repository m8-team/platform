import {useState} from 'react'
import {ArrowRotateRight, Cloud, Magnifier, ShieldCheck, Speedometer} from '@gravity-ui/icons'
import {Button, Icon, Text} from '@gravity-ui/uikit'
import {Link} from '@tanstack/react-router'

import {
  resetGatewayDemo,saveDemoService,saveDemoRoute,saveDemoConsumer,toggleDemoConsumer,simulateGatewayRequest,useGatewayDemo,
} from './demo'
import type {ApiService,GatewayRoute,GatewayRouteKind,GatewayConsumer,AuthType} from './demo'
import {GatewayPage,Panel,Metric,Status,Pill,Drawer,Selector,TextField,DescriptionField,Detail,Notice,Empty,formatNumber} from './GatewayShell'
import type {T} from './GatewayShell'

const methods=['ANY','GET','POST','PUT','DELETE'] as const
const methodOptions=methods.map(method=>({value:method,label:method}))

function ServiceEditor({service,close,t}:{service?:ApiService;close:()=>void;t:T}){
  const [name,setName]=useState(service?.name??'')
  const [upstream,setUpstream]=useState(service?.upstream??'')
  const [port,setPort]=useState(String(service?.port??8080))
  const [protocol,setProtocol]=useState<ApiService['protocol']>(service?.protocol??'HTTP')
  const [environment,setEnvironment]=useState<ApiService['environment']>(service?.environment??'Production')
  const [description,setDescription]=useState(service?.description??'')
  const [error,setError]=useState('')
  return <Drawer title={service?t('Редактировать API-сервис','Edit API Service'):t('Создать API-сервис','Create API Service')} subtitle="M8 Gateway / upstream service" close={close}>
    <form className="gw-form" onSubmit={event=>{event.preventDefault();const issue=saveDemoService({name,upstream,port:Number(port),protocol,environment,description},service?.id);if(issue)setError(issue);else close()}}>
      <TextField label={t('Имя сервиса','Service name')} value={name} onChange={setName} placeholder="my-api"/>
      <TextField label={t('Upstream hostname','Upstream hostname')} value={upstream} onChange={setUpstream} placeholder="my-api.production.svc.cluster.local"/>
      <div className="gw-form__two"><Selector label="Protocol" value={protocol} onChange={v=>setProtocol(v as ApiService['protocol'])} options={[{value:'HTTP',label:'HTTP'},{value:'gRPC',label:'gRPC'}]}/><TextField label={t('Порт','Port')} type="number" value={port} onChange={setPort}/></div>
      <Selector label={t('Окружение','Environment')} value={environment} onChange={v=>setEnvironment(v as ApiService['environment'])} options={[{value:'Production',label:'Production'},{value:'Development',label:'Development'}]}/>
      <DescriptionField label={t('Описание','Description')} value={description} onChange={setDescription}/>
      <Notice kind="info">{t('Сервис создаётся только в демореестре. Секреты и параметры подключения не проверяются.','Created in demo registry only. Connectivity and credentials are not verified.')}</Notice>
      {error?<Notice kind="danger">{error}</Notice>:null}
      <div className="gw-form__actions"><Button type="button" view="outlined" onClick={close}>{t('Отмена','Cancel')}</Button><Button type="submit" view="action">{t('Сохранить в демо','Save demo')}</Button></div>
    </form>
  </Drawer>
}
export function GatewayApiServicesPage(){
  const demo=useGatewayDemo()
  const [search,setSearch]=useState('')
  const [health,setHealth]=useState('All')
  const [opened,setOpened]=useState<string|null>(null)
  const [edit,setEdit]=useState<string|null>(null)
  const [creating,setCreating]=useState(false)
  const visible=demo.services.filter(s=>(health==='All'||s.health===health)&&`${s.name} ${s.upstream} ${s.description}`.toLowerCase().includes(search.toLowerCase()))
  const selected=demo.services.find(s=>s.id===opened)
  const editing=demo.services.find(s=>s.id===edit)
  return <GatewayPage active="services" title="API Services" description="Upstreams, health, protocols and route ownership">
    {t=><>
      <div className="gw-section-title"><div><Text as="h2" variant="header-2">{t('Каталог API-сервисов','API Service Catalog')}</Text><Text variant="body-2" color="secondary">{t('Сервисы, опубликованные через Gateway','Backend services exposed through the Gateway')}</Text></div><div className="gw-actions"><Button view="outlined" onClick={resetGatewayDemo}><Icon data={ArrowRotateRight} size={16}/>{t('Сбросить демо','Reset demo')}</Button><Button view="action" onClick={()=>setCreating(true)}>+ {t('Новый сервис','New service')}</Button></div></div>
      <div className="gw-kpis">
        <Metric label={t('Сервисов','Services')} value={demo.services.length} caption={t('HTTP + gRPC','HTTP + gRPC')}/>
        <Metric label={t('Healthy','Healthy')} value={demo.services.filter(s=>s.health==='Healthy').length} tone="success"/>
        <Metric label={t('Degraded / Unknown','Degraded / Unknown')} value={demo.services.filter(s=>s.health!=='Healthy').length} tone="warning"/>
        <Metric label={t('Запросов за 24 часа','Requests · 24h')} value={formatNumber(demo.services.reduce((sum,s)=>sum+s.requests24h,0))} caption={t('Синтетическая статистика','Illustrative metrics')}/>
      </div>
      <Panel title={t('Upstream-сервисы','Upstream services')} subtitle={t('Нажмите на сервис для просмотра связей и параметров','Open a service for configuration and dependencies')}
        actions={<div className="gw-toolbar"><div className="gw-search"><Icon data={Magnifier} size={16}/><input aria-label="Search services" value={search} onChange={e=>setSearch(e.target.value)} placeholder={t('Имя или адрес сервиса','Name or upstream hostname')}/></div><Selector label={t('Здоровье','Health')} value={health} onChange={setHealth} options={['All','Healthy','Degraded','Unknown'].map(value=>({value,label:value}))}/></div>}>
        <div className="gw-table-scroll"><table className="gw-table"><thead><tr><th>{t('Сервис','Service')}</th><th>{t('Upstream','Upstream')}</th><th>{t('Протокол','Protocol')}</th><th>{t('Маршрутов','Routes')}</th><th>p95</th><th>{t('Состояние','Health')}</th><th/></tr></thead><tbody>{visible.map(s=><tr key={s.id}><td><strong>{s.name}</strong><span className="gw-muted">{s.description}</span></td><td className="gw-mono">{s.upstream}:{s.port}</td><td><Pill>{s.protocol}</Pill></td><td>{demo.routes.filter(r=>r.serviceId===s.id).length}</td><td>{s.latencyP95? `${s.latencyP95} ms`:'—'}</td><td><Status status={s.health}/></td><td><Button size="s" view="outlined" onClick={()=>setOpened(s.id)}>{t('Детали','Details')}</Button></td></tr>)}</tbody></table></div>
        {!visible.length?<Empty>{t('Сервисы не найдены','No services found')}</Empty>:null}
      </Panel>
      <div className="gw-columns">
        <Panel title={t('Связи сервисов','Service routing')} subtitle={t('Маршруты используют upstream-сервисы как backend','Routes point to registered backend services')}>
          <div className="gw-entity-list">{demo.services.map(s=><div key={s.id} className="gw-entity"><div className="gw-entity__icon"><Icon data={Cloud} size={19}/></div><div><strong>{s.name}</strong><span>{s.protocol} · {s.environment}</span></div><Pill>{demo.routes.filter(r=>r.serviceId===s.id).length} routes</Pill></div>)}</div>
        </Panel>
        <Panel title={t('Точки контроля','Security and capacity')} subtitle={t('Декларативные привязки, без вызова провайдера','Declarative relations, no provider calls')}>
          <div className="gw-feature-list"><div><Icon data={ShieldCheck} size={20}/><span>{t('Политики применяются на уровне Gateway, Route и Service','Policies can target Gateway, Route or Service')}</span></div><div><Icon data={Speedometer} size={20}/><span>{t('Лимиты запросов задаются независимо от Quota Manager','Rate limits are separate from Quota Manager resource quotas')}</span></div><div><Icon data={Cloud} size={20}/><span>{t('Фактическое состояние определяется reconciler провайдера','Observed state is reported by the provider reconciler')}</span></div></div>
        </Panel>
      </div>
      {selected?<Drawer title={selected.name} subtitle={selected.id} close={()=>setOpened(null)}>
        <div className="gw-details"><Detail label="Health" value={<Status status={selected.health}/>}/><Detail label="Protocol" value={selected.protocol}/><Detail label="Upstream" value={<span className="gw-mono">{selected.upstream}:{selected.port}</span>}/><Detail label="Environment" value={selected.environment}/><Detail label={t('Запросов за 24 часа','Requests · 24h')} value={formatNumber(selected.requests24h)}/><Detail label="p95" value={`${selected.latencyP95} ms`}/></div>
        <Text variant="body-2" color="secondary">{selected.description}</Text>
        <strong>{t('Связанные маршруты','Related routes')}</strong>
        <div className="gw-entity-list">{demo.routes.filter(r=>r.serviceId===selected.id).map(r=><div key={r.id} className="gw-entity"><span>{r.name}</span><Status status={r.status}/></div>)}{!demo.routes.some(r=>r.serviceId===selected.id)?<Empty>{t('Нет маршрутов','No routes')}</Empty>:null}</div>
        <div className="gw-form__actions"><Button view="outlined" onClick={()=>{setEdit(selected.id);setOpened(null)}}>{t('Редактировать','Edit')}</Button><Link to="/gateway/routes" className="gw-link">{t('Открыть маршруты','View routes')}</Link></div>
      </Drawer>:null}
      {creating||editing?<ServiceEditor key={editing?.id??'new'} service={editing} close={()=>{setCreating(false);setEdit(null)}} t={t}/>:null}
    </>}
  </GatewayPage>
}

function RouteEditor({route,close,t}:{route?:GatewayRoute;close:()=>void;t:T}){
  const demo=useGatewayDemo()
  const [name,setName]=useState(route?.name??'')
  const [kind,setKind]=useState<GatewayRouteKind>(route?.kind??'HTTPRoute')
  const [hostname,setHostname]=useState(route?.hostname??'api.acme.example')
  const [pathPrefix,setPathPrefix]=useState(route?.pathPrefix??'/v1/')
  const [method,setMethod]=useState<GatewayRoute['method']>(route?.method??'ANY')
  const [serviceId,setServiceId]=useState(route?.serviceId??demo.services[0]?.id??'')
  const [description,setDescription]=useState(route?.description??'')
  const [error,setError]=useState('')
  return <Drawer title={route?t('Изменение маршрута','Edit Route'):t('Новый маршрут','Create Route')} subtitle="HTTPRoute / AIGatewayRoute" close={close}>
    <form className="gw-form" onSubmit={event=>{event.preventDefault();const issue=saveDemoRoute({name,kind,hostname,pathPrefix,method,serviceId,description},route?.id);if(issue)setError(issue);else close()}}>
      <TextField label={t('Имя маршрута','Route name')} value={name} onChange={setName} placeholder="my-public-route"/>
      <Selector label="Resource type" value={kind} onChange={v=>setKind(v as GatewayRouteKind)} options={[{value:'HTTPRoute',label:'Gateway API / HTTPRoute'},{value:'AIGatewayRoute',label:'AI Gateway / AIGatewayRoute'}]}/>
      <TextField label="Hostname" value={hostname} onChange={setHostname} placeholder="api.acme.example"/>
      <div className="gw-form__two"><TextField label="PathPrefix" value={pathPrefix} onChange={setPathPrefix} placeholder="/v1/orders"/><Selector label="Method" value={method} onChange={v=>setMethod(v as GatewayRoute['method'])} options={methodOptions.map(x=>({...x}))}/></div>
      <Selector label={t('Backend Service','Backend Service')} value={serviceId} onChange={setServiceId} options={demo.services.map(s=>({value:s.id,label:`${s.name} (${s.protocol})`}))}/>
      <DescriptionField label={t('Описание','Description')} value={description} onChange={setDescription}/>
      <Notice kind="info">{t('После сохранения новый маршрут получает Pending до подтверждения провайдера. В демо reconciliation не выполняется.','A saved route is Pending until provider confirmation. No reconciliation is performed in demo mode.')}</Notice>
      {error?<Notice kind="danger">{error}</Notice>:null}
      <div className="gw-form__actions"><Button type="button" view="outlined" onClick={close}>{t('Отмена','Cancel')}</Button><Button type="submit" view="action">{t('Сохранить маршрут','Save route')}</Button></div>
    </form>
  </Drawer>
}

function RouteTester({t}:{t:T}){
  const demo=useGatewayDemo()
  const [hostname,setHostname]=useState('ai.acme.example')
  const [path,setPath]=useState('/v1/chat/completions')
  const [method,setMethod]=useState('POST')
  const [consumerId,setConsumerId]=useState('consumer-ai')
  const [result,setResult]=useState<ReturnType<typeof simulateGatewayRequest>|null>(null)
  return <Panel title={t('Проверка маршрутизации','Route simulator')} subtitle={t('Проверка совпадения Host + Path + Method; реальные запросы не отправляются','Match host, path and method without sending real traffic')}>
    <form className="gw-simulator" onSubmit={event=>{event.preventDefault();setResult(simulateGatewayRequest({hostname,path,method,consumerId}))}}>
      <TextField label="Hostname" value={hostname} onChange={v=>{setHostname(v);setResult(null)}}/>
      <TextField label="Request path" value={path} onChange={v=>{setPath(v);setResult(null)}}/>
      <Selector label="HTTP method" value={method} onChange={v=>{setMethod(v);setResult(null)}} options={methodOptions.filter(x=>x.value!=='ANY').map(x=>({...x}))}/>
      <Selector label="Consumer" value={consumerId} onChange={v=>{setConsumerId(v);setResult(null)}} options={demo.consumers.map(c=>({value:c.id,label:c.name}))}/>
      <div className="gw-simulator__submit"><Button view="action" type="submit">{t('Симулировать','Simulate')}</Button></div>
    </form>
    {result?<div className="gw-simulator__result" role="status">
      <div className="gw-simulator__status"><Pill tone={result.status>=400?'warning':'success'}>{result.status} · DEMO</Pill><strong>{result.outcome}</strong></div>
      {result.route?<div className="gw-details"><Detail label="Matched route" value={result.route.name}/><Detail label="Backend" value={result.service?.name??'Unknown'}/><Detail label="Applied policies" value={result.appliedPolicies.map(p=>p.name).join(', ')||'None'}/><Detail label="Applicable limits" value={result.limits.map(l=>l.name).join(', ')||'None'}/></div>:null}
      <Notice kind="info">{t('Это модель совпадения и привязки политик, не проверка JWT, API Key, mTLS или фактического rate limiting.','This is route/policy matching only, not JWT, API Key, mTLS validation or real rate limiting.')}</Notice>
    </div>:null}
  </Panel>
}

export function GatewayRoutesPage(){
  const demo=useGatewayDemo()
  const [search,setSearch]=useState('')
  const [kind,setKind]=useState('All')
  const [status,setStatus]=useState('All')
  const [opened,setOpened]=useState<string|null>(null)
  const [edit,setEdit]=useState<string|null>(null)
  const [creating,setCreating]=useState(false)
  const visible=demo.routes.filter(r=>(kind==='All'||r.kind===kind)&&(status==='All'||r.status===status)&&`${r.name} ${r.hostname} ${r.pathPrefix}`.toLowerCase().includes(search.toLowerCase()))
  const selected=demo.routes.find(r=>r.id===opened)
  const editing=demo.routes.find(r=>r.id===edit)
  return <GatewayPage active="routes" title={titledRoutes} description="Envoy Gateway HTTPRoute and Agent Router AIGatewayRoute management">
    {t=><>
      <div className="gw-section-title"><div><Text as="h2" variant="header-2">{t('Маршруты Gateway','Gateway Routes')}</Text><Text color="secondary" variant="body-2">{t('Опубликованные и ожидающие применения маршруты','Desired routes and provider acceptance status')}</Text></div><Button view="action" onClick={()=>setCreating(true)}>+ {t('Новый маршрут','New route')}</Button></div>
      <div className="gw-kpis"><Metric label="HTTPRoute" value={demo.routes.filter(r=>r.kind==='HTTPRoute').length}/><Metric label="AIGatewayRoute" value={demo.routes.filter(r=>r.kind==='AIGatewayRoute').length}/><Metric label="Accepted" value={demo.routes.filter(r=>r.status==='Accepted').length} tone="success"/><Metric label={t('Pending / Degraded','Pending / Degraded')} value={demo.routes.filter(r=>r.status==='Pending'||r.status==='Degraded').length} tone="warning"/></div>
      <Panel title={t('Маршруты','Routes')} subtitle={t('Перейдите в детали для редактирования, просмотра политик или манифеста','Open details to edit, inspect policies or view manifests')} actions={<div className="gw-toolbar"><div className="gw-search"><Icon data={Magnifier} size={16}/><input aria-label="Search routes" placeholder={t('Имя, host или path','Name, host or path')} value={search} onChange={e=>setSearch(e.target.value)}/></div><Selector label="Type" value={kind} onChange={setKind} options={['All','HTTPRoute','AIGatewayRoute'].map(v=>({value:v,label:v}))}/><Selector label="Status" value={status} onChange={setStatus} options={['All','Accepted','Pending','Degraded','Rejected'].map(v=>({value:v,label:v}))}/></div>}>
        <div className="gw-table-scroll"><table className="gw-table"><thead><tr><th>{t('Маршрут','Route')}</th><th>Host / Path</th><th>Backend</th><th>{t('Провайдер','Provider')}</th><th>{t('Состояние','Status')}</th><th>{t('Запросов · 24h','Requests · 24h')}</th><th/></tr></thead><tbody>{visible.map(r=><tr key={r.id}><td><strong>{r.name}</strong><span className="gw-muted">{r.kind} · v{r.version}</span></td><td><span className="gw-mono">{r.hostname}</span><span className="gw-muted gw-mono">{r.method} {r.pathPrefix}</span></td><td>{demo.services.find(s=>s.id===r.serviceId)?.name??'—'}</td><td>{r.provider}</td><td><Status status={r.status}/></td><td>{formatNumber(r.requests24h)}</td><td><Button size="s" view="outlined" onClick={()=>setOpened(r.id)}>{t('Детали','Details')}</Button></td></tr>)}</tbody></table></div>
        {!visible.length?<Empty>{t('Маршруты не найдены','No routes found')}</Empty>:null}
      </Panel>
      <RouteTester t={t}/>
      {selected?<Drawer title={selected.name} subtitle={`${selected.kind} · v${selected.version}`} close={()=>setOpened(null)}>
        <div className="gw-details"><Detail label="Status" value={<Status status={selected.status}/>}/><Detail label="Provider" value={selected.provider}/><Detail label="Hostname" value={selected.hostname}/><Detail label="PathPrefix" value={selected.pathPrefix}/><Detail label="Method" value={selected.method}/><Detail label="Backend Service" value={demo.services.find(s=>s.id===selected.serviceId)?.name??'—'}/><Detail label="Requests / 24h" value={formatNumber(selected.requests24h)}/><Detail label="Latency p95" value={`${selected.latencyP95} ms`}/></div>
        <Text variant="body-2" color="secondary">{selected.description}</Text>
        <strong>{t('Применимые политики','Applicable policies')}</strong>
        <div className="gw-entity-list">{demo.policies.filter(p=>p.status==='Enabled'&&(p.targetType==='Gateway'||p.targetType==='Route'&&p.targetId===selected.id||p.targetType==='Service'&&p.targetId===selected.serviceId)).map(p=><div className="gw-entity" key={p.id}><strong>{p.name}</strong><Pill>{p.kind}</Pill></div>)}</div>
        <strong>{t('Лимиты','Rate limits')}</strong>
        <div className="gw-entity-list">{demo.limits.filter(l=>l.status==='Enabled'&&(l.targetType==='Gateway'||l.targetType==='Route'&&l.targetId===selected.id||l.targetType==='Service'&&l.targetId===selected.serviceId)).map(l=><div className="gw-entity" key={l.id}><strong>{l.name}</strong><span>{l.requests}/{l.period}</span></div>)}</div>
        <div className="gw-form__actions"><Button view="outlined" onClick={()=>{setEdit(selected.id);setOpened(null)}}>{t('Редактировать','Edit')}</Button><Link className="gw-link" to="/gateway/yaml">{t('Открыть YAML','View YAML')}</Link></div>
      </Drawer>:null}
      {creating||editing?<RouteEditor key={editing?.id??'new'} route={editing} t={t} close={()=>{setCreating(false);setEdit(null)}}/>:null}
    </>}
  </GatewayPage>
}
const titledRoutes='Routes'

function ConsumerEditor({close,t}:{close:()=>void;t:T}){
  const [name,setName]=useState('')
  const [application,setApplication]=useState('')
  const [auth,setAuth]=useState<AuthType>('OAuth2')
  const [description,setDescription]=useState('')
  const [error,setError]=useState('')
  return <Drawer title={t('Новый потребитель','New consumer')} subtitle={t('Регистрация клиента API','API client registration')} close={close}>
    <form className="gw-form" onSubmit={event=>{event.preventDefault();const issue=saveDemoConsumer({name,application,auth,description});if(issue)setError(issue);else close()}}>
      <TextField label={t('Имя потребителя','Consumer name')} value={name} onChange={setName} placeholder="partner-client"/>
      <TextField label={t('Приложение','Application')} value={application} onChange={setApplication} placeholder="Partner Portal"/>
      <Selector label={t('Метод аутентификации','Authentication method')} value={auth} onChange={v=>setAuth(v as AuthType)} options={(['OAuth2','API Key','mTLS'] as AuthType[]).map(v=>({value:v,label:v}))}/>
      <DescriptionField label={t('Назначение','Description')} value={description} onChange={setDescription}/>
      <Notice kind="warning">{t('Создаётся только метадата клиента. API-ключи, сертификаты и OAuth-credentials не создаются и не отображаются.','Client metadata only: no API keys, certificates or OAuth credentials are generated or displayed.')}</Notice>
      {error?<Notice kind="danger">{error}</Notice>:null}
      <div className="gw-form__actions"><Button view="outlined" onClick={close} type="button">{t('Отмена','Cancel')}</Button><Button view="action" type="submit">{t('Создать','Create')}</Button></div>
    </form>
  </Drawer>
}

export function GatewayConsumersPage(){
  const demo=useGatewayDemo()
  const [search,setSearch]=useState('')
  const [filter,setFilter]=useState('All')
  const [opened,setOpened]=useState<string|null>(null)
  const [creating,setCreating]=useState(false)
  const selected=demo.consumers.find(c=>c.id===opened)
  const visible=demo.consumers.filter(c=>(filter==='All'||c.status===filter)&&`${c.name} ${c.application}`.toLowerCase().includes(search.toLowerCase()))
  return <GatewayPage active="consumers" title="Consumers" description="API clients, authentication methods and access lifecycle">
    {t=><>
      <div className="gw-section-title"><div><Text as="h2" variant="header-2">{t('API-потребители','API Consumers')}</Text><Text variant="body-2" color="secondary">{t('Приложения и клиенты, обращающиеся к Gateway','Applications and identities calling the Gateway')}</Text></div><Button view="action" onClick={()=>setCreating(true)}>+ {t('Новый потребитель','New consumer')}</Button></div>
      <div className="gw-kpis"><Metric label={t('Потребителей','Consumers')} value={demo.consumers.length}/><Metric label="Active" value={demo.consumers.filter(c=>c.status==='Active').length} tone="success"/><Metric label="Suspended" value={demo.consumers.filter(c=>c.status==='Suspended').length} tone="warning"/><Metric label={t('Запросы · 24h','Requests · 24h')} value={formatNumber(demo.consumers.reduce((sum,c)=>sum+c.requests24h,0))}/></div>
      <Panel title={t('Каталог клиентов','Consumer registry')} subtitle={t('Метаданные без секретов и ключевого материала','Metadata only; credential material is never exposed')} actions={<div className="gw-toolbar"><div className="gw-search"><Icon data={Magnifier} size={16}/><input aria-label="Search consumers" value={search} onChange={e=>setSearch(e.target.value)} placeholder={t('Поиск по имени','Search by name')}/></div><Selector label="Status" value={filter} onChange={setFilter} options={['All','Active','Suspended'].map(v=>({value:v,label:v}))}/></div>}>
        <div className="gw-table-scroll"><table className="gw-table"><thead><tr><th>Consumer</th><th>{t('Приложение','Application')}</th><th>Auth</th><th>{t('Состояние','Status')}</th><th>{t('Запросы · 24h','Requests · 24h')}</th><th>{t('Последний запрос','Last seen')}</th><th/></tr></thead><tbody>{visible.map(c=><tr key={c.id}><td><strong>{c.name}</strong><span className="gw-muted">{c.id}</span></td><td>{c.application}</td><td><Pill>{c.auth}</Pill></td><td><Status status={c.status}/></td><td>{formatNumber(c.requests24h)}</td><td>{c.lastSeen}</td><td><Button view="outlined" size="s" onClick={()=>setOpened(c.id)}>{t('Детали','Details')}</Button></td></tr>)}</tbody></table></div>
        {!visible.length?<Empty>{t('Клиенты не найдены','No consumers match filters')}</Empty>:null}
      </Panel>
      {selected?<Drawer title={selected.name} subtitle={selected.id} close={()=>setOpened(null)}>
        <div className="gw-details"><Detail label="Status" value={<Status status={selected.status}/>}/><Detail label="Application" value={selected.application}/><Detail label="Authentication" value={selected.auth}/><Detail label="Requests / 24h" value={formatNumber(selected.requests24h)}/><Detail label="Last activity" value={selected.lastSeen}/></div>
        <Text variant="body-2" color="secondary">{selected.description}</Text>
        <Notice kind="info">{t('Секреты в Console не хранятся и не выдаются. Права доступа должны проверяться через M8 Access.','No secrets are stored or issued by the Console. Permissions must be validated through M8 Access.')}</Notice>
        <div className="gw-form__actions"><Button view={selected.status==='Active'?'outlined-danger':'action'} onClick={()=>toggleDemoConsumer(selected.id)}>{selected.status==='Active'?t('Приостановить (демо)','Suspend (demo)'):t('Активировать (демо)','Activate (demo)')}</Button></div>
      </Drawer>:null}
      {creating?<ConsumerEditor t={t} close={()=>setCreating(false)}/>:null}
    </>}
  </GatewayPage>
}
