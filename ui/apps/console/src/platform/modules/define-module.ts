import type {ModuleDefinition} from './types';
import {moduleSchema} from './module-schema';

export function defineModule<const TModule extends ModuleDefinition>(
  module: TModule,
): TModule {
  moduleSchema.parse(module);
  return module;
}
