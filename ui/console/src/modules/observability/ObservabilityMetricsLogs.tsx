import {useMemo,useState} from 'react'
import {ArrowRotateRight,Check,Clock,TriangleExclamation} from '@gravity-ui/icons'
import {Button,Icon,Text} from '@gravity-ui/uikit'
import {Link,useParams} from '@tanstack/react-router'

import {
  currentMetric,formatMetric,metricInfo,metricSeries,resetObservabilityDemo,searchLogs,serviceNames,useObservabilityDemo,
} from './demo'
import type {LogLevel,TelemetryMetric,TelemetryService} from './demo'
import {
  Bar,Detail,Empty,Input,Label,LogBadge,Notice,Page,Panel,Select,Sparkline,Stat,timeLabel,utc,
} from './ObservabilityShell'

const metricTypes:TelemetryMetric[]=['rps','errorRate','latencyP95','cpu']
const metricLabels:Record<TelemetryMetric,string>={rps:'RPS',errorRate:'Error rate',latencyP95:'p95 latency',cpu:'CPU'}
const metricColor=(metric:TelemetryMetric,value:number):'default'|'danger'=>metric==='errorRate'&&value>=2||metric==='latencyP95'&&value>=1200?'danger':'default'

export function ObservabilityMetricsPage(){
  const demo=useObservabilityDemo()
  const [service,setService]=useState<TelemetryService>('api-gateway')
  const [selectedMetric,setSelectedMetric]=useState<TelemetryMetric>('rps')
  const [range,setRange]=useState('60m')
  const [environment,setEnvironment]=useState('production')
  const all=serviceNames.map(s=>({
    service:s,
    rps:currentMetric(s,'rps',demo.metrics),
    errorRate:currentMetric(s,'errorRate',demo.metrics),
    latencyP95:currentMetric(s,'latencyP95',demo.metrics),
    cpu:currentMetric(s,'cpu',demo.metrics),
  }))
  const selected=all.find(s=>s.service===service)!
  const points=range==='15m'?8:range==='30m'?16:24
  return <Page tab="metrics" title="Metrics" subtitle="M8 Platform service health, request volume, latency and resource saturation">
    {t=><>
      <div className="ob-section-head"><div><Text as="h2" variant="header-2">{t('Обзор метрик','Metrics overview')}</Text><Text variant="body-2" color="secondary">{t('Синтетическая телеметрия, ориентированная на Mimir/Prometheus','Synthetic Prometheus/Mimir-style telemetry for M8 services')}</Text></div>
        <Button view="outlined" onClick={resetObservabilityDemo}><Icon data={ArrowRotateRight} size={15}/>{t('Сброс демоданных','Reset demo')}</Button>
      </div>
      <div className="ob-filters ob-metrics-filters">
        <Select label={t('Сервис','Service')} value={service} onChange={v=>setService(v as TelemetryService)} options={serviceNames.map(s=>({value:s,label:s}))}/>
        <Select label={t('Период','Time range')} value={range} onChange={setRange} options={[{value:'15m',label:'Last 15m'},{value:'30m',label:'Last 30m'},{value:'60m',label:'Last 60m'}]}/>
        <Select label={t('Окружение','Environment')} value={environment} onChange={setEnvironment} options={[{value:'production',label:'Production / eu-1'}]}/>
      </div>
      <div className="ob-stats">
        <Stat label="Request rate" value={formatMetric(selected.rps,'rps')} note={t('Запросов в секунду','Requests per second')}/>
        <Stat label="Error rate" value={formatMetric(selected.errorRate,'errorRate')} note="5xx / total" tone={selected.errorRate>=2?'danger':undefined}/>
        <Stat label="Latency p95" value={formatMetric(selected.latencyP95,'latencyP95')} note={t('95-й перцентиль','95th percentile')} tone={selected.latencyP95>=1200?'warning':undefined}/>
        <Stat label="CPU utilization" value={formatMetric(selected.cpu,'cpu')} note={t('Использовано CPU','CPU use')}/>
      </div>
      <div className="ob-grid ob-grid_2">
        <Panel title={t('Метрики сервиса','Service performance')} subtitle={t('Выберите метрику для детализации временного ряда','Select a metric for a detailed time series')} actions={<Select label={t('Метрика','Metric')} value={selectedMetric} onChange={v=>setSelectedMetric(v as TelemetryMetric)} options={metricTypes.map(m=>({value:m,label:metricLabels[m]}))}/>}>
          <div className="ob-chart-main">
            <div><Text variant="caption-2" color="secondary">{metricInfo[selectedMetric].description}</Text><strong>{formatMetric(currentMetric(service,selectedMetric,demo.metrics),selectedMetric)}</strong></div>
            <Sparkline values={metricSeries(service,selectedMetric,demo.metrics).slice(-points)} height={190} variant={metricColor(selectedMetric,currentMetric(service,selectedMetric,demo.metrics))}/>
            <div className="ob-chart-axis"><span>{range==='15m'?'21:32':range==='30m'?'21:16':'21:00'}</span><span>21:46 UTC</span></div>
          </div>
        </Panel>
        <Panel title={t('Четыре ключевых сигнала','Four golden signals')} subtitle={t('Согласованный с выбранным сервисом период','Matching the selected service and time range')}>
          <div className="ob-signal-list">{metricTypes.map(metric=><button type="button" key={metric} className={`ob-signal ${selectedMetric===metric?'ob-signal_active':''}`} onClick={()=>setSelectedMetric(metric)}>
            <div><strong>{metricLabels[metric]}</strong><span>{formatMetric(currentMetric(service,metric,demo.metrics),metric)}</span></div>
            <Sparkline values={metricSeries(service,metric,demo.metrics).slice(-points)} height={50} variant={metricColor(metric,currentMetric(service,metric,demo.metrics))}/>
          </button>)}</div>
        </Panel>
      </div>
      <Panel title={t('Здоровье сервисов','Service health overview')} subtitle={t('Общее состояние по основным сигналам. Цвета используются только как подсказка.','Service-wide metrics; status coloring is illustrative only.')}>
        <div className="ob-table-scroll"><table className="ob-table"><thead><tr><th>Service</th><th>RPS</th><th>Error rate</th><th>Latency p95</th><th>CPU</th><th>{t('Состояние','Health')}</th><th/></tr></thead>
          <tbody>{all.map(s=>{
            const degraded=s.errorRate>=2||s.latencyP95>=1200
            return <tr key={s.service}><td><strong>{s.service}</strong></td>
              <td>{formatMetric(s.rps,'rps')}</td><td>{formatMetric(s.errorRate,'errorRate')}</td><td>{formatMetric(s.latencyP95,'latencyP95')}</td><td><div className="ob-table-meter">{s.cpu.toFixed(1)}%<Bar value={s.cpu} max={100}/></div></td>
              <td><Label tone={degraded?'warning':'positive'}>{degraded?'Degraded':'Healthy'}</Label></td>
              <td><Button size="s" view="outlined" onClick={()=>{setService(s.service);setSelectedMetric('rps')}}>{t('Просмотр','Inspect')}</Button></td>
            </tr>
          })}</tbody>
        </table></div>
      </Panel>
      <div className="ob-grid ob-grid_2">
        <Panel title={t('Связанные сигналы','Related signals')} subtitle={t('Логи, трассировки и алерты коррелируются через trace/operation IDs','Logs, traces and alerts use trace and operation correlation IDs')}>
          <div className="ob-link-stack">
            <Link to="/observability/logs" className="ob-link"><Icon data={Clock} size={16}/>{t('Открыть логи','Explore logs')}</Link>
            <Link to="/observability/traces" className="ob-link"><Icon data={Check} size={16}/>{t('Открыть трассировки','Explore traces')}</Link>
            <Link to="/observability/alerts" className="ob-link"><Icon data={TriangleExclamation} size={16}/>{t('Открыть алерты','Open alerts')}</Link>
          </div>
        </Panel>
        <Panel title={t('План подключения','Production sources')}>
          <div className="ob-guidance">
            <div><strong>Prometheus / Mimir</strong><span>{t('Запросы PromQL через серверный шлюз с tenant scoping','PromQL queries through a tenant-aware backend gateway')}</span></div>
            <div><strong>OpenTelemetry Collector</strong><span>{t('Нормализация метрик и labels, лимиты кардинальности','Normalize metrics/labels and guard cardinality')}</span></div>
            <div><strong>Resource Manager</strong><span>{t('Маппинг service/resource labels на M8 Resource IDs','Map telemetry service/resource labels to M8 resource IDs')}</span></div>
          </div>
        </Panel>
      </div>
    </>}
  </Page>
}

