import {defineModules} from './define-modules';
import {resourceManagerModule} from '@/modules/resource-manager/module';

export const moduleRegistry = defineModules([
  resourceManagerModule,
] as const);
