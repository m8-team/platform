import {useSyncExternalStore} from 'react'

export type QuotaScope = 'project' | 'organization'
export type QuotaStatus = 'Pending' | 'Approved' | 'Applied' | 'Rejected'
export type ReservationStatus = 'Reserved' | 'Reconciling' | 'Committed' | 'Released'
export type FindingStatus = 'Critical' | 'Warning' | 'Healthy' | 'Resolved'
export type QuotaMode = 'Dedicated' | 'Shared'
export type QuotaCategory = 'Compute' | 'Storage' | 'Network' | 'AI'

export interface Quota {
  id: string
  title: string
  category: QuotaCategory
  unit: string
  scope: QuotaScope
  used: number
  reserved: number
  limit: number
  source: 'Direct' | 'Inherited'
  mode: QuotaMode
  otherAllocated: number
  description: string
}

export interface Reservation {
  id: string
  resource: string
  kind: string
  quantity: string
  status: ReservationStatus
  operation: string
  provider: string
  age: string
  timeline: string[]
}

export interface QuotaRequest {
  id: string
  quotaId: string
  scope: QuotaScope
  requestedLimit: number
  reason: string
  requester: string
  status: QuotaStatus
  created: string
}

export interface Finding {
  id: string
  resource: string
  kind: string
  expected: string
  observed: string
  status: FindingStatus
  detail: string
  provider: string
}

interface DemoState {
  scope: QuotaScope
  quotas: Quota[]
  reservations: Reservation[]
  requests: QuotaRequest[]
  findings: Finding[]
  scanCount: number
  activity: string[]
}

const initialQuotas: Quota[] = [
  {id: 'compute.vcpu.allocated', title: 'CPU', category: 'Compute', unit: 'vCPU', scope: 'project', used: 28, reserved: 4, limit: 40, source: 'Direct', mode: 'Dedicated', otherAllocated: 48, description: 'Requested virtual CPU cores allocated to workloads'},
  {id: 'compute.memory.allocated', title: 'Memory', category: 'Compute', unit: 'GiB', scope: 'project', used: 82, reserved: 8, limit: 128, source: 'Direct', mode: 'Dedicated', otherAllocated: 128, description: 'Allocated memory across virtual machines and workloads'},
  {id: 'storage.capacity.allocated', title: 'Storage', category: 'Storage', unit: 'GiB', scope: 'project', used: 340, reserved: 20, limit: 500, source: 'Inherited', mode: 'Shared', otherAllocated: 1000, description: 'Provisioned persistent storage capacity'},
  {id: 'network.public_ips.count', title: 'Public IPs', category: 'Network', unit: 'IPs', scope: 'project', used: 3, reserved: 1, limit: 8, source: 'Direct', mode: 'Dedicated', otherAllocated: 12, description: 'Public IP address allocations'},
  {id: 'ai.gateway.routes.count', title: 'AI Gateway Routes', category: 'AI', unit: 'routes', scope: 'project', used: 12, reserved: 3, limit: 20, source: 'Direct', mode: 'Dedicated', otherAllocated: 20, description: 'Agent Router / AI Gateway route resources'},
  {id: 'ai.tokens.monthly', title: 'AI Tokens', category: 'AI', unit: 'tokens/month', scope: 'project', used: 7200000, reserved: 0, limit: 10000000, source: 'Direct', mode: 'Shared', otherAllocated: 10000000, description: 'Monthly AI token consumption budget'},
  {id: 'compute.vcpu.allocated', title: 'CPU', category: 'Compute', unit: 'vCPU', scope: 'organization', used: 72, reserved: 8, limit: 120, source: 'Direct', mode: 'Dedicated', otherAllocated: 0, description: 'Organization-wide vCPU allocation'},
  {id: 'compute.memory.allocated', title: 'Memory', category: 'Compute', unit: 'GiB', scope: 'organization', used: 220, reserved: 28, limit: 384, source: 'Direct', mode: 'Dedicated', otherAllocated: 0, description: 'Organization-wide memory'},
  {id: 'storage.capacity.allocated', title: 'Storage', category: 'Storage', unit: 'GiB', scope: 'organization', used: 1020, reserved: 80, limit: 2000, source: 'Direct', mode: 'Dedicated', otherAllocated: 0, description: 'Organization-wide persistent storage'},
  {id: 'network.public_ips.count', title: 'Public IPs', category: 'Network', unit: 'IPs', scope: 'organization', used: 9, reserved: 2, limit: 32, source: 'Direct', mode: 'Dedicated', otherAllocated: 0, description: 'Organization-wide public addresses'},
  {id: 'ai.gateway.routes.count', title: 'AI Gateway Routes', category: 'AI', unit: 'routes', scope: 'organization', used: 24, reserved: 4, limit: 50, source: 'Direct', mode: 'Dedicated', otherAllocated: 0, description: 'Organization-wide route resources'},
  {id: 'ai.tokens.monthly', title: 'AI Tokens', category: 'AI', unit: 'tokens/month', scope: 'organization', used: 14800000, reserved: 0, limit: 30000000, source: 'Direct', mode: 'Shared', otherAllocated: 0, description: 'Monthly AI token budget'},
]

