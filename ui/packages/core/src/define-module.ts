import {moduleSchema} from './module-schema';
import type {M8ModuleDefinition} from './types';

export function defineModule<const TModule extends M8ModuleDefinition>(
  module: TModule,
): TModule {
  moduleSchema.parse(module);
  return module;
}
