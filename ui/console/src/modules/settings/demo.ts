import {useSyncExternalStore} from 'react'

export interface ProjectSettings {
  id:string
  organization:string
  displayName:string
  slug:string
  description:string
  environment:'production'|'development'|'staging'
  region:'eu-1'|'eu-2'|'us-1'
  defaultNamespace:string
  labels:Record<string,string>
  version:number
}
export type ModuleTier='Community'|'Professional'|'Enterprise'
export type ModuleCategory='Core'|'Infrastructure'|'Business'|'AI'
export interface ModuleDefinition {
  id:string
  title:string
  description:string
  category:ModuleCategory
  tier:ModuleTier
  dependencies:string[]
  capabilities:string[]
  enabled:boolean
  required:boolean
}
export type IntegrationKind='Kubernetes'|'Temporal'|'Mimir'|'Loki'|'Tempo'|'ClickHouse'|'YDB'|'Keycloak'|'Agent Router'
export interface Integration {
  id:string
  name:string
  kind:IntegrationKind
  endpoint:string
  secretRef:string
  status:'Configured'|'Needs setup'|'Disabled'
  environment:'production'|'development'
  lastChecked:string
  testNote:string
}
export type WebhookEvent='resource.created'|'resource.updated'|'resource.deleted'|'quota.exceeded'|'operation.failed'|'gateway.route.changed'|'audit.security_event'
export interface Webhook {
  id:string
  name:string
  url:string
  events:WebhookEvent[]
  status:'Enabled'|'Disabled'
  signing:'Not configured'|'External Secret'
  secretRef:string
  deliverySuccessPercent:number
  createdAt:string
}
export interface WebhookDelivery {
  id:string
  webhookId:string
  event:WebhookEvent
  status:'Delivered'|'Failed'|'Pending'|'Simulated'
  attempt:number
  timestamp:string
  responseCode?:number
  operationId?:string
}
export type TokenScope='resources:read'|'resources:write'|'quotas:read'|'gateway:manage'|'audit:read'|'metrics:read'
export interface ApiToken {
  id:string
  name:string
  owner:string
  scopes:TokenScope[]
  status:'Active'|'Revoked'|'Draft'
  createdAt:string
  expiresAt:string
  lastUsed:string
  fingerprint:string
}
interface SettingsState {
  project:ProjectSettings
  modules:ModuleDefinition[]
  integrations:Integration[]
  webhooks:Webhook[]
  deliveries:WebhookDelivery[]
  tokens:ApiToken[]
  activity:string[]
}
export const allowedWebhookEvents:WebhookEvent[]=[
  'resource.created','resource.updated','resource.deleted','quota.exceeded',
  'operation.failed','gateway.route.changed','audit.security_event',
]
export const tokenScopes:TokenScope[]=[
  'resources:read','resources:write','quotas:read','gateway:manage','audit:read','metrics:read',
]
export const integrationKinds:IntegrationKind[]=[
  'Kubernetes','Temporal','Mimir','Loki','Tempo','ClickHouse','YDB','Keycloak','Agent Router',
]
export const initialSettings:SettingsState={
  project:{
    id:'project-ai-studio',organization:'Acme / Production',displayName:'AI Studio',slug:'ai-studio',
    description:'Internal AI applications, gateway routes and model workloads',
    environment:'production',region:'eu-1',defaultNamespace:'ai-studio',
    labels:{team:'ml-platform',cost_center:'engineering'},version:12,
  },
  modules:[
    {id:'resource-manager',title:'Resource Manager',description:'Resource Catalog, organizations and project ownership',category:'Core',tier:'Community',dependencies:[],capabilities:['resources.read','resources.write'],enabled:true,required:true},
    {id:'identity',title:'Identity',description:'Users, service accounts and directory',category:'Core',tier:'Community',dependencies:['resource-manager'],capabilities:['identity.users','identity.service_accounts'],enabled:true,required:true},
    {id:'authentication',title:'Authentication',description:'OIDC, CIBA, passkeys and step-up policies',category:'Core',tier:'Community',dependencies:['identity'],capabilities:['auth.sessions','auth.step_up'],enabled:true,required:true},
    {id:'access',title:'Access',description:'Roles, permissions and access policy decisions',category:'Core',tier:'Professional',dependencies:['identity','resource-manager'],capabilities:['access.check','access.policies'],enabled:true,required:true},
    {id:'audit',title:'Audit',description:'Append-only platform events and exports',category:'Core',tier:'Community',dependencies:['resource-manager'],capabilities:['audit.events','audit.exports'],enabled:true,required:true},
    {id:'quota-manager',title:'Quota Manager',description:'Allocated quotas, reservations and reconciliation',category:'Infrastructure',tier:'Professional',dependencies:['resource-manager'],capabilities:['quota.check','quota.reserve'],enabled:true,required:false},
    {id:'workflow',title:'Workflow Engine',description:'Temporal provisioning and long-running operations',category:'Infrastructure',tier:'Professional',dependencies:['resource-manager'],capabilities:['operations.execute'],enabled:true,required:false},
    {id:'gateway',title:'Gateway',description:'Envoy Gateway API services, routes and policies',category:'Infrastructure',tier:'Professional',dependencies:['resource-manager','access'],capabilities:['gateway.routes','gateway.policies'],enabled:true,required:false},
    {id:'observability',title:'Observability',description:'Mimir, Loki, Tempo and SLO diagnostics',category:'Infrastructure',tier:'Community',dependencies:['resource-manager'],capabilities:['metrics.query','logs.query','traces.query'],enabled:true,required:false},
    {id:'provider-api',title:'Provider API',description:'Resource type registration and cloud adapters',category:'Infrastructure',tier:'Professional',dependencies:['resource-manager','workflow'],capabilities:['providers.reconcile'],enabled:true,required:false},
    {id:'billing',title:'Billing',description:'Metering, pricing, usage ledger and invoices',category:'Business',tier:'Enterprise',dependencies:['resource-manager','quota-manager'],capabilities:['billing.usage','billing.invoices'],enabled:false,required:false},
    {id:'ai-gateway',title:'AI Gateway',description:'Agent Router, providers, models and inference routes',category:'AI',tier:'Professional',dependencies:['gateway','quota-manager','provider-api'],capabilities:['ai.routes','ai.usage'],enabled:true,required:false},
    {id:'ai-studio',title:'AI Studio',description:'Agents, tools and inference workspaces',category:'AI',tier:'Enterprise',dependencies:['ai-gateway','identity'],capabilities:['ai.agents','ai.tools'],enabled:false,required:false},
    {id:'risk-decision',title:'Risk Decision',description:'Adaptive access policies and fraud signals',category:'Business',tier:'Enterprise',dependencies:['authentication','access'],capabilities:['security.risk.decide'],enabled:false,required:false},
  ],
  integrations:[
    {id:'int-kubernetes',name:'Cozystack eu-1',kind:'Kubernetes',endpoint:'https://kubernetes.demo.example',secretRef:'gateway/kubernetes-eu1',status:'Configured',environment:'production',lastChecked:'2026-11-07 21:30',testNote:'Demo reference only; not connected'},
    {id:'int-temporal',name:'Temporal Cluster',kind:'Temporal',endpoint:'https://temporal.demo.example',secretRef:'workflow/temporal',status:'Configured',environment:'production',lastChecked:'2026-11-07 21:28',testNote:'Demo reference only; not connected'},
    {id:'int-mimir',name:'Mimir Metrics',kind:'Mimir',endpoint:'https://mimir.demo.example',secretRef:'observability/mimir',status:'Configured',environment:'production',lastChecked:'2026-11-07 21:22',testNote:'Demo reference only; not connected'},
    {id:'int-loki',name:'Loki Logs',kind:'Loki',endpoint:'https://loki.demo.example',secretRef:'observability/loki',status:'Configured',environment:'production',lastChecked:'2026-11-07 21:20',testNote:'Demo reference only; not connected'},
    {id:'int-tempo',name:'Tempo Traces',kind:'Tempo',endpoint:'https://tempo.demo.example',secretRef:'observability/tempo',status:'Needs setup',environment:'production',lastChecked:'Never',testNote:'Example target not configured'},
    {id:'int-keycloak',name:'Keycloak Identity',kind:'Keycloak',endpoint:'https://sso.demo.example',secretRef:'identity/keycloak',status:'Configured',environment:'production',lastChecked:'2026-11-07 21:15',testNote:'Demo reference only; not connected'},
    {id:'int-router',name:'Agent Router',kind:'Agent Router',endpoint:'https://agent-router.demo.example',secretRef:'ai/agent-router',status:'Needs setup',environment:'development',lastChecked:'Never',testNote:'Example target not configured'},
  ],
  webhooks:[
    {id:'hook-001',name:'Platform events',url:'https://hooks.demo.example/m8/events',events:['resource.created','resource.updated','resource.deleted'],status:'Enabled',signing:'External Secret',secretRef:'webhooks/platform-events',deliverySuccessPercent:99.5,createdAt:'2026-11-03'},
    {id:'hook-002',name:'Quota notifications',url:'https://alerts.demo.example/quota',events:['quota.exceeded','operation.failed'],status:'Enabled',signing:'External Secret',secretRef:'webhooks/quotas',deliverySuccessPercent:94.1,createdAt:'2026-11-04'},
    {id:'hook-003',name:'Security feed',url:'https://security.demo.example/ingest',events:['audit.security_event'],status:'Disabled',signing:'Not configured',secretRef:'',deliverySuccessPercent:0,createdAt:'2026-11-06'},
  ],
  deliveries:[
    {id:'delivery-004',webhookId:'hook-001',event:'resource.created',status:'Delivered',attempt:1,timestamp:'2026-11-07 21:42',responseCode:202,operationId:'op-2026-1040'},
    {id:'delivery-003',webhookId:'hook-002',event:'operation.failed',status:'Failed',attempt:3,timestamp:'2026-11-07 21:35',responseCode:503,operationId:'op-2026-1038'},
    {id:'delivery-002',webhookId:'hook-002',event:'quota.exceeded',status:'Delivered',attempt:1,timestamp:'2026-11-07 20:17',responseCode:200},
    {id:'delivery-001',webhookId:'hook-001',event:'resource.updated',status:'Delivered',attempt:1,timestamp:'2026-11-07 19:42',responseCode:200},
  ],
  tokens:[
    {id:'tok-001',name:'console-automation',owner:'platform-admin',scopes:['resources:read','quotas:read','metrics:read'],status:'Active',createdAt:'2026-10-11',expiresAt:'2027-01-11',lastUsed:'2026-11-07 21:40',fingerprint:'m8_demo_••••f4c2'},
    {id:'tok-002',name:'gateway-ci',owner:'svc-ci',scopes:['resources:read','gateway:manage'],status:'Active',createdAt:'2026-09-15',expiresAt:'2026-12-15',lastUsed:'2026-11-07 18:20',fingerprint:'m8_demo_••••b81a'},
    {id:'tok-003',name:'legacy-export',owner:'platform-admin',scopes:['audit:read'],status:'Revoked',createdAt:'2026-08-18',expiresAt:'2026-11-18',lastUsed:'2026-10-29 12:11',fingerprint:'m8_demo_••••0d9a'},
  ],
  activity:[],
}
let state:SettingsState=initialSettings
const listeners=new Set<()=>void>()
const subscribe=(listener:()=>void)=>{listeners.add(listener);return()=>{listeners.delete(listener)}}
function publish(next:SettingsState){state=next;listeners.forEach(fn=>fn())}
function emit(next:SettingsState,entry:string){publish({...next,activity:[entry,...state.activity].slice(0,25)})}
export function useSettingsDemo(){return useSyncExternalStore(subscribe,()=>state,()=>initialSettings)}
export function resetSettingsDemo(){publish(initialSettings)}
export function validateProject(input:Pick<ProjectSettings,'displayName'|'slug'|'description'|'region'|'environment'|'defaultNamespace'>):string|null{
  if(input.displayName.trim().length<3||input.displayName.trim().length>80)return 'Name must be 3–80 characters'
  if(!/^[a-z][a-z0-9-]{2,49}$/.test(input.slug)||input.slug.endsWith('-'))return 'Slug must be 3–50 lowercase letters, digits or hyphens'
  if(!/^[a-z][a-z0-9-]{2,62}$/.test(input.defaultNamespace))return 'Invalid Kubernetes namespace'
  if(input.description.length>500)return 'Description must be at most 500 characters'
  if(!['eu-1','eu-2','us-1'].includes(input.region))return 'Invalid region'
  if(!['production','development','staging'].includes(input.environment))return 'Invalid environment'
  return null
}
export function saveProject(input:Pick<ProjectSettings,'displayName'|'slug'|'description'|'region'|'environment'|'defaultNamespace'>):string|null{
  const error=validateProject(input);if(error)return error
  emit({...state,project:{...state.project,...input,version:state.project.version+1}},'Project metadata updated (demo)')
  return null
}
const tiers:ModuleTier[]=['Community','Professional','Enterprise']
const demoTier:ModuleTier='Professional'
function hasLicense(tier:ModuleTier){return tiers.indexOf(demoTier)>=tiers.indexOf(tier)}
export function validateModuleToggle(id:string,enable:boolean):string|null{
  const module=state.modules.find(m=>m.id===id)
  if(!module)return 'Module not found'
  if(module.enabled===enable)return 'Module is already in the requested state'
  if(enable){
    if(!hasLicense(module.tier))return `${module.tier} license is required (demo license: ${demoTier})`
    const missing=module.dependencies.filter(dep=>!state.modules.find(m=>m.id===dep)?.enabled)
    if(missing.length)return `Enable dependencies first: ${missing.join(', ')}`
  }else{
    if(module.required)return 'Core required module cannot be disabled'
    const dependents=state.modules.filter(m=>m.enabled&&m.dependencies.includes(id))
    if(dependents.length)return `Module is required by: ${dependents.map(m=>m.title).join(', ')}`
  }
  return null
}
export function toggleModule(id:string):string|null{
  const item=state.modules.find(m=>m.id===id)
  if(!item)return 'Module not found'
  const problem=validateModuleToggle(id,!item.enabled)
  if(problem)return problem
  emit({...state,modules:state.modules.map(m=>m.id===id?{...m,enabled:!m.enabled}:m)},`Module ${item.title}: ${item.enabled?'disabled':'enabled'} (demo)`)
  return null
}
const validHost=(hostname:string)=>hostname.includes('.')&&!hostname.includes('..')&&!/\s/.test(hostname)
const forbiddenHost=(host:string)=>{
  const h=host.toLowerCase().replace(/\.$/,'')
  if(h==='localhost'||h.endsWith('.localhost')||h.endsWith('.local')||h.endsWith('.internal'))return true
  if(h==='0.0.0.0'||h==='::1'||h==='[::1]')return true
  const octets=h.split('.').map(Number)
  if(octets.length===4&&octets.every((n,i)=>h.split('.')[i]!==''&&Number.isInteger(n)&&n>=0&&n<=255)){
    const [a,b]=octets
    return a===0||a===10||a===127||a===169&&b===254||a===172&&b>=16&&b<=31||a===192&&b===168||a===100&&b>=64&&b<=127||a>=224||a===198&&b===18
  }
  return h.includes(':')
}
export function validateHttpsEndpoint(raw:string):string|null{
  try{
    const u=new URL(raw)
    if(u.protocol!=='https:')return 'HTTPS is required'
    if(u.username||u.password)return 'URLs cannot contain credentials'
    if(!validHost(u.hostname)||forbiddenHost(u.hostname))return 'Private, local or invalid hostnames are not allowed'
    if(u.hash)return 'URL fragments are not allowed'
    return null
  }catch{return 'Invalid HTTPS URL'}
}
export function validateIntegration(input:Pick<Integration,'name'|'kind'|'endpoint'|'secretRef'>):string|null{
  if(input.name.trim().length<3)return 'Integration name must have at least 3 characters'
  if(!integrationKinds.includes(input.kind))return 'Unknown integration kind'
  const urlError=validateHttpsEndpoint(input.endpoint)
  if(urlError)return urlError
  if(input.secretRef&&!/^[a-z0-9][a-z0-9/_-]{2,127}$/.test(input.secretRef))return 'Secret reference must use a stable vault path'
  return null
}
export function saveIntegration(input:Pick<Integration,'name'|'kind'|'endpoint'|'secretRef'|'environment'>,id?:string):string|null{
  const error=validateIntegration(input)
  if(error)return error
  if(id&&!state.integrations.some(i=>i.id===id))return 'Integration not found'
  const target=id??`int-demo-${state.integrations.filter(i=>i.id.startsWith('int-demo-')).length+1}`
  const item:Integration={...input,id:target,status:'Needs setup',lastChecked:'Never',testNote:'Not verified; simulated metadata'}
  emit({...state,integrations:id?state.integrations.map(i=>i.id===id?item:i):[...state.integrations,item]},`Integration ${input.name} metadata saved (demo)`)
  return null
}
export function testIntegration(id:string):boolean{
  if(!state.integrations.some(i=>i.id===id))return false
  emit({...state,integrations:state.integrations.map(i=>i.id===id?{...i,testNote:'UI simulation only — no connection test made',lastChecked:'Demo check only'}:i)},`Integration ${id} demo check; no outbound request`)
  return true
}
export function toggleIntegration(id:string):boolean{
  const item=state.integrations.find(i=>i.id===id)
  if(!item)return false
  emit({...state,integrations:state.integrations.map(i=>i.id===id?{...i,status:item.status==='Disabled'?'Needs setup':'Disabled'}:i)},`Integration ${id} status changed (demo)`)
  return true
}
export function validateWebhook(input:Pick<Webhook,'name'|'url'|'events'|'secretRef'>):string|null{
  if(input.name.trim().length<3)return 'Webhook name must have at least 3 characters'
  const urlError=validateHttpsEndpoint(input.url);if(urlError)return urlError
  if(!input.events.length||input.events.some(e=>!allowedWebhookEvents.includes(e)))return 'Choose one or more supported events'
  if(input.secretRef&&!/^[a-z0-9][a-z0-9/_-]{2,127}$/.test(input.secretRef))return 'Secret reference must use a stable vault path'
  return null
}
export function saveWebhook(input:Pick<Webhook,'name'|'url'|'events'|'secretRef'>,id?:string):string|null{
  const error=validateWebhook(input);if(error)return error
  const old=state.webhooks.find(h=>h.id===id)
  if(id&&!old)return 'Webhook not found'
  const item:Webhook={
    ...input,id:id??`hook-demo-${state.webhooks.filter(h=>h.id.startsWith('hook-demo-')).length+1}`,
    status:old?.status??'Disabled',signing:input.secretRef?'External Secret':'Not configured',
    deliverySuccessPercent:old?.deliverySuccessPercent??0,createdAt:old?.createdAt??'2026-11-07',
  }
  emit({...state,webhooks:id?state.webhooks.map(h=>h.id===id?item:h):[...state.webhooks,item]},`Webhook ${item.name} saved (demo)`)
  return null
}
export function toggleWebhook(id:string):boolean{
  const item=state.webhooks.find(h=>h.id===id)
  if(!item)return false
  // Require an external signing secret for enabling a webhook.
  if(item.status==='Disabled'&&!item.secretRef)return false
  emit({...state,webhooks:state.webhooks.map(h=>h.id===id?{...h,status:item.status==='Enabled'?'Disabled':'Enabled'}:h)},`Webhook ${id} status updated (demo)`)
  return true
}
export function retryDelivery(id:string):string|null{
  const old=state.deliveries.find(d=>d.id===id)
  if(!old||old.status!=='Failed')return null
  const hook=state.webhooks.find(h=>h.id===old.webhookId)
  if(!hook||hook.status!=='Enabled')return null
  const nextId=`delivery-demo-${state.deliveries.filter(d=>d.id.startsWith('delivery-demo-')).length+1}`
  const item:WebhookDelivery={...old,id:nextId,status:'Simulated',attempt:old.attempt+1,timestamp:'2026-11-07 21:50',responseCode:undefined}
  emit({...state,deliveries:[item,...state.deliveries]},`Webhook delivery ${nextId} simulated from ${id}; no HTTP request`)
  return nextId
}
export function validateTokenRequest(input:{name:string;scopes:TokenScope[];expiryDays:number}):string|null{
  if(!/^[a-z][a-z0-9-]{2,49}$/.test(input.name))return 'Token name must be 3–50 lowercase letters, digits or hyphens'
  if(!input.scopes.length||input.scopes.some(s=>!tokenScopes.includes(s)))return 'Choose valid scopes'
  if(![7,30,90].includes(input.expiryDays))return 'Choose 7, 30 or 90 days'
  if(state.tokens.some(t=>t.name===input.name&&t.status!=='Revoked'))return 'Token name already in use'
  return null
}
export function requestDemoToken(input:{name:string;scopes:TokenScope[];expiryDays:number}):string|null{
  const error=validateTokenRequest(input);if(error)return error
  const expiry=new Date(Date.UTC(2026,10,7+input.expiryDays)).toISOString().slice(0,10)
  const item:ApiToken={
    id:`tok-demo-${state.tokens.filter(t=>t.id.startsWith('tok-demo-')).length+1}`,
    name:input.name,owner:'platform-admin',scopes:input.scopes,status:'Draft',
    createdAt:'2026-11-07',expiresAt:expiry,lastUsed:'Never',
    fingerprint:'No credential issued',
  }
  emit({...state,tokens:[item,...state.tokens]},`API token draft ${item.name} created; no credential issued`)
  return null
}
export function revokeDemoToken(id:string):boolean{
  const item=state.tokens.find(t=>t.id===id)
  if(!item||item.status==='Revoked')return false
  emit({...state,tokens:state.tokens.map(t=>t.id===id?{...t,status:'Revoked'}:t)},`API token ${id} marked revoked (demo only)`)
  return true
}
