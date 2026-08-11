import type {M8PlatformRegistrySnapshot} from '../registry/PlatformRegistry';
import {findMountPoint} from '../registry/PlatformRegistry';
import {interpolatePathTemplate, joinPaths} from '../utils/path';
import type {M8BuildPathInput, M8RouterRuntime, M8ScopeRuntime} from './ModuleRuntimeContext';

export function createRouterRuntime(input: {
  registry: M8PlatformRegistrySnapshot;
  scopes: M8ScopeRuntime;
  navigate: M8RouterRuntime['navigate'];
}): M8RouterRuntime {
  return {
    navigate: input.navigate,
    buildPath: (buildInput) => buildPathFromRegistry(input.registry, input.scopes, buildInput),
  };
}

export function buildPathFromRegistry(
  registry: M8PlatformRegistrySnapshot,
  scopes: M8ScopeRuntime,
  input: M8BuildPathInput,
): string {
  const mountPoint = findMountPoint(registry, input.mountPointId);

  if (!mountPoint) {
    throw new Error(`Unknown mount point: ${input.mountPointId}`);
  }

  const params = {
    ...scopes.current,
    ...input.params,
  };

  const mountedPath = interpolatePathTemplate(mountPoint.pathTemplate, params);

  return joinPaths(mountedPath, input.moduleBasePath, input.relativePath);
}
