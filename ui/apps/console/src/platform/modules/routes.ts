export function normalizePath(path: string): string {
  if (!path) {
    return '/';
  }

  let normalized = path.replace(/\/+/g, '/').replace(/\/$/, '');

  if (!normalized.startsWith('/')) {
    normalized = `/${normalized}`;
  }

  return normalized || '/';
}

export function joinRoute(basePath: string, routePath: string): string {
  const base = normalizePath(basePath);
  const route = normalizePath(routePath);

  if (route === '/') {
    return base;
  }

  if (base === '/') {
    return route;
  }

  return normalizePath(`${base}/${route.slice(1)}`);
}
