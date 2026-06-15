import {
  type Experiment,
  type ExperimentListResponse,
  type ExperimentListParams,
  type CreateExperimentRequest,
  type Policy,
  type PolicyListResponse,
  type CreatePolicyRequest,
  type OnboardingProgressResponse,
  type OnboardingPatchRequest,
  type AgentTestConnectionResponse,
  type QuickSetupResponse,
  type InvitationListResponse,
  type CreateInvitationRequest,
  type Invitation,
  type TeamMemberListResponse,
  type APIKeyListResponse,
  type CreateAPIKeyRequest,
  type CreateAPIKeyResponse,
  type BillingInfo,
  type NotificationChannelListResponse,
  type CreateNotificationChannelRequest,
  type NotificationChannel,
  type NotificationRuleListResponse,
  type CreateNotificationRuleRequest,
  type NotificationRule,
  type HierarchyResponse,
  type CreateOrganizationRequest,
  type CreateWorkspaceRequest,
  type CreateProjectRequest,
  type CreateEnvironmentRequest,
  type PatchOrganizationRequest,
  type PatchWorkspaceRequest,
  type PatchProjectRequest,
  type PatchEnvironmentRequest,
  type Organization,
  type Workspace,
  type Project,
  type Environment,
  type ResilienceScoreResponse,
  type ResilienceScore,
  type ResilienceScoreParams,
  type VulnerabilityListResponse,
  type VulnerabilityListParams,
  type SuggestionListParams,
  type GameDay,
  type GameDayListResponse,
  type GameDayDetailResponse,
  type CreateGameDayRequest,
  type CreateGameDayEventRequest,
  type GameDayEvent,
  type GameDayPostmortem,
  type CreatePostmortemRequest,
  type WorkflowTemplate,
  type WorkflowTemplateListResponse,
  type CreateWorkflowTemplateRequest,
  type AuditLogListResponse,
  type AuditLogListParams,
  type AuditExport,
  type AuditExportListResponse,
  type SSOProvider,
  type SSOProviderListResponse,
  type CreateSSOProviderRequest,
  type ABACPolicy,
  type ABACPolicyListResponse,
  type CreateABACPolicyRequest,
  type MFARecoveryCodes,
  type ActiveSessionListResponse,
  type ServiceDependencyListResponse,
  type TopologyDriftListResponse,
  type TopologyMetricsListResponse,
  type SuggestionWithConfidenceListResponse,
  type ResultAnalysisListResponse,
  type ResultAnalysis,
  type TriggerAnalysisRequest,
  type ChatSessionListResponse,
  type ChatMessageListResponse,
  type SendMessageRequest,
  type SendMessageResponse,
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
  get: (id: string) => apiFetch<Policy>(`/api/v1/policies/${id}`),
  create: (data: CreatePolicyRequest) =>
    apiFetch<Policy>('/api/v1/policies', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  delete: (id: string) =>
    apiFetch<void>(`/api/v1/policies/${id}`, { method: 'DELETE' }),
};

export const onboardingApi = {
  get: () => apiFetch<OnboardingProgressResponse>('/api/v1/onboarding'),

  update: (data: OnboardingPatchRequest) =>
    apiFetch<OnboardingProgressResponse>('/api/v1/onboarding', {
      method: 'PATCH',
      body: JSON.stringify(data),
    }),

  skip: () =>
    apiFetch<OnboardingProgressResponse>('/api/v1/onboarding/skip', { method: 'POST' }),

  complete: () =>
    apiFetch<OnboardingProgressResponse>('/api/v1/onboarding/complete', { method: 'POST' }),

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

export const hierarchyApi = {
  list: () => apiFetch<HierarchyResponse>('/api/v1/hierarchy'),
  createOrg: (data: CreateOrganizationRequest) =>
    apiFetch<Organization>('/api/v1/organizations', { method: 'POST', body: JSON.stringify(data) }),
  patchOrg: (id: string, data: PatchOrganizationRequest) =>
    apiFetch<Organization>(`/api/v1/organizations/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
  createWorkspace: (data: CreateWorkspaceRequest) =>
    apiFetch<Workspace>('/api/v1/workspaces', { method: 'POST', body: JSON.stringify(data) }),
  patchWorkspace: (id: string, data: PatchWorkspaceRequest) =>
    apiFetch<Workspace>(`/api/v1/workspaces/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
  createProject: (data: CreateProjectRequest) =>
    apiFetch<Project>('/api/v1/projects', { method: 'POST', body: JSON.stringify(data) }),
  patchProject: (id: string, data: PatchProjectRequest) =>
    apiFetch<Project>(`/api/v1/projects/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
  createEnvironment: (data: CreateEnvironmentRequest) =>
    apiFetch<Environment>('/api/v1/environments', { method: 'POST', body: JSON.stringify(data) }),
  patchEnvironment: (id: string, data: PatchEnvironmentRequest) =>
    apiFetch<Environment>(`/api/v1/environments/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
};

export const resilienceApi = {
  get: (params?: ResilienceScoreParams) => {
    const query = new URLSearchParams();
    if (params?.environmentId) query.set('environmentId', params.environmentId);
    const qs = query.toString();
    return apiFetch<ResilienceScoreResponse>(`/api/v1/resilience-score${qs ? `?${qs}` : ''}`);
  },
  calculate: (environmentId: string) =>
    apiFetch<ResilienceScore>('/api/v1/resilience-score/calculate', {
      method: 'POST',
      body: JSON.stringify({ environmentId }),
    }),
};

export const vulnerabilitiesApi = {
  list: (params?: VulnerabilityListParams) => {
    const query = new URLSearchParams();
    if (params?.limit) query.set('limit', String(params.limit));
    if (params?.offset) query.set('offset', String(params.offset));
    if (params?.severity) query.set('severity', params.severity);
    if (params?.status) query.set('status', params.status);
    if (params?.environmentId) query.set('environmentId', params.environmentId);
    const qs = query.toString();
    return apiFetch<VulnerabilityListResponse>(`/api/v1/vulnerabilities${qs ? `?${qs}` : ''}`);
  },
  updateStatus: (id: string, status: string) =>
    apiFetch<void>(`/api/v1/vulnerabilities/${id}`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    }),
  scan: (environmentId: string) =>
    apiFetch<{ findingsCreated: number; findingsUpdated: number }>('/api/v1/vulnerabilities/scan', {
      method: 'POST',
      body: JSON.stringify({ environmentId }),
    }),
};

export const suggestionsApi = {
  list: (params?: SuggestionListParams) => {
    const query = new URLSearchParams();
    if (params?.limit) query.set('limit', String(params.limit));
    if (params?.offset) query.set('offset', String(params.offset));
    if (params?.environmentId) query.set('environmentId', params.environmentId);
    const qs = query.toString();
    return apiFetch<SuggestionWithConfidenceListResponse>(`/api/v1/suggestions${qs ? `?${qs}` : ''}`);
  },
  generate: (environmentId: string) =>
    apiFetch<{ generated: number }>('/api/v1/suggestions/generate', {
      method: 'POST',
      body: JSON.stringify({ environmentId }),
    }),
  delete: (id: string) =>
    apiFetch<void>(`/api/v1/suggestions/${id}`, { method: 'DELETE' }),
};

export const topologyApi = {
  dependencies: (environmentId: string) =>
    apiFetch<ServiceDependencyListResponse>(`/api/v1/topology/dependencies?environmentId=${encodeURIComponent(environmentId)}`),
  drifts: (environmentId: string) =>
    apiFetch<TopologyDriftListResponse>(`/api/v1/topology/drifts?environmentId=${encodeURIComponent(environmentId)}`),
  metrics: (environmentId: string) =>
    apiFetch<TopologyMetricsListResponse>(`/api/v1/topology/metrics?environmentId=${encodeURIComponent(environmentId)}`),
  acknowledgeDrift: (id: string) =>
    apiFetch<void>(`/api/v1/topology/drifts/${id}/acknowledge`, { method: 'POST' }),
};

export const resultAnalysisApi = {
  list: () => apiFetch<ResultAnalysisListResponse>('/api/v1/result-analysis'),
  get: (id: string) => apiFetch<ResultAnalysis>(`/api/v1/result-analysis/${id}`),
  trigger: (data: TriggerAnalysisRequest) =>
    apiFetch<ResultAnalysis>('/api/v1/result-analysis', { method: 'POST', body: JSON.stringify(data) }),
};

export const aiChatApi = {
  listSessions: () => apiFetch<ChatSessionListResponse>('/api/v1/ai/chat/sessions'),
  createSession: () => apiFetch<import('./types').ChatSession>('/api/v1/ai/chat/sessions', { method: 'POST' }),
  deleteSession: (id: string) => apiFetch<void>(`/api/v1/ai/chat/sessions/${id}`, { method: 'DELETE' }),
  listMessages: (sessionId: string) =>
    apiFetch<ChatMessageListResponse>(`/api/v1/ai/chat/sessions/${sessionId}/messages`),
  sendMessage: (sessionId: string, data: SendMessageRequest) =>
    apiFetch<SendMessageResponse>(`/api/v1/ai/chat/sessions/${sessionId}/messages`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),
};

export const gameDaysApi = {
  list: () => apiFetch<GameDayListResponse>('/api/v1/gamedays'),
  get: (id: string) => apiFetch<GameDayDetailResponse>(`/api/v1/gamedays/${id}`),
  create: (data: CreateGameDayRequest) =>
    apiFetch<GameDay>('/api/v1/gamedays', { method: 'POST', body: JSON.stringify(data) }),
  updateStatus: (id: string, status: string) =>
    apiFetch<GameDay>(`/api/v1/gamedays/${id}/status`, { method: 'PATCH', body: JSON.stringify({ status }) }),
  addEvent: (id: string, data: CreateGameDayEventRequest) =>
    apiFetch<GameDayEvent>(`/api/v1/gamedays/${id}/events`, { method: 'POST', body: JSON.stringify(data) }),
  createPostmortem: (id: string, data: CreatePostmortemRequest) =>
    apiFetch<GameDayPostmortem>(`/api/v1/gamedays/${id}/postmortem`, { method: 'POST', body: JSON.stringify(data) }),
  updatePostmortem: (id: string, data: CreatePostmortemRequest) =>
    apiFetch<GameDayPostmortem>(`/api/v1/gamedays/${id}/postmortem`, { method: 'PUT', body: JSON.stringify(data) }),
};

export const workflowsApi = {
  list: () => apiFetch<WorkflowTemplateListResponse>('/api/v1/workflow-templates'),
  get: (id: string) => apiFetch<WorkflowTemplate>(`/api/v1/workflow-templates/${id}`),
  create: (data: CreateWorkflowTemplateRequest) =>
    apiFetch<WorkflowTemplate>('/api/v1/workflow-templates', { method: 'POST', body: JSON.stringify(data) }),
  delete: (id: string) =>
    apiFetch<void>(`/api/v1/workflow-templates/${id}`, { method: 'DELETE' }),
};

export const auditApi = {
  list: (params?: AuditLogListParams) => {
    const query = new URLSearchParams();
    if (params?.limit) query.set('limit', String(params.limit));
    if (params?.offset) query.set('offset', String(params.offset));
    if (params?.action) query.set('action', params.action);
    if (params?.resource) query.set('resource', params.resource);
    if (params?.userId) query.set('userId', params.userId);
    if (params?.from) query.set('from', params.from);
    if (params?.to) query.set('to', params.to);
    const qs = query.toString();
    return apiFetch<AuditLogListResponse>(`/api/v1/audit-logs${qs ? `?${qs}` : ''}`);
  },
  createExport: () =>
    apiFetch<AuditExport>('/api/v1/audit-exports', { method: 'POST' }),
  listExports: () => apiFetch<AuditExportListResponse>('/api/v1/audit-exports'),
};

export const securityApi = {
  listSSO: () => apiFetch<SSOProviderListResponse>('/api/v1/saml-providers'),
  createSSO: (data: CreateSSOProviderRequest) =>
    apiFetch<SSOProvider>('/api/v1/saml-providers', { method: 'POST', body: JSON.stringify(data) }),
  deleteSSO: (id: string) =>
    apiFetch<void>(`/api/v1/saml-providers/${id}`, { method: 'DELETE' }),
  listABACPolicies: () => apiFetch<ABACPolicyListResponse>('/api/v1/abac-policies'),
  createABACPolicy: (data: CreateABACPolicyRequest) =>
    apiFetch<ABACPolicy>('/api/v1/abac-policies', { method: 'POST', body: JSON.stringify(data) }),
  deleteABACPolicy: (id: string) =>
    apiFetch<void>(`/api/v1/abac-policies/${id}`, { method: 'DELETE' }),
  getMFACodes: () => apiFetch<MFARecoveryCodes>('/api/v1/mfa/recovery-codes'),
  generateMFACodes: () =>
    apiFetch<MFARecoveryCodes>('/api/v1/mfa/recovery-codes', { method: 'POST' }),
  listSessions: () => apiFetch<ActiveSessionListResponse>('/api/v1/sessions'),
  revokeSession: (id: string) =>
    apiFetch<void>(`/api/v1/sessions/${id}`, { method: 'DELETE' }),
  revokeAllSessions: () =>
    apiFetch<void>('/api/v1/sessions', { method: 'DELETE' }),
  requestEmailChange: (newEmail: string) =>
    apiFetch<void>('/api/v1/account/change-email', { method: 'POST', body: JSON.stringify({ newEmail }) }),
  requestAccountDeletion: () =>
    apiFetch<void>('/api/v1/account/request-deletion', { method: 'POST' }),
};

export const marketplaceApi = {
  list: (params?: import('./types').MarketplaceListParams) => {
    const query = new URLSearchParams();
    if (params?.category) query.set('category', params.category);
    if (params?.limit) query.set('limit', String(params.limit));
    if (params?.offset) query.set('offset', String(params.offset));
    const qs = query.toString();
    return apiFetch<import('./types').MarketplaceListResponse>(`/api/v1/marketplace${qs ? `?${qs}` : ''}`);
  },
  install: (id: string) =>
    apiFetch<void>('/api/v1/marketplace/install', { method: 'POST', body: JSON.stringify({ pluginId: id }) }),
  uninstall: (id: string) =>
    apiFetch<void>(`/api/v1/marketplace/${id}`, { method: 'DELETE' }),
};

export const federationApi = {
  list: () => apiFetch<import('./types').FederatedClusterListResponse>('/api/v1/federation/clusters'),
  register: (data: import('./types').RegisterClusterRequest) =>
    apiFetch<import('./types').FederatedCluster>('/api/v1/federation/clusters', { method: 'POST', body: JSON.stringify(data) }),
  remove: (id: string) =>
    apiFetch<void>(`/api/v1/federation/clusters/${id}`, { method: 'DELETE' }),
};

export const cicdApi = {
  list: () => apiFetch<import('./types').CICDIntegrationListResponse>('/api/v1/cicd-integrations'),
  create: (data: import('./types').CreateCICDIntegrationRequest) =>
    apiFetch<import('./types').CICDIntegration>('/api/v1/cicd-integrations', { method: 'POST', body: JSON.stringify(data) }),
  delete: (id: string) =>
    apiFetch<void>(`/api/v1/cicd-integrations/${id}`, { method: 'DELETE' }),
};

export const predictionsApi = {
  list: () => apiFetch<import('./types').PredictionListResponse>('/api/v1/predictions'),
  run: () => apiFetch<void>('/api/v1/predictions/run', { method: 'POST' }),
  patchStatus: (id: string, data: import('./types').PatchPredictionStatusRequest) =>
    apiFetch<import('./types').Prediction>(`/api/v1/predictions/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
};
