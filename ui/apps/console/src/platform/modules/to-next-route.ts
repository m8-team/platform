import type {NextRouteSpec} from '@json-render/next';
import type {M8RouteSpec} from '@m8/module-sdk';

/**
 * Remove M8 platform metadata before passing routes to @json-render/next.
 * json-render/next receives only its native NextRouteSpec fields.
 */
export function toNextRouteSpec(route: M8RouteSpec): NextRouteSpec {
  const nextRoute = {...route};

  delete nextRoute.navigation;
  delete nextRoute.access;
  delete nextRoute.queries;

  return nextRoute;
}
