import {env} from '@/shared/config/env';

type QueryValue = string | number | boolean | undefined | null;

function buildQuery(params?: Record<string, QueryValue>): string {
  if (!params) {
    return '';
  }

  const search = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value === undefined || value === null || value === '') {
      return;
    }
    search.set(key, String(value));
  });

  const serialized = search.toString();
  return serialized ? `?${serialized}` : '';
}

export class ApiError extends Error {
  readonly status: number;
  readonly payload?: unknown;

  constructor(message: string, status: number, payload?: unknown) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.payload = payload;
  }
}

export class ApiClient {
  async request<T>(
    path: string,
    init?: RequestInit,
    query?: Record<string, QueryValue>,
  ): Promise<T> {
    const response = await fetch(`${env.apiBaseUrl}${path}${buildQuery(query)}`, {
      headers: {
        'Content-Type': 'application/json',
        ...(init?.headers ?? {}),
      },
      ...init,
    });

    if (!response.ok) {
      const payload = await safeJson(response);
      throw new ApiError(`Request failed with status ${response.status}`, response.status, payload);
    }

    if (response.status === 204) {
      return undefined as T;
    }

    return (await safeJson(response)) as T;
  }

  get<T>(path: string, query?: Record<string, QueryValue>) {
    return this.request<T>(path, undefined, query);
  }

  post<T>(path: string, body?: unknown) {
    return this.request<T>(path, {
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  patch<T>(path: string, body?: unknown) {
    return this.request<T>(path, {
      method: 'PATCH',
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  delete<T>(path: string, query?: Record<string, QueryValue>) {
    return this.request<T>(path, {method: 'DELETE'}, query);
  }
}

async function safeJson(response: Response): Promise<unknown> {
  const contentType = response.headers.get('content-type') || '';
  if (!contentType.includes('application/json')) {
    return response.text();
  }
  return response.json();
}

export const apiClient = new ApiClient();
