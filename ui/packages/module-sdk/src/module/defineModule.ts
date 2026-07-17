import type {M8MaybePromise} from '../primitives';
import type {M8ModuleManifest} from '../manifest/ModuleManifest';
import type {M8ModuleRuntimeContext} from '../runtime/ModuleRuntimeContext';
import type {M8RemoteModule} from './RemoteModule';

export type M8ModuleDefinition<TManifest extends M8ModuleManifest = M8ModuleManifest> = {
  manifest: TManifest;
  initialize?: (ctx: M8ModuleRuntimeContext) => M8MaybePromise<void>;
  dispose?: () => M8MaybePromise<void>;
};

export function defineModule<TManifest extends M8ModuleManifest>(
  module: M8ModuleDefinition<TManifest>,
): M8ModuleDefinition<TManifest> & M8RemoteModule {
  return {
    ...module,
    getManifest: () => module.manifest,
  };
}
