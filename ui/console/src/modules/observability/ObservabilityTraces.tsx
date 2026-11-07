import {useMemo,useState} from 'react'
import {Text} from '@gravity-ui/uikit'
import {Link,useParams} from '@tanstack/react-router'

import {searchTraces,useObservabilityDemo,validateTrace,serviceNames} from './demo'
import type {TelemetryTrace,TraceSpan} from './demo'
import {
  Detail,Empty,Input,Label,Notice,Page,Panel,Select,Stat,TraceBadge,utc,
} from './ObservabilityShell'

function spanDepth(span:TraceSpan,trace:TelemetryTrace):number{
  let parent=span.parentId,depth=0
  const visited=new Set([span.id])
  while(parent&&depth<20){
    if(visited.has(parent))break
    visited.add(parent)
    const candidate=trace.spans.find(s=>s.id===parent)
    if(!candidate)break
    depth++
    parent=candidate.parentId
  }
  return depth
}
export function ObservabilityTracesPage(){
  const demo=useObservabilityDemo()
  const [service,setService]=useState('All')
  const [status,setStatus]=useState('All')
  const [query,setQuery]=useState('')
  const [slowOnly,setSlowOnly]=useState(false)
  const filtered=useMemo(()=>searchTraces({service,status,query},demo.traces)
    .filter(t=>!slowOnly||t.durationMs>=1000),[demo.traces,service,status,query,slowOnly])
  return <Page tab="traces" title="Traces" subtitle="Tempo-compatible trace explorer, spans, dependencies and request correlation">
    {t=><>
      <div className="ob-section-head"><div><Text as="h2" variant="header-2">{t('Распределённые трассировки','Distributed Tracing')}</Text><Text variant="body-2" color="secondary">{t('Задержки и ошибки в цепочках запросов разных сервисов','Latency and failures across multi-service request paths')}</Text></div><Label tone="info">{filtered.length} traces</Label></div>
      <div className="ob-stats">
        <Stat label={t('Трассировок','Total traces')} value={filtered.length}/>
        <Stat label={t('Ошибки','Failed traces')} value={filtered.filter(tr=>tr.status==='ERROR').length} tone="danger"/>
        <Stat label={t('Медленные > 1с','Slow > 1s')} value={filtered.filter(tr=>tr.durationMs>1000).length} tone="warning"/>
        <Stat label={t('Сервисов','Services')} value={new Set(filtered.flatMap(tr=>tr.spans.map(s=>s.service))).size}/>
      </div>
      <Panel title={t('Фильтры трассировок','Trace filters')} subtitle={t('Данные синтетические; поисковые выражения Tempo TraceQL не исполняются','Synthetic data; Tempo TraceQL queries are not executed')}>
        <div className="ob-filters">
          <Input label={t('Поиск по trace/operation/span','Search trace, operation or span')} value={query} onChange={setQuery} placeholder="trace-3f29 / op-2026-1041"/>
          <Select label="Service" value={service} onChange={setService} options={[{value:'All',label:t('Все сервисы','All services')},...serviceNames.map(s=>({value:s,label:s}))]}/>
          <Select label={t('Статус','Status')} value={status} onChange={setStatus} options={['All','OK','ERROR'].map(x=>({value:x,label:x}))}/>
          <label className="ob-check-filter"><input type="checkbox" checked={slowOnly} onChange={e=>setSlowOnly(e.target.checked)}/>{t('Только длительные ≥1с','Slow traces ≥1s')}</label>
        </div>
      </Panel>
      <Panel title={t('Список трассировок','Trace explorer')} subtitle={t('Открыть waterfall, spans и связанные логи','Open waterfall, span attributes and correlated logs')}>
        <div className="ob-table-scroll"><table className="ob-table">
          <thead><tr><th>{t('Время UTC','Time UTC')}</th><th>{t('Трассировка','Trace')}</th><th>{t('Сервисы','Services')}</th><th>{t('Длительность','Duration')}</th><th>{t('Спанов','Spans')}</th><th>Status</th><th/></tr></thead>
          <tbody>{filtered.map(trace=><tr key={trace.id}>
            <td className="ob-nowrap ob-mono">{trace.timestamp.slice(11,19)}</td>
            <td><Link className="ob-link" to="/observability/traces/$traceId" params={{traceId:trace.id}}>{trace.name}</Link><span className="ob-muted ob-mono">{trace.id}</span></td>
            <td>{new Set(trace.spans.map(s=>s.service)).size} <span className="ob-muted">{trace.rootService}</span></td>
            <td><strong>{trace.durationMs.toLocaleString('en-US')} ms</strong></td>
            <td>{trace.spans.length}</td><td><TraceBadge status={trace.status}/></td>
            <td><Link className="ob-link" to="/observability/traces/$traceId" params={{traceId:trace.id}}>{t('Детали','Details')}</Link></td>
          </tr>)}</tbody>
        </table></div>
        {!filtered.length?<Empty>{t('Трассировки не найдены','No matching traces')}</Empty>:null}
      </Panel>
      <div className="ob-grid ob-grid_2">
        <Panel title={t('Карта связей сервисов','Service dependencies')} subtitle={t('Агрегировано по демонстрационным traces','Derived from synthetic spans')}>
          <div className="ob-dependency-list">
            {[...new Map(demo.traces.flatMap(tr=>tr.spans.filter(s=>s.parentId).map(span=>{
              const parent=tr.spans.find(p=>p.id===span.parentId)
              return [parent?`${parent.service} → ${span.service}`:'unknown',1] as const
            }))).keys()].map(key=><div key={key}><strong>{key}</strong><Label>{t('Наблюдаемая связь','Observed link')}</Label></div>)}
          </div>
        </Panel>
        <Panel title={t('Диагностика','Troubleshooting workflow')}>
          <div className="ob-guidance"><div><strong>Trace ID</strong><span>{t('Связывает запрос через Gateway, Identity, Resource Manager и Temporal','Correlates Gateway, Identity, Resource Manager and Temporal')}</span></div><div><strong>Span ID</strong><span>{t('Позволяет найти узкий участок и его атрибуты','Pinpoints slow or failing spans and attributes')}</span></div><div><strong>Operation ID</strong><span>{t('Связывает асинхронное выполнение с M8 Long-running Operations','Links asynchronous provisioning to M8 Long-running Operations')}</span></div></div>
        </Panel>
      </div>
    </>}
  </Page>
}

