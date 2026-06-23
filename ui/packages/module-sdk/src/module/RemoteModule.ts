import type {M8MaybePromise} from '../primitives';
import type {M8ModuleManifest} from '../manifest/ModuleManifest';
import type {M8ModuleRuntimeContext} from '../runtime/ModuleRuntimeContext';

export type M8RemoteModule = {
  getManifest: () => M8MaybePromise<M8ModuleManifest>;
  initialize?: (ctx: M8ModuleRuntimeContext) => M8MaybePromise<void>;
  dispose?: () => M8MaybePromise<void>;
};

export function defineRemoteModule<TRemoteModule extends M8RemoteModule>(
  remoteModule: TRemoteModule,
): TRemoteModule {
  return remoteModule;
}
