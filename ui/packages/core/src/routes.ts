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

const dynamicSegment = /^\[[^\[\]/]+\]$/;
const catchAllSegment = /^\[\.\.\.[^\[\]/]+\]$/;
const optionalCatchAllSegment = /^\[\[\.\.\.[^\[\]/]+\]\]$/;

export function isValidRoute(path: string): boolean {
  if (!path.startsWith('/') || /[?#\\]/.test(path)) return false;

  const segments = normalizePath(path).split('/').slice(1);
  return segments.every((segment, index) => {
    if (!segment.includes('[') && !segment.includes(']')) return true;
    const catchAll = catchAllSegment.test(segment) ||
      optionalCatchAllSegment.test(segment);
    if (catchAll) return index === segments.length - 1;
    return dynamicSegment.test(segment);
  });
}

export function canonicalizeRoute(path: string): string {
  return normalizePath(path).split('/').map(segment => {
    if (optionalCatchAllSegment.test(segment)) return '[[...]]';
    if (catchAllSegment.test(segment)) return '[...]';
    if (dynamicSegment.test(segment)) return '[]';
    return segment;
  }).join('/') || '/';
}
