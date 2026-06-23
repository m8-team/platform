import {defineModule} from '../src';

export const gatewayModule = defineModule({
  manifest: {
    id: 'gateway',
    title: 'Gateway',
    description: 'API gateway, routes, listeners, policies, consumers and products.',
    version: '0.1.0',
    moduleApiVersion: '1.0.0',
    kind: 'domain',
    lifecycle: 'preview',
    basePath: 'gateway',
    mountPointId: 'project.main',
    order: 400,
    requiredCapabilities: ['gateway-api'],
    requiredPermissions: ['gateway.read'],

    routes: [
      {
        id: 'gateway.overview',
        path: '/',
        title: 'Gateway Overview',
        component: () => import('./pages/GatewayOverviewPage'),
        requiredPermissions: ['gateway.read'],
        requiredScopes: ['organization', 'workspace', 'project'],
      },
      {
        id: 'gateway.api-services.list',
        path: '/api-services',
        title: 'API Services',
        component: () => import('./pages/ApiServicesPage'),
        requiredPermissions: ['gateway.api-service.read'],
        requiredScopes: ['organization', 'workspace', 'project'],
      },
    ],

    navigation: [
      {
        id: 'gateway',
        title: 'Gateway',
        to: '/',
        mountPointId: 'project.main',
        order: 400,
        requiredPermissions: ['gateway.read'],
        children: [
          {
            id: 'gateway.overview',
            title: 'Overview',
            to: '/',
            exact: true,
            order: 10,
          },
          {
            id: 'gateway.api-services',
            title: 'API Services',
            to: '/api-services',
            order: 20,
            requiredPermissions: ['gateway.api-service.read'],
          },
        ],
      },
    ],

    widgets: [
      {
        id: 'gateway.traffic-card',
        slotId: 'project.overview',
        title: 'Gateway Traffic',
        component: () => import('./widgets/GatewayTrafficCard'),
        order: 100,
        requiredPermissions: ['gateway.metrics.read'],
        requiredScopes: ['organization', 'workspace', 'project'],
      },
    ],

    queryNamespace: 'gateway',
  },
});

export const getManifest = gatewayModule.getManifest;
export const initialize = gatewayModule.initialize;
export const dispose = gatewayModule.dispose;
