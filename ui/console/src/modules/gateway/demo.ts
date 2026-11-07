import {useSyncExternalStore} from 'react'

export type GatewayProtocol = 'HTTP' | 'gRPC'
export type ServiceHealth = 'Healthy' | 'Degraded' | 'Unknown'
export type GatewayRouteKind = 'HTTPRoute' | 'AIGatewayRoute'
export type GatewayRouteStatus = 'Accepted' | 'Degraded' | 'Pending' | 'Rejected'
export type ConsumerStatus = 'Active' | 'Suspended'
export type AuthType = 'API Key' | 'OAuth2' | 'mTLS'
export type GatewayPolicyKind = 'JWT' | 'API Key' | 'mTLS' | 'CORS' | 'WAF'
export type PolicyStatus = 'Enabled' | 'Disabled'
export type RateTargetType = 'Gateway' | 'Service' | 'Route' | 'Consumer'
export type RatePeriod = 'minute' | 'hour' | 'day'

export interface ApiService {
  id: string
  name: string
  protocol: GatewayProtocol
  upstream: string
  port: number
  environment: 'Production' | 'Development'
  health: ServiceHealth
  requests24h: number
  latencyP95: number
  description: string
}

export interface GatewayRoute {
  id: string
  name: string
  kind: GatewayRouteKind
  hostname: string
  pathPrefix: string
  method: 'ANY' | 'GET' | 'POST' | 'PUT' | 'DELETE'
  serviceId: string
  status: GatewayRouteStatus
  provider: 'Envoy Gateway' | 'Agent Router'
  requests24h: number
  latencyP95: number
  version: number
  description: string
}

export interface GatewayConsumer {
  id: string
  name: string
  application: string
  auth: AuthType
  status: ConsumerStatus
  requests24h: number
  lastSeen: string
  description: string
}

export interface GatewayPolicy {
  id: string
  name: string
  kind: GatewayPolicyKind
  targetType: 'Gateway' | 'Service' | 'Route'
  targetId: string
  status: PolicyStatus
  description: string
  mode: 'Enforce' | 'Monitor'
}

export interface GatewayRateLimit {
  id: string
  name: string
  targetType: RateTargetType
  targetId: string
  requests: number
  period: RatePeriod
  burst: number
  used: number
  mode: 'Enforce' | 'Monitor'
  status: PolicyStatus
}

export interface GatewayAudit {
  id: number
  event: string
  actor: string
  at: string
}
export interface DemoGateway {
  services: ApiService[]
  routes: GatewayRoute[]
  consumers: GatewayConsumer[]
  policies: GatewayPolicy[]
  limits: GatewayRateLimit[]
  activity: GatewayAudit[]
}

