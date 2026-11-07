import assert from 'node:assert/strict'
import {afterEach,test} from 'node:test'
import {
  initialSettings,resetSettingsDemo,saveProject,validateProject,
  validateModuleToggle,toggleModule,
  validateHttpsEndpoint,saveIntegration,testIntegration,toggleIntegration,
  validateWebhook,saveWebhook,toggleWebhook,retryDelivery,
  validateTokenRequest,requestDemoToken,revokeDemoToken,
} from './demo.ts'

afterEach(()=>resetSettingsDemo())

test('Project Settings validate identifiers and update only demo metadata',()=>{
  const original=JSON.stringify(initialSettings.project)
  const good={
    displayName:'AI Studio Production',slug:'ai-studio',description:'Updated locally',
    environment:'production',region:'eu-1',defaultNamespace:'ai-studio',
  }
  assert.equal(validateProject(good),null)
  assert.match(validateProject({...good,slug:'Upper_Case'})??'',/Slug/)
  assert.match(validateProject({...good,defaultNamespace:'Bad.Namespace'})??'',/namespace/)
  assert.equal(saveProject(good),null)
  assert.equal(JSON.stringify(initialSettings.project),original)
  assert.equal(saveProject({...good,region:'invalid'}),'Invalid region')
})

test('Module Registry enforces required core and active dependencies',()=>{
  assert.match(validateModuleToggle('resource-manager',false)??'',/required/)
  assert.match(validateModuleToggle('gateway',false)??'',/required by/)
  assert.equal(toggleModule('observability'),null)
  assert.equal(toggleModule('observability'),null)
  assert.match(validateModuleToggle('billing',true)??'',/Enterprise license/)
  assert.match(validateModuleToggle('missing-module',true)??'',/not found/)
})

test('HTTPS URL validation rejects local, private and credential-bearing endpoints',()=>{
  assert.equal(validateHttpsEndpoint('https://hooks.example.com/notify'),null)
  for(const url of [
    'http://hooks.example.com/notify',
    'https://localhost/hook',
    'https://127.0.0.1/hook',
    'https://10.3.4.5/hook',
    'https://192.168.1.23/hook',
    'https://172.18.0.3/hook',
    'https://169.254.169.254/latest/meta-data/',
    'https://internal.local/hook',
    'https://user:password@hooks.example.com/hook',
  ])assert.ok(validateHttpsEndpoint(url),url)
  assert.equal(validateHttpsEndpoint('not a URL'),'Invalid HTTPS URL')
})

test('Integration creates metadata, simulates connection check without claiming success',()=>{
  const item={
    name:'Example ClickHouse',kind:'ClickHouse',endpoint:'https://clickhouse.example.com',
    secretRef:'integrations/clickhouse',environment:'production',
  }
  assert.equal(saveIntegration(item),null)
  assert.equal(testIntegration('int-demo-1'),true)
  assert.equal(toggleIntegration('int-demo-1'),true)
  assert.equal(testIntegration('not-found'),false)
  assert.match(saveIntegration({...item,endpoint:'http://localhost:8123'})??'',/HTTPS/)
})

test('Webhooks require topics and a signing reference before enabling',()=>{
  assert.match(validateWebhook({name:'Webhook',url:'https://hooks.example.com',events:[],secretRef:''})??'',/one or more/)
  assert.equal(saveWebhook({name:'Test outbound',url:'https://hooks.example.com/notifications',events:['resource.created'],secretRef:''}),null)
  assert.equal(toggleWebhook('hook-demo-1'),false)
  assert.equal(saveWebhook({name:'Test outbound',url:'https://hooks.example.com/notifications',events:['resource.created'],secretRef:'hooks/my-key'},'hook-demo-1'),null)
  assert.equal(toggleWebhook('hook-demo-1'),true)
})

test('Retry creates a new simulated delivery without altering failed original',()=>{
  const id=retryDelivery('delivery-003')
  assert.equal(id,'delivery-demo-1')
  assert.equal(retryDelivery('delivery-003'),'delivery-demo-2')
  assert.equal(retryDelivery('delivery-004'),null)
  assert.equal(retryDelivery('not-found'),null)
})

test('Token requests store Draft metadata only, never usable token secrets',()=>{
  const request={name:'ci-readonly',scopes:['resources:read','metrics:read'],expiryDays:30}
  assert.equal(validateTokenRequest(request),null)
  assert.equal(requestDemoToken(request),null)
  assert.match(validateTokenRequest(request)??'',/already in use/)
  assert.ok(initialSettings.tokens.every(t=>!('secret' in t)))
  assert.equal(revokeDemoToken('tok-demo-1'),true)
  assert.equal(revokeDemoToken('tok-demo-1'),false)
  assert.equal(requestDemoToken({...request,name:'missing-scopes',scopes:[]}), 'Choose valid scopes')
  assert.equal(requestDemoToken({...request,name:'invalid-expiry',expiryDays:365}), 'Choose 7, 30 or 90 days')
})

test('Reset keeps immutable reference data and restores initial demo state',()=>{
  const baseline=JSON.stringify(initialSettings)
  toggleModule('observability')
  saveProject({displayName:'New project',slug:'new-project',description:'Updated',region:'eu-2',environment:'staging',defaultNamespace:'new-project'})
  requestDemoToken({name:'temp-reader',scopes:['resources:read'],expiryDays:7})
  resetSettingsDemo()
  assert.equal(JSON.stringify(initialSettings),baseline)
  assert.equal(validateModuleToggle('observability',false),null)
  assert.equal(validateTokenRequest({name:'temp-reader',scopes:['resources:read'],expiryDays:7}),null)
})
