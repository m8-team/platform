import type {ApiService,GatewayRoute} from './demo'

/** YAML double-quoted scalar, to avoid ambiguous values and unexpected newlines. */
const scalar=(value:string|number)=>JSON.stringify(String(value))
export function httpRouteManifest(route:GatewayRoute,service:ApiService):string{
  const lines=[
    '# DEMO ONLY: review namespaces, Gateway listeners and backend refs before use.',
    'apiVersion: gateway.networking.k8s.io/v1',
    'kind: HTTPRoute',
    'metadata:',
    `  name: ${scalar(route.name)}`,
    '  namespace: production',
    'spec:',
    '  parentRefs:',
    '    - name: edge-gateway',
    '      namespace: gateway-system',
    '  hostnames:',
    `    - ${scalar(route.hostname)}`,
    '  rules:',
    '    - matches:',
    '        - path:',
    '            type: PathPrefix',
    `            value: ${scalar(route.pathPrefix)}`,
  ]
  if(route.method!=='ANY')lines.push(`          method: ${route.method}`)
  lines.push(
    '      backendRefs:',
    `        - name: ${scalar(service.name)}`,
    `          port: ${service.port}`,
    '          weight: 1',
  )
  return lines.join('\n')+'\n'
}
export function aiRouteDescriptor(route:GatewayRoute,service:ApiService):string{
  return [
    '# M8 CONCEPTUAL RESOURCE TYPE DESCRIPTOR — NOT A DEPLOYABLE CRD.',
    '# Verify the actual Agent Router CRD schema before provisioning.',
    'resourceType: m8.ai.gateway.v1.AIGatewayRoute',
    'metadata:',
    `  name: ${scalar(route.name)}`,
    '  project: ai-studio',
    'spec:',
    `  hostname: ${scalar(route.hostname)}`,
    `  pathPrefix: ${scalar(route.pathPrefix)}`,
    `  method: ${scalar(route.method)}`,
    '  target:',
    '    type: ApiService',
    `    ref: ${scalar(service.id)}`,
    '  provider: agent-router',
    `  desiredVersion: ${route.version}`,
  ].join('\n')+'\n'
}
export function apiServiceDescriptor(service:ApiService):string{
  return [
    '# M8 CONCEPTUAL RESOURCE TYPE DESCRIPTOR — NOT A KUBERNETES SERVICE MANIFEST.',
    'resourceType: m8.gateway.v1.ApiService',
    'metadata:',
    `  name: ${scalar(service.name)}`,
    'spec:',
    `  protocol: ${scalar(service.protocol)}`,
    `  upstreamHostname: ${scalar(service.upstream)}`,
    `  upstreamPort: ${service.port}`,
    `  environment: ${scalar(service.environment)}`,
  ].join('\n')+'\n'
}
export function gatewayManifest(route:GatewayRoute,service:ApiService):string{
  return route.kind==='HTTPRoute'?httpRouteManifest(route,service):aiRouteDescriptor(route,service)
}
