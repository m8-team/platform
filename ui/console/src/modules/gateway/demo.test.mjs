import assert from 'node:assert/strict'
import {afterEach,test} from 'node:test'
import {
  resetGatewayDemo,saveDemoService,saveDemoRoute,saveDemoConsumer,toggleDemoConsumer,
  saveDemoPolicy,toggleDemoPolicy,saveDemoLimit,simulateGatewayRequest,initialGateway,
} from './demo.ts'
import {gatewayManifest,apiServiceDescriptor} from './manifest.ts'

afterEach(()=>resetGatewayDemo())

test('service creation checks upstream and rejects duplicate names',()=>{
  assert.equal(saveDemoService({name:'orders-api',protocol:'HTTP',upstream:'other.example',port:8080,environment:'Production',description:'Duplicate'}),'A service with this name already exists')
  assert.equal(saveDemoService({name:'new-api',protocol:'HTTP',upstream:'new-api.production.svc.cluster.local',port:8081,environment:'Development',description:'Demo'}),null)
  assert.equal(saveDemoService({name:'invalid',protocol:'HTTP',upstream:'https://example.com',port:8080,environment:'Production',description:''}),'Enter a DNS-style upstream hostname (without URL scheme)')
})

test('route validation rejects duplicate host/path/method and invalid backends',()=>{
  assert.match(saveDemoRoute({name:'duplicate-route',kind:'HTTPRoute',hostname:'api.acme.example',pathPrefix:'/v1/orders',method:'GET',serviceId:'svc-orders',description:'Duplicate'})??'',/already claims/)
  assert.equal(saveDemoRoute({name:'new-thing',kind:'HTTPRoute',hostname:'api.acme.example',pathPrefix:'/v2/orders',method:'GET',serviceId:'missing',description:'Broken'}),'Select a valid backend service')
  assert.equal(saveDemoRoute({name:'new-thing',kind:'HTTPRoute',hostname:'api.acme.example',pathPrefix:'/v2/orders',method:'GET',serviceId:'svc-orders',description:'Valid'}),null)
})

test('route simulator uses segment-aware longest prefix match and detects pending routes',()=>{
  const success=simulateGatewayRequest({hostname:'ai.acme.example',path:'/v1/chat/completions',method:'POST',consumerId:'consumer-ai'})
  assert.equal(success.status,200)
  assert.equal(success.route?.id,'route-chat')
  assert.ok(success.appliedPolicies.some(policy=>policy.id==='policy-api-key'))
  assert.equal(simulateGatewayRequest({hostname:'ai.acme.example',path:'/v1/embeddings',method:'POST',consumerId:'consumer-ai'}).status,503)
  assert.equal(simulateGatewayRequest({hostname:'api.acme.example',path:'/v1/orders-v2',method:'GET',consumerId:'consumer-web'}).status,404)
  assert.equal(simulateGatewayRequest({hostname:'api.acme.example',path:'/v1/orders',method:'GET',consumerId:'consumer-partner'}).status,403)
})

test('consumer suspension is reversible and API key material is never issued',()=>{
  assert.equal(saveDemoConsumer({name:'test-client',application:'Sandbox',auth:'API Key',description:'Metadata'}),null)
  assert.equal(toggleDemoConsumer('consumer-ai'),true)
  assert.equal(simulateGatewayRequest({hostname:'ai.acme.example',path:'/v1/chat/completions',method:'POST',consumerId:'consumer-ai'}).status,403)
})

test('policy updates are isolated demo metadata',()=>{
  assert.equal(saveDemoPolicy({name:'Require JWT',kind:'JWT',targetType:'Route',targetId:'route-orders',mode:'Enforce',description:'Demo'}),null)
  assert.equal(toggleDemoPolicy('policy-oidc'),true)
  assert.equal(saveDemoPolicy({name:'Broken target',kind:'WAF',targetType:'Route',targetId:'not-real',mode:'Monitor',description:''}),'Route target not found')
})

test('rate limits validate quantities and target existence',()=>{
  assert.equal(saveDemoLimit({name:'Bad',targetType:'Route',targetId:'route-chat',requests:0,period:'minute',burst:10,mode:'Enforce'}),'Requests must be a positive integer')
  assert.equal(saveDemoLimit({name:'Bad target',targetType:'Consumer',targetId:'invalid',requests:100,period:'minute',burst:10,mode:'Enforce'}),'Select a valid target')
  assert.equal(saveDemoLimit({name:'New limit',targetType:'Consumer',targetId:'consumer-ai',requests:100,period:'minute',burst:10,mode:'Monitor'}),null)
})

test('HTTPRoute YAML is Gateway API and AI descriptor is explicitly non-deployable',()=>{
  const service=initialGateway.services.find(s=>s.id==='svc-orders')
  const route=initialGateway.routes.find(r=>r.id==='route-orders')
  assert.ok(route&&service)
  assert.match(gatewayManifest(route,service),/apiVersion: gateway.networking.k8s.io\/v1/)
  assert.match(gatewayManifest(route,service),/kind: HTTPRoute/)
  assert.match(gatewayManifest(route,service),/backendRefs:/)
  const aiRoute=initialGateway.routes.find(r=>r.id==='route-chat')
  const aiService=initialGateway.services.find(s=>s.id==='svc-agent-router')
  assert.ok(aiRoute&&aiService)
  assert.match(gatewayManifest(aiRoute,aiService),/NOT A DEPLOYABLE CRD/)
  assert.match(apiServiceDescriptor(service),/NOT A KUBERNETES SERVICE MANIFEST/)
})
