import {useSyncExternalStore} from 'react'

export type TelemetryService = 'api-gateway'|'resource-manager'|'quota-manager'|'identity'|'ai-gateway'|'temporal-worker'
export type TelemetryMetric = 'rps'|'errorRate'|'latencyP95'|'cpu'
export type LogLevel = 'DEBUG'|'INFO'|'WARN'|'ERROR'
export type TraceStatus = 'OK'|'ERROR'
export type AlertSeverity = 'Critical'|'Warning'|'Info'
export type AlertStatus = 'Firing'|'Acknowledged'|'Resolved'|'Muted'
export type AlertRuleMetric = 'errorRate'|'latencyP95'|'cpu'

export interface MetricSeries {
  service:TelemetryService
  metric:TelemetryMetric
  unit:string
  points:number[]
}
export interface TelemetryLog {
  id:string
  timestamp:string
  level:LogLevel
  service:TelemetryService
  message:string
  environment:'production'|'development'
  traceId?:string
  spanId?:string
  operationId?:string
  resource?:string
  attributes:Record<string,string|number|boolean>
}
export interface TraceSpan {
  id:string
  parentId?:string
  name:string
  service:TelemetryService
  startMs:number
  durationMs:number
  status:TraceStatus
  attributes:Record<string,string|number|boolean>
}
export interface TelemetryTrace {
  id:string
  name:string
  rootService:TelemetryService
  timestamp:string
  durationMs:number
  status:TraceStatus
  httpStatus:number
  operationId?:string
  resource?:string
  spans:TraceSpan[]
}
export interface AlertRule {
  id:string
  name:string
  service:TelemetryService
  metric:AlertRuleMetric
  comparator:'gt'|'lt'
  threshold:number
  durationMinutes:number
  severity:AlertSeverity
  enabled:boolean
  description:string
}
export interface AlertIncident {
  id:string
  ruleId:string
  service:TelemetryService
  title:string
  severity:AlertSeverity
  status:AlertStatus
  startedAt:string
  updatedAt:string
  value:number
  summary:string
  operationId?:string
  traceId?:string
  timeline:string[]
}
export interface ServiceSlo {
  id:string
  service:TelemetryService
  name:string
  target:number
  achieved:number
  window:'7d'|'30d'
  good:number
  total:number
  metric:'availability'|'latency'
  description:string
}
export interface ObservabilityState {
  metrics:MetricSeries[]
  logs:TelemetryLog[]
  traces:TelemetryTrace[]
  rules:AlertRule[]
  incidents:AlertIncident[]
  slos:ServiceSlo[]
  activity:string[]
}

export const serviceNames:TelemetryService[]=['api-gateway','resource-manager','quota-manager','identity','ai-gateway','temporal-worker']
export const metricInfo:Record<TelemetryMetric,{name:string;unit:string;description:string}>={
  rps:{name:'Request rate',unit:'req/s',description:'Requests per second'},
  errorRate:{name:'Error rate',unit:'%',description:'Share of failed requests'},
  latencyP95:{name:'Latency p95',unit:'ms',description:'95th percentile latency'},
  cpu:{name:'CPU utilization',unit:'%',description:'CPU usage fraction'},
}
const profiles:Record<TelemetryService,{rps:number;errorRate:number;latencyP95:number;cpu:number}>={
  'api-gateway':{rps:1250,errorRate:.34,latencyP95:95,cpu:52},
  'resource-manager':{rps:340,errorRate:.41,latencyP95:150,cpu:42},
  'quota-manager':{rps:85,errorRate:.08,latencyP95:38,cpu:29},
  identity:{rps:530,errorRate:.62,latencyP95:220,cpu:55},
  'ai-gateway':{rps:185,errorRate:2.8,latencyP95:1240,cpu:78},
  'temporal-worker':{rps:58,errorRate:1.2,latencyP95:580,cpu:62},
}
const pattern=[.83,.88,.91,.87,.96,1.01,.98,1.07,1.04,.95,1.02,1.08,1.1,.99,1.05,.94,.91,1.0,1.04,1.12,1.08,1.06,1.13,1.08]
export const demoMetricTimeLabels=pattern.map((_,i)=>`21:${String(0+i*2).padStart(2,'0')}`)
const metrics:MetricSeries[]=serviceNames.flatMap(service=>(Object.keys(metricInfo) as TelemetryMetric[]).map(metric=>{
  const baseline=profiles[service][metric]
  return {service,metric,unit:metricInfo[metric].unit,points:pattern.map((mult,index)=>{
    const pressure=service==='ai-gateway'&&index>=18&&metric!=='rps'?1.25:1
    const spike=service==='temporal-worker'&&index>=12&&index<=17&&metric==='latencyP95'?1.6:1
    return Number((baseline*mult*pressure*spike).toFixed(metric==='errorRate'?2:1))
  })}
}))

