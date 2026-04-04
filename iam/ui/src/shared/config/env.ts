export type ApiMode = 'mock' | 'live';

const apiMode = (import.meta.env.VITE_IAM_API_MODE as ApiMode | undefined) === 'live' ? 'live' : 'mock';

export const env = {
  apiMode,
  apiBaseUrl: import.meta.env.VITE_IAM_API_BASE_URL?.trim() ?? '',
  defaultTenantId: import.meta.env.VITE_IAM_DEFAULT_TENANT_ID?.trim() || 'tenant-demo',
  enableFallbackToMock: import.meta.env.VITE_IAM_FALLBACK_TO_MOCK !== 'false',
} as const;
