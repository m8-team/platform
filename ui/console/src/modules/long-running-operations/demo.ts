import {useSyncExternalStore} from 'react'

export type OperationStatus = 'Queued' | 'Running' | 'Waiting' | 'CancelRequested' | 'Succeeded' | 'Failed' | 'Cancelled'
export type StepStatus = 'Pending' | 'Running' | 'Succeeded' | 'Failed'
export type OperationScope = 'project' | 'organization'

export interface OperationStep {
  id: string
  name: string
  status: StepStatus
  duration?: string
  detail?: string
}

export interface OperationLog {
  at: string
  level: 'INFO' | 'WARN' | 'ERROR'
  message: string
}

export interface OperationError {
  code: string
  message: string
  retryable: boolean
}

export interface LongRunningOperation {
  id: string
  title: string
  description: string
  service: string
  action: string
  resource: string
  resourceType: string
  scope: OperationScope
  status: OperationStatus
  progress: number
  submittedAt: string
  updatedAt: string
  duration: string
  owner: string
  provider: string
  workflowId: string
  requestId: string
  correlationId: string
  reservationId?: string
  attempt: number
  parentOperationId?: string
  cancellable: boolean
  retryable: boolean
  steps: OperationStep[]
  logs: OperationLog[]
  error?: OperationError
  result?: string
}

