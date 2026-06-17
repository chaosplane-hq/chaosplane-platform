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
  id: string;
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

// Mirrors service.ScenarioStep (apps/api .../service/scenario.go): one node in a
// chaos DAG. `dependsOn` references other steps by `name`; the API rejects cycles.
export interface ScenarioStep {
  name: string;
  dependsOn?: string[];
  action: ExperimentAction;
  target: ExperimentTarget;
  duration?: string;
}

// Workflow shape of the create endpoint. The API rejects a body carrying both a
// top-level action and steps, so this deliberately omits action.
export interface CreateScenarioRequest {
  name: string;
  namespace?: string;
  steps: ScenarioStep[];
  duration?: string;
}

// Mirrors service.ParamSpec from GET /api/v1/fault-catalog. `required` follows
// the web convention: a param without a default value is required.
export interface FaultParamSpec {
  key: string;
  required: boolean;
}

export interface FaultCatalogType {
  type: string;
  group: string;
  params?: FaultParamSpec[];
}

export interface FaultCatalogGroup {
  group: string;
  types: FaultCatalogType[];
}

export interface FaultCatalogResponse {
  groups: FaultCatalogGroup[];
}

export type PolicyEnforcement = 'enforce' | 'audit' | 'disabled';

export interface Policy {
  id: string;
  name: string;
  description?: string;
  enforcement: PolicyEnforcement;
  maxConcurrent?: number;
  maxTargets?: number;
  allowedNamespaces: string[];
  blockedNamespaces: string[];
  createdAt: string;
}

export interface PolicyListResponse {
  policies: Policy[];
  total: number;
}

export interface CreatePolicyRequest {
  name: string;
  description?: string;
  enforcement: PolicyEnforcement;
  maxConcurrent?: number;
  maxTargets?: number;
  allowedNamespaces?: string[];
  blockedNamespaces?: string[];
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
  tenantId: string;
  organizationId: string;
  teamId?: string;
  email: string;
  role: MemberRole;
  status: InvitationStatus;
  expiresAt: string;
  acceptedAt?: string;
  createdAt: string;
  inviteToken?: string;
}

export interface InvitationListResponse {
  items: Invitation[];
}

export interface CreateInvitationRequest {
  email: string;
  organizationId: string;
  teamId?: string;
  role?: MemberRole;
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
  items: TeamMember[];
}

export interface APIKey {
  id: string;
  name: string;
  lastUsedAt?: string;
  expiresAt?: string;
  revokedAt?: string;
  createdAt: string;
  plaintext?: string;
}

export interface APIKeyListResponse {
  items: APIKey[];
}

export interface CreateAPIKeyRequest {
  name: string;
  expiresIn?: string;
}

export type CreateAPIKeyResponse = APIKey;

export type BillingPlan = 'free' | 'pro' | 'enterprise';

export type BillingStatus = 'active' | 'past_due' | 'canceled' | 'trialing' | 'suspended';

export interface Subscription {
  id: string;
  tenantId: string;
  plan: BillingPlan;
  status: BillingStatus;
  gateway: string;
  currentPeriodStart?: string;
  currentPeriodEnd?: string;
  trialEndsAt?: string;
  cancelledAt?: string;
  suspendedAt?: string;
  createdAt: string;
}

export interface BillingUsage {
  experiments: number;
  agents: number;
  apiCalls: number;
}

export interface BillingLimits {
  maxExperiments: number;
  maxAgents: number;
  maxApiCalls: number;
}

export interface BillingInfo {
  subscription: Subscription | null;
  usage: BillingUsage | null;
  limits: BillingLimits | null;
}

export type NotificationChannelType = 'slack' | 'webhook' | 'email' | 'pagerduty';

export interface NotificationChannel {
  id: string;
  tenantId: string;
  type: NotificationChannelType;
  name: string;
  config: Record<string, unknown>;
  enabled: boolean;
  createdAt: string;
}

export interface NotificationChannelListResponse {
  items: NotificationChannel[];
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
  tenantId: string;
  channelId: string;
  eventType: string;
  filters: Record<string, unknown> | null;
  enabled: boolean;
  createdAt: string;
}

export interface NotificationRuleListResponse {
  items: NotificationRule[];
}

