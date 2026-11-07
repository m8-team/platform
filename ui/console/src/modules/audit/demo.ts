import {useSyncExternalStore} from 'react'

export type AuditCategory = 'Resource' | 'Gateway' | 'Quota' | 'Identity' | 'Security' | 'System'
export type AuditOutcome = 'Success' | 'Failure' | 'Denied'
export type AuditSeverity = 'Info' | 'Warning' | 'Critical'
export type ActorType = 'User' | 'ServiceAccount' | 'System'
export type AuditScope = 'Organization' | 'Project'
export type ExportFormat = 'CSV' | 'JSONL'
export type ExportStatus = 'Queued' | 'Ready'

export interface AuditActor {
  id:string
  displayName:string
  type:ActorType
}
export interface AuditEvent {
  id:string
  timestamp:string
  category:AuditCategory
  action:string
  outcome:AuditOutcome
  severity:AuditSeverity
  actor:AuditActor
  scope:AuditScope
  scopeName:string
  resourceType:string
  resourceId:string
  resourceName:string
  source:string
  description:string
  requestId:string
  correlationId:string
  traceId?:string
  operationId?:string
  ipAddress?:string
  changes?:{before?:Record<string,unknown>;after?:Record<string,unknown>}
  metadata?:Record<string,unknown>
}
export interface AuditFilters {
  query?:string
  category?:string
  outcome?:string
  severity?:string
  scope?:string
  actorType?:string
  fromDate?:string
  toDate?:string
}
export interface AuditExport {
  id:string
  createdAt:string
  requestedBy:string
  status:ExportStatus
  format:ExportFormat
  filters:AuditFilters
  eventIds:string[]
  rowCount:number
  redaction:'Safe'
  filename:string
}
interface AuditState {
  events:AuditEvent[]
  exports:AuditExport[]
  draft:AuditFilters
}

