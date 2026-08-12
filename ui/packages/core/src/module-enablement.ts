import type {ModuleDefinition} from './types';

export interface ModuleEnablementOptions {
  readonly enabledModuleIds?: readonly string[];
}

export function isModuleEnabled(
  moduleDefinition: ModuleDefinition,
  options: ModuleEnablementOptions = {},
): boolean {
  return !options.enabledModuleIds ||
    options.enabledModuleIds.includes(moduleDefinition.id);
}

export function selectEnabledModules<
  const TModules extends readonly ModuleDefinition[],
>(
  installedModules: TModules,
  options: ModuleEnablementOptions = {},
): readonly TModules[number][] {
  const candidates = installedModules.filter(moduleDefinition =>
    isModuleEnabled(moduleDefinition, options));
  const enabledIds = new Set(candidates.map(moduleDefinition => moduleDefinition.id));

  let changed = true;
  while (changed) {
    changed = false;
    for (const moduleDefinition of candidates) {
      const hasMissingDependency = (
        moduleDefinition.dependencies?.required ?? []
      ).some(dependencyId => !enabledIds.has(dependencyId));
      if (enabledIds.has(moduleDefinition.id) && hasMissingDependency) {
        enabledIds.delete(moduleDefinition.id);
        changed = true;
      }
    }
  }

  return candidates.filter(moduleDefinition => enabledIds.has(moduleDefinition.id));
}
