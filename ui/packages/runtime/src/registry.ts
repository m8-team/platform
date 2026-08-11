import type {ComponentRegistry} from '@json-render/react';

export function composeComponentRegistries(
  core: ComponentRegistry,
  extensions: readonly ComponentRegistry[] = [],
): ComponentRegistry {
  const registry: ComponentRegistry = {...core};
  for (const extension of extensions) {
    for (const [name, component] of Object.entries(extension)) {
      if (registry[name]) throw new Error(`Duplicate component renderer: "${name}".`);
      registry[name] = component;
    }
  }
  return registry;
}