const log=(id:string,timestamp:string,level:LogLevel,service:TelemetryService,message:string,
  extra:Partial<TelemetryLog>={}):TelemetryLog=>({
  id,timestamp,level,service,message,environment:'production',attributes:{cluster:'eu-1',namespace:service,region:'eu-1'},...extra,
})
const logs:TelemetryLog[]=[
  log('log-0022','2026-11-07T21:46:10Z','ERROR','ai-gateway','Provider upstream deadline exceeded',{traceId:'trace-3f29',spanId:'span-3',resource:'route-gpt-5',attributes:{provider:'agent-router',durationMs:1524,status:504}}),
  log('log-0021','2026-11-07T21:45:43Z','WARN','ai-gateway','AI Gateway p95 latency above 1200ms',{traceId:'trace-3f29',spanId:'span-2',resource:'route-gpt-5',attributes:{latencyP95:1510,thresholdMs:1200}}),
  log('log-0020','2026-11-07T21:43:12Z','INFO','api-gateway','Request completed',{traceId:'trace-4a10',spanId:'span-1',attributes:{status:200,method:'GET',path:'/v1/orders'}}),
  log('log-0019','2026-11-07T21:41:18Z','WARN','temporal-worker','Activity retry scheduled for provider reconciliation',{traceId:'trace-f712',operationId:'op-2026-1041',resource:'route-gpt-5',attributes:{attempt:3,backoffSeconds:12}}),
  log('log-0018','2026-11-07T21:38:15Z','INFO','quota-manager','Quota reservation committed',{operationId:'op-2026-1042',resource:'vm-inference-02',attributes:{reservationId:'res-1042',cpu:4,memoryGiB:8}}),
  log('log-0017','2026-11-07T21:36:42Z','ERROR','resource-manager','Provisioning confirmation timed out',{traceId:'trace-b730',operationId:'op-2026-1038',resource:'vm-batch-03',attributes:{errorCode:'PROVIDER_RECONCILIATION_TIMEOUT'}}),
  log('log-0016','2026-11-07T21:32:09Z','INFO','identity','OIDC token verified',{traceId:'trace-4a10',spanId:'span-5',attributes:{issuer:'m8-identity',acr:'aal2',method:'passkey'}}),
  log('log-0015','2026-11-07T21:29:45Z','WARN','identity','Step-up authentication required',{traceId:'trace-81fa',attributes:{acrRequired:'aal2',acrProvided:'aal1'}}),
  log('log-0014','2026-11-07T21:26:30Z','ERROR','api-gateway','Downstream request failed with 503',{traceId:'trace-c001',spanId:'span-1',attributes:{status:503,path:'/v1/identity'}}),
  log('log-0013','2026-11-07T21:23:10Z','INFO','temporal-worker','Workflow completed',{operationId:'op-2026-1039',resource:'route-embeddings',attributes:{workflowId:'ai-route-1039'}}),
  log('log-0012','2026-11-07T21:20:55Z','INFO','resource-manager','Desired resource version accepted',{operationId:'op-2026-1040',resource:'ws-production',attributes:{resourceVersion:12}}),
  log('log-0011','2026-11-07T21:18:33Z','INFO','ai-gateway','Embeddings route returned success',{traceId:'trace-9d21',attributes:{status:200,tokens:3072}}),
  log('log-0010','2026-11-07T21:15:26Z','DEBUG','quota-manager','Admission budget check passed',{operationId:'op-2026-1042',attributes:{availableCpu:12}}),
  log('log-0009','2026-11-07T21:13:12Z','WARN','api-gateway','Rate limit nearing threshold',{traceId:'trace-f912',attributes:{route:'orders-public',usagePercent:88}}),
  log('log-0008','2026-11-07T21:10:43Z','INFO','identity','Access token issued',{traceId:'trace-4a10',attributes:{clientId:'commerce-web',scopes:'orders:read'}}),
  log('log-0007','2026-11-07T21:08:21Z','INFO','ai-gateway','Route registered in desired state',{operationId:'op-2026-1041',resource:'route-gpt-5',attributes:{version:9}}),
  log('log-0006','2026-11-07T21:05:09Z','ERROR','temporal-worker','Activity reached retry limit',{traceId:'trace-f712',operationId:'op-2026-1038',attributes:{attempts:5,activityType:'ConfirmResourceDeletion'}}),
  log('log-0005','2026-11-07T21:02:53Z','INFO','quota-manager','Quota policy updated',{operationId:'op-2026-1032',resource:'AI Studio',attributes:{cpuLimit:40}}),
  log('log-0004','2026-11-07T20:59:14Z','DEBUG','resource-manager','Outbox event delivered',{attributes:{topic:'resource.events',attempt:1}}),
  log('log-0003','2026-11-07T20:56:18Z','INFO','api-gateway','Gateway listener ready',{attributes:{listener:'https',port:443}}),
  log('log-0002','2026-11-07T20:53:07Z','WARN','ai-gateway','Provider quota approaching limit',{traceId:'trace-3f29',attributes:{tokensUsed:8750000,tokenLimit:10000000}}),
  log('log-0001','2026-11-07T20:50:06Z','INFO','temporal-worker','Worker initialized',{attributes:{taskQueue:'m8-provisioning',pollers:4}}),
]
const span=(id:string,name:string,service:TelemetryService,startMs:number,durationMs:number,status:TraceStatus='OK',
  parentId?:string,attributes:Record<string,string|number|boolean>={}):TraceSpan=>({id,name,service,startMs,durationMs,status,parentId,attributes})
