import type {ModuleDefinition} from '@m8/module-sdk';
import {ModuleRegistry} from './module-registry';

export function defineModules<
  const TModules extends readonly ModuleDefinition[],
>(modules: TModules): ModuleRegistry<TModules> {
  return new ModuleRegistry(modules);
}
