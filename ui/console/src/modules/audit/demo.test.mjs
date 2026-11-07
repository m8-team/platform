import assert from 'node:assert/strict'
import {afterEach,test} from 'node:test'
import {
  initialAuditEvents,listAuditEvents,redactAuditData,serializeAuditExport,
  createAuditExport,prepareAuditExport,getExportContent,getAuditEvent,
  resetAuditDemo,setExportDraft,
} from './demo.ts'

afterEach(()=>resetAuditDemo())

test('filters match actor, category, outcome and UTC date range',()=>{
  const result=listAuditEvents({category:'Security',outcome:'Denied',fromDate:'2026-11-07',toDate:'2026-11-07'})
  assert.equal(result.length,2)
  assert.ok(result.every(item=>item.category==='Security'&&item.outcome==='Denied'))
  assert.equal(listAuditEvents({query:'op-2026-1042'}).length,1)
  assert.equal(listAuditEvents({query:'nonexistent-id'}).length,0)
  assert.equal(listAuditEvents({fromDate:'2026-11-08'}).length,0)
  assert.equal(listAuditEvents({fromDate:'2026-11-08',toDate:'2026-11-07'}).length,0)
})

test('audit redaction removes sensitive nested values without modifying input',()=>{
  const source={password:'raw-password',nested:{apiKey:'raw-key',token:'raw-token',limit:32,credentials:['x']},array:[{privateKey:'secret'}]}
  const safe=redactAuditData(source)
  assert.equal(safe.password,'[REDACTED]')
  assert.equal(safe.nested.apiKey,'[REDACTED]')
  assert.equal(safe.nested.limit,32)
  assert.equal(safe.array[0].privateKey,'[REDACTED]')
  assert.equal(source.nested.token,'raw-token')
})

test('CSV output escapes quotes and protects spreadsheet formulas',()=>{
  const example={
    ...initialAuditEvents[0],id:'evt-csv',resourceName:'=HYPERLINK("https://attacker.invalid")',
    metadata:{apiKey:'raw-secret',notes:'safe'},
  }
  const csv=serializeAuditExport([example],'CSV')
  assert.ok(csv.includes('"\'=HYPERLINK(""https://attacker.invalid"")"'))
  assert.ok(csv.includes('[REDACTED]'))
  assert.ok(!csv.includes('raw-secret'))
  assert.ok(!csv.includes('https://attacker.invalid")"'))
})

test('JSON Lines preserves useful event fields but removes sensitive payloads',()=>{
  const output=serializeAuditExport(initialAuditEvents.filter(item=>item.id==='evt-20261107-018'),'JSONL')
  const record=JSON.parse(output.trim())
  assert.equal(record.action,'gateway.route.created')
  assert.equal(record.changes.after.apiKey,'[REDACTED]')
  assert.equal(record.operationId,'op-2026-1041')
})

test('export request snapshots selection; preparation is audited append-only',()=>{
  const original=initialAuditEvents.map(item=>JSON.stringify(item))
  const record=createAuditExport({category:'Gateway'},'JSONL')
  assert.equal(record.status,'Queued')
  assert.equal(record.eventIds.length,4)
  assert.equal(getExportContent(record.id),null)
  assert.equal(prepareAuditExport(record.id),true)
  assert.equal(prepareAuditExport(record.id),false)
  const data=getExportContent(record.id)
  assert.ok(data)
  const lines=data.trim().split('\n').map(line=>JSON.parse(line))
  assert.equal(lines.length,4)
  assert.ok(lines.every(item=>item.category==='Gateway'))
  assert.ok(getAuditEvent(`evt-20261107-export-${record.id}`))
  assert.ok(getAuditEvent(`evt-20261107-ready-${record.id}`))
  assert.deepEqual(initialAuditEvents.map(item=>JSON.stringify(item)),original)
})

test('new export events are not included in preceding snapshot',()=>{
  const record=createAuditExport({},'CSV')
  assert.equal(record.eventIds.length,initialAuditEvents.length)
  prepareAuditExport(record.id)
  const output=getExportContent(record.id)
  assert.ok(output)
  assert.ok(!output.includes('audit.export.requested'))
  assert.ok(!output.includes('audit.export.ready'))
})

test('reset clears export jobs and retains initial immutable demo history',()=>{
  const record=createAuditExport({},'CSV')
  setExportDraft({outcome:'Success'})
  resetAuditDemo()
  assert.equal(getExportContent(record.id),null)
  assert.equal(listAuditEvents().length,initialAuditEvents.length)
})
