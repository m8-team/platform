import {defineModules, selectEnabledModules} from '@m8/core';
import {resourceManagerModule} from '@m8/resource-manager';

export const installedModules = [
  resourceManagerModule,
] as const;

export const enabledModules = selectEnabledModules(installedModules);

export const moduleRegistry = defineModules(enabledModules);