export interface CreateNotificationRuleRequest {
  channelId: string;
  eventType: string;
  filters?: Record<string, unknown>;
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

export interface OnboardingProgressResponse {
  userId: string;
  tenantId: string;
  stepOrgCreated: boolean;
  stepWorkspaceCreated: boolean;
  stepTeamCreated: boolean;
  stepMemberInvited: boolean;
  stepClusterConnected: boolean;
  stepFirstExperiment: boolean;
  stepResultViewed: boolean;
  completedAt?: string;
  skippedAt?: string;
  updatedAt: string;
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

export interface OnboardingPatchRequest {
  stepOrgCreated?: boolean;
  stepWorkspaceCreated?: boolean;
  stepTeamCreated?: boolean;
  stepMemberInvited?: boolean;
  stepClusterConnected?: boolean;
  stepFirstExperiment?: boolean;
  stepResultViewed?: boolean;
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
  tenantId?: string;
  name: string;
  slug?: string;
  type?: string;
  projectId: string;
  agentStatus: AgentStatus;
}

export interface Project {
  id: string;
  tenantId?: string;
  name: string;
  slug?: string;
  workspaceId: string;
  environments: Environment[];
}

export interface Workspace {
  id: string;
  tenantId?: string;
  name: string;
  slug?: string;
  organizationId: string;
  projects: Project[];
}

export interface Organization {
  id: string;
  tenantId?: string;
  name: string;
  slug?: string;
  workspaces: Workspace[];
}

export interface RawEnvironment {
  id: string;
  tenantId: string;
  projectId: string;
  name: string;
  slug: string;
  type: string;
  agentStatus: AgentStatus;
}

export interface RawProject {
  id: string;
  tenantId: string;
  workspaceId: string;
  name: string;
  slug: string;
  description?: string;
}

export interface RawWorkspace {
  id: string;
  tenantId: string;
  organizationId: string;
  name: string;
  slug: string;
  description?: string;
}

export interface RawOrganization {
  id: string;
  tenantId: string;
  name: string;
  slug: string;
  description?: string;
}

export interface RawTeam {
  id: string;
  tenantId: string;
  workspaceId: string;
  name: string;
  slug: string;
  description?: string;
}

export interface HierarchyResponse {
  organizations: RawOrganization[];
  workspaces: RawWorkspace[];
  teams: RawTeam[];
  projects: RawProject[];
  environments: RawEnvironment[];
}

export interface HierarchyTree {
  organizations: Organization[];
}

export interface CreateOrganizationRequest { name: string }
export interface CreateWorkspaceRequest { name: string; organizationId: string }
export interface CreateProjectRequest { name: string; workspaceId: string }
export interface CreateEnvironmentRequest { name: string; projectId: string }

export interface PatchOrganizationRequest { name?: string }
export interface PatchWorkspaceRequest { name?: string; slug?: string }
export interface PatchProjectRequest { name?: string }
export interface PatchEnvironmentRequest { name?: string }

// Resilience score types
export type ResilienceGrade = 'A' | 'B' | 'C' | 'D' | 'F';

export interface ResilienceScore {
  id: string;
  environmentId: string;
  overallGrade: ResilienceGrade;
  overallScore: number;
  availability: number;
  faultTolerance: number;
  recoverability: number;
  details: Record<string, unknown>;
  calculatedAt: string;
}

export interface ResilienceScoreResponse {
  current: ResilienceScore | null;
  history: ResilienceScore[];
}

export interface ResilienceScoreParams {
  environmentId?: string;
}

// Vulnerability types
export type VulnerabilitySeverity = 'critical' | 'high' | 'medium' | 'low' | 'info';
export type VulnerabilityStatus = 'open' | 'acknowledged' | 'resolved' | 'false_positive';

export interface Vulnerability {
  id: string;
  environmentId: string;
  category: string;
  severity: VulnerabilitySeverity;
  title: string;
  description: string;
  resourceKind: string;
  resourceName: string;
  resourceNamespace: string;
  remediation?: string;
  suggestedExperiment?: Record<string, unknown>;
  status: VulnerabilityStatus;
  detectedAt: string;
  resolvedAt?: string;
}

export interface VulnerabilitySummary {
  critical: number;
  high: number;
  medium: number;
  low: number;
}

export interface VulnerabilityListResponse {
  items: Vulnerability[];
  totalCount: number;
  bySeverity: Record<string, number>;
  byCategory: Record<string, number>;
}

export interface VulnerabilityListParams {
  limit?: number;
  offset?: number;
  severity?: VulnerabilitySeverity;
  status?: VulnerabilityStatus;
  environmentId?: string;
}

// Suggestion types
export interface Suggestion {
  id: string;
  environmentId: string;
  findingId?: string;
  source: string;
  title: string;
  description: string;
  actionType: ActionType | string;
  targetNamespace: string;
  targetName: string;
  duration: string;
  parameters: Record<string, unknown>;
  confidence: number;
  createdAt: string;
}

export interface SuggestionListResponse {
  items: Suggestion[];
}

export interface SuggestionListParams {
  limit?: number;
  offset?: number;
  environmentId?: string;
}

export interface ServiceDependency {
  id: string;
  sourceKind: string;
  sourceName: string;
  sourceNamespace: string;
  targetKind: string;
  targetName: string;
  targetNamespace: string;
  protocol?: string;
  port?: number;
  lastSeenAt: string;
}

export interface ServiceDependencyListResponse {
  dependencies: ServiceDependency[];
  nodeCount: number;
  edgeCount: number;
}

export type DriftSeverity = 'critical' | 'high' | 'medium' | 'low';

export interface TopologyDrift {
  id: string;
  driftType: string;
  resourceKind: string;
  resourceName: string;
  resourceNamespace: string;
  previousState?: Record<string, unknown>;
  currentState?: Record<string, unknown>;
  detectedAt: string;
  acknowledgedAt?: string;
}

export interface TopologyDriftListResponse {
  items: TopologyDrift[];
  totalCount: number;
}

export interface TopologyMetric {
  metricName: string;
  metricValue: number;
  labels: Record<string, unknown>;
  collectedAt: string;
}

export interface TopologyMetricsListResponse {
  items: TopologyMetric[];
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

export type SuggestionWithConfidence = Suggestion;

export interface SuggestionWithConfidenceListResponse {
  items: SuggestionWithConfidence[];
}

export interface ResultAnalysis {
  id: string;
  experimentName: string;
  environmentId?: string;
  summary: string;
  impactAnalysis?: string;
  recommendations?: string;
  severityAssessment?: string;
  affectedServices: Record<string, unknown>;
  metricsImpact: Record<string, unknown>;
  analyzedAt: string;
}

export interface ResultAnalysisListResponse {
  items: ResultAnalysis[];
}

export interface TriggerAnalysisRequest {
  experimentName: string;
  environmentId?: string;
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
}

export interface ChatSessionListResponse {
  items: ChatSession[];
}

export interface ChatMessageListResponse {
  items: ChatMessage[];
}

export interface SendMessageRequest {
  content: string;
}

export interface SendMessageResponse {
  userMessage: ChatMessage;
  assistantMessage: ChatMessage;
}

// Marketplace types
export type PluginCategory = 'chaos_action' | 'workflow_template' | 'integration' | 'monitoring';

export interface MarketplacePlugin {
  id: string;
  name: string;
  displayName: string;
  description?: string;
  author: string;
  version: string;
  category: PluginCategory;
  downloads: number;
  rating: number;
  verified: boolean;
  publishedAt: string;
  installed?: boolean;
}

export interface MarketplaceListResponse {
  items: MarketplacePlugin[];
}

export interface MarketplaceListParams {
  category?: PluginCategory;
  limit?: number;
  offset?: number;
}

// Federation types
export type ClusterProvider = 'aws' | 'gcp' | 'azure' | 'on-premise' | 'other';
export type ClusterStatus = 'connected' | 'disconnected' | 'pending' | 'error';

export interface FederatedCluster {
  id: string;
  name: string;
  region: string;
  provider: ClusterProvider;
  status: ClusterStatus;
  apiEndpoint: string;
  metadata?: Record<string, unknown>;
  createdAt: string;
}

export interface FederatedClusterListResponse {
  items: FederatedCluster[];
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
  config: Record<string, unknown>;
  lastTriggered?: string;
  createdAt: string;
}

export interface CICDIntegrationListResponse {
  items: CICDIntegration[];
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
  environmentId: string;
  predictionType: string;
  severity: PredictionSeverity;
  title: string;
  description: string;
  confidence: number;
  recommendedAction?: string;
  autoRemediation?: Record<string, unknown>;
  status: PredictionStatus;
  predictedAt: string;
}

export interface PredictionListResponse {
  items: Prediction[];
}

export interface PatchPredictionStatusRequest {
  status: 'acknowledged' | 'resolved' | 'dismissed';
}

export type GameDayStatus = 'planned' | 'in_progress' | 'completed' | 'cancelled';

export interface GameDayEvent {
  id: string;
  eventType: string;
  title: string;
  description?: string;
  userId?: string;
  metadata?: Record<string, unknown>;
  createdAt: string;
  occurredAt?: string;
}

export interface GameDayPostmortem {
  id: string;
  summary: string;
  whatWentWell?: string;
  whatWentWrong?: string;
  actionItems?: string;
  lessonsLearned?: string;
  createdBy?: string;
  createdAt: string;
}

export interface GameDay {
  id: string;
  environmentId?: string;
  title: string;
  description?: string;
  status: GameDayStatus;
  scheduledAt?: string;
  startedAt?: string;
  endedAt?: string;
  completedAt?: string;
  createdBy?: string;
  createdAt: string;
  events?: GameDayEvent[];
  postmortem?: GameDayPostmortem;
}

export interface RawGameDayPostmortem {
  id: string;
  summary: string;
  whatWentWell?: string;
  whatWentWrong?: string;
  actionItems: unknown;
  createdBy: string;
  createdAt: string;
}

export interface GameDayDetailResponse {
  gameday: GameDay;
  events: GameDayEvent[];
  postmortem?: RawGameDayPostmortem;
}

export interface GameDayListResponse {
  items: GameDay[];
}

export interface CreateGameDayRequest {
  environmentId: string;
  title: string;
  description?: string;
  scheduledAt?: string;
}

export interface CreateGameDayEventRequest {
  eventType: string;
  title: string;
  description?: string;
  metadata?: Record<string, unknown>;
}

export interface CreatePostmortemRequest {
  summary: string;
  whatWentWell?: string;
  whatWentWrong?: string;
  actionItems?: Record<string, unknown> | unknown[];
}

export type WorkflowCategory = 'chaos' | 'load' | 'security' | 'custom';

export interface WorkflowTemplate {
  id: string;
  tenantId?: string;
  name: string;
  description?: string;
  category: 'chaos' | 'load' | 'security' | 'custom' | string;
  isPublic: boolean;
  spec: Record<string, unknown>;
  createdBy?: string;
  createdAt: string;
}

export interface WorkflowTemplateListResponse {
  items: WorkflowTemplate[];
}

export interface CreateWorkflowTemplateRequest {
  name: string;
  description?: string;
  category?: 'chaos' | 'load' | 'security' | 'custom';
  isPublic?: boolean;
  spec?: Record<string, unknown>;
}

export interface AuditLog {
  id: string;
  tenantId: string;
  userId?: string;
  action: string;
  resourceType: string;
  resourceId?: string;
  ipAddress?: string;
  requestMethod?: string;
  requestPath?: string;
  responseStatus?: number;
  createdAt: string;
}

export interface AuditLogListResponse {
  items: AuditLog[];
  totalCount: number;
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
  destination: string;
  config: Record<string, unknown>;
  status: string;
  recordsExported: number;
  errorMessage?: string;
  createdAt: string;
}

export interface AuditExportListResponse {
  items: AuditExport[];
}

export interface SSOProvider {
  id: string;
  tenantId: string;
  name: string;
  type: 'saml' | 'oidc';
  entityId: string;
  ssoUrl: string;
  metadataUrl?: string;
  enabled: boolean;
  jitProvisioning: boolean;
  defaultRole: string;
  createdAt: string;
}

export interface SSOProviderListResponse {
  items: SSOProvider[];
}

export interface CreateSSOProviderRequest {
  name: string;
  entityId: string;
  ssoUrl: string;
  certificate?: string;
  type?: 'saml' | 'oidc';
  config?: Record<string, string>;
}

export interface ABACPolicy {
  id: string;
  tenantId: string;
  name: string;
  description?: string;
  effect: 'allow' | 'deny';
  subjects: unknown;
  resources: unknown;
  actions: unknown;
  conditions: unknown;
  priority: number;
  enabled: boolean;
  createdAt: string;
}

export interface ABACPolicyListResponse {
  items: ABACPolicy[];
}

export interface CreateABACPolicyRequest {
  name: string;
  description?: string;
  effect: 'allow' | 'deny';
  subjects?: unknown;
  resources?: unknown;
  actions?: unknown;
  conditions?: unknown;
  priority?: number;
}

export interface MFARecoveryCodes {
  codes: string[];
  remaining: number;
  generatedAt?: string;
}

export interface ActiveSession {
  id: string;
  ipAddress?: string;
  userAgent?: string;
  lastActivity: string;
  createdAt: string;
  isCurrent?: boolean;
}

export interface ActiveSessionListResponse {
  items: ActiveSession[];
}