const admin:AuditActor={id:'usr-platform-admin',displayName:'Platform Admin',type:'User'}
const platform:AuditActor={id:'sa-resource-manager',displayName:'resource-manager',type:'ServiceAccount'}
const quota:AuditActor={id:'sa-quota-manager',displayName:'quota-manager',type:'ServiceAccount'}
const gateway:AuditActor={id:'sa-gateway',displayName:'gateway-controller',type:'ServiceAccount'}
const security:AuditActor={id:'sa-risk-engine',displayName:'risk-decision',type:'ServiceAccount'}
const ml:AuditActor={id:'usr-ml-owner',displayName:'ML Platform Owner',type:'User'}
const system:AuditActor={id:'system-audit',displayName:'M8 Audit',type:'System'}
function event(
  id:string,timestamp:string,category:AuditCategory,action:string,outcome:AuditOutcome,severity:AuditSeverity,
  actor:AuditActor,scope:AuditScope,resourceType:string,resourceId:string,resourceName:string,source:string,
  description:string,extra:Partial<AuditEvent>={},
):AuditEvent{
  return {
    id,timestamp,category,action,outcome,severity,actor,scope,
    scopeName:scope==='Project'?'Acme / Production / AI Studio':'Acme',
    resourceType,resourceId,resourceName,source,description,
    requestId:`req-${id.slice(-4)}`,correlationId:`corr-${id.slice(-4)}`,
    ...extra,
  }
}
export const initialAuditEvents:AuditEvent[]=[
  event('evt-20261107-018','2026-11-07T21:11:44Z','Gateway','gateway.route.created','Success','Info',ml,'Project','AIGatewayRoute','route-gpt-5','route-gpt-5','Gateway API','Route desired state was accepted', {
    operationId:'op-2026-1041',traceId:'tr-9801',ipAddress:'198.51.100.21',
    changes:{before:{},after:{hostname:'ai.acme.example',path:'/v1/chat/completions',backend:'svc-agent-router',apiKey:'[REDACTED]'}},
  }),
  event('evt-20261107-017','2026-11-07T21:10:33Z','Quota','quota.reservation.created','Success','Info',quota,'Project','QuotaReservation','res-1042','res-1042','Quota Manager','Four vCPUs and eight GiB reserved for provisioning',{
    operationId:'op-2026-1042',metadata:{cpu:4,memoryGiB:8,allocation:'dedicated'},
  }),
  event('evt-20261107-016','2026-11-07T21:07:51Z','Gateway','gateway.route.reconciliation.delayed','Failure','Warning',gateway,'Project','AIGatewayRoute','route-gpt-5','route-gpt-5','Agent Router','Provider has not confirmed Ready condition',{
    operationId:'op-2026-1041',metadata:{condition:'Ready=False',errorCode:'PROVIDER_TIMEOUT'},
  }),
  event('evt-20261107-015','2026-11-07T21:04:21Z','Security','access.request.denied','Denied','Critical',security,'Organization','AccessRequest','access-9102','production-ssh','M8 Access','Step-up authentication requirement was not met',{
    traceId:'tr-denied-12',ipAddress:'203.0.113.42',metadata:{reason:'ACR_TOO_LOW',requiredAcr:'aal2',providedAcr:'aal1'},
  }),
  event('evt-20261107-014','2026-11-07T20:59:19Z','Identity','user.session.created','Success','Info',admin,'Organization','Session','session-7310','console-login','M8 Identity','Console session created using passkey',{
    ipAddress:'198.51.100.21',metadata:{amr:['webauthn'],credential:'[REDACTED]'},
  }),
  event('evt-20261107-013','2026-11-07T20:45:09Z','Quota','quota.policy.updated','Success','Warning',admin,'Project','QuotaPolicy','compute.vcpu.allocated','CPU limit','Quota Manager','Project CPU limit increased after approval',{
    operationId:'op-2026-1032',changes:{before:{limit:32,mode:'Dedicated'},after:{limit:40,mode:'Dedicated'}},metadata:{quotaRequestId:'QR-108'},
  }),
  event('evt-20261107-012','2026-11-07T20:29:27Z','Gateway','gateway.policy.updated','Success','Warning',admin,'Organization','GatewayPolicy','policy-waf','Coraza WAF','Gateway Policy Studio','WAF policy switched to monitoring mode',{
    changes:{before:{mode:'Enforce'},after:{mode:'Monitor'}},metadata:{reviewId:'sec-482'},
  }),
  event('evt-20261107-011','2026-11-07T20:19:46Z','Resource','resource.delete.failed','Failure','Critical',platform,'Project','VirtualMachine','vm-batch-03','vm-batch-03','Resource Manager','Deletion was not confirmed by the provider',{
    operationId:'op-2026-1038',metadata:{errorCode:'PROVIDER_RECONCILIATION_TIMEOUT',resourceVersion:'1294'},
  }),
  event('evt-20261107-010','2026-11-07T19:57:12Z','Resource','workspace.suspend.requested','Success','Info',admin,'Organization','Workspace','ws-development','Development','Resource Manager','Drain of workspace resources requested',{
    operationId:'op-2026-1035',metadata:{reason:'Scheduled maintenance'},
  }),
  event('evt-20261107-009','2026-11-07T19:33:04Z','Security','rate_limit.enforced','Denied','Warning',gateway,'Project','AIGatewayRoute','route-chat','openai-chat','Envoy Gateway','Excess traffic rejected by rate limiter',{
    ipAddress:'203.0.113.73',metadata:{rule:'limit-chat',limitPerMinute:1200},
  }),
  event('evt-20261107-008','2026-11-07T18:51:43Z','Gateway','gateway.consumer.suspended','Success','Warning',admin,'Organization','Consumer','consumer-partner','partner-sandbox','Gateway','API consumer access suspended',{
    changes:{before:{status:'Active'},after:{status:'Suspended'}},
  }),
  event('evt-20261107-007','2026-11-07T18:25:31Z','Resource','resource.created','Success','Info',platform,'Project','Bucket','ml-models-artifacts','ml-models-artifacts','Provider API','Object storage resource allocated',{
    operationId:'op-2026-1033',metadata:{provider:'Yandex Cloud',bucketRegion:'eu-1'},
  }),
  event('evt-20261107-006','2026-11-07T17:48:13Z','Identity','service_account.created','Success','Info',admin,'Organization','ServiceAccount','sa-analytics','analytics-reader','M8 Identity','Service account metadata registered',{
    changes:{before:{},after:{name:'analytics-reader',scopes:['read:analytics'],privateKey:'[REDACTED]'}},
  }),
  event('evt-20261107-005','2026-11-07T16:12:21Z','Quota','quota.request.approved','Success','Info',admin,'Project','QuotaChangeRequest','QR-108','CPU increase','Quota Manager','Quota increase approved by organization administrator',{
    metadata:{requestedLimit:40,approvalId:'approval-3981'},
  }),
  event('evt-20261107-004','2026-11-07T15:02:34Z','Security','authentication.challenge.failed','Failure','Warning',security,'Organization','AuthenticationChallenge','challenge-2101','step-up','M8 Authentication','Second-factor challenge failed',{
    ipAddress:'192.0.2.10',metadata:{method:'push',reason:'DENIED_BY_USER'},
  }),
  event('evt-20261107-003','2026-11-07T13:19:02Z','Resource','organization.created','Success','Info',admin,'Organization','Organization','acme-sandbox','Acme Sandbox','Resource Manager','Sandbox organization registered',{
    changes:{before:{},after:{name:'Acme Sandbox',state:'Creating'}},
  }),
  event('evt-20261107-002','2026-11-07T12:14:42Z','System','reconciliation.run.completed','Success','Info',system,'Organization','ReconciliationRun','recon-201','gateway-eu1','Gateway Reconciler','Gateway reconciliation run completed',{
    metadata:{checked:42,drift:2},
  }),
  event('evt-20261107-001','2026-11-07T09:01:07Z','System','audit.retention.checked','Success','Info',system,'Organization','RetentionPolicy','audit-retention','Audit retention','M8 Audit','Retention policy configuration checked (demo only)',{
    metadata:{configuredDays:365,attestation:'NOT_VERIFIED'},
  }),
]