const logLevels:LogLevel[]=['DEBUG','INFO','WARN','ERROR']
export function ObservabilityLogsPage(){
  const demo=useObservabilityDemo()
  const [service,setService]=useState('All')
  const [level,setLevel]=useState('All')
  const [query,setQuery]=useState('')
  const [limit,setLimit]=useState('20')
  const rows=useMemo(()=>searchLogs({service,level,query},demo.logs),[demo.logs,service,level,query])
  const visible=rows.slice(0,Number(limit))
  return <Page tab="logs" title="Logs" subtitle="Loki-style structured logs and correlation with resources, traces and operations">
    {t=><>
      <div className="ob-section-head"><div><Text as="h2" variant="header-2">{t('Поток логов','Log Explorer')}</Text><Text variant="body-2" color="secondary">{t('Структурированные события с метками сервиса, severity и trace ID','Structured logs labeled by service, severity and trace ID')}</Text></div><Label tone="info">{rows.length} {t('событий','entries')}</Label></div>
      <Panel title={t('Поиск логов','Log query')} subtitle={t('Запрос выполняется локально, LogQL не вызывается','In-memory filtering; LogQL is not executed')}>
        <div className="ob-filters">
          <Input label={t('Поиск: текст, trace, operation','Search: text, trace, operation')} value={query} onChange={setQuery} placeholder="trace-3f29 / op-2026-1041"/>
          <Select label="Service" value={service} onChange={setService} options={[{value:'All',label:t('Все сервисы','All services')},...serviceNames.map(s=>({value:s,label:s}))]}/>
          <Select label={t('Уровень','Level')} value={level} onChange={setLevel} options={[{value:'All',label:t('Все','All')},...logLevels.map(v=>({value:v,label:v}))]}/>
          <Select label={t('Макс. событий','Max entries')} value={limit} onChange={setLimit} options={['10','20','50'].map(v=>({value:v,label:v}))}/>
        </div>
      </Panel>
      <div className="ob-stats">
        <Stat label="Total matching" value={rows.length}/>
        <Stat label="ERROR" value={rows.filter(l=>l.level==='ERROR').length} tone="danger"/>
        <Stat label="WARN" value={rows.filter(l=>l.level==='WARN').length} tone="warning"/>
        <Stat label={t('С trace ID','With trace ID')} value={rows.filter(l=>!!l.traceId).length}/>
      </div>
      <Panel title={t('Результаты','Log entries')} subtitle={t('Синтетические логи · 7 ноября 2026 · UTC','Synthetic log records · 7 Nov 2026 · UTC')}>
        <div className="ob-table-scroll"><table className="ob-table ob-table_logs"><thead><tr><th>UTC</th><th>Level</th><th>Service</th><th>{t('Сообщение','Message')}</th><th>Trace ID</th><th/></tr></thead><tbody>{visible.map(l=><tr key={l.id}>
          <td className="ob-nowrap ob-mono">{l.timestamp.slice(11,19)}</td><td><LogBadge level={l.level}/></td><td>{l.service}</td>
          <td><strong>{l.message}</strong><span className="ob-muted">{l.resource??l.operationId??l.id}</span></td>
          <td>{l.traceId?<Link className="ob-link" to="/observability/traces/$traceId" params={{traceId:l.traceId}}>{l.traceId}</Link>:<span className="ob-muted">—</span>}</td>
          <td><Link className="ob-link" to="/observability/logs/$logId" params={{logId:l.id}}>{t('Детали','Details')}</Link></td>
        </tr>)}</tbody></table></div>
        {!visible.length?<Empty>{t('Логи не найдены','No matching logs')}</Empty>:null}
      </Panel>
      <Notice>{t('Для production нужны M8 Access/RBAC, сквозная корреляция и ограничение хранения. Возможность просматривать сырой payload требует отдельного permission.','Production log access requires M8 Access/RBAC, trace correlation and retention controls. Raw payload access requires a separate permission.')}</Notice>
    </>}
  </Page>
}

