import {useMemo,useState} from 'react'
import {ArrowRotateRight,Check,Cloud,Database,Gear,Magnifier} from '@gravity-ui/icons'
import {Button,Icon,Text} from '@gravity-ui/uikit'

import {
  initialSettings,integrationKinds,resetSettingsDemo,saveIntegration,saveProject,testIntegration,
  toggleIntegration,toggleModule,useSettingsDemo,validateModuleToggle,
} from './demo'
import type {Integration,IntegrationKind,ModuleCategory,ModuleDefinition,ProjectSettings} from './demo'
import {Detail,Empty,Input,Metric,Modal,Notice,Panel,Pill,Select,SettingsPage,Textarea} from './SettingsShell'
import type {T} from './SettingsShell'

const categories:ModuleCategory[]=['Core','Infrastructure','Business','AI']

export function SettingsProjectPage(){
  const demo=useSettingsDemo()
  const project=demo.project
  const [name,setName]=useState(project.displayName)
  const [slug,setSlug]=useState(project.slug)
  const [description,setDescription]=useState(project.description)
  const [region,setRegion]=useState<ProjectSettings['region']>(project.region)
  const [environment,setEnvironment]=useState<ProjectSettings['environment']>(project.environment)
  const [namespace,setNamespace]=useState(project.defaultNamespace)
  const [error,setError]=useState('')
  const [success,setSuccess]=useState('')
  const changed=name!==project.displayName||slug!==project.slug||description!==project.description||region!==project.region||environment!==project.environment||namespace!==project.defaultNamespace
  const resetFields=()=>{
    setName(project.displayName);setSlug(project.slug);setDescription(project.description)
    setRegion(project.region);setEnvironment(project.environment);setNamespace(project.defaultNamespace)
    setError('');setSuccess('')
  }
  const save=()=>{
    const issue=saveProject({displayName:name,slug,description,region,environment,defaultNamespace:namespace})
    if(issue){setError(issue);setSuccess('')}
    else{setError('');setSuccess('Saved to in-memory demo store')}
  }
  return <SettingsPage tab="project" title="Project Settings" subtitle="Project metadata, default placement and organizational ownership">
    {t=><>
      <div className="st-section-heading"><div><Text as="h2" variant="header-2">{t('Настройки AI Studio','AI Studio settings')}</Text><Text color="secondary" variant="body-2">{t('Метаданные проекта и параметры размещения ресурсов','Project identity and resource placement defaults')}</Text></div>
        <Button view="outlined" onClick={()=>{resetSettingsDemo();const p=initialSettings.project;setName(p.displayName);setSlug(p.slug);setDescription(p.description);setRegion(p.region);setEnvironment(p.environment);setNamespace(p.defaultNamespace);setError('');setSuccess('')}}><Icon data={ArrowRotateRight} size={16}/>{t('Сбросить демоданные','Reset demo')}</Button>
      </div>
      <div className="st-kpis">
        <Metric label={t('Активные модули','Enabled modules')} value={demo.modules.filter(m=>m.enabled).length} caption={t('Примерный каталог','Sample registry')}/>
        <Metric label={t('Интеграции','Integrations')} value={demo.integrations.length}/>
        <Metric label={t('Вебхуки','Webhooks')} value={demo.webhooks.filter(h=>h.status==='Enabled').length} caption={t('Включено','Enabled')}/>
        <Metric label={t('Версия настроек','Configuration version')} value={project.version} caption={t('Демо-счётчик изменений','Local demo version')}/>
      </div>
      <div className="st-columns">
        <Panel title={t('Свойства проекта','Project properties')} subtitle={t('Редактирование не изменяет реальные организации и окружения','Changes are local; no production resources are updated')}>
          <form className="st-form" onSubmit={e=>{e.preventDefault();save()}}>
            <div className="st-form__two">
              <Input label={t('Название','Display name')} value={name} onChange={v=>{setName(v);setSuccess('')}}/>
              <Input label="Project Slug" value={slug} onChange={v=>{setSlug(v);setSuccess('')}}/>
            </div>
            <Textarea label={t('Описание','Description')} value={description} onChange={v=>{setDescription(v);setSuccess('')}}/>
            <div className="st-form__two">
              <Select label={t('Основной регион','Default region')} value={region} onChange={v=>setRegion(v as ProjectSettings['region'])} options={['eu-1','eu-2','us-1'].map(v=>({value:v,label:v}))}/>
              <Select label={t('Окружение','Environment')} value={environment} onChange={v=>setEnvironment(v as ProjectSettings['environment'])} options={['production','staging','development'].map(v=>({value:v,label:v}))}/>
            </div>
            <Input label={t('Kubernetes namespace по умолчанию','Default Kubernetes namespace')} value={namespace} onChange={setNamespace}/>
            <Notice tone="warning">{t('Изменение region, namespace или slug в реальной платформе требует отдельного workflow миграции. В демо они меняют только поля UI.','Changing region, namespace or slug in production requires a migration workflow. In demo mode only UI fields change.')}</Notice>
            {error?<Notice tone="danger">{error}</Notice>:null}
            {success?<Notice tone="positive">{t('Настройки сохранены в демомодели','Saved to demo model')}</Notice>:null}
            <div className="st-actions st-actions_end"><Button view="outlined" type="button" disabled={!changed} onClick={resetFields}>{t('Отменить изменения','Discard')}</Button><Button view="action" type="submit" disabled={!changed}>{t('Сохранить','Save settings')}</Button></div>
          </form>
        </Panel>
        <Panel title={t('Владение и контекст','Ownership & scope')} subtitle={t('Идентификаторы ресурсов и зависимости модулей','Resource identifiers and module dependencies')}>
          <div className="st-details">
            <Detail label="Project ID" value={<span className="st-mono">{project.id}</span>}/>
            <Detail label={t('Организация','Organization')} value={project.organization}/>
            <Detail label="Scope" value="Project"/>
            <Detail label="Region" value={project.region}/>
            <Detail label={t('Лицензия','Demo license')} value={<Pill tone="info">Professional (demo)</Pill>}/>
            <Detail label={t('Состояние','Status')} value={<Pill tone="positive">Active</Pill>}/>
          </div>
          <Text as="h3" variant="subheader-2">{t('Метки проекта','Project labels')}</Text>
          <div className="st-label-list">{Object.entries(project.labels).map(([k,v])=><div key={k}><span className="st-mono">{k}</span><strong className="st-mono">{v}</strong></div>)}</div>
          <Text as="h3" variant="subheader-2">{t('Связанные сервисы','Connected services')}</Text>
          <div className="st-compact-list">{demo.modules.filter(m=>m.enabled).slice(0,6).map(m=><div key={m.id}><span>{m.title}</span><Pill tone="positive">Enabled</Pill></div>)}</div>
          <Notice>{t('В production источник истины — Resource Manager/YDB; изменение настроек проверяется M8 Access и записывается в Audit.','In production, Resource Manager/YDB owns project metadata; updates require M8 Access authorization and Audit records.')}</Notice>
        </Panel>
      </div>
      <Panel title={t('Изменения в демосессии','Demo session activity')} subtitle={t('История существует только в памяти вкладки','Only retained in current browser memory')}>
        {demo.activity.length?<div className="st-activity">{demo.activity.slice(0,6).map((entry,i)=><div key={i}><Icon data={Check} size={16}/>{entry}</div>)}</div>:<Empty>{t('Изменений пока нет','No changes yet')}</Empty>}
      </Panel>
    </>}
  </SettingsPage>
}