const initialState:AuditState={events:initialAuditEvents,exports:[],draft:{}}
let state:AuditState=initialState
const listeners=new Set<()=>void>()
const subscribe=(listener:()=>void)=>{listeners.add(listener);return()=>{listeners.delete(listener)}}
function publish(next:AuditState){state=next;listeners.forEach(fn=>fn())}
export function useAuditDemo(){return useSyncExternalStore(subscribe,()=>state,()=>initialState)}
export function setExportDraft(filters:AuditFilters){publish({...state,draft:{...filters}})}
export function listAuditEvents(filters:AuditFilters={},events:AuditEvent[]=state.events):AuditEvent[]{
  const q=(filters.query??'').trim().toLowerCase()
  const from=filters.fromDate?Date.parse(`${filters.fromDate}T00:00:00Z`):null
  const to=filters.toDate?Date.parse(`${filters.toDate}T23:59:59.999Z`):null
  return events.filter(e=>{
    if(filters.category&&filters.category!=='All'&&filters.category!==e.category)return false
    if(filters.outcome&&filters.outcome!=='All'&&filters.outcome!==e.outcome)return false
    if(filters.severity&&filters.severity!=='All'&&filters.severity!==e.severity)return false
    if(filters.scope&&filters.scope!=='All'&&filters.scope!==e.scope)return false
    if(filters.actorType&&filters.actorType!=='All'&&filters.actorType!==e.actor.type)return false
    const t=Date.parse(e.timestamp)
    if(from!==null&&(Number.isNaN(from)||t<from))return false
    if(to!==null&&(Number.isNaN(to)||t>to))return false
    if(!q)return true
    return [e.id,e.action,e.actor.displayName,e.actor.id,e.resourceId,e.resourceName,e.source,e.correlationId,e.requestId,e.operationId??'',e.description].some(s=>s.toLowerCase().includes(q))
  }).sort((a,b)=>b.timestamp.localeCompare(a.timestamp)||b.id.localeCompare(a.id))
}
const sensitiveKeys=/password|secret|token|private.?key|api.?key|credential|authorization|cookie|session.?id/i
/** Ensures preview and downloadable export never contain raw secrets from structured event payloads. */
export function redactAuditData(value:unknown,key=''):unknown{
  if(sensitiveKeys.test(key))return '[REDACTED]'
  if(Array.isArray(value))return value.map(v=>redactAuditData(v))
  if(value&&typeof value==='object')return Object.fromEntries(Object.entries(value as Record<string,unknown>).map(([k,v])=>[k,redactAuditData(v,k)]))
  return value
}
function safeEvent(event:AuditEvent):Record<string,unknown>{
  return {
    id:event.id,timestamp:event.timestamp,category:event.category,action:event.action,outcome:event.outcome,
    severity:event.severity,actor:event.actor.displayName,actorType:event.actor.type,
    scope:event.scope,scopeName:event.scopeName,resourceType:event.resourceType,resourceId:event.resourceId,
    resourceName:event.resourceName,source:event.source,description:event.description,
    requestId:event.requestId,correlationId:event.correlationId,operationId:event.operationId??'',
    ipAddress:event.ipAddress??'',changes:redactAuditData(event.changes??{}),metadata:redactAuditData(event.metadata??{}),
  }
}
const exportKeys=['id','timestamp','category','action','outcome','severity','actor','actorType','scope','scopeName','resourceType','resourceId','resourceName','source','description','requestId','correlationId','operationId','ipAddress','changes','metadata'] as const
function escapeCsv(value:unknown):string{
  const str=typeof value==='object'?JSON.stringify(value):String(value??'')
  // Escape spreadsheet formula injection, including leading whitespace.
  const protectedValue=/^\s*[=+@\-]/.test(str)?`'${str}`:str
  return `"${protectedValue.replaceAll('"','""')}"`
}
export function serializeAuditExport(events:AuditEvent[],format:ExportFormat):string{
  const records=events.map(safeEvent)
  if(format==='JSONL')return records.map(e=>JSON.stringify(e)).join('\n')+(records.length?'\n':'')
  return [exportKeys.join(','),...records.map(record=>exportKeys.map(k=>escapeCsv(record[k])).join(','))].join('\r\n')+'\r\n'
}
export function createAuditExport(filters:AuditFilters,format:ExportFormat):AuditExport{
  const events=listAuditEvents(filters)
  const id=`export-${String(state.exports.length+1).padStart(4,'0')}`
  const ext=format==='CSV'?'csv':'jsonl'
  const entry:AuditExport={
    id,format,status:'Queued',filters:{...filters},createdAt:'2026-11-07T21:16:00Z',
    requestedBy:'Platform Admin',eventIds:events.map(e=>e.id),rowCount:events.length,redaction:'Safe',
    filename:`m8-audit-${id}.${ext}`,
  }
  const logged=event('evt-20261107-export-'+id,'2026-11-07T21:16:01Z','System','audit.export.requested','Success','Info',admin,'Organization','AuditExport',id,id,'M8 Audit','A demo export request was created',{
    metadata:{format,rowCount:entry.rowCount,redaction:'Safe'},
  })
  publish({events:[logged,...state.events],exports:[entry,...state.exports]})
  return entry
}
export function prepareAuditExport(id:string):boolean{
  const found=state.exports.find(e=>e.id===id)
  if(!found||found.status!=='Queued')return false
  const logged=event(`evt-20261107-ready-${id}`,'2026-11-07T21:16:10Z','System','audit.export.ready','Success','Info',system,'Organization','AuditExport',id,id,'M8 Audit','Redacted demo export prepared',{metadata:{exportId:id,rowCount:found.rowCount}})
  publish({events:[logged,...state.events],exports:state.exports.map(e=>e.id===id?{...e,status:'Ready'}:e)})
  return true
}
export function getExportContent(id:string):string|null{
  const job=state.exports.find(e=>e.id===id)
  if(!job||job.status!=='Ready')return null
  const wanted=new Set(job.eventIds)
  return serializeAuditExport(job.eventIds.map(eventId=>state.events.find(e=>e.id===eventId)).filter((e):e is AuditEvent=>!!e&&wanted.has(e.id)),job.format)
}
export function getAuditEvent(id:string){return state.events.find(e=>e.id===id)}
export function resetAuditDemo(){publish(initialState)}