export function ObservabilityLogDetailsPage(){
  const {logId}=useParams({from:'/observability/logs/$logId'})
  const demo=useObservabilityDemo()
  const l=demo.logs.find(log=>log.id===logId)
  return <Page tab="logs" title={l?.message??'Log Entry'} subtitle={l?timeLabel(l.timestamp):'Log record not found'}>{t=>
    !l?<Panel title={t('Лог не найден','Log not found')}><Link to="/observability/logs" className="ob-link">{t('Вернуться к логам','Back to logs')}</Link></Panel>:<>
      <div className="ob-section-head"><Link className="ob-link" to="/observability/logs">‹ {t('Все логи','All logs')}</Link><LogBadge level={l.level}/></div>
      <div className="ob-grid ob-grid_2">
        <Panel title={t('Запись лога','Log entry')}>
          <div className="ob-details">
            <Detail label="Log ID" value={<span className="ob-mono">{l.id}</span>}/>
            <Detail label={t('Время UTC','Timestamp UTC')} value={utc(l.timestamp)}/>
            <Detail label="Service" value={l.service}/>
            <Detail label="Level" value={<LogBadge level={l.level}/>}/>
            <Detail label="Environment" value={l.environment}/>
            <Detail label="Message" value={l.message}/>
            <Detail label="Resource" value={l.resource??'—'}/>
          </div>
        </Panel>
        <Panel title={t('Корреляция','Correlation')}>
          <div className="ob-details">
            <Detail label="Trace ID" value={l.traceId?<Link className="ob-link" to="/observability/traces/$traceId" params={{traceId:l.traceId}}>{l.traceId}</Link>:'—'}/>
            <Detail label="Span ID" value={l.spanId??'—'}/>
            <Detail label="Operation ID" value={l.operationId?<Link className="ob-link" to="/platform-operations/long-running/$operationId" params={{operationId:l.operationId}}>{l.operationId}</Link>:'—'}/>
          </div>
          <Notice>{t('Корреляция с ресурсами использует демонстрационные идентификаторы.','Correlation links refer to demo resource and operation identifiers.')}</Notice>
        </Panel>
      </div>
      <Panel title={t('Структурированные атрибуты','Structured attributes')}>
        <pre className="ob-code">{JSON.stringify(l.attributes,null,2)}</pre>
      </Panel>
    </>
  }</Page>
}