function makeSteps(names: string[], completed: number, active: 'Running' | 'Failed' | 'Pending' = 'Running'): OperationStep[] {
  return names.map((name, i) => ({
    id: `step-${i+1}`,
    name,
    status: i < completed ? 'Succeeded' : i === completed ? active : 'Pending',
    duration: i < completed ? `${(i+1)*2}s` : undefined,
  }))
}
function makeLogs(lines: Array<[OperationLog['level'],string]>): OperationLog[] {
  return lines.map(([level,message],i) => ({at:`22:${String(8+i).padStart(2,'0')}:12`,level,message}))
}
const initialOperations: LongRunningOperation[] = [
  {
    id:'op-2026-1042',title:'Provision inference VM',description:'Create inference worker and attach private storage',service:'Compute',action:'Create',resource:'vm-inference-02',resourceType:'VirtualMachine',scope:'project',status:'Running',progress:62,submittedAt:'2026-11-07 22:04',updatedAt:'2026-11-07 22:11',duration:'7m 12s',owner:'ML Platform',provider:'Cozystack / eu-1',workflowId:'provision-vm-1042',requestId:'req-72b8',correlationId:'corr-8c07',reservationId:'res-1042',attempt:1,cancellable:true,retryable:false,
    steps:makeSteps(['Validate request','Reserve quotas','Create VM Custom Resource','Wait for reconciliation','Register observed state'],3),
    logs:makeLogs([['INFO','Quota reserved: 4 vCPU, 8 GiB'],['INFO','VirtualMachine CR created in tenant ai-studio'],['INFO','Waiting for provider Ready condition']]),
  },
  {
    id:'op-2026-1041',title:'Deploy AI Gateway route',description:'Create AIGatewayRoute and reconcile agent-router provider',service:'AI Gateway',action:'Create',resource:'route-gpt-5',resourceType:'AIGatewayRoute',scope:'project',status:'Waiting',progress:75,submittedAt:'2026-11-07 21:53',updatedAt:'2026-11-07 22:07',duration:'19m',owner:'AI Studio',provider:'Agent Router / eu-1',workflowId:'ai-route-1041',requestId:'req-504e',correlationId:'corr-b11d',reservationId:'res-1041',attempt:1,cancellable:true,retryable:false,
    steps:makeSteps(['Validate route specification','Reserve route quota','Apply route CRD','Observe provider state','Commit allocation'],3,'Running'),
    logs:makeLogs([['INFO','Quota reservation res-1041 created'],['INFO','AIGatewayRoute submitted'],['WARN','Provider result not yet confirmed; reconciliation pending']]),
  },
  {
    id:'op-2026-1040',title:'Create workspace',description:'Create workspace with default access policies',service:'Resource Manager',action:'Create',resource:'ws-production',resourceType:'Workspace',scope:'organization',status:'Running',progress:34,submittedAt:'2026-11-07 22:01',updatedAt:'2026-11-07 22:09',duration:'8m 40s',owner:'Platform Team',provider:'M8 Control Plane',workflowId:'workspace-create-1040',requestId:'req-99d2',correlationId:'corr-f31a',attempt:1,cancellable:true,retryable:false,
    steps:makeSteps(['Create resource record','Initialize workspace structure','Provision namespace','Configure Access permissions','Mark workspace Active'],1),
    logs:makeLogs([['INFO','Workspace record created'],['INFO','Namespace provisioning started']]),
  },
  {
    id:'op-2026-1039',title:'Deploy embeddings route',description:'Provision completed and allocation committed',service:'AI Gateway',action:'Create',resource:'route-embeddings',resourceType:'AIGatewayRoute',scope:'project',status:'Succeeded',progress:100,submittedAt:'2026-11-07 21:42',updatedAt:'2026-11-07 21:51',duration:'9m 02s',owner:'AI Studio',provider:'Agent Router / eu-1',workflowId:'ai-route-1039',requestId:'req-1a78',correlationId:'corr-ead0',reservationId:'res-1039',attempt:1,cancellable:false,retryable:false,
    steps:makeSteps(['Validate route','Reserve quota','Apply CRD','Observe Ready','Commit quota'],5,'Pending'),
    logs:makeLogs([['INFO','Route created'],['INFO','Provider reported Ready'],['INFO','Quota reservation committed']]),
    result:'Route ready; reservation res-1039 committed',
  },
  {
    id:'op-2026-1038',title:'Delete obsolete VM',description:'Destroy VM and safely release its allocated resources',service:'Compute',action:'Delete',resource:'vm-batch-03',resourceType:'VirtualMachine',scope:'project',status:'Failed',progress:46,submittedAt:'2026-11-07 21:31',updatedAt:'2026-11-07 21:47',duration:'16m 06s',owner:'SRE Team',provider:'Cozystack / eu-1',workflowId:'vm-delete-1038',requestId:'req-0fb1',correlationId:'corr-7e1e',attempt:2,cancellable:false,retryable:true,
    steps:makeSteps(['Validate deletion','Set deletion intent','Delete provider CR','Confirm actual deletion','Release allocation'],2,'Failed'),
    logs:makeLogs([['INFO','Deletion requested'],['WARN','Provider finalizer is still present'],['ERROR','Timeout: resource deletion was not confirmed']]),
    error:{code:'PROVIDER_RECONCILIATION_TIMEOUT',message:'Provider did not confirm deletion before deadline. Do not release allocated quota until absence is verified.',retryable:true},
  },
  {
    id:'op-2026-1037',title:'Scale Kubernetes cluster',description:'Increase worker pool capacity for production cluster',service:'Infrastructure',action:'Update',resource:'k8s-eu1-production',resourceType:'Cluster',scope:'organization',status:'Queued',progress:0,submittedAt:'2026-11-07 22:10',updatedAt:'2026-11-07 22:10',duration:'2m',owner:'Platform Team',provider:'Yandex Cloud / eu-1',workflowId:'cluster-scale-1037',requestId:'req-ab85',correlationId:'corr-8bc9',attempt:1,cancellable:true,retryable:false,
    steps:makeSteps(['Validate capacity','Create placement plan','Request additional nodes','Wait for nodes Ready','Update capacity snapshot'],0,'Pending'),
    logs:makeLogs([['INFO','Operation queued; waiting for provider concurrency slot']]),
  },
  {
    id:'op-2026-1036',title:'Reconcile ClickHouse schemas',description:'Compare replicated database objects after migration',service:'Data Platform',action:'Reconcile',resource:'zendendb',resourceType:'ClickHouseDatabase',scope:'organization',status:'Failed',progress:91,submittedAt:'2026-11-07 20:15',updatedAt:'2026-11-07 21:03',duration:'48m',owner:'Data Engineering',provider:'Managed ClickHouse',workflowId:'clickhouse-reconcile-1036',requestId:'req-932a',correlationId:'corr-5d12',attempt:1,cancellable:false,retryable:true,
    steps:makeSteps(['Discover source objects','Compare database schemas','Verify table engines','Validate row checksums','Publish reconciliation report'],3,'Failed'),
    logs:makeLogs([['INFO','Validated 254 compatible tables'],['WARN','Detected table with unexpected engine'],['ERROR','Checksum validation incomplete']]),
    error:{code:'DATA_VALIDATION_FAILED',message:'One or more tables need manual reconciliation before migration can be marked verified.',retryable:true},
  },
  {
    id:'op-2026-1035',title:'Suspend development workspace',description:'Apply safe drain and suspend selected workspace',service:'Resource Manager',action:'Suspend',resource:'ws-development',resourceType:'Workspace',scope:'organization',status:'CancelRequested',progress:58,submittedAt:'2026-11-07 21:37',updatedAt:'2026-11-07 22:05',duration:'35m',owner:'Platform Team',provider:'M8 Control Plane',workflowId:'workspace-suspend-1035',requestId:'req-2edc',correlationId:'corr-d042',attempt:1,cancellable:false,retryable:false,
    steps:makeSteps(['Validate permissions','Pause new workloads','Drain workspace resources','Confirm suspension'],2,'Running'),
    logs:makeLogs([['INFO','Workspace drain started'],['WARN','Cancellation requested; awaiting workflow confirmation']]),
  },
  {
    id:'op-2026-1034',title:'Rotate Gateway certificates',description:'Issue replacement certificate and roll out gateway configuration',service:'Gateway',action:'Rotate',resource:'edge-gateway',resourceType:'Gateway',scope:'organization',status:'Succeeded',progress:100,submittedAt:'2026-11-07 19:42',updatedAt:'2026-11-07 20:04',duration:'22m',owner:'Security Team',provider:'Envoy Gateway',workflowId:'rotate-edge-1034',requestId:'req-2b13',correlationId:'corr-d948',attempt:1,cancellable:false,retryable:false,
    steps:makeSteps(['Issue certificate','Publish Secret','Apply Gateway config','Verify health'],4,'Pending'),
    logs:makeLogs([['INFO','New certificate issued'],['INFO','Envoy Gateway rollout successful']]),result:'Certificate rotation completed',
  },
  {
    id:'op-2026-1033',title:'Provision Object Storage',description:'Create project bucket and access credentials',service:'Storage',action:'Create',resource:'ml-models-artifacts',resourceType:'Bucket',scope:'project',status:'Cancelled',progress:24,submittedAt:'2026-11-07 18:22',updatedAt:'2026-11-07 18:27',duration:'5m',owner:'ML Platform',provider:'Yandex Cloud / eu-1',workflowId:'bucket-create-1033',requestId:'req-50da',correlationId:'corr-a0c9',attempt:1,cancellable:false,retryable:false,
    steps:makeSteps(['Validate name','Check quota','Create bucket','Apply ACL'],1,'Pending'),
    logs:makeLogs([['INFO','Bucket provisioning requested'],['WARN','Cancellation confirmed; compensation completed']]),
    result:'Cancelled; resource and reservation cleaned up',
  },
  {
    id:'op-2026-1032',title:'Update project quotas',description:'Apply approved CPU limit to project AI Studio',service:'Quota Manager',action:'Update',resource:'AI Studio',resourceType:'QuotaPolicy',scope:'project',status:'Succeeded',progress:100,submittedAt:'2026-11-07 18:12',updatedAt:'2026-11-07 18:13',duration:'43s',owner:'Platform Admin',provider:'M8 YDB',workflowId:'quota-policy-update-1032',requestId:'req-3bf0',correlationId:'corr-b6c9',attempt:1,cancellable:false,retryable:false,
    steps:makeSteps(['Check parent allocation','Apply quota policy','Emit policy event'],3,'Pending'),
    logs:makeLogs([['INFO','Policy version updated'],['INFO','Quota policy event published']]),result:'Quota policy version 12 applied',
  },
]

