import {moduleManifestSchema} from './module-schema';
import type {M8ModuleDefinition} from './types';

export function defineModule<const TModule extends M8ModuleDefinition>(
  module: TModule,
): TModule {
  moduleManifestSchema.parse(module);
  const frozen = {
    ...module,
    dependencies: module.dependencies ? Object.freeze({
      ...module.dependencies,
      required: Object.freeze([...(module.dependencies.required ?? [])]),
      optional: Object.freeze([...(module.dependencies.optional ?? [])]),
    }) : undefined,
    routes: module.routes ? Object.freeze({...module.routes}) : undefined,
    queries: Object.freeze([...(module.queries ?? [])]),
    operations: Object.freeze([...(module.operations ?? [])]),
  };
  return Object.freeze(frozen) as unknown as TModule;
}