function ModuleDetails({module,close,t}:{module:ModuleDefinition;close:()=>void;t:T}){
  const demo=useSettingsDemo()
  return <Modal title={module.title} close={close}>
    <div className="st-details">
      <Detail label="Module ID" value={<span className="st-mono">{module.id}</span>}/>
      <Detail label="Bounded context" value={module.category}/>
      <Detail label="Licensing tier" value={<Pill>{module.tier}</Pill>}/>
      <Detail label="State" value={<Pill tone={module.enabled?'positive':'warning'}>{module.enabled?'Enabled':'Disabled'}</Pill>}/>
      <Detail label="Required" value={module.required?'Yes':'No'}/>
    </div>
    <Text color="secondary" variant="body-2">{module.description}</Text>
    <Text as="h3" variant="subheader-2">{t('Capability IDs','Capability IDs')}</Text>
    <div className="st-tags">{module.capabilities.map(c=><Pill key={c}>{c}</Pill>)}</div>
    <Text as="h3" variant="subheader-2">{t('Зависимости','Dependencies')}</Text>
    {module.dependencies.length?<div className="st-compact-list">{module.dependencies.map(dep=><div key={dep}><span>{demo.modules.find(m=>m.id===dep)?.title??dep}</span><Pill tone={demo.modules.find(m=>m.id===dep)?.enabled?'positive':'warning'}>{demo.modules.find(m=>m.id===dep)?.enabled?'Enabled':'Disabled'}</Pill></div>)}</div>:<Text color="secondary">{t('Нет зависимостей','No dependencies')}</Text>}
    <Notice tone="warning">{t('Реальное включение модуля требует лицензирования, provisioning, миграций и проверки API/CRD. Здесь переключается лишь демосостояние.','Actual enablement requires licensing, provisioning, migrations and API/CRD checks. This only changes local demo state.')}</Notice>
    <div className="st-actions st-actions_end"><Button view="outlined" onClick={close}>{t('Закрыть','Close')}</Button></div>
  </Modal>
}
export function SettingsModulesPage(){
  const demo=useSettingsDemo()
  const [query,setQuery]=useState('')
  const [category,setCategory]=useState('All')
  const [status,setStatus]=useState('All')
  const [selected,setSelected]=useState<string|null>(null)
  const [error,setError]=useState('')
  const filtered=useMemo(()=>demo.modules.filter(m=>
    (category==='All'||m.category===category)&&(status==='All'||(status==='Enabled')===m.enabled)&&
    [m.title,m.id,m.description,...m.capabilities].some(str=>str.toLowerCase().includes(query.toLowerCase()))
  ),[demo.modules,query,category,status])
  const toggle=(id:string)=>{
    const error=toggleModule(id)
    setError(error??'')
  }
  return <SettingsPage tab="modules" title="Modules" subtitle="M8 Module Registry, capabilities, dependencies and licensing">
    {t=><>
      <div className="st-section-heading"><div><Text as="h2" variant="header-2">{t('Каталог модулей','Module catalog')}</Text><Text variant="body-2" color="secondary">{t('Демо-подмножество Module Registry / Capability Map','Demo subset of the Module Registry / Capability Map')}</Text></div><Pill tone="info">Professional license (demo)</Pill></div>
      <div className="st-kpis">
        <Metric label={t('В каталоге','Catalog entries')} value={demo.modules.length}/>
        <Metric label="Enabled" value={demo.modules.filter(m=>m.enabled).length}/>
        <Metric label={t('Не включено','Disabled')} value={demo.modules.filter(m=>!m.enabled).length}/>
        <Metric label={t('Категорий','Categories')} value={categories.length}/>
      </div>
      <Panel title={t('Выбор модулей','Module selection')} subtitle={t('Для каждого модуля показаны tier, capabilities и обязательные зависимости','Each module declares tier, capabilities and required dependencies')}>
        <div className="st-toolbar">
          <div className="st-search"><Icon data={Magnifier} size={16}/><input value={query} onChange={e=>setQuery(e.target.value)} placeholder={t('Имя, ID или capability','Name, ID or capability')} aria-label="Search modules"/></div>
          <Select label={t('Категория','Category')} value={category} onChange={setCategory} options={[{value:'All',label:t('Все','All')},...categories.map(x=>({value:x,label:x}))]}/>
          <Select label="Status" value={status} onChange={setStatus} options={['All','Enabled','Disabled'].map(x=>({value:x,label:x}))}/>
        </div>
        {error?<Notice tone="warning">{error}</Notice>:null}
        <div className="st-module-grid">{filtered.map(module=>{
          const changeProblem=validateModuleToggle(module.id,!module.enabled)
          return <div className="st-module" key={module.id}>
            <div className="st-module__head"><div className="st-module__icon"><Icon data={module.category==='Infrastructure'?Cloud:module.category==='Core'?Gear:Database} size={19}/></div><div><strong>{module.title}</strong><span className="st-mono">{module.id}</span></div><Pill tone={module.enabled?'positive':'neutral'}>{module.enabled?'Enabled':'Disabled'}</Pill></div>
            <p>{module.description}</p>
            <div className="st-tags"><Pill>{module.category}</Pill><Pill tone={module.tier==='Enterprise'?'warning':'neutral'}>{module.tier}</Pill>{module.required?<Pill tone="info">Required</Pill>:null}</div>
            <div className="st-module__foot"><span>{module.capabilities.length} capabilities · {module.dependencies.length} deps</span><div className="st-actions">
              <Button size="s" view="outlined" onClick={()=>setSelected(module.id)}>{t('Детали','Details')}</Button>
              <Button size="s" view={module.enabled?'outlined':'action'} disabled={!!changeProblem} onClick={()=>toggle(module.id)}>{module.enabled?t('Отключить','Disable'):t('Включить','Enable')}</Button>
            </div></div>
            {changeProblem?<span className="st-module__reason">{changeProblem}</span>:null}
          </div>
        })}</div>
        {!filtered.length?<Empty>{t('Модули не найдены','No modules found')}</Empty>:null}
      </Panel>
      <Notice tone="warning">{t('Каталог содержит 14 примеров, а не полный перечень M8. Состояние в UI не подтверждает наличие установленного оператора или лицензии.','The catalog contains 14 examples, not the full M8 registry. Enabled in demo does not mean the operator is deployed or licensed.')}</Notice>
      {selected&&demo.modules.find(m=>m.id===selected)?<ModuleDetails module={demo.modules.find(m=>m.id===selected)!} close={()=>setSelected(null)} t={t}/>:null}
    </>}
  </SettingsPage>
}

