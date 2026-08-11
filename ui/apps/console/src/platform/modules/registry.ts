import {defineModules} from '@m8/core';
import {resourceManagerModule} from '@m8/resource-manager';

export const moduleRegistry = defineModules([
  resourceManagerModule,
] as const);
