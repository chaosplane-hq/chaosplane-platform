import {
  type Experiment,
  type ExperimentListResponse,
  type ExperimentListParams,
  type CreateExperimentRequest,
  type Policy,
  type PolicyListResponse,
  type OnboardingState,
  type OnboardingUpdateRequest,
  type AgentTestConnectionResponse,
  type QuickSetupResponse,
  type InvitationListResponse,
  type CreateInvitationRequest,
  type Invitation,
  type TeamMemberListResponse,
  type APIKeyListResponse,
  type CreateAPIKeyRequest,
  type CreateAPIKeyResponse,
  type APIKey,
  type BillingInfo,
  type NotificationChannelListResponse,
  type CreateNotificationChannelRequest,
  type NotificationChannel,
  type NotificationRuleListResponse,
  type CreateNotificationRuleRequest,
  type NotificationRule,
} from './types';

import { authHeaders } from './auth';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080';

export async function apiFetch<T>(
  path: string,
  options?: RequestInit,
): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...authHeaders(),
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

export const onboardingApi = {
  get: () => apiFetch<OnboardingState>('/api/v1/onboarding'),

  update: (data: OnboardingUpdateRequest) =>
    apiFetch<OnboardingState>('/api/v1/onboarding', {
      method: 'PATCH',
      body: JSON.stringify(data),
    }),

  skip: () =>
    apiFetch<OnboardingState>('/api/v1/onboarding/skip', { method: 'POST' }),

  complete: () =>
    apiFetch<OnboardingState>('/api/v1/onboarding/complete', { method: 'POST' }),

  quickSetup: () =>
    apiFetch<QuickSetupResponse>('/auth/quick-setup', { method: 'POST' }),
};

export const agentsApi = {
  testConnection: () =>
    apiFetch<AgentTestConnectionResponse>('/api/v1/agents/test-connection', { method: 'POST' }),
};

export const invitationsApi = {
  listMembers: () => apiFetch<TeamMemberListResponse>('/api/v1/members'),
  list: () => apiFetch<InvitationListResponse>('/api/v1/invitations'),
  create: (data: CreateInvitationRequest) =>
    apiFetch<Invitation>('/api/v1/invitations', { method: 'POST', body: JSON.stringify(data) }),
  resend: (id: string) =>
    apiFetch<Invitation>(`/api/v1/invitations/${id}/resend`, { method: 'POST' }),
  revoke: (id: string) =>
    apiFetch<void>(`/api/v1/invitations/${id}`, { method: 'DELETE' }),
};

export const apiKeysApi = {
  list: () => apiFetch<APIKeyListResponse>('/api/v1/api-keys'),
  create: (data: CreateAPIKeyRequest) =>
    apiFetch<CreateAPIKeyResponse>('/api/v1/api-keys', { method: 'POST', body: JSON.stringify(data) }),
  rotate: (id: string) =>
    apiFetch<CreateAPIKeyResponse>(`/api/v1/api-keys/${id}/rotate`, { method: 'POST' }),
  revoke: (id: string) =>
    apiFetch<void>(`/api/v1/api-keys/${id}`, { method: 'DELETE' }),
};

export const billingApi = {
  get: () => apiFetch<BillingInfo>('/api/v1/billing'),
  upgrade: (plan: string) =>
    apiFetch<BillingInfo>('/api/v1/billing/upgrade', { method: 'POST', body: JSON.stringify({ plan }) }),
  cancel: () =>
    apiFetch<BillingInfo>('/api/v1/billing/cancel', { method: 'POST' }),
};

export const notificationsApi = {
  listChannels: () => apiFetch<NotificationChannelListResponse>('/api/v1/notification-channels'),
  createChannel: (data: CreateNotificationChannelRequest) =>
    apiFetch<NotificationChannel>('/api/v1/notification-channels', { method: 'POST', body: JSON.stringify(data) }),
  deleteChannel: (id: string) =>
    apiFetch<void>(`/api/v1/notification-channels/${id}`, { method: 'DELETE' }),
  listRules: () => apiFetch<NotificationRuleListResponse>('/api/v1/notification-rules'),
  createRule: (data: CreateNotificationRuleRequest) =>
    apiFetch<NotificationRule>('/api/v1/notification-rules', { method: 'POST', body: JSON.stringify(data) }),
  deleteRule: (id: string) =>
    apiFetch<void>(`/api/v1/notification-rules/${id}`, { method: 'DELETE' }),
};
