import {definePlatformRegistry} from '../src';

export const platformRegistry = definePlatformRegistry({
  uiApiVersion: '1.0.0',
  revision: 'dev',

  scopes: [
    {
      id: 'global',
      title: 'Global',
      level: 0,
    },
    {
      id: 'organization',
      title: 'Organization',
      level: 10,
      paramName: 'orgId',
      parentScopeId: 'global',
    },
    {
      id: 'workspace',
      title: 'Workspace',
      level: 20,
      paramName: 'workspaceId',
      parentScopeId: 'organization',
    },
    {
      id: 'project',
      title: 'Project',
      level: 30,
      paramName: 'projectId',
      parentScopeId: 'workspace',
    },
  ],

  mountPoints: [
    {
      id: 'global.main',
      scopeId: 'global',
      pathTemplate: '/',
    },
    {
      id: 'organization.main',
      scopeId: 'organization',
      pathTemplate: '/o/:organization',
    },
    {
      id: 'workspace.main',
      scopeId: 'workspace',
      pathTemplate: '/o/:organization/w/:workspace',
    },
    {
      id: 'project.main',
      scopeId: 'project',
      pathTemplate: '/o/:organization/w/:workspace/p/:project',
    },
  ],

  slots: [
    {
      id: 'platform.overview',
      title: 'Platform Overview',
      scopeId: 'global',
    },
    {
      id: 'project.overview',
      title: 'Project Overview',
      scopeId: 'project',
    },
  ],

  modules: [
    {
      id: 'gateway',
      enabled: true,
      title: 'Gateway',
      version: '0.1.0',
      moduleApiVersion: '1.0.0',
    },
  ],
});
