import {
  type Experiment,
  type ExperimentListResponse,
  type ExperimentListParams,
  type CreateExperimentRequest,
  type Policy,
  type PolicyListResponse,
} from './types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080';

export async function apiFetch<T>(
  path: string,
  options?: RequestInit,
): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    ...options,
  });

  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`);
  }

  return res.json() as Promise<T>;
}

export function getWsUrl(path: string): string {
  const base = (process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080')
    .replace(/^http/, 'ws');
  return `${base}${path}`;
}

export const experimentsApi = {
  list: (params?: ExperimentListParams) => {
    const query = new URLSearchParams();
    if (params?.limit) query.set('limit', String(params.limit));
    if (params?.offset) query.set('offset', String(params.offset));
    if (params?.status) query.set('status', params.status);
    if (params?.namespace) query.set('namespace', params.namespace);
    if (params?.action) query.set('action', params.action);
    const qs = query.toString();
    return apiFetch<ExperimentListResponse>(`/api/v1/experiments${qs ? `?${qs}` : ''}`);
  },

  get: (name: string) =>
    apiFetch<Experiment>(`/api/v1/experiments/${name}`),

  create: (data: CreateExperimentRequest) =>
    apiFetch<Experiment>('/api/v1/experiments', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  delete: (name: string) =>
    apiFetch<void>(`/api/v1/experiments/${name}`, { method: 'DELETE' }),

  abort: (name: string) =>
    apiFetch<Experiment>(`/api/v1/experiments/${name}/abort`, { method: 'POST' }),
};

export const policiesApi = {
  list: () => apiFetch<PolicyListResponse>('/api/v1/policies'),
  get: (name: string) => apiFetch<Policy>(`/api/v1/policies/${name}`),
};