export const initialGateway: DemoGateway = {
  services: [
    {id:'svc-orders',name:'orders-api',protocol:'HTTP',upstream:'orders-api.production.svc.cluster.local',port:8080,environment:'Production',health:'Healthy',requests24h:125420,latencyP95:84,description:'Orders API for omnichannel commerce'},
    {id:'svc-inventory',name:'inventory-api',protocol:'gRPC',upstream:'inventory.production.svc.cluster.local',port:9090,environment:'Production',health:'Healthy',requests24h:88320,latencyP95:42,description:'Inventory availability and stock reservation'},
    {id:'svc-agent-router',name:'agent-router',protocol:'HTTP',upstream:'agent-router.ai-studio.svc.cluster.local',port:8080,environment:'Production',health:'Healthy',requests24h:51700,latencyP95:731,description:'Inference routing for AI Studio'},
    {id:'svc-identity',name:'identity-api',protocol:'HTTP',upstream:'identity.platform.svc.cluster.local',port:8080,environment:'Production',health:'Degraded',requests24h:19840,latencyP95:338,description:'Identity and directory services'},
  ],
  routes: [
    {id:'route-orders',name:'orders-public',kind:'HTTPRoute',hostname:'api.acme.example',pathPrefix:'/v1/orders',method:'ANY',serviceId:'svc-orders',status:'Accepted',provider:'Envoy Gateway',requests24h:112340,latencyP95:82,version:14,description:'Public order lifecycle endpoints'},
    {id:'route-inventory',name:'inventory-grpc',kind:'HTTPRoute',hostname:'grpc.acme.example',pathPrefix:'/inventory.v1.InventoryService',method:'POST',serviceId:'svc-inventory',status:'Accepted',provider:'Envoy Gateway',requests24h:88320,latencyP95:42,version:3,description:'gRPC inventory service'},
    {id:'route-chat',name:'openai-chat',kind:'AIGatewayRoute',hostname:'ai.acme.example',pathPrefix:'/v1/chat/completions',method:'POST',serviceId:'svc-agent-router',status:'Accepted',provider:'Agent Router',requests24h:39100,latencyP95:812,version:9,description:'Chat completion inference workloads'},
    {id:'route-embeddings',name:'embeddings',kind:'AIGatewayRoute',hostname:'ai.acme.example',pathPrefix:'/v1/embeddings',method:'POST',serviceId:'svc-agent-router',status:'Pending',provider:'Agent Router',requests24h:7200,latencyP95:393,version:2,description:'Vector embedding inference'},
    {id:'route-identity',name:'identity-public',kind:'HTTPRoute',hostname:'api.acme.example',pathPrefix:'/v1/identity',method:'ANY',serviceId:'svc-identity',status:'Degraded',provider:'Envoy Gateway',requests24h:19840,latencyP95:338,version:6,description:'Public identity endpoints'},
  ],
  consumers: [
    {id:'consumer-web',name:'commerce-web',application:'Storefront',auth:'OAuth2',status:'Active',requests24h:89440,lastSeen:'2 min ago',description:'Public web customer application'},
    {id:'consumer-erp',name:'erp-sync',application:'ERP integration',auth:'mTLS',status:'Active',requests24h:28400,lastSeen:'7 min ago',description:'Back-office integration from ERP'},
    {id:'consumer-ai',name:'ai-studio-console',application:'AI Studio',auth:'API Key',status:'Active',requests24h:42700,lastSeen:'1 min ago',description:'Internal AI Studio API client; key material is never displayed'},
    {id:'consumer-partner',name:'partner-sandbox',application:'Partner Portal',auth:'OAuth2',status:'Suspended',requests24h:0,lastSeen:'12 days ago',description:'Sandbox consumer currently suspended'},
  ],
  policies: [
    {id:'policy-oidc',name:'OIDC access tokens',kind:'JWT',targetType:'Gateway',targetId:'gateway-production',status:'Enabled',description:'Validate OIDC JWT audience, issuer and expiration',mode:'Enforce'},
    {id:'policy-api-key',name:'AI client API keys',kind:'API Key',targetType:'Route',targetId:'route-chat',status:'Enabled',description:'Validate client credentials via control-plane identity',mode:'Enforce'},
    {id:'policy-cors',name:'Storefront CORS',kind:'CORS',targetType:'Route',targetId:'route-orders',status:'Enabled',description:'Allow listed web origins for store applications',mode:'Enforce'},
    {id:'policy-waf',name:'Coraza WAF (observe)',kind:'WAF',targetType:'Gateway',targetId:'gateway-production',status:'Enabled',description:'Detect common attack patterns; detection mode only',mode:'Monitor'},
    {id:'policy-mtls',name:'ERP client certificates',kind:'mTLS',targetType:'Service',targetId:'svc-inventory',status:'Enabled',description:'Require validated client certificates',mode:'Enforce'},
  ],
  limits: [
    {id:'limit-gateway',name:'Gateway protection',targetType:'Gateway',targetId:'gateway-production',requests:6000,period:'minute',burst:500,used:2470,mode:'Enforce',status:'Enabled'},
    {id:'limit-chat',name:'Chat completions',targetType:'Route',targetId:'route-chat',requests:1200,period:'minute',burst:150,used:850,mode:'Enforce',status:'Enabled'},
    {id:'limit-orders',name:'Orders API',targetType:'Route',targetId:'route-orders',requests:2000,period:'minute',burst:250,used:1560,mode:'Enforce',status:'Enabled'},
    {id:'limit-consumer-ai',name:'AI Studio consumer',targetType:'Consumer',targetId:'consumer-ai',requests:25000,period:'hour',burst:500,used:8200,mode:'Monitor',status:'Enabled'},
  ],
  activity: [
    {id:4,event:'AIGatewayRoute embeddings observed as Pending',actor:'agent-router reconciler',at:'22:14'},
    {id:3,event:'Coraza WAF policy entered Monitor mode',actor:'platform-admin',at:'21:51'},
    {id:2,event:'Quota reservation committed for openai-chat',actor:'quota-manager',at:'20:46'},
    {id:1,event:'Orders API route version 14 accepted',actor:'envoy-gateway',at:'20:31'},
  ],
}

