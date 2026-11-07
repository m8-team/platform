import {useMemo,useState} from 'react'
import {ArrowRotateRight,Check,Clock,TriangleExclamation} from '@gravity-ui/icons'
import {Button,Icon,Text} from '@gravity-ui/uikit'
import {Link,useParams} from '@tanstack/react-router'

import {
  changeDemoIncident,createDemoRule,currentMetric,formatMetric,metricInfo,resetObservabilityDemo,
  serviceNames,sloBudget,toggleDemoRule,useObservabilityDemo,validateRule,
} from './demo'
import type {AlertRule,AlertRuleMetric,AlertSeverity,AlertStatus,ServiceSlo,TelemetryService} from './demo'
import {
  Bar,Detail,Empty,IncidentBadge,Input,Label,Notice,Page,Panel,Select,SeverityBadge,Stat,
  Sparkline,utc,
} from './ObservabilityShell'
import type {T} from './ObservabilityShell'

const severities:AlertSeverity[]=['Critical','Warning','Info']
const metrics:AlertRuleMetric[]=['errorRate','latencyP95','cpu']

function RuleForm({t,close}:{t:T;close:()=>void}){
  const [name,setName]=useState('')
  const [service,setService]=useState<TelemetryService>('ai-gateway')
  const [metric,setMetric]=useState<AlertRuleMetric>('errorRate')
  const [comparator,setComparator]=useState<AlertRule['comparator']>('gt')
  const [threshold,setThreshold]=useState('2')
  const [duration,setDuration]=useState('5')
  const [severity,setSeverity]=useState<AlertSeverity>('Warning')
  const [description,setDescription]=useState('')
  const [error,setError]=useState('')
  const draft={name,service,metric,comparator,threshold:Number(threshold),durationMinutes:Number(duration)}
  const previewError=validateRule(draft)
  return <div className="ob-modal-host">
    <button type="button" className="ob-modal__backdrop" onClick={close} aria-label="Close rule editor"/>
    <section className="ob-modal" role="dialog" aria-modal="true" aria-label={t('Создать алерт','Create alert rule')}>
      <header className="ob-modal__head"><Text as="h2" variant="header-2">{t('Новая алерт-политика','New alert rule')}</Text><Button view="outlined" onClick={close}>✕</Button></header>
      <form className="ob-form" onSubmit={e=>{
        e.preventDefault()
        const issue=createDemoRule({...draft,description,severity})
        if(issue)setError(issue);else close()
      }}>
        <Input label={t('Название','Name')} value={name} onChange={setName} placeholder="Critical error rate"/>
        <div className="ob-form__two">
          <Select label="Service" value={service} onChange={v=>setService(v as TelemetryService)} options={serviceNames.map(s=>({value:s,label:s}))}/>
          <Select label={t('Метрика','Metric')} value={metric} onChange={v=>setMetric(v as AlertRuleMetric)} options={metrics.map(m=>({value:m,label:metricInfo[m].name}))}/>
        </div>
        <div className="ob-form__two">
          <Select label={t('Условие','Condition')} value={comparator} onChange={v=>setComparator(v as AlertRule['comparator'])} options={[{value:'gt',label:'Greater than'},{value:'lt',label:'Less than'}]}/>
          <Input label={`${t('Порог','Threshold')} (${metricInfo[metric].unit})`} type="number" value={threshold} onChange={setThreshold}/>
        </div>
        <div className="ob-form__two">
          <Input label={t('Длительность, минуты','For (minutes)')} type="number" value={duration} onChange={setDuration}/>
          <Select label="Severity" value={severity} onChange={v=>setSeverity(v as AlertSeverity)} options={severities.map(s=>({value:s,label:s}))}/>
        </div>
        <label className="ob-field"><span>{t('Описание','Description')}</span><textarea className="ob-textarea" rows={3} value={description} onChange={e=>setDescription(e.target.value)}/></label>
        <Notice>{t('Политика добавляется только в локальную демомодель; Alertmanager правила и уведомления не создаются.','This only adds local demo metadata; no Alertmanager rule or notification is created.')}</Notice>
        {error?<Notice tone="danger">{error}</Notice>:null}
        {previewError&&name.length>0?<Notice tone="warning">{previewError}</Notice>:null}
        <div className="ob-modal__actions"><Button type="button" view="outlined" onClick={close}>{t('Отмена','Cancel')}</Button><Button type="submit" view="action" disabled={!!previewError}>{t('Создать демо-правило','Create demo rule')}</Button></div>
      </form>
    </section>
  </div>
}

