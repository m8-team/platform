import type {ModuleDefinition} from '@m8/core';

export interface ModuleEnablementOptions {
  readonly enabledModuleIds?: readonly string[];
}

export function isModuleEnabled(
  module: ModuleDefinition,
  options: ModuleEnablementOptions,
): boolean {
  const {enabledModuleIds} = options;
  if (enabledModuleIds && !enabledModuleIds.includes(module.id)) return false;
  return true;
}

export function selectEnabledModules(
  modules: readonly ModuleDefinition[],
  options: ModuleEnablementOptions,
): readonly ModuleDefinition[] {
  const candidates = modules.filter(module => isModuleEnabled(module, options));
  const enabledIds = new Set(candidates.map(module => module.id));
  let changed = true;
  while (changed) {
    changed = false;
    for (const module of candidates) {
      if (enabledIds.has(module.id) &&
          (module.dependencies?.required ?? []).some(dependency => !enabledIds.has(dependency))) {
        enabledIds.delete(module.id);
        changed = true;
      }
    }
  }
  return candidates.filter(module => enabledIds.has(module.id));
}
