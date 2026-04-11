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
  | 'stress-memory'
  | 'ebpf-network-delay'
  | 'ebpf-network-loss'
  | 'ebpf-dns-chaos'
  | 'aws-ec2-stop'
  | 'aws-ec2-terminate'
  | 'aws-rds-failover'
  | 'aws-ecs-stop-task'
  | 'aws-az-failure'
  | 'azure-vm-stop'
  | 'azure-aks-scale'
  | 'azure-cosmosdb-failover'
  | 'gcp-gke-scale'
  | 'gcp-cloudsql-failover'
  | 'gcp-cloudrun-stop'
  | 'vm-cpu-stress'
  | 'vm-memory-stress'
  | 'vm-disk-stress'
  | 'vm-network-delay'
  | 'vm-process-kill'
  | 'vm-process-suspend';

export const ACTION_TYPES: ActionType[] = [
  'pod-kill', 'container-kill', 'pod-cpu-stress', 'pod-memory-stress',
  'pod-io-stress', 'pod-dns-error', 'pod-http-abort', 'pod-http-delay',
  'network-delay', 'network-loss', 'network-corrupt', 'network-duplicate',
  'network-partition', 'network-bandwidth', 'node-drain', 'node-taint',
  'node-restart', 'node-cpu-stress', 'stress-cpu', 'stress-memory',
  'ebpf-network-delay', 'ebpf-network-loss', 'ebpf-dns-chaos',
  'aws-ec2-stop', 'aws-ec2-terminate', 'aws-rds-failover', 'aws-ecs-stop-task', 'aws-az-failure',
  'azure-vm-stop', 'azure-aks-scale', 'azure-cosmosdb-failover',
  'gcp-gke-scale', 'gcp-cloudsql-failover', 'gcp-cloudrun-stop',
  'vm-cpu-stress', 'vm-memory-stress', 'vm-disk-stress', 'vm-network-delay', 'vm-process-kill', 'vm-process-suspend',
];

