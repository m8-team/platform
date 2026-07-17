import type {M8ModuleManifest} from '../manifest/ModuleManifest';
import type {M8ModuleRuntimeContext} from '../runtime/ModuleRuntimeContext';

export function isModuleVisible(
  manifest: M8ModuleManifest,
  runtime: M8ModuleRuntimeContext,
): boolean {
  if (manifest.requiredPermissions?.length && !runtime.permissions.hasAll(manifest.requiredPermissions)) {
    return false;
  }

  if (manifest.requiredFeatureFlags?.length) {
    return manifest.requiredFeatureFlags.every((flag) => runtime.featureFlags.enabled(flag));
  }

  return true;
}