function IntegrationEditor({current,close,t}:{current?:Integration;close:()=>void;t:T}){
  const [name,setName]=useState(current?.name??'')
  const [kind,setKind]=useState<IntegrationKind>(current?.kind??'Kubernetes')
  const [endpoint,setEndpoint]=useState(current?.endpoint??'https://api.demo.example')
  const [secretRef,setSecretRef]=useState(current?.secretRef??'')
  const [environment,setEnvironment]=useState<Integration['environment']>(current?.environment??'production')
  const [error,setError]=useState('')
  return <Modal title={current?t('Изменить интеграцию','Edit integration'):t('Добавить интеграцию','Add integration')} close={close}>
    <form className="st-form" onSubmit={e=>{
      e.preventDefault()
      const issue=saveIntegration({name,kind,endpoint,secretRef,environment},current?.id)
      if(issue)setError(issue);else close()
    }}>
      <Input label={t('Название','Name')} value={name} onChange={setName} placeholder="My Kubernetes Cluster"/>
      <div className="st-form__two">
        <Select label={t('Тип','Type')} value={kind} onChange={v=>setKind(v as IntegrationKind)} options={integrationKinds.map(v=>({value:v,label:v}))}/>
        <Select label={t('Окружение','Environment')} value={environment} onChange={v=>setEnvironment(v as Integration['environment'])} options={['production','development'].map(v=>({value:v,label:v}))}/>
      </div>
      <Input label={t('Адрес endpoint (HTTPS)','Endpoint URL (HTTPS)')} value={endpoint} onChange={setEndpoint} placeholder="https://service.example.com"/>
      <Input label={t('Ссылка на секрет в Vault','Secret reference (Vault path)')} value={secretRef} onChange={setSecretRef} placeholder="integrations/my-connector"/>
      <Notice tone="warning">{t('Сюда вводится только путь к секрету, не токен/пароль. Не выполняются запросы к указанному адресу.','Enter only a vault secret reference, never a password or token. No requests will be sent to this URL.')}</Notice>
      {error?<Notice tone="danger">{error}</Notice>:null}
      <div className="st-actions st-actions_end"><Button type="button" view="outlined" onClick={close}>{t('Отмена','Cancel')}</Button><Button type="submit" view="action">{t('Сохранить демоданные','Save demo')}</Button></div>
    </form>
  </Modal>
}
export function SettingsIntegrationsPage(){
  const demo=useSettingsDemo()
  const [query,setQuery]=useState('')
  const [kind,setKind]=useState('All')
  const [selected,setSelected]=useState<string|null>(null)
  const [editor,setEditor]=useState<string|null>(null)
  const [creating,setCreating]=useState(false)
  const [notice,setNotice]=useState('')
  const visible=demo.integrations.filter(i=>(kind==='All'||i.kind===kind)&&[i.name,i.kind,i.endpoint].some(x=>x.toLowerCase().includes(query.toLowerCase())))
  const selectedValue=demo.integrations.find(i=>i.id===selected)
  const editing=demo.integrations.find(i=>i.id===editor)
  return <SettingsPage tab="integrations" title="Integrations" subtitle="Provider connectors, observability backends and external service metadata">
    {t=><>
      <div className="st-section-heading"><div><Text as="h2" variant="header-2">{t('Внешние интеграции','External integrations')}</Text><Text variant="body-2" color="secondary">{t('Подключения описаны метаданными без токенов и паролей','Connector metadata without exposed credentials')}</Text></div><Button view="action" onClick={()=>setCreating(true)}>+ {t('Добавить интеграцию','Add integration')}</Button></div>
      <div className="st-kpis">
        <Metric label={t('Всего','Total integrations')} value={demo.integrations.length}/>
        <Metric label="Configured" value={demo.integrations.filter(i=>i.status==='Configured').length}/>
        <Metric label={t('Требуют настройки','Needs setup')} value={demo.integrations.filter(i=>i.status==='Needs setup').length} tone="warning"/>
        <Metric label="Disabled" value={demo.integrations.filter(i=>i.status==='Disabled').length}/>
      </div>
      <Panel title={t('Каталог интеграций','Connector registry')} subtitle={t('Configured — только демостатус, соединение не проверялось','Configured indicates demo metadata only; connectivity is unverified')}>
        <div className="st-toolbar"><div className="st-search"><Icon data={Magnifier} size={16}/><input aria-label="Search integrations" value={query} onChange={e=>setQuery(e.target.value)} placeholder={t('Поиск по имени / endpoint','Search name / endpoint')}/></div><Select label="Type" value={kind} onChange={setKind} options={[{value:'All',label:'All'},...integrationKinds.map(v=>({value:v,label:v}))]}/></div>
        <div className="st-table-scroll"><table className="st-table"><thead><tr><th>{t('Интеграция','Integration')}</th><th>Type</th><th>Endpoint</th><th>{t('Состояние','Status')}</th><th>{t('Проверка','Checked')}</th><th/></tr></thead>
        <tbody>{visible.map(i=><tr key={i.id}><td><strong>{i.name}</strong><span className="st-muted">{i.id}</span></td><td>{i.kind}</td><td className="st-mono">{i.endpoint}</td><td><Pill tone={i.status==='Configured'?'positive':i.status==='Disabled'?'neutral':'warning'}>{i.status}</Pill></td><td>{i.lastChecked}</td><td><Button view="outlined" size="s" onClick={()=>setSelected(i.id)}>{t('Детали','Details')}</Button></td></tr>)}</tbody></table></div>
        {!visible.length?<Empty>{t('Нет подходящих интеграций','No matching integrations')}</Empty>:null}
      </Panel>
      {notice?<Notice tone="info">{notice}</Notice>:null}
      <div className="st-columns">
        <Panel title={t('Секреты и доступ','Credentials and access')}>
          <div className="st-guidance"><div><strong>Vault SecretRef</strong><span>{t('Сохраняются только ссылки на секреты; значения не попадают в UI или audit payload.','Only secret references are held in UI; values must never enter UI or audit payloads.')}</span></div><div><strong>M8 Access / RBAC</strong><span>{t('Изменения интеграции требуют разрешения на уровне organization/project.','Organization/project permissions authorize connector changes.')}</span></div></div>
        </Panel>
        <Panel title={t('Provisioning и сверка','Provider reconciliation')}>
          <div className="st-guidance"><div><strong>Resource Type Registry</strong><span>{t('Проверка схемы коннектора и поддерживаемых capability IDs.','Validate connector schema and supported capability IDs.')}</span></div><div><strong>Temporal</strong><span>{t('Асинхронные операции подключения, health-check и backoff.','Durable connection jobs, health checks and backoff policies.')}</span></div></div>
        </Panel>
      </div>
      {selectedValue?<Modal title={selectedValue.name} close={()=>setSelected(null)}>
        <div className="st-details"><Detail label="Kind" value={selectedValue.kind}/><Detail label="Endpoint" value={<span className="st-mono">{selectedValue.endpoint}</span>}/><Detail label="Status" value={<Pill>{selectedValue.status}</Pill>}/><Detail label="Environment" value={selectedValue.environment}/><Detail label="SecretRef" value={<span className="st-mono">{selectedValue.secretRef||'Not configured'}</span>}/><Detail label={t('Последняя проверка','Last checked')} value={selectedValue.lastChecked}/></div>
        <Notice tone="warning">{selectedValue.testNote}. {t('Действие Test выполняет только локальную симуляцию без сетевого подключения.','Test action only changes local demo state; it never connects to an endpoint.')}</Notice>
        <div className="st-actions st-actions_end"><Button view="outlined" onClick={()=>{testIntegration(selectedValue.id);setNotice(t('Локальная симуляция выполнена — соединение не проверялось','Demo check completed — no connection was verified'))}}>{t('Тест (демо)','Test (demo)')}</Button><Button view="outlined" onClick={()=>{toggleIntegration(selectedValue.id);setNotice(t('Изменено демосостояние','Demo state changed'))}}>{selectedValue.status==='Disabled'?t('Включить','Enable'):t('Отключить','Disable')}</Button><Button view="action" onClick={()=>{setEditor(selectedValue.id);setSelected(null)}}>{t('Редактировать','Edit')}</Button></div>
      </Modal>:null}
      {(creating||editing)?<IntegrationEditor key={editing?.id??'new'} current={editing} t={t} close={()=>{setCreating(false);setEditor(null)}}/>:null}
    </>}
  </SettingsPage>
}