export function ObservabilityAlertsPage(){
  const demo=useObservabilityDemo()
  const [service,setService]=useState('All')
  const [status,setStatus]=useState('All')
  const [tab,setTab]=useState<'incidents'|'rules'>('incidents')
  const [creating,setCreating]=useState(false)
  const incidents=useMemo(()=>demo.incidents.filter(i=>(service==='All'||i.service===service)&&(status==='All'||i.status===status)),[demo.incidents,service,status])
  const rules=demo.rules.filter(r=>service==='All'||r.service===service)
  return <Page tab="alerts" title="Alerts" subtitle="M8 incident response, alert rule catalog and on-call investigation workflow">
    {t=><>
      <div className="ob-section-head">
        <div><Text as="h2" variant="header-2">{t('Алерты и инциденты','Alerts & incidents')}</Text><Text variant="body-2" color="secondary">{t('Обнаружение, подтверждение и разбор событий доступности','Detect, acknowledge and investigate reliability problems')}</Text></div>
        <div className="ob-actions"><Button view="outlined" onClick={resetObservabilityDemo}><Icon data={ArrowRotateRight} size={15}/>{t('Сбросить демо','Reset demo')}</Button><Button view="action" onClick={()=>setCreating(true)}>+ {t('Создать правило','New alert rule')}</Button></div>
      </div>
      <div className="ob-stats">
        <Stat label="Firing" value={demo.incidents.filter(i=>i.status==='Firing').length} tone="danger"/>
        <Stat label="Acknowledged" value={demo.incidents.filter(i=>i.status==='Acknowledged').length} tone="warning"/>
        <Stat label="Resolved" value={demo.incidents.filter(i=>i.status==='Resolved').length} tone="positive"/>
        <Stat label={t('Активные правила','Enabled rules')} value={demo.rules.filter(r=>r.enabled).length}/>
      </div>
      <div className="ob-filters">
        <Select label={t('Сервис','Service')} value={service} onChange={setService} options={[{value:'All',label:t('Все сервисы','All services')},...serviceNames.map(s=>({value:s,label:s}))]}/>
        {tab==='incidents'?<Select label={t('Состояние','Status')} value={status} onChange={setStatus} options={['All','Firing','Acknowledged','Resolved','Muted'].map(s=>({value:s,label:s}))}/>:null}
      </div>
      <div className="ob-inner-tabs" role="tablist">
        <button type="button" role="tab" aria-selected={tab==='incidents'} className={tab==='incidents'?'ob-inner-tab_active':''} onClick={()=>setTab('incidents')}>{t('Инциденты','Incidents')} ({incidents.length})</button>
        <button type="button" role="tab" aria-selected={tab==='rules'} className={tab==='rules'?'ob-inner-tab_active':''} onClick={()=>setTab('rules')}>{t('Правила алертов','Alert rules')} ({rules.length})</button>
      </div>
      {tab==='incidents'?<Panel title={t('Инциденты','Incidents')} subtitle={t('Состояния изменяются только в демореестре','Status actions affect only the demo store')}>
        <div className="ob-table-scroll"><table className="ob-table"><thead><tr><th>{t('Инцидент','Incident')}</th><th>Service</th><th>Severity</th><th>{t('Статус','Status')}</th><th>{t('Начало UTC','Started UTC')}</th><th/></tr></thead>
          <tbody>{incidents.map(i=><tr key={i.id}>
            <td><Link className="ob-link" to="/observability/alerts/$incidentId" params={{incidentId:i.id}}>{i.title}</Link><span className="ob-muted ob-mono">{i.id}</span></td>
            <td>{i.service}</td><td><SeverityBadge severity={i.severity}/></td><td><IncidentBadge status={i.status}/></td><td className="ob-nowrap">{i.startedAt.slice(11,19)}</td>
            <td><Link className="ob-link" to="/observability/alerts/$incidentId" params={{incidentId:i.id}}>{t('Детали','Details')}</Link></td>
          </tr>)}</tbody></table></div>
        {!incidents.length?<Empty>{t('Инциденты не найдены','No matching incidents')}</Empty>:null}
      </Panel>:
        <Panel title={t('Правила мониторинга','Alert rule catalog')} subtitle={t('Синтетические метрики и локальные переключатели Enabled/Disabled','Synthetic metrics and local enable/disable switches')}>
          <div className="ob-table-scroll"><table className="ob-table"><thead><tr><th>{t('Правило','Rule')}</th><th>Service</th><th>{t('Условие','Condition')}</th><th>Severity</th><th>Status</th><th/></tr></thead>
            <tbody>{rules.map(r=><tr key={r.id}>
              <td><strong>{r.name}</strong><span className="ob-muted">{r.description}</span></td><td>{r.service}</td>
              <td className="ob-nowrap">{metricInfo[r.metric].name} {r.comparator==='gt'?' > ':' < '}{r.threshold}{metricInfo[r.metric].unit} / {r.durationMinutes}m</td>
              <td><SeverityBadge severity={r.severity}/></td><td><Label tone={r.enabled?'positive':'neutral'}>{r.enabled?'Enabled':'Disabled'}</Label></td>
              <td><Button size="s" view="outlined" onClick={()=>toggleDemoRule(r.id)}>{r.enabled?t('Отключить','Disable'):t('Включить','Enable')}</Button></td>
            </tr>)}</tbody></table></div>
        </Panel>}
      <div className="ob-grid ob-grid_2">
        <Panel title={t('Процесс реагирования','Incident response')}>
          <div className="ob-guidance">
            <div><strong>Firing</strong><span>{t('Условие алерта наблюдается и требует диагностики','Alert is firing and needs attention')}</span></div>
            <div><strong>Acknowledged</strong><span>{t('Оператор принял сигнал, это не означает восстановление','An operator owns the incident; service may still be impaired')}</span></div>
            <div><strong>Resolved</strong><span>{t('Демонстрационная отметка, не подтверждённая провайдером','Marked resolved in demo; not verified by providers')}</span></div>
          </div>
        </Panel>
        <Panel title={t('Источники алертов','Alert backends')}>
          <div className="ob-guidance"><div><strong>Mimir Ruler / Prometheus</strong><span>{t('Вычисление правил PromQL в production','Evaluate PromQL expressions in production')}</span></div>
          <div><strong>Alertmanager</strong><span>{t('Группировка, маршрутизация, mute и on-call','Alert routing, grouping, silences and on-call')}</span></div>
          <div><strong>M8 Access</strong><span>{t('Права на подтверждение, изменение правил и эскалацию','Authorize incident response, rule edits and escalation')}</span></div></div>
        </Panel>
      </div>
      {creating?<RuleForm t={t} close={()=>setCreating(false)}/>:null}
    </>}
  </Page>
}

