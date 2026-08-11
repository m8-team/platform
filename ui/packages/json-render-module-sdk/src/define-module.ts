import {moduleSchema} from './module-schema';
import type {ModuleDefinition} from './types';

export function defineModule<const TModule extends ModuleDefinition>(
  module: TModule,
): TModule {
  moduleSchema.parse(module);
  return module;
}