export const ACTION_TYPE_GROUPS: Record<string, ActionType[]> = {
  Pod: ['pod-kill', 'container-kill', 'pod-cpu-stress', 'pod-memory-stress', 'pod-io-stress', 'pod-dns-error', 'pod-http-abort', 'pod-http-delay'],
  Network: ['network-delay', 'network-loss', 'network-corrupt', 'network-duplicate', 'network-partition', 'network-bandwidth'],
  Node: ['node-drain', 'node-taint', 'node-restart', 'node-cpu-stress'],
  Stress: ['stress-cpu', 'stress-memory'],
  eBPF: ['ebpf-network-delay', 'ebpf-network-loss', 'ebpf-dns-chaos'],
  AWS: ['aws-ec2-stop', 'aws-ec2-terminate', 'aws-rds-failover', 'aws-ecs-stop-task', 'aws-az-failure'],
  Azure: ['azure-vm-stop', 'azure-aks-scale', 'azure-cosmosdb-failover'],
  GCP: ['gcp-gke-scale', 'gcp-cloudsql-failover', 'gcp-cloudrun-stop'],
  VM: ['vm-cpu-stress', 'vm-memory-stress', 'vm-disk-stress', 'vm-network-delay', 'vm-process-kill', 'vm-process-suspend'],
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

// Hierarchy types
export type AgentStatus = 'connected' | 'disconnected' | 'degraded';

export interface Environment {
  id: string;
  name: string;
  projectId: string;
  agentStatus: AgentStatus;
  createdAt: string;
  updatedAt?: string;
}

export interface Project {
  id: string;
  name: string;
  workspaceId: string;
  environments: Environment[];
  createdAt: string;
}

export interface Workspace {
  id: string;
  name: string;
  organizationId: string;
  projects: Project[];
  createdAt: string;
}

export interface Organization {
  id: string;
  name: string;
  workspaces: Workspace[];
  createdAt: string;
}

export interface HierarchyResponse {
  organizations: Organization[];
}

export interface CreateOrganizationRequest { name: string }
export interface CreateWorkspaceRequest { name: string; organizationId: string }
export interface CreateProjectRequest { name: string; workspaceId: string }
export interface CreateEnvironmentRequest { name: string; projectId: string }

export interface PatchOrganizationRequest { name?: string }
export interface PatchWorkspaceRequest { name?: string }
export interface PatchProjectRequest { name?: string }
export interface PatchEnvironmentRequest { name?: string }

// Resilience score types
export type ResilienceGrade = 'A' | 'B' | 'C' | 'D' | 'F';

export interface ResilienceScore {
  grade: ResilienceGrade;
  score: number;
  environmentId: string;
  calculatedAt: string;
  breakdown?: Record<string, number>;
}

export interface ResilienceScoreParams {
  environmentId?: string;
}

// Vulnerability types
export type VulnerabilitySeverity = 'critical' | 'high' | 'medium' | 'low' | 'info';
export type VulnerabilityStatus = 'open' | 'acknowledged' | 'resolved' | 'false_positive';

export interface Vulnerability {
  id: string;
  title: string;
  severity: VulnerabilitySeverity;
  status: VulnerabilityStatus;
  description?: string;
  environmentId?: string;
  detectedAt: string;
  resolvedAt?: string;
}

export interface VulnerabilityListResponse {
  vulnerabilities: Vulnerability[];
  total: number;
}

export interface VulnerabilityListParams {
  limit?: number;
  offset?: number;
  severity?: VulnerabilitySeverity;
  status?: VulnerabilityStatus;
  environmentId?: string;
}

// Suggestion types
export type SuggestionPriority = 'high' | 'medium' | 'low';

export interface Suggestion {
  id: string;
  title: string;
  description: string;
  priority: SuggestionPriority;
  category?: string;
  createdAt: string;
}

export interface SuggestionListResponse {
  suggestions: Suggestion[];
  total: number;
}

export interface SuggestionListParams {
  limit?: number;
  offset?: number;
}

export interface ServiceDependency {
  id: string;
  source: string;
  target: string;
  protocol: string;
  latencyP99?: number;
  errorRate?: number;
  requestsPerSecond?: number;
}

export interface ServiceDependencyListResponse {
  dependencies: ServiceDependency[];
  total: number;
}

export type DriftSeverity = 'critical' | 'high' | 'medium' | 'low';

export interface TopologyDrift {
  id: string;
  service: string;
  type: string;
  description: string;
  severity: DriftSeverity;
  detectedAt: string;
  acknowledgedAt?: string;
  acknowledgedBy?: string;
}

export interface TopologyDriftListResponse {
  drifts: TopologyDrift[];
  total: number;
}

export interface TopologyMetric {
  id: string;
  service: string;
  metric: string;
  value: number;
  unit: string;
  timestamp: string;
  trend?: 'up' | 'down' | 'stable';
}

export interface TopologyMetricsListResponse {
  metrics: TopologyMetric[];
  total: number;
}

export interface VulnerabilitySummary {
  critical: number;
  high: number;
  medium: number;
  low: number;
}

export interface VulnerabilityListWithSummaryResponse extends VulnerabilityListResponse {
  summary: VulnerabilitySummary;
}

export interface UpdateVulnerabilityStatusRequest {
  status: VulnerabilityStatus;
}

export interface SuggestionWithConfidence extends Suggestion {
  confidence: number;
  targetService?: string;
  experimentParams?: Record<string, string | number | boolean>;
}

export interface SuggestionWithConfidenceListResponse {
  suggestions: SuggestionWithConfidence[];
  total: number;
}

export type AnalysisStatus = 'pending' | 'running' | 'completed' | 'failed';

export interface ResultAnalysis {
  id: string;
  experimentName: string;
  status: AnalysisStatus;
  summary?: string;
  impact?: string;
  recommendations?: string[];
  createdAt: string;
  completedAt?: string;
}

export interface ResultAnalysisListResponse {
  analyses: ResultAnalysis[];
  total: number;
}

export interface TriggerAnalysisRequest {
  experimentName: string;
}

export type ChatMessageRole = 'user' | 'assistant';

export interface ChatMessage {
  id: string;
  role: ChatMessageRole;
  content: string;
  createdAt: string;
}

export interface ChatSession {
  id: string;
  title: string;
  createdAt: string;
  updatedAt: string;
  messageCount: number;
}

export interface ChatSessionListResponse {
  sessions: ChatSession[];
  total: number;
}

export interface ChatMessageListResponse {
  messages: ChatMessage[];
  total: number;
}

export interface SendMessageRequest {
  content: string;
}

export interface SendMessageResponse {
  message: ChatMessage;
  reply: ChatMessage;
}

// Marketplace types
export type PluginCategory = 'chaos_action' | 'workflow_template' | 'integration' | 'monitoring';

export interface MarketplacePlugin {
  id: string;
  name: string;
  description: string;
  author: string;
  downloads: number;
  rating: number;
  verified: boolean;
  category: PluginCategory;
  version: string;
  installed?: boolean;
  tags?: string[];
}

export interface MarketplaceListResponse {
  plugins: MarketplacePlugin[];
  total: number;
}

export interface MarketplaceListParams {
  category?: PluginCategory;
  limit?: number;
  offset?: number;
}

// Federation types
export type ClusterProvider = 'aws' | 'gcp' | 'azure' | 'on-premise' | 'other';
export type ClusterStatus = 'active' | 'inactive' | 'pending' | 'error';

export interface FederatedCluster {
  id: string;
  name: string;
  region: string;
  provider: ClusterProvider;
  status: ClusterStatus;
  apiEndpoint: string;
  registeredAt: string;
}

export interface FederatedClusterListResponse {
  clusters: FederatedCluster[];
  total: number;
}

export interface RegisterClusterRequest {
  name: string;
  region: string;
  provider: ClusterProvider;
  apiEndpoint: string;
}

// CI/CD types
export type CICDProvider = 'github_actions' | 'gitlab_ci' | 'jenkins';

export interface CICDIntegration {
  id: string;
  name: string;
  provider: CICDProvider;
  enabled: boolean;
  config: Record<string, string>;
  createdAt: string;
}

export interface CICDIntegrationListResponse {
  integrations: CICDIntegration[];
  total: number;
}

export interface CreateCICDIntegrationRequest {
  name: string;
  provider: CICDProvider;
  config: Record<string, string>;
}

// Prediction types
export type PredictionSeverity = 'critical' | 'high' | 'medium' | 'low';
export type PredictionStatus = 'active' | 'acknowledged' | 'resolved' | 'dismissed';

export interface Prediction {
  id: string;
  title: string;
  description: string;
  severity: PredictionSeverity;
  confidence: number;
  recommendedAction: string;
  status: PredictionStatus;
  createdAt: string;
  acknowledgedAt?: string;
  resolvedAt?: string;
}

export interface PredictionListResponse {
  predictions: Prediction[];
  total: number;
}

export interface PatchPredictionStatusRequest {
  status: 'acknowledged' | 'resolved' | 'dismissed';
}

export interface ResilienceScoreHistory {
  scores: ResilienceScore[];
  environmentId: string;
}

export type GameDayStatus = 'planned' | 'in_progress' | 'completed' | 'cancelled';

export interface GameDayEvent {
  id: string;
  gameDayId: string;
  title?: string;
  description: string;
  occurredAt?: string;
  timestamp: string;
  type?: string;
}

export interface GameDayPostmortem {
  id: string;
  gameDayId: string;
  summary: string;
  findings: string;
  lessonsLearned?: string;
  actionItems: string;
  createdAt: string;
  updatedAt?: string;
}

export interface GameDay {
  id: string;
  title: string;
  description?: string;
  status: GameDayStatus;
  scheduledAt?: string;
  completedAt?: string;
  events: GameDayEvent[];
  postmortem?: GameDayPostmortem;
  createdAt: string;
}

export interface GameDayListResponse {
  gameDays: GameDay[];
  total: number;
}

export interface CreateGameDayRequest {
  title: string;
  description?: string;
  scheduledAt?: string;
}

export interface CreateGameDayEventRequest {
  description: string;
  type?: string;
  timestamp?: string;
}

export interface CreatePostmortemRequest {
  summary: string;
  findings: string;
  actionItems: string;
}

export interface WorkflowTemplate {
  id: string;
  name: string;
  description?: string;
  category?: 'chaos' | 'load' | 'security' | 'custom';
  isPublic?: boolean;
  spec?: Record<string, unknown>;
  steps: Record<string, unknown>[];
  createdAt: string;
  updatedAt?: string;
}

export interface WorkflowTemplateListResponse {
  templates: WorkflowTemplate[];
  total: number;
}

export interface CreateWorkflowTemplateRequest {
  name: string;
  description?: string;
  category?: 'chaos' | 'load' | 'security' | 'custom';
  isPublic?: boolean;
  spec?: Record<string, unknown>;
  steps?: Record<string, unknown>[];
}

export interface AuditLog {
  id: string;
  action: string;
  resource: string;
  resourceId?: string;
  userId: string;
  userEmail?: string;
  details?: Record<string, unknown>;
  createdAt: string;
}

export interface AuditLogListResponse {
  logs: AuditLog[];
  total: number;
}

export interface AuditLogListParams {
  limit?: number;
  offset?: number;
  action?: string;
  resource?: string;
  userId?: string;
  from?: string;
  to?: string;
}

export interface AuditExport {
  id: string;
  status: 'pending' | 'ready' | 'completed' | 'failed';
  downloadUrl?: string;
  createdAt: string;
}

export interface AuditExportListResponse {
  exports: AuditExport[];
  total: number;
}

export interface SSOProvider {
  id: string;
  name: string;
  type: 'saml' | 'oidc';
  entityId?: string;
  ssoUrl?: string;
  enabled: boolean;
  config: Record<string, string>;
  createdAt: string;
}

export interface SSOProviderListResponse {
  providers: SSOProvider[];
  total: number;
}

export interface CreateSSOProviderRequest {
  name: string;
  type: 'saml' | 'oidc';
  entityId?: string;
  ssoUrl?: string;
  config?: Record<string, string>;
}

export interface ABACPolicy {
  id: string;
  name: string;
  description?: string;
  effect: 'allow' | 'deny';
  actions: string[];
  resources: string[];
  rules: Record<string, unknown>[];
  createdAt: string;
}

export interface ABACPolicyListResponse {
  policies: ABACPolicy[];
  total: number;
}

export interface CreateABACPolicyRequest {
  name: string;
  description?: string;
  effect?: 'allow' | 'deny';
  actions?: string[];
  resources?: string[];
  rules?: Record<string, unknown>[];
}

export interface MFARecoveryCodes {
  codes: string[];
  remaining: number;
  generatedAt: string;
}

export interface ActiveSession {
  id: string;
  userAgent: string;
  ipAddress: string;
  createdAt: string;
  lastActiveAt: string;
  isCurrent: boolean;
}