export function ObservabilityTraceDetailsPage(){
  const {traceId}=useParams({from:'/observability/traces/$traceId'})
  const demo=useObservabilityDemo()
  const trace=demo.traces.find(tr=>tr.id===traceId)
  const [selectedId,setSelectedId]=useState('')
  const selected=trace?.spans.find(s=>s.id===selectedId)??trace?.spans[0]
  const relatedLogs=demo.logs.filter(l=>l.traceId===traceId)
  return <Page tab="traces" title={trace?.name??'Trace'} subtitle={trace?utc(trace.timestamp):'Trace not found'}>{t=>
    !trace?<Panel title={t('Трассировка не найдена','Trace not found')}><Link to="/observability/traces" className="ob-link">{t('Вернуться к трассировкам','Back to traces')}</Link></Panel>:<>
      <div className="ob-section-head"><Link to="/observability/traces" className="ob-link">‹ {t('Все трассировки','All traces')}</Link><TraceBadge status={trace.status}/></div>
      {validateTrace(trace)?<Notice tone="danger">{validateTrace(trace)}</Notice>:null}
      <div className="ob-stats">
        <Stat label="Duration" value={`${trace.durationMs} ms`}/>
        <Stat label="Spans" value={trace.spans.length}/>
        <Stat label="Services" value={new Set(trace.spans.map(s=>s.service)).size}/>
        <Stat label="HTTP status" value={trace.httpStatus} tone={trace.httpStatus>=500?'danger':undefined}/>
      </div>
      <div className="ob-grid ob-grid_2">
        <Panel title={t('Информация о запросе','Request context')}>
          <div className="ob-details">
            <Detail label="Trace ID" value={<span className="ob-mono">{trace.id}</span>}/>
            <Detail label={t('Корневой сервис','Root service')} value={trace.rootService}/>
            <Detail label={t('Ресурс','Resource')} value={trace.resource??'—'}/>
            <Detail label={t('Время (UTC)','Time (UTC)')} value={utc(trace.timestamp)}/>
          </div>
        </Panel>
        <Panel title={t('Связанная операция','Linked operation')}>
          <div className="ob-details">
            <Detail label="Operation ID" value={trace.operationId?<Link className="ob-link" to="/platform-operations/long-running/$operationId" params={{operationId:trace.operationId}}>{trace.operationId}</Link>:'—'}/>
            <Detail label="Logs" value={relatedLogs.length}/>
            <Detail label={t('Состояние','Status')} value={<TraceBadge status={trace.status}/>}/>
          </div>
          <Notice>{t('Бары отражают относительный start/duration span, а не измеренный живой трафик.','Span waterfall reflects synthetic relative start/duration, not live traces.')}</Notice>
        </Panel>
      </div>
      <Panel title={t('Waterfall диаграмма','Span waterfall')} subtitle={t('Нажмите на span для подробностей','Select a span to see its attributes')}>
        <div className="ob-waterfall">
          <div className="ob-waterfall__header"><span>Span</span><span>{t('Шкала времени','Timeline')} · 0 → {trace.durationMs}ms</span></div>
          {trace.spans.map(span=>{
            const depth=spanDepth(span,trace)
            const start=Math.min(100,Math.max(0,span.startMs/trace.durationMs*100))
            const width=Math.max(.7,span.durationMs/trace.durationMs*100)
            return <button type="button" key={span.id} className={`ob-waterfall__row ${span.id===selected?.id?'ob-waterfall__row_selected':''}`} onClick={()=>setSelectedId(span.id)}>
              <div className="ob-waterfall__label" style={{paddingLeft:`${depth*15}px`}}><strong>{span.name}</strong><span>{span.service}</span></div>
              <div className="ob-waterfall__track"><i className={span.status==='ERROR'?'ob-waterfall__bar_error':''} style={{left:`${start}%`,width:`${Math.min(width,100-start)}%`}}/></div>
              <span className="ob-waterfall__duration">{span.durationMs}ms</span>
            </button>
          })}
        </div>
      </Panel>
      {selected?<Panel title={t('Выбранный span','Selected span')} subtitle={selected.id}>
        <div className="ob-grid ob-grid_2">
          <div className="ob-details"><Detail label="Operation" value={selected.name}/><Detail label="Service" value={selected.service}/><Detail label="Parent span" value={selected.parentId??'ROOT'}/><Detail label="Status" value={<TraceBadge status={selected.status}/>}/></div>
          <div className="ob-details"><Detail label="Start offset" value={`${selected.startMs} ms`}/><Detail label="Duration" value={`${selected.durationMs} ms`}/><Detail label="Span ID" value={<span className="ob-mono">{selected.id}</span>}/></div>
        </div>
        <pre className="ob-code">{JSON.stringify(selected.attributes,null,2)}</pre>
      </Panel>:null}
      <Panel title={t('Логи, связанные с Trace ID','Correlated logs')} subtitle={t('Открыть структурированное событие в Loki-style explorer','Inspect structured log entries with the same trace ID')}>
        {relatedLogs.length?<div className="ob-link-stack">{relatedLogs.map(log=><Link key={log.id} to="/observability/logs/$logId" params={{logId:log.id}} className="ob-link"><Label tone={log.level==='ERROR'?'danger':log.level==='WARN'?'warning':'neutral'}>{log.level}</Label><span>{log.message}</span></Link>)}</div>:<Empty>{t('С этим trace не связаны логи','No demo logs linked to this trace')}</Empty>}
      </Panel>
    </>
  }</Page>
}