let state = initialGateway
const subscribers = new Set<() => void>()
const subscribe = (listener: () => void) => { subscribers.add(listener); return () => {subscribers.delete(listener)} }
function publish(next: DemoGateway) {state = next; subscribers.forEach(listener => listener())}
export function useGatewayDemo() {return useSyncExternalStore(subscribe,() => state,() => initialGateway)}
export function resetGatewayDemo() {publish(initialGateway)}
function log(next: DemoGateway,event:string):DemoGateway {
  return {...next,activity:[{id:Math.max(0,...state.activity.map(item=>item.id))+1,event,actor:'Console demo',at:'22:16'},...state.activity].slice(0,20)}
}
function nextId(prefix:string,ids:string[]):string {
  const used=new Set(ids)
  let counter=1
  while(used.has(`${prefix}-demo-${counter}`))counter+=1
  return `${prefix}-demo-${counter}`
}
export function validateService(values:Pick<ApiService,'name'|'protocol'|'upstream'|'port'>,existingId?:string):string|null {
  if(!/^[a-z][a-z0-9-]{2,49}$/.test(values.name))return 'Service name must be 3–50 lowercase letters, digits or hyphens'
  if(!/^[a-z0-9][a-z0-9.-]+$/.test(values.upstream)||values.upstream.includes('..'))return 'Enter a DNS-style upstream hostname (without URL scheme)'
  if(!Number.isInteger(values.port)||values.port<1||values.port>65535)return 'Port must be 1–65535'
  if(state.services.some(s=>s.name===values.name&&s.id!==existingId))return 'A service with this name already exists'
  return null
}
export function saveDemoService(values:Pick<ApiService,'name'|'protocol'|'upstream'|'port'|'environment'|'description'>,id?:string):string|null {
  const error=validateService(values,id)
  if(error)return error
  if(id&&!state.services.some(s=>s.id===id))return 'Service not found'
  const service:ApiService={...values,id:id??nextId('svc',state.services.map(s=>s.id)),health:'Unknown',requests24h:0,latencyP95:0}
  const services=id?state.services.map(s=>s.id===id?{...s,...values}:s):[...state.services,service]
  publish(log({...state,services},`Service ${service.name} ${id?'updated':'created'} (demo)`))
  return null
}
function validHostname(host:string):boolean{
  return /^[a-z0-9](?:[a-z0-9.-]*[a-z0-9])?$/.test(host)&&host.includes('.')&&!host.includes('..')&&host.length<=253
}
export function validateRoute(values:Pick<GatewayRoute,'name'|'kind'|'hostname'|'pathPrefix'|'method'|'serviceId'>,existingId?:string):string|null {
  if(!/^[a-z][a-z0-9-]{2,49}$/.test(values.name))return 'Route name must be 3–50 lowercase letters, digits or hyphens'
  if(!validHostname(values.hostname))return 'Enter a valid hostname'
  if(!values.pathPrefix.startsWith('/')||values.pathPrefix.includes('?')||values.pathPrefix.includes('#')||values.pathPrefix.includes('//'))return 'Path prefix must start with / and cannot contain query, fragment or //'
  if(!state.services.some(s=>s.id===values.serviceId))return 'Select a valid backend service'
  if(state.routes.some(r=>r.id!==existingId&&r.name===values.name))return 'Route name already exists'
  if(state.routes.some(r=>r.id!==existingId&&r.hostname===values.hostname&&r.pathPrefix===values.pathPrefix&&(r.method==='ANY'||values.method==='ANY'||r.method===values.method)))return 'A route already claims this hostname, path and method'
  return null
}
export function saveDemoRoute(values:Pick<GatewayRoute,'name'|'kind'|'hostname'|'pathPrefix'|'method'|'serviceId'|'description'>,id?:string):string|null {
  const error=validateRoute(values,id)
  if(error)return error
  const old=state.routes.find(r=>r.id===id)
  if(id&&!old)return 'Route not found'
  const route:GatewayRoute={...values,id:id??nextId('route',state.routes.map(r=>r.id)),status:id?old?.status??'Pending':'Pending',provider:values.kind==='HTTPRoute'?'Envoy Gateway':'Agent Router',requests24h:old?.requests24h??0,latencyP95:old?.latencyP95??0,version:(old?.version??0)+1}
  const routes=id?state.routes.map(r=>r.id===id?route:r):[...state.routes,route]
  publish(log({...state,routes},`Route ${route.name} ${id?'updated':'created'}; reconciliation pending`))
  return null
}
export function simulateGatewayRequest(input:{hostname:string;path:string;method:string;consumerId:string}):{status:number;outcome:string;route?:GatewayRoute;service?:ApiService;appliedPolicies:GatewayPolicy[];limits:GatewayRateLimit[]} {
  const route=[...state.routes].filter(r=>r.hostname===input.hostname&&
    (input.path===r.pathPrefix||input.path.startsWith(r.pathPrefix.endsWith('/')?r.pathPrefix:`${r.pathPrefix}/`))&&
    (r.method==='ANY'||r.method===input.method)).sort((a,b)=>b.pathPrefix.length-a.pathPrefix.length)[0]
  if(!route)return{status:404,outcome:'No matching host + path + method rule',appliedPolicies:[],limits:[]}
  const service=state.services.find(s=>s.id===route.serviceId)
  const appliedPolicies=state.policies.filter(p=>p.status==='Enabled'&&(p.targetType==='Gateway'||p.targetType==='Route'&&p.targetId===route.id||p.targetType==='Service'&&p.targetId===route.serviceId))
  const limits=state.limits.filter(l=>l.status==='Enabled'&&(l.targetType==='Gateway'||l.targetType==='Route'&&l.targetId===route.id||l.targetType==='Service'&&l.targetId===route.serviceId||l.targetType==='Consumer'&&l.targetId===input.consumerId))
  const consumer=state.consumers.find(c=>c.id===input.consumerId)
  if(!consumer||consumer.status!=='Active')return{status:403,outcome:'Consumer is missing or suspended',route,service,appliedPolicies,limits}
  if(route.status==='Rejected'||route.status==='Pending')return{status:503,outcome:route.status==='Pending'?'Route is awaiting provider acceptance':'Route is rejected by the provider',route,service,appliedPolicies,limits}
  const blocked=limits.find(l=>l.mode==='Enforce'&&l.used>=l.requests)
  if(blocked)return{status:429,outcome:`Enforced limit exceeded: ${blocked.name}`,route,service,appliedPolicies,limits}
  return{status:200,outcome:'Simulated routing match; no upstream request sent and no credentials validated',route,service,appliedPolicies,limits}
}
export function saveDemoConsumer(values:Pick<GatewayConsumer,'name'|'application'|'auth'|'description'>):string|null {
  if(!/^[a-z][a-z0-9-]{2,49}$/.test(values.name))return 'Consumer name must be 3–50 lowercase letters, digits or hyphens'
  if(state.consumers.some(c=>c.name===values.name))return 'Consumer already exists'
  if(!values.application.trim())return 'Application is required'
  const consumer:GatewayConsumer={...values,id:nextId('consumer',state.consumers.map(c=>c.id)),status:'Active',requests24h:0,lastSeen:'Never'}
  publish(log({...state,consumers:[...state.consumers,consumer]},`Consumer ${consumer.name} created (metadata only)`))
  return null
}
export function toggleDemoConsumer(id:string):boolean {
  if(!state.consumers.some(c=>c.id===id))return false
  const consumers=state.consumers.map(c=>c.id===id?{...c,status:c.status==='Active'?'Suspended':'Active'} as GatewayConsumer:c)
  publish(log({...state,consumers},`Consumer ${id} changed status (demo)`))
  return true
}
export function saveDemoPolicy(values:Pick<GatewayPolicy,'name'|'kind'|'targetType'|'targetId'|'mode'|'description'>,id?:string):string|null {
  if(!values.name.trim())return 'Policy name is required'
  if(!values.targetId.trim())return 'Policy target is required'
  if(values.targetType==='Service'&&!state.services.some(s=>s.id===values.targetId))return 'Service target not found'
  if(values.targetType==='Route'&&!state.routes.some(r=>r.id===values.targetId))return 'Route target not found'
  const old=state.policies.find(p=>p.id===id)
  if(id&&!old)return 'Policy not found'
  const policy:GatewayPolicy={...values,id:id??nextId('policy',state.policies.map(p=>p.id)),status:old?.status??'Enabled'}
  publish(log({...state,policies:id?state.policies.map(p=>p.id===id?policy:p):[...state.policies,policy]},`Policy ${policy.name} saved (demo)`))
  return null
}
export function toggleDemoPolicy(id:string):boolean{
  if(!state.policies.some(p=>p.id===id))return false
  const policies=state.policies.map(p=>p.id===id?{...p,status:p.status==='Enabled'?'Disabled':'Enabled'} as GatewayPolicy:p)
  publish(log({...state,policies},`Policy ${id} status updated (demo)`))
  return true
}
export function saveDemoLimit(values:Pick<GatewayRateLimit,'name'|'targetType'|'targetId'|'requests'|'period'|'burst'|'mode'>,id?:string):string|null {
  if(!values.name.trim())return 'Limit name is required'
  if(!Number.isSafeInteger(values.requests)||values.requests<1)return 'Requests must be a positive integer'
  if(!Number.isSafeInteger(values.burst)||values.burst<0)return 'Burst must be a nonnegative integer'
  const targets:Record<RateTargetType,string[]>={
    Gateway:['gateway-production'],Service:state.services.map(s=>s.id),Route:state.routes.map(r=>r.id),Consumer:state.consumers.map(c=>c.id),
  }
  if(!targets[values.targetType].includes(values.targetId))return 'Select a valid target'
  const old=state.limits.find(l=>l.id===id)
  if(id&&!old)return 'Rate limit not found'
  const limit:GatewayRateLimit={...values,id:id??nextId('limit',state.limits.map(l=>l.id)),status:old?.status??'Enabled',used:old?.used??0}
  publish(log({...state,limits:id?state.limits.map(l=>l.id===id?limit:l):[...state.limits,limit]},`Rate limit ${limit.name} saved (demo)`))
  return null
}
export function toggleDemoLimit(id:string):boolean{
  if(!state.limits.some(l=>l.id===id))return false
  const limits=state.limits.map(l=>l.id===id?{...l,status:l.status==='Enabled'?'Disabled':'Enabled'} as GatewayRateLimit:l)
  publish(log({...state,limits},`Rate limit ${id} status updated (demo)`))
  return true
}
