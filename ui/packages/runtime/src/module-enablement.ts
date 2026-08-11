import type {M8ModuleDefinition, M8RuntimeContext} from '@m8/core';

export interface ModuleEnablementOptions {
  readonly context: M8RuntimeContext;
  readonly enabledModuleIds?: readonly string[];
}

export function isModuleEnabled(
  module: M8ModuleDefinition,
  options: ModuleEnablementOptions,
): boolean {
  const {context, enabledModuleIds} = options;
  if (enabledModuleIds && !enabledModuleIds.includes(module.id)) return false;
  if (module.access?.permission && !context.permissions?.includes(module.access.permission)) return false;
  if (module.access?.feature && !context.features?.includes(module.access.feature)) return false;
  if (module.access?.editions?.length &&
      (!context.edition || !module.access.editions.includes(context.edition))) return false;
  return true;
}

export function selectEnabledModules(
  modules: readonly M8ModuleDefinition[],
  options: ModuleEnablementOptions,
): readonly M8ModuleDefinition[] {
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
