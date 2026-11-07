import assert from 'node:assert/strict'
import {afterEach,test} from 'node:test'
import {
  initialObservability,initialTraces,initialSlos,serviceNames,
  currentMetric,formatMetric,metricSeries,searchLogs,searchTraces,validateTrace,
  validateRule,createDemoRule,toggleDemoRule,changeDemoIncident,
  sloBudget,validateSlo,resetObservabilityDemo,
} from './demo.ts'

afterEach(()=>resetObservabilityDemo())

test('metrics expose coherent time series for all six services',()=>{
  assert.equal(serviceNames.length,6)
  for(const service of serviceNames){
    for(const metric of ['rps','errorRate','latencyP95','cpu']){
      const values=metricSeries(service,metric)
      assert.equal(values.length,24)
      assert.ok(values.every(v=>Number.isFinite(v)&&v>=0))
      assert.equal(values.at(-1),currentMetric(service,metric))
    }
  }
  assert.ok(currentMetric('ai-gateway','latencyP95')>1200)
  assert.match(formatMetric(42.123,'errorRate'),/^42\.12 %$/)
})

test('log search combines service, severity and correlation text',()=>{
  const rows=searchLogs({service:'ai-gateway',level:'WARN',query:'trace-3f29'})
  assert.ok(rows.length>=1)
  assert.ok(rows.every(row=>row.level==='WARN'&&row.service==='ai-gateway'&&row.traceId==='trace-3f29'))
  assert.ok(searchLogs({query:'op-2026-1042'}).length>=1)
  assert.equal(searchLogs({query:'does-not-exist'}).length,0)
})

test('trace search matches descendant service, status and operation IDs',()=>{
  const rows=searchTraces({service:'identity',status:'OK',query:'orders'})
  assert.equal(rows.length,1)
  assert.equal(rows[0].id,'trace-4a10')
  assert.ok(searchTraces({query:'op-2026-1041'}).length>=1)
})

test('all demo trace waterfalls have valid span identities and bounds',()=>{
  for(const trace of initialTraces){
    assert.equal(validateTrace(trace),null,trace.id)
  }
  const invalid={...initialTraces[0],spans:initialTraces[0].spans.map((s,i)=>i===1?{...s,parentId:'no-span'}:s)}
  assert.match(validateTrace(invalid)??'',/Missing parent/)
})

test('alert rule validation rejects invalid values and creates only demo metadata',()=>{
  const base={name:'Custom high CPU',service:'identity',metric:'cpu',comparator:'gt',threshold:85,durationMinutes:5,severity:'Warning',description:'Synthetic alert'}
  assert.equal(validateRule(base),null)
  assert.match(validateRule({...base,threshold:105})??'',/100%/)
  assert.match(validateRule({...base,durationMinutes:0})??'',/1–1440/)
  assert.equal(createDemoRule(base),null)
  assert.equal(toggleDemoRule('rule-demo-1'),true)
  assert.equal(toggleDemoRule('unknown'),false)
})

test('acknowledge, mute and resolve enforce incident lifecycle',()=>{
  assert.equal(changeDemoIncident('inc-301','acknowledge'),true)
  assert.equal(changeDemoIncident('inc-301','acknowledge'),false)
  assert.equal(changeDemoIncident('inc-301','resolve'),true)
  assert.equal(changeDemoIncident('inc-301','mute'),false)
  assert.equal(changeDemoIncident('inc-300','mute'),true)
  assert.equal(changeDemoIncident('inc-300','acknowledge'),false)
  assert.equal(changeDemoIncident('not-found','resolve'),false)
})

test('SLO error budgets are derived from good/total and detect exhaustion',()=>{
  for(const slo of initialSlos)assert.equal(validateSlo(slo),null)
  const gateway=initialSlos.find(s=>s.id==='slo-gateway-availability')
  assert.ok(gateway)
  assert.equal(sloBudget(gateway).allowedBad,500)
  assert.equal(sloBudget(gateway).bad,300)
  assert.equal(sloBudget(gateway).remaining,200)
  const ai=initialSlos.find(s=>s.id==='slo-ai-availability')
  assert.ok(ai)
  assert.equal(sloBudget(ai).remaining,-1800)
  assert.equal(validateSlo({...ai,good:ai.total+1}),'Invalid good/total counts')
})

test('reset preserves initial reference data and no demo values are permanently changed',()=>{
  const start=JSON.stringify(initialObservability)
  changeDemoIncident('inc-301','resolve')
  createDemoRule({name:'Test rule',service:'ai-gateway',metric:'cpu',comparator:'gt',threshold:90,durationMinutes:5,severity:'Info',description:''})
  resetObservabilityDemo()
  assert.equal(JSON.stringify(initialObservability),start)
})
