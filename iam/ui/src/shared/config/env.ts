export type ApiMode = 'mock' | 'live';

const configuredApiMode = import.meta.env.VITE_IAM_API_MODE as ApiMode | undefined;
const apiMode: ApiMode = configuredApiMode === 'mock' ? 'mock' : 'live';
const fallbackToMockEnv = import.meta.env.VITE_IAM_FALLBACK_TO_MOCK;

export const env = {
  apiMode,
  apiBaseUrl: import.meta.env.VITE_IAM_API_BASE_URL?.trim() ?? '',
  defaultTenantId: import.meta.env.VITE_IAM_DEFAULT_TENANT_ID?.trim() || 'tenant-demo',
  enableFallbackToMock:
    fallbackToMockEnv === undefined ? apiMode === 'mock' : fallbackToMockEnv !== 'false',
} as const;