const initialState: DemoState = {
  scope: 'project',
  quotas: initialQuotas,
  reservations: [
    {id:'res-1042',resource:'vm-inference-02',kind:'VirtualMachine',quantity:'+4 vCPU · +8 GiB',status:'Reserved',operation:'op-2026-1042',provider:'cozystack/eu-1',age:'2 min',timeline:['Request accepted','YDB quota reserved','Temporal provisioning started']},
    {id:'res-1041',resource:'route-gpt-5',kind:'AIGatewayRoute',quantity:'+1 route',status:'Reconciling',operation:'op-2026-1041',provider:'agent-router/eu-1',age:'19 min',timeline:['Request accepted','Quota reserved','Provider outcome unknown','Reconciliation required']},
    {id:'res-1039',resource:'route-embeddings',kind:'AIGatewayRoute',quantity:'+1 route',status:'Committed',operation:'op-2026-1039',provider:'agent-router/eu-1',age:'31 min',timeline:['Request accepted','Quota reserved','Provider confirmed resource','Allocation committed']},
    {id:'res-1038',resource:'vm-batch-03',kind:'VirtualMachine',quantity:'+8 vCPU',status:'Released',operation:'op-2026-1038',provider:'cozystack/eu-1',age:'1 h',timeline:['Request accepted','Quota reserved','Provider confirmed absence','Reservation released']},
  ],
  requests: [
    {id:'QR-108',quotaId:'compute.vcpu.allocated',scope:'project',requestedLimit:64,reason:'Additional inference workloads',requester:'ML Platform Team',status:'Pending',created:'2026-11-07'},
    {id:'QR-107',quotaId:'ai.gateway.routes.count',scope:'project',requestedLimit:40,reason:'New customer AI Gateway routes',requester:'Platform Team',status:'Pending',created:'2026-11-06'},
    {id:'QR-106',quotaId:'storage.capacity.allocated',scope:'project',requestedLimit:800,reason:'Test dataset storage',requester:'Development Team',status:'Approved',created:'2026-11-05'},
  ],
  findings: [
    {id:'DR-201',resource:'route-gpt-5',kind:'AIGatewayRoute',expected:'1 route / reserved',observed:'CRD present',status:'Critical',detail:'Reservation still pending despite observed resource. Commit requires a verified resource identity.',provider:'agent-router/eu-1'},
    {id:'DR-202',resource:'tenant-ai',kind:'ResourceQuota',expected:'hard.cpu: 40',observed:'hard.cpu: 48',status:'Warning',detail:'The cluster hard limit differs from the desired M8 policy. Compare resource version before repair.',provider:'cozystack/eu-1'},
    {id:'DR-203',resource:'vm-inference-01',kind:'VirtualMachine',expected:'8 vCPU',observed:'8 vCPU',status:'Healthy',detail:'Provider allocation matches YDB records.',provider:'cozystack/eu-1'},
    {id:'DR-204',resource:'pv-models',kind:'PersistentVolumeClaim',expected:'200 GiB',observed:'200 GiB',status:'Healthy',detail:'Persistent volume request matches recorded allocation.',provider:'cozystack/eu-1'},
  ],
  scanCount: 3,
  activity: ['Quota admission reserved 4 vCPU for vm-inference-02','DR-201 requires reconciliation','CPU quota increase requested by ML Platform Team'],
}