export function ObservabilityAlertDetailsPage(){
  const {incidentId}=useParams({from:'/observability/alerts/$incidentId'})
  const demo=useObservabilityDemo()
  const incident=demo.incidents.find(i=>i.id===incidentId)
  const rule=incident?demo.rules.find(r=>r.id===incident.ruleId):undefined
  return <Page tab="alerts" title={incident?.title??'Incident'} subtitle={incident?utc(incident.startedAt):'Incident not found'}>{t=>
    !incident?<Panel title={t('Инцидент не найден','Incident not found')}><Link className="ob-link" to="/observability/alerts">{t('Назад к алертам','Back to alerts')}</Link></Panel>:<>
      <div className="ob-section-head"><Link to="/observability/alerts" className="ob-link">‹ {t('Все алерты','All alerts')}</Link>
        <div className="ob-actions"><SeverityBadge severity={incident.severity}/><IncidentBadge status={incident.status}/></div>
      </div>
      <div className="ob-grid ob-grid_2">
        <Panel title={t('Инцидент','Incident details')}>
          <div className="ob-details">
            <Detail label="Incident ID" value={<span className="ob-mono">{incident.id}</span>}/>
            <Detail label="Service" value={incident.service}/>
            <Detail label={t('Начало','Started')} value={utc(incident.startedAt)}/>
            <Detail label={t('Изменён','Updated')} value={utc(incident.updatedAt)}/>
            <Detail label={t('Наблюдаемое значение','Observed value')} value={incident.value}/>
            <Detail label={t('Описание','Description')} value={incident.summary}/>
          </div>
        </Panel>
        <Panel title={t('Корреляции и правило','Correlations & rule')}>
          <div className="ob-details">
            <Detail label="Rule" value={rule?.name??incident.ruleId}/>
            <Detail label="Threshold" value={rule?`${rule.comparator==='gt'?'>':'<'} ${rule.threshold} ${metricInfo[rule.metric].unit}`:'—'}/>
            <Detail label="Trace ID" value={incident.traceId?<Link className="ob-link" to="/observability/traces/$traceId" params={{traceId:incident.traceId}}>{incident.traceId}</Link>:'—'}/>
            <Detail label="Operation ID" value={incident.operationId?<Link className="ob-link" to="/platform-operations/long-running/$operationId" params={{operationId:incident.operationId}}>{incident.operationId}</Link>:'—'}/>
          </div>
        </Panel>
      </div>
      <Panel title={t('История инцидента','Incident timeline')}>
        <ol className="ob-timeline">{incident.timeline.map((entry,i)=><li key={i}>{entry}</li>)}</ol>
        <div className="ob-actions ob-actions_end">
          {incident.status==='Firing'?<Button view="outlined" onClick={()=>changeDemoIncident(incident.id,'acknowledge')}><Icon data={Clock} size={16}/>{t('Подтвердить (демо)','Acknowledge (demo)')}</Button>:null}
          {(incident.status==='Firing'||incident.status==='Acknowledged')?<Button view="outlined" onClick={()=>changeDemoIncident(incident.id,'mute')}><Icon data={TriangleExclamation} size={16}/>{t('Заглушить (демо)','Mute (demo)')}</Button>:null}
          {incident.status!=='Resolved'?<Button view="action" onClick={()=>changeDemoIncident(incident.id,'resolve')}><Icon data={Check} size={16}/>{t('Разрешить (демо)','Resolve (demo)')}</Button>:null}
        </div>
      </Panel>
      <Notice tone="warning">{t('Демо-действия не отключают реальные уведомления и не подтверждают восстановление сервиса. Продакшен должен проверять права и хранить аудит действий.','Demo actions do not silence real alerts or verify service recovery. Production must enforce permissions and audit every incident action.')}</Notice>
    </>
  }</Page>
}

