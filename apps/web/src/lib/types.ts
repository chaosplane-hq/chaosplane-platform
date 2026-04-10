export type ExperimentPhase =
  | 'Pending'
  | 'Running'
  | 'Completed'
  | 'Failed'
  | 'Aborted';

export type ActionType =
  | 'pod-kill'
  | 'container-kill'
  | 'pod-cpu-stress'
  | 'pod-memory-stress'
  | 'pod-io-stress'
  | 'pod-dns-error'
  | 'pod-http-abort'
  | 'pod-http-delay'
  | 'network-delay'
  | 'network-loss'
  | 'network-corrupt'
  | 'network-duplicate'
  | 'network-partition'
  | 'network-bandwidth'
  | 'node-drain'
  | 'node-taint'
  | 'node-restart'
  | 'node-cpu-stress'
  | 'stress-cpu'
  | 'stress-memory';

export const ACTION_TYPES: ActionType[] = [
  'pod-kill',
  'container-kill',
  'pod-cpu-stress',
  'pod-memory-stress',
  'pod-io-stress',
  'pod-dns-error',
  'pod-http-abort',
  'pod-http-delay',
  'network-delay',
  'network-loss',
  'network-corrupt',
  'network-duplicate',
  'network-partition',
  'network-bandwidth',
  'node-drain',
  'node-taint',
  'node-restart',
  'node-cpu-stress',
  'stress-cpu',
  'stress-memory',
];

export const ACTION_TYPE_GROUPS: Record<string, ActionType[]> = {
  Pod: [
    'pod-kill',
    'container-kill',
    'pod-cpu-stress',
    'pod-memory-stress',
    'pod-io-stress',
    'pod-dns-error',
    'pod-http-abort',
    'pod-http-delay',
  ],
  Network: [
    'network-delay',
    'network-loss',
    'network-corrupt',
    'network-duplicate',
    'network-partition',
    'network-bandwidth',
  ],
  Node: ['node-drain', 'node-taint', 'node-restart', 'node-cpu-stress'],
  Stress: ['stress-cpu', 'stress-memory'],
};

export interface ExperimentTarget {
  namespace: string;
  labelSelector?: Record<string, string>;
  mode?: 'one' | 'all' | 'fixed' | 'fixed-percent' | 'random-max-percent';
  value?: string;
}

export interface ExperimentAction {
  type: ActionType;
  parameters?: Record<string, string | number | boolean>;
}

export interface ExperimentStatus {
  phase: ExperimentPhase;
  startTime?: string;
  completionTime?: string;
  affectedResources?: string[];
  message?: string;
}

export interface Experiment {
  name: string;
  namespace: string;
  action: ExperimentAction;
  target: ExperimentTarget;
  status: ExperimentStatus;
  duration?: string;
  createdAt?: string;
}

export interface ExperimentListResponse {
  experiments: Experiment[];
  total: number;
  limit: number;
  offset: number;
}

export interface ExperimentListParams {
  limit?: number;
  offset?: number;
  status?: ExperimentPhase;
  namespace?: string;
  action?: ActionType;
}

export interface CreateExperimentRequest {
  name: string;
  namespace: string;
  action: ExperimentAction;
  target: ExperimentTarget;
  duration?: string;
}

export interface PolicyRule {
  type: string;
  maxConcurrent?: number;
  allowedNamespaces?: string[];
  blockedNamespaces?: string[];
}

export interface Policy {
  name: string;
  namespace: string;
  rules: PolicyRule[];
  createdAt?: string;
}

export interface PolicyListResponse {
  policies: Policy[];
  total: number;
}

export interface ExperimentStatusMessage {
  name: string;
  namespace: string;
  phase: ExperimentPhase;
  message?: string;
  affectedResources?: string[];
  timestamp: string;
}

export type InvitationStatus = 'pending' | 'accepted' | 'expired' | 'revoked';
export type MemberRole = 'owner' | 'admin' | 'member' | 'viewer';

export interface Invitation {
  id: string;
  email: string;
  role: MemberRole;
  status: InvitationStatus;
  invitedBy: string;
  createdAt: string;
  expiresAt: string;
}

export interface InvitationListResponse {
  invitations: Invitation[];
  total: number;
}

export interface CreateInvitationRequest {
  email: string;
  role: MemberRole;
}

export interface TeamMember {
  id: string;
  email: string;
  name: string;
  role: MemberRole;
  joinedAt: string;
  avatarUrl?: string;
}

export interface TeamMemberListResponse {
  members: TeamMember[];
  total: number;
}

export interface APIKey {
  id: string;
  name: string;
  prefix: string;
  scopes: string[];
  createdAt: string;
  lastUsedAt?: string;
  expiresAt?: string;
}

export interface APIKeyListResponse {
  keys: APIKey[];
  total: number;
}

export interface CreateAPIKeyRequest {
  name: string;
  scopes: string[];
  expiresAt?: string;
}

export interface CreateAPIKeyResponse {
  key: APIKey;
  plaintext: string;
}

export type BillingPlan = 'free' | 'pro' | 'enterprise';

export interface BillingUsage {
  experimentsRun: number;
  experimentsLimit: number;
  membersCount: number;
  membersLimit: number;
  apiCallsCount: number;
  apiCallsLimit: number;
}

export interface BillingInfo {
  plan: BillingPlan;
  status: 'active' | 'past_due' | 'canceled' | 'trialing';
  currentPeriodEnd?: string;
  usage: BillingUsage;
  nextInvoiceAmount?: number;
}

export type NotificationChannelType = 'slack' | 'webhook' | 'email' | 'pagerduty';

export interface NotificationChannel {
  id: string;
  name: string;
  type: NotificationChannelType;
  config: Record<string, string>;
  createdAt: string;
}

export interface NotificationChannelListResponse {
  channels: NotificationChannel[];
  total: number;
}

export interface CreateNotificationChannelRequest {
  name: string;
  type: NotificationChannelType;
  config: Record<string, string>;
}

export type NotificationEvent =
  | 'experiment.started'
  | 'experiment.completed'
  | 'experiment.failed'
  | 'experiment.aborted';

export interface NotificationRule {
  id: string;
  channelId: string;
  events: NotificationEvent[];
  namespaceFilter?: string;
  createdAt: string;
}

export interface NotificationRuleListResponse {
  rules: NotificationRule[];
  total: number;
}

export interface CreateNotificationRuleRequest {
  channelId: string;
  events: NotificationEvent[];
  namespaceFilter?: string;
}

export type OnboardingStepId =
  | 'org'
  | 'workspace'
  | 'team'
  | 'invite_member'
  | 'connect_cluster'
  | 'first_experiment'
  | 'view_results';

export interface OnboardingStep {
  id: OnboardingStepId;
  completed: boolean;
  skipped?: boolean;
}

export interface OnboardingState {
  completed: boolean;
  skipped: boolean;
  currentStep: OnboardingStepId;
  steps: OnboardingStep[];
}

export interface OnboardingUpdateRequest {
  currentStep?: OnboardingStepId;
  stepId?: OnboardingStepId;
  stepCompleted?: boolean;
}

export interface QuickSetupResponse {
  success: boolean;
  message?: string;
}

export interface AgentTestConnectionResponse {
  connected: boolean;
  agentVersion?: string;
  message?: string;
}
