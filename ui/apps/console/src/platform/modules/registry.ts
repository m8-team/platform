import {defineModules} from './define-modules';
import {resourceManagerModule} from '@m8/resource-manager-module';

export const moduleRegistry = defineModules([
  resourceManagerModule,
] as const);
