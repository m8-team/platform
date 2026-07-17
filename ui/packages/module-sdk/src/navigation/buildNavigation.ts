import type {M8ModuleManifest} from '../manifest/ModuleManifest';
import type {M8NavigationContribution} from './NavContribution';
import type {M8ModuleRuntimeContext} from '../runtime/ModuleRuntimeContext';
import {isModuleVisible} from '../validation/isModuleVisible';

export function buildNavigation(
  manifests: M8ModuleManifest[],
  runtime: M8ModuleRuntimeContext,
): M8NavigationContribution[] {
  return manifests
    .filter((manifest) => isModuleVisible(manifest, runtime))
    .flatMap((manifest) => filterNavigationItems(manifest.navigation, runtime))
    .sort(sortByOrder);
}

export function filterNavigationItems(
  items: M8NavigationContribution[],
  runtime: M8ModuleRuntimeContext,
): M8NavigationContribution[] {
  return items
    .filter((item) => isNavigationItemVisible(item, runtime))
    .map((item) => {
      const children = item.children ? filterNavigationItems(item.children, runtime).sort(sortByOrder) : undefined;

      if (!children) {
        return item;
      }

      return {
        ...item,
        children,
      };
    })
    .sort(sortByOrder);
}

export function isNavigationItemVisible(
  item: M8NavigationContribution,
  runtime: M8ModuleRuntimeContext,
): boolean {
  if (item.requiredPermissions?.length && !runtime.permissions.hasAll(item.requiredPermissions)) {
    return false;
  }

  if (item.requiredFeatureFlags?.length) {
    return item.requiredFeatureFlags.every((flag) => runtime.featureFlags.enabled(flag));
  }

  return true;
}

function sortByOrder<TItem extends {order?: number; title?: string}>(a: TItem, b: TItem): number {
  const orderDiff = (a.order ?? 0) - (b.order ?? 0);

  if (orderDiff !== 0) {
    return orderDiff;
  }

  return (a.title ?? '').localeCompare(b.title ?? '');
}
