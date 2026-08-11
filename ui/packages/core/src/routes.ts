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

export function canonicalizeRoute(path: string): string {
  return normalizePath(path).split('/').map(segment => {
    if (/^\[\[\.\.\.[^\]]+\]\]$/.test(segment)) return '[[...]]';
    if (/^\[\.\.\.[^\]]+\]$/.test(segment)) return '[...]';
    if (/^\[[^\]]+\]$/.test(segment)) return '[]';
    return segment;
  }).join('/') || '/';
}
