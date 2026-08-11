import {defineModule} from '@/platform/modules/define-module';

export const resourceManagerModule = defineModule({
  id: 'resource-manager',
  title: 'Resource Manager',
  icon: 'FolderTree',
  basePath: '/resource-manager',
  order: 10,
  access: {
    permission: 'resource-manager.read',
    feature: 'resource-manager',
  },
  routes: {},
  queries: [],
  operations: [],
});
