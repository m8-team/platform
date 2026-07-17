export function normalizePath(path: string | undefined | null): string {
  if (!path) {
    return '';
  }

  const trimmed = path.trim();
  if (!trimmed || trimmed === '/') {
    return '';
  }

  return trimmed.startsWith('/') ? trimmed : `/${trimmed}`;
}

export function joinPaths(...parts: Array<string | undefined | null>): string {
  const result = parts
    .map((part) => normalizePath(part))
    .filter(Boolean)
    .join('');

  return result || '/';
}

export function interpolatePathTemplate(
  template: string,
  params: Record<string, string | number | undefined>,
): string {
  return template.replace(/:([A-Za-z0-9_]+)/g, (_, key: string) => {
    const value = params[key];

    if (value === undefined || value === null || value === '') {
      throw new Error(`Missing required path parameter: ${key}`);
    }

    return encodeURIComponent(String(value));
  });
}

export function stripTrailingSlash(path: string): string {
  if (path === '/') {
    return path;
  }

  return path.endsWith('/') ? path.slice(0, -1) : path;
}
