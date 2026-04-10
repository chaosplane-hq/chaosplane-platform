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