function SloCard({slo,t}:{slo:ServiceSlo;t:T}){
  const budget=sloBudget(slo)
  const passing=slo.achieved>=slo.target
  return <div className="ob-slo-card">
    <div className="ob-slo-card__head">
      <div><strong>{slo.name}</strong><span>{slo.service} · {slo.window} · {slo.metric}</span></div>
      <Label tone={passing?'positive':'danger'}>{passing?'Healthy':'Breached'}</Label>
    </div>
    <div className="ob-slo-card__numbers"><span><small>SLI observed</small><strong>{slo.achieved.toFixed(3)}%</strong></span><span><small>SLO target</small><strong>{slo.target.toFixed(3)}%</strong></span></div>
    <Bar value={slo.achieved} max={100} variant={passing?'positive':'danger'}/>
    <div className="ob-slo-card__bottom"><span>Error budget remaining</span><strong className={budget.remaining<0?'ob-negative':''}>{budget.remaining.toLocaleString('en-US')} requests</strong></div>
  </div>
}

export function ObservabilitySloPage(){
  const demo=useObservabilityDemo()
  const [service,setService]=useState('All')
  const [window,setWindow]=useState('All')
  const [expanded,setExpanded]=useState('')
  const filtered=useMemo(()=>demo.slos.filter(s=>(service==='All'||s.service===service)&&(window==='All'||s.window===window)),[demo.slos,service,window])
  const breached=filtered.filter(s=>s.achieved<s.target)
  const selected=filtered.find(s=>s.id===expanded)??filtered[0]
  return <Page tab="slo" title="Service Level Objectives" subtitle="Availability and latency reliability objectives, error-budget burn and service ownership">
    {t=><>
      <div className="ob-section-head"><div><Text as="h2" variant="header-2">{t('Надёжность сервисов','Service reliability')}</Text><Text variant="body-2" color="secondary">{t('Цели SLO, SLI и оставшийся бюджет ошибок за окно','SLO targets, SLI results and remaining error budgets per window')}</Text></div><Label tone="info">Synthetic SLI data</Label></div>
      <div className="ob-filters">
        <Select label="Service" value={service} onChange={setService} options={[{value:'All',label:t('Все сервисы','All services')},...serviceNames.map(s=>({value:s,label:s}))]}/>
        <Select label={t('Период','Window')} value={window} onChange={setWindow} options={[{value:'All',label:t('Все','All')},{value:'7d',label:'7 days'},{value:'30d',label:'30 days'}]}/>
      </div>
      <div className="ob-stats">
        <Stat label={t('Целей SLO','Tracked SLOs')} value={filtered.length}/>
        <Stat label={t('В пределах цели','Meeting target')} value={filtered.length-breached.length} tone="positive"/>
        <Stat label={t('Нарушено','Breached')} value={breached.length} tone="danger"/>
        <Stat label={t('Полностью исчерпан budget','Budget exhausted')} value={filtered.filter(s=>sloBudget(s).remaining<0).length} tone="warning"/>
      </div>
      <Panel title={t('Цели надёжности','Reliability objectives')} subtitle={t('Для каждого окна бюджет рассчитывается по good/total events','Error budget is derived from good/total events within each window')}>
        <div className="ob-slo-grid">{filtered.map(s=><button type="button" key={s.id} className={`ob-slo-pressable ${selected?.id===s.id?'ob-slo-pressable_active':''}`} onClick={()=>setExpanded(s.id)}><SloCard slo={s} t={t}/></button>)}</div>
        {!filtered.length?<Empty>{t('Нет SLO для выбранного фильтра','No SLO matches the selected filters')}</Empty>:null}
      </Panel>
      {selected?<div className="ob-grid ob-grid_2">
        <Panel title={t('Расчёт бюджета ошибок','Error budget calculation')} subtitle={selected.name}>
          <div className="ob-details">
            <Detail label="SLO target" value={`${selected.target}%`}/>
            <Detail label="Observed SLI" value={`${selected.achieved}%`}/>
            <Detail label={t('Всего событий','Total events')} value={selected.total.toLocaleString('en-US')}/>
            <Detail label={t('Хороших событий','Good events')} value={selected.good.toLocaleString('en-US')}/>
            <Detail label={t('Допустимые ошибки','Allowed bad events')} value={sloBudget(selected).allowedBad.toLocaleString('en-US')}/>
            <Detail label={t('Фактические ошибки','Actual bad events')} value={sloBudget(selected).bad.toLocaleString('en-US')}/>
            <Detail label={t('Остаток бюджета','Remaining budget')} value={sloBudget(selected).remaining.toLocaleString('en-US')}/>
          </div>
          <Notice tone={sloBudget(selected).remaining<0?'danger':'info'}>{selected.description}</Notice>
        </Panel>
        <Panel title={t('Тенденция и burn rate','Trend & burn rate')} subtitle={t('Иллюстрация; не рассчитано из реальных окон наблюдения','Illustrative trend; not measured from real rolling windows')}>
          <Sparkline values={[99.97,99.94,99.91,99.88,99.9,99.84,99.76,selected.achieved]} height={140} variant={selected.achieved<selected.target?'danger':'positive'}/>
          <div className="ob-guidance">
            <div><strong>Current budget use</strong><span>{sloBudget(selected).usedPercent.toFixed(1)}% {t('из доступного бюджета ошибок','of the allocated error budget')}</span></div>
            <div><strong>Burn-rate alerts</strong><span>{t('Рекомендация: multi-window burn rate, Alertmanager и on-call','Recommended: multi-window burn-rate alerts and Alertmanager on-call')}</span></div>
          </div>
          <Link className="ob-link" to="/observability/alerts">{t('Открыть алерты','View alerts')}</Link>
        </Panel>
      </div>:null}
      <Notice tone="warning">{t('SLO demo: значения агрегированы и не являются доказательством доступности. Целевые показатели должны вычисляться сервером по утверждённым правилам SLI и календарным окнам.','SLO demo: aggregated values are illustrative, not availability evidence. Production SLI must be computed server-side from approved definitions and window boundaries.')}</Notice>
    </>}
  </Page>
}