interface DemoState {operations:LongRunningOperation[];events:string[]}
const initialState: DemoState={operations:initialOperations,events:[]}
let state=initialState
const listeners=new Set<()=>void>()
const subscribe=(listener:()=>void)=>{listeners.add(listener);return()=>{listeners.delete(listener)}}
function publish(next:DemoState){state=next;listeners.forEach(listener=>listener())}
export function useOperationsDemo(){return useSyncExternalStore(subscribe,()=>state,()=>initialState)}
export function getDemoOperation(id:string){return state.operations.find(item=>item.id===id)}
const demoTime='2026-11-07 22:12'
function log(message:string,level:OperationLog['level']='INFO'):OperationLog {return{at:'22:12:00',level,message}}
function update(id:string,change:(op:LongRunningOperation)=>LongRunningOperation):boolean{
  if(!state.operations.some(op=>op.id===id))return false
  const changed=state.operations.map(op=>op.id===id?change(op):op)
  publish({...state,operations:changed,events:[`${id} updated (demo)`,...state.events].slice(0,20)})
  return true
}
export function cancelDemoOperation(id:string):boolean{
  const op=getDemoOperation(id)
  if(!op||!op.cancellable||!(['Queued','Running','Waiting'] as OperationStatus[]).includes(op.status))return false
  return update(id,item=>({...item,status:'CancelRequested',cancellable:false,updatedAt:demoTime,logs:[...item.logs,log('Cancellation requested; awaiting workflow acknowledgement','WARN')]}))
}
export function advanceDemoOperation(id:string):boolean{
  const op=getDemoOperation(id)
  if(!op||!(['Queued','Running','Waiting','CancelRequested'] as OperationStatus[]).includes(op.status))return false
  if(op.status==='CancelRequested'){
    return update(id,item=>({...item,status:'Cancelled',cancellable:false,updatedAt:demoTime,logs:[...item.logs,log('Workflow confirmed cancellation; compensation finished')],result:'Cancelled after workflow acknowledgement'}))
  }
  const nextStatus:OperationStatus=op.progress>=80?'Succeeded':'Running'
  const nextProgress=nextStatus==='Succeeded'?100:Math.max(20,Math.min(95,op.progress+24))
  return update(id,item=>{
    const nextSteps=item.steps.map((step,index)=>{
      const currentIndex=nextStatus==='Succeeded'?item.steps.length:Math.min(item.steps.length-1,Math.floor(nextProgress/100*item.steps.length))
      return {...step,status:index<currentIndex?'Succeeded':index===currentIndex?'Running':'Pending'} as OperationStep
    })
    return {...item,status:nextStatus,progress:nextProgress,updatedAt:demoTime,cancellable:nextStatus==='Running',steps:nextSteps,logs:[...item.logs,log(nextStatus==='Succeeded'?'Operation finished successfully':`Demo progress advanced to ${nextProgress}%`)],result:nextStatus==='Succeeded'?'Demo operation completed':undefined}
  })
}
export function retryDemoOperation(id:string):string|null{
  const op=getDemoOperation(id)
  if(!op||op.status!=='Failed'||!op.retryable||!op.error?.retryable)return null
  const numeric=state.operations.map(x=>Number(x.id.split('-').pop())).filter(Number.isFinite)
  const newId=`op-2026-${Math.max(1042,...numeric)+1}`
  const next:LongRunningOperation={...op,id:newId,status:'Queued',progress:0,submittedAt:demoTime,updatedAt:demoTime,duration:'0s',attempt:op.attempt+1,parentOperationId:op.id,workflowId:`${op.workflowId}-attempt-${op.attempt+1}`,requestId:`retry-${newId}`,cancellable:true,retryable:false,steps:op.steps.map(step=>({...step,status:'Pending',duration:undefined})),logs:[log(`New retry operation created from ${op.id}; original operation remains Failed`)],error:undefined,result:undefined}
  publish({...state,operations:[next,...state.operations],events:[`Retry created ${newId} from ${id}`,...state.events].slice(0,20)})
  return newId
}
export function resetOperationsDemo(){publish({operations:initialOperations,events:[]})}
