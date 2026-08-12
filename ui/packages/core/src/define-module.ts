import {moduleDefinitionSchema} from './module-schema';
import type {ModuleDefinition} from './types';

export function defineModule<const TModule extends ModuleDefinition>(
  module: TModule & Record<Exclude<keyof TModule, keyof ModuleDefinition>, never>,
): TModule {
  moduleDefinitionSchema.parse(module);
  const frozen = {
    ...module,
    routes: module.routes ? Object.freeze({...module.routes}) : undefined,
    queries: Object.freeze([...(module.queries ?? [])]),
    operations: Object.freeze([...(module.operations ?? [])]),
  };
  return Object.freeze(frozen) as TModule;
}