export const initialTraces:TelemetryTrace[]=[
  {id:'trace-3f29',name:'POST /v1/chat/completions',rootService:'api-gateway',timestamp:'2026-11-07T21:46:09Z',durationMs:1544,status:'ERROR',httpStatus:504,operationId:'op-2026-1041',resource:'route-gpt-5',spans:[
    span('span-1','POST /v1/chat/completions','api-gateway',0,1544,'ERROR',undefined,{httpStatus:504}),
    span('span-2','route-ai-request','ai-gateway',11,1527,'ERROR','span-1',{model:'gpt-demo'}),
    span('span-3','agent-router.proxy','ai-gateway',33,1491,'ERROR','span-2',{status:504,upstream:'model-provider'}),
    span('span-4','authorize-consumer','identity',12,12,'OK','span-1',{client:'ai-studio-console'}),
  ]},
  {id:'trace-4a10',name:'GET /v1/orders',rootService:'api-gateway',timestamp:'2026-11-07T21:43:11Z',durationMs:112,status:'OK',httpStatus:200,resource:'orders-public',spans:[
    span('span-1','GET /v1/orders','api-gateway',0,112),
    span('span-5','jwt.verify','identity',4,17,'OK','span-1'),
    span('span-2','orders.list','resource-manager',23,78,'OK','span-1'),
    span('span-3','db.query.orders','resource-manager',32,46,'OK','span-2'),
  ]},
  {id:'trace-f712',name:'Reconcile AIGatewayRoute',rootService:'temporal-worker',timestamp:'2026-11-07T21:41:17Z',durationMs:1204,status:'ERROR',httpStatus:500,operationId:'op-2026-1041',resource:'route-gpt-5',spans:[
    span('span-1','workflow.run','temporal-worker',0,1204,'ERROR'),
    span('span-2','provider.apply','resource-manager',36,671,'OK','span-1'),
    span('span-3','provider.observe','temporal-worker',720,480,'ERROR','span-1'),
  ]},
  {id:'trace-b730',name:'Delete VirtualMachine',rootService:'resource-manager',timestamp:'2026-11-07T21:36:41Z',durationMs:2008,status:'ERROR',httpStatus:500,operationId:'op-2026-1038',resource:'vm-batch-03',spans:[
    span('span-1','resource.delete','resource-manager',0,2008,'ERROR'),
    span('span-2','quota.read','quota-manager',30,23,'OK','span-1'),
    span('span-3','provider.delete','resource-manager',68,1910,'ERROR','span-1'),
  ]},
  {id:'trace-c001',name:'GET /v1/identity',rootService:'api-gateway',timestamp:'2026-11-07T21:26:29Z',durationMs:480,status:'ERROR',httpStatus:503,resource:'identity-public',spans:[
    span('span-1','GET /v1/identity','api-gateway',0,480,'ERROR'),
    span('span-2','identity.lookup','identity',14,456,'ERROR','span-1'),
  ]},
  {id:'trace-9d21',name:'POST /v1/embeddings',rootService:'ai-gateway',timestamp:'2026-11-07T21:18:31Z',durationMs:320,status:'OK',httpStatus:200,resource:'route-embeddings',spans:[
    span('span-1','POST /v1/embeddings','ai-gateway',0,320),
    span('span-2','provider.predict','ai-gateway',30,273,'OK','span-1'),
  ]},
  {id:'trace-81fa',name:'POST /auth/step-up',rootService:'identity',timestamp:'2026-11-07T21:29:43Z',durationMs:91,status:'OK',httpStatus:200,spans:[
    span('span-1','check.amr.acr','identity',0,91),
    span('span-2','evaluate-risk','identity',10,63,'OK','span-1'),
  ]},
  {id:'trace-f912',name:'GET /v1/orders (rate check)',rootService:'api-gateway',timestamp:'2026-11-07T21:13:11Z',durationMs:73,status:'OK',httpStatus:200,spans:[
    span('span-1','GET /v1/orders','api-gateway',0,73),
    span('span-2','rate-limit.check','api-gateway',5,9,'OK','span-1'),
    span('span-3','service.proxy','api-gateway',15,54,'OK','span-1'),
  ]},
]
const initialRules:AlertRule[]=[
  {id:'rule-ai-latency',name:'AI gateway high p95 latency',service:'ai-gateway',metric:'latencyP95',comparator:'gt',threshold:1200,durationMinutes:5,severity:'Critical',enabled:true,description:'p95 response latency over 1200 ms'},
  {id:'rule-ai-errors',name:'AI Gateway error budget burn',service:'ai-gateway',metric:'errorRate',comparator:'gt',threshold:2,durationMinutes:10,severity:'Warning',enabled:true,description:'Errors above 2% of requests'},
  {id:'rule-temporal-errors',name:'Provisioning activity failures',service:'temporal-worker',metric:'errorRate',comparator:'gt',threshold:1,durationMinutes:10,severity:'Critical',enabled:true,description:'Worker failure rate above 1%'},
  {id:'rule-identity-cpu',name:'Identity CPU saturation',service:'identity',metric:'cpu',comparator:'gt',threshold:85,durationMinutes:5,severity:'Warning',enabled:true,description:'CPU sustained above 85%'},
]
const initialIncidents:AlertIncident[]=[
  {id:'inc-301',ruleId:'rule-ai-latency',service:'ai-gateway',title:'AI Gateway latency above SLO',severity:'Critical',status:'Firing',startedAt:'2026-11-07T21:44:00Z',updatedAt:'2026-11-07T21:46:00Z',value:1510,summary:'Upstream model-provider requests timing out',traceId:'trace-3f29',operationId:'op-2026-1041',timeline:['Firing: latency threshold exceeded','Correlated trace trace-3f29']},
  {id:'inc-300',ruleId:'rule-temporal-errors',service:'temporal-worker',title:'Provisioning retry exhaustion',severity:'Critical',status:'Firing',startedAt:'2026-11-07T21:34:00Z',updatedAt:'2026-11-07T21:42:00Z',value:1.2,summary:'Provider reconciliation confirmations are timing out',traceId:'trace-f712',operationId:'op-2026-1038',timeline:['Firing: workflow activity retry budget exceeded','Error correlated with op-2026-1038']},
  {id:'inc-299',ruleId:'rule-ai-errors',service:'ai-gateway',title:'AI Gateway elevated 5xx rate',severity:'Warning',status:'Acknowledged',startedAt:'2026-11-07T20:59:00Z',updatedAt:'2026-11-07T21:20:00Z',value:2.8,summary:'Elevated provider 5xx rate, investigation in progress',traceId:'trace-3f29',timeline:['Firing: error rate > 2%','Acknowledged by on-call operator']},
  {id:'inc-298',ruleId:'rule-identity-cpu',service:'identity',title:'Identity CPU saturation recovered',severity:'Warning',status:'Resolved',startedAt:'2026-11-07T17:02:00Z',updatedAt:'2026-11-07T18:18:00Z',value:58,summary:'CPU returned below 85% after scale-out',timeline:['Firing: CPU > 85%','Resolved after workload redistribution']},
]
export const initialSlos:ServiceSlo[]=[
  {id:'slo-gateway-availability',service:'api-gateway',name:'Gateway availability',target:99.95,achieved:99.97,window:'30d',good:999700,total:1000000,metric:'availability',description:'Requests with non-5xx responses'},
  {id:'slo-resource-availability',service:'resource-manager',name:'Resource Manager availability',target:99.9,achieved:99.94,window:'30d',good:999400,total:1000000,metric:'availability',description:'Successful control-plane API responses'},
  {id:'slo-ai-availability',service:'ai-gateway',name:'AI Gateway availability',target:99.9,achieved:99.72,window:'30d',good:997200,total:1000000,metric:'availability',description:'Inference requests without server errors'},
  {id:'slo-ai-latency',service:'ai-gateway',name:'AI Gateway latency under 1.5s',target:99,achieved:98.65,window:'7d',good:986500,total:1000000,metric:'latency',description:'Inference responses completed in < 1500 ms'},
  {id:'slo-quota-availability',service:'quota-manager',name:'Quota admission availability',target:99.95,achieved:99.99,window:'30d',good:999900,total:1000000,metric:'availability',description:'Admission operations completed without 5xx'},
  {id:'slo-identity-availability',service:'identity',name:'Identity API availability',target:99.9,achieved:99.91,window:'30d',good:999100,total:1000000,metric:'availability',description:'Identity verification success excluding user errors'},
]
export const initialObservability:ObservabilityState={
  metrics,logs,traces:initialTraces,rules:initialRules,incidents:initialIncidents,slos:initialSlos,activity:[],
}
let state=initialObservability
const listeners=new Set<()=>void>()
const subscribe=(listener:()=>void)=>{listeners.add(listener);return()=>{listeners.delete(listener)}}
function publish(next:ObservabilityState){state=next;listeners.forEach(fn=>fn())}
export function useObservabilityDemo(){return useSyncExternalStore(subscribe,()=>state,()=>initialObservability)}
export function resetObservabilityDemo(){publish(initialObservability)}
export function metricSeries(service:TelemetryService,metric:TelemetryMetric,points:MetricSeries[]=state.metrics){
  return points.find(s=>s.service===service&&s.metric===metric)?.points??[]
}
export function currentMetric(service:TelemetryService,metric:TelemetryMetric,points:MetricSeries[]=state.metrics):number{
  const series=metricSeries(service,metric,points)
  return series.length?series[series.length-1]:0
}
export function formatMetric(value:number,metric:TelemetryMetric):string{
  const num=metric==='errorRate'?value.toFixed(2):metric==='rps'?Math.round(value).toLocaleString('en-US'):value.toFixed(1)
  return `${num} ${metricInfo[metric].unit}`
}
export function searchLogs({service,level,query}:{service?:string;level?:string;query?:string},source:TelemetryLog[]=state.logs):TelemetryLog[]{
  const q=(query??'').toLowerCase().trim()
  return source.filter(log=>
    (!service||service==='All'||log.service===service)&&(!level||level==='All'||log.level===level)&&
    (!q||[log.message,log.service,log.id,log.traceId??'',log.operationId??'',log.resource??'',JSON.stringify(log.attributes)].some(x=>x.toLowerCase().includes(q)))
  ).sort((a,b)=>b.timestamp.localeCompare(a.timestamp))
}
export function searchTraces({service,status,query}:{service?:string;status?:string;query?:string},source:TelemetryTrace[]=state.traces):TelemetryTrace[]{
  const q=(query??'').toLowerCase().trim()
  return source.filter(t=>(!service||service==='All'||t.spans.some(s=>s.service===service))&&
    (!status||status==='All'||t.status===status)&&
    (!q||[t.id,t.name,t.rootService,t.operationId??'',t.resource??'',...t.spans.map(s=>s.name)].some(x=>x.toLowerCase().includes(q)))
  ).sort((a,b)=>b.timestamp.localeCompare(a.timestamp))
}
export function validateTrace(trace:TelemetryTrace):string|null{
  if(!trace.spans.length)return 'Trace must contain a root span'
  const ids=new Set(trace.spans.map(s=>s.id))
  if(ids.size!==trace.spans.length)return 'Duplicate span identifiers'
  if(trace.spans.filter(s=>!s.parentId).length!==1)return 'Trace requires exactly one root span'
  for(const s of trace.spans){
    if(s.parentId&&!ids.has(s.parentId))return `Missing parent for ${s.id}`
    if(s.durationMs<0||s.startMs<0||s.startMs+s.durationMs>trace.durationMs)return `Invalid span bounds for ${s.id}`
    let parent=s.parentId
    const visited=new Set([s.id])
    while(parent){
      if(visited.has(parent))return 'Span parent cycle detected'
      visited.add(parent)
      parent=trace.spans.find(item=>item.id===parent)?.parentId
    }
  }
  return null
}
export function validateRule(rule:Pick<AlertRule,'name'|'service'|'metric'|'comparator'|'threshold'|'durationMinutes'>):string|null{
  if(rule.name.trim().length<4)return 'Rule name must have at least 4 characters'
  if(!serviceNames.includes(rule.service))return 'Invalid service'
  if(!Number.isFinite(rule.threshold)||rule.threshold<0)return 'Threshold must be a nonnegative number'
  if(['errorRate','cpu'].includes(rule.metric)&&rule.threshold>100)return 'Threshold cannot exceed 100%'
  if(!Number.isSafeInteger(rule.durationMinutes)||rule.durationMinutes<1||rule.durationMinutes>1440)return 'Duration must be 1–1440 minutes'
  return null
}
export function createDemoRule(input:Omit<AlertRule,'id'|'enabled'>):string|null{
  const error=validateRule(input)
  if(error)return error
  const nextId=`rule-demo-${state.rules.filter(r=>r.id.startsWith('rule-demo-')).length+1}`
  publish({...state,rules:[{...input,id:nextId,enabled:true},...state.rules],activity:[`Rule ${input.name} created (demo)`,...state.activity]})
  return null
}
export function toggleDemoRule(id:string):boolean{
  if(!state.rules.some(r=>r.id===id))return false
  publish({...state,rules:state.rules.map(r=>r.id===id?{...r,enabled:!r.enabled}:r),activity:[`Rule ${id} toggled`,...state.activity]})
  return true
}
export function changeDemoIncident(id:string,action:'acknowledge'|'resolve'|'mute'):boolean{
  const incident=state.incidents.find(x=>x.id===id)
  if(!incident)return false
  if(incident.status==='Resolved'||incident.status==='Muted'&&action==='acknowledge')return false
  if(action==='acknowledge'&&incident.status!=='Firing')return false
  if(action==='mute'&&incident.status!=='Firing'&&incident.status!=='Acknowledged')return false
  const status:AlertStatus=action==='acknowledge'?'Acknowledged':action==='resolve'?'Resolved':'Muted'
  publish({...state,incidents:state.incidents.map(x=>x.id===id?{...x,status,updatedAt:'2026-11-07T21:48:00Z',timeline:[...x.timeline,`${status} by Console demo operator`]}:x),activity:[`Incident ${id} marked ${status}`,...state.activity]})
  return true
}
export function validateSlo(slo:ServiceSlo):string|null{
  if(slo.target<=0||slo.target>100||!Number.isFinite(slo.target))return 'Invalid SLO target'
  if(slo.total<=0||slo.good<0||slo.good>slo.total)return 'Invalid good/total counts'
  if(Math.abs(slo.achieved-slo.good/slo.total*100)>.01)return 'Achieved SLI inconsistent with good/total counts'
  return null
}
export function sloBudget(slo:ServiceSlo){
  const allowedBad=Math.floor(slo.total*(1-slo.target/100)+1e-8)
  const bad=slo.total-slo.good
  return {allowedBad,bad,remaining:allowedBad-bad,usedPercent:allowedBad>0?bad/allowedBad*100:bad?100:0}
}