let state = initialState
const listeners = new Set<() => void>()
function publish(next: DemoState) {
  state = next
  listeners.forEach((listener) => listener())
}
const subscribe = (listener: () => void) => {
  listeners.add(listener)
  return () => {listeners.delete(listener)}
}
export function useQuotaDemo() {
  return useSyncExternalStore(subscribe, () => state, () => initialState)
}
export function setDemoScope(scope: QuotaScope) {
  publish({...state, scope})
}
function activity(message: string) {
  return [message, ...state.activity].slice(0, 8)
}
export function validateQuotaLimit(scope: QuotaScope, id: string, value: number, mode: QuotaMode): string | null {
  const quota = state.quotas.find((item) => item.scope === scope && item.id === id)
  if (!quota) return 'Quota definition not found'
  if (!Number.isSafeInteger(value) || value <= 0) return 'The limit must be a positive whole number'
  if (scope === 'project' && mode === 'Dedicated') {
    const parent = state.quotas.find((item) => item.scope === 'organization' && item.id === id)
    if (parent && quota.otherAllocated + value > parent.limit) {
      return `Parent budget exceeded: only ${Math.max(0, parent.limit - quota.otherAllocated).toLocaleString('en-US')} ${quota.unit} available for this project`
    }
  }
  return null
}
export function updateDemoPolicy(scope: QuotaScope, id: string, value: number, mode: QuotaMode): string | null {
  const error = validateQuotaLimit(scope, id, value, mode)
  if (error) return error
  publish({...state,
    quotas: state.quotas.map((q) => q.scope === scope && q.id === id ? {...q, limit: value, mode, source: 'Direct'} : q),
    activity: activity(`Quota policy ${id} changed to ${value} (${scope})`),
  })
  return null
}
export function createDemoRequest(scope: QuotaScope, quotaId: string, requestedLimit: number, reason: string): string | null {
  if (!Number.isSafeInteger(requestedLimit) || requestedLimit <= 0) return 'Enter a valid positive requested limit'
  const quota = state.quotas.find((q) => q.scope === scope && q.id === quotaId)
  if (!quota) return 'Select a quota'
  if (requestedLimit <= quota.limit) return 'Requested limit must exceed the current limit'
  if (reason.trim().length < 10) return 'Please provide at least 10 characters of justification'
  const id = `QR-${Math.max(108, ...state.requests.map((r) => Number(r.id.slice(3)) || 0)) + 1}`
  publish({...state, requests: [{id,quotaId,scope,requestedLimit,reason:reason.trim(),requester:'Console Demo User',status:'Pending',created:'2026-11-07'}, ...state.requests], activity: activity(`${id} quota increase requested`)})
  return null
}
export function reviewDemoRequest(id: string, approved: boolean) {
  publish({...state, requests: state.requests.map((r) => r.id === id && r.status === 'Pending' ? {...r, status: approved ? 'Approved' : 'Rejected'} : r), activity: activity(`${id} ${approved ? 'approved' : 'rejected'} (demo)`)})
}
export function applyDemoRequest(id: string): string | null {
  const request = state.requests.find((r) => r.id === id && r.status === 'Approved')
  if (!request) return 'Only approved requests can be applied'
  const quota = state.quotas.find((q) => q.scope === request.scope && q.id === request.quotaId)
  const error = updateDemoPolicy(request.scope, request.quotaId, request.requestedLimit, quota?.mode ?? 'Dedicated')
  if (error) return error
  publish({...state, requests: state.requests.map((r) => r.id === id ? {...r, status: 'Applied'} : r),activity: activity(`${id} applied to demo quota policy`)})
  return null
}
export function inspectDemoReservation(id: string) {
  publish({...state,reservations:state.reservations.map((r)=>r.id===id&&r.status==='Reconciling'?{...r,timeline:[...r.timeline,'Manual re-check simulated; provider state still unverified']}:r),activity:activity(`Reservation ${id} re-check simulated`)})
}
export function runDemoScan() {
  publish({...state, scanCount: state.scanCount + 1, activity:activity('Demo reconciliation scan completed; no provider contacted')})
}
export function reviewDemoFinding(id: string) {
  publish({...state, findings:state.findings.map((f)=>f.id===id&&f.status!=='Healthy'?{...f,status:'Resolved',detail:`${f.detail} Demo finding marked reviewed; no provider resource was changed.`}:f),activity:activity(`Finding ${id} marked reviewed in demo`)})
}
