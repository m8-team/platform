import type {NextRouteSpec} from '@json-render/next';
import type {M8RouteSpec} from './types';

/**
 * Remove M8 platform metadata before passing routes to @json-render/next.
 * json-render/next receives only its native NextRouteSpec fields.
 */
export function toNextRouteSpec(route: M8RouteSpec): NextRouteSpec {
  const {
    navigation: _navigation,
    access: _access,
    queries: _queries,
    ...nextRoute
  } = route;

  return nextRoute;
}
