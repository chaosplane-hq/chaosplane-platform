import type { ActionType, FaultCatalogGroup } from './types';
import { ACTION_TYPE_GROUPS } from './types';

export interface ParamDef {
  key: string;
  label: string;
  placeholder?: string;
  defaultValue?: string;
}

// Per-fault parameter schemas. A param WITHOUT a defaultValue is treated as
// required — the same convention the API's fault registry uses to derive
// ParamSpec.Required, so this local table and the catalog stay consistent.
export const ACTION_PARAMS: Partial<Record<ActionType, ParamDef[]>> = {
  'container-kill': [{ key: 'containerName', label: 'Container Name', placeholder: 'app' }],
  'pod-cpu-stress': [
    { key: 'workers', label: 'Workers', placeholder: '1', defaultValue: '1' },
    { key: 'load', label: 'CPU Load (%)', placeholder: '50', defaultValue: '50' },
  ],
  'pod-memory-stress': [
    { key: 'workers', label: 'Workers', placeholder: '1', defaultValue: '1' },
    { key: 'size', label: 'Memory Size', placeholder: '256MB', defaultValue: '256MB' },
  ],
  'pod-io-stress': [
    { key: 'workers', label: 'Workers', placeholder: '1', defaultValue: '1' },
    { key: 'volumePath', label: 'Volume Path', placeholder: '/data' },
  ],
  'pod-dns-error': [{ key: 'patterns', label: 'DNS Patterns (comma-separated)', placeholder: 'example.com' }],
  'pod-http-abort': [
    { key: 'port', label: 'Port', placeholder: '8080', defaultValue: '8080' },
    { key: 'path', label: 'Path', placeholder: '/', defaultValue: '/' },
    { key: 'method', label: 'Method', placeholder: 'GET', defaultValue: 'GET' },
  ],
  'pod-http-delay': [
    { key: 'port', label: 'Port', placeholder: '8080', defaultValue: '8080' },
    { key: 'path', label: 'Path', placeholder: '/', defaultValue: '/' },
    { key: 'delay', label: 'Delay (ms)', placeholder: '1000', defaultValue: '1000' },
    { key: 'method', label: 'Method', placeholder: 'GET', defaultValue: 'GET' },
  ],
  'network-delay': [
    { key: 'latency', label: 'Latency', placeholder: '100ms', defaultValue: '100ms' },
    { key: 'jitter', label: 'Jitter', placeholder: '10ms', defaultValue: '10ms' },
    { key: 'correlation', label: 'Correlation (%)', placeholder: '0', defaultValue: '0' },
  ],
  'network-loss': [
    { key: 'loss', label: 'Loss (%)', placeholder: '10', defaultValue: '10' },
    { key: 'correlation', label: 'Correlation (%)', placeholder: '0', defaultValue: '0' },
  ],
  'network-corrupt': [
    { key: 'corrupt', label: 'Corrupt (%)', placeholder: '10', defaultValue: '10' },
    { key: 'correlation', label: 'Correlation (%)', placeholder: '0', defaultValue: '0' },
  ],
  'network-duplicate': [
    { key: 'duplicate', label: 'Duplicate (%)', placeholder: '10', defaultValue: '10' },
    { key: 'correlation', label: 'Correlation (%)', placeholder: '0', defaultValue: '0' },
  ],
  'network-partition': [{ key: 'direction', label: 'Direction', placeholder: 'both', defaultValue: 'both' }],
  'network-bandwidth': [
    { key: 'rate', label: 'Rate', placeholder: '1mbps', defaultValue: '1mbps' },
    { key: 'limit', label: 'Limit (bytes)', placeholder: '10000', defaultValue: '10000' },
    { key: 'buffer', label: 'Buffer (bytes)', placeholder: '10000', defaultValue: '10000' },
  ],
  'node-taint': [
    { key: 'key', label: 'Taint Key', placeholder: 'chaos' },
    { key: 'value', label: 'Taint Value', placeholder: 'true' },
    { key: 'effect', label: 'Effect', placeholder: 'NoSchedule', defaultValue: 'NoSchedule' },
  ],
  'node-cpu-stress': [
    { key: 'workers', label: 'Workers', placeholder: '1', defaultValue: '1' },
    { key: 'load', label: 'CPU Load (%)', placeholder: '50', defaultValue: '50' },
  ],
  'stress-cpu': [
    { key: 'workers', label: 'Workers', placeholder: '1', defaultValue: '1' },
    { key: 'load', label: 'CPU Load (%)', placeholder: '50', defaultValue: '50' },
  ],
  'stress-memory': [
    { key: 'workers', label: 'Workers', placeholder: '1', defaultValue: '1' },
    { key: 'size', label: 'Memory Size', placeholder: '256MB', defaultValue: '256MB' },
  ],
  'ebpf-network-delay': [
    { key: 'latency', label: 'Latency', placeholder: '100ms', defaultValue: '100ms' },
    { key: 'interface', label: 'Network Interface', placeholder: 'eth0', defaultValue: 'eth0' },
  ],
  'ebpf-network-loss': [
    { key: 'loss', label: 'Loss (%)', placeholder: '10', defaultValue: '10' },
    { key: 'interface', label: 'Network Interface', placeholder: 'eth0', defaultValue: 'eth0' },
  ],
  'ebpf-dns-chaos': [
    { key: 'patterns', label: 'DNS Patterns (comma-separated)', placeholder: 'example.com' },
    { key: 'action', label: 'Action', placeholder: 'error', defaultValue: 'error' },
  ],
  'aws-ec2-stop': [
    { key: 'instanceId', label: 'Instance ID', placeholder: 'i-1234567890abcdef0' },
    { key: 'region', label: 'Region', placeholder: 'us-east-1', defaultValue: 'us-east-1' },
  ],
  'aws-ec2-terminate': [
    { key: 'instanceId', label: 'Instance ID', placeholder: 'i-1234567890abcdef0' },
    { key: 'region', label: 'Region', placeholder: 'us-east-1', defaultValue: 'us-east-1' },
  ],
  'aws-rds-failover': [
    { key: 'dbClusterIdentifier', label: 'DB Cluster Identifier', placeholder: 'my-aurora-cluster' },
    { key: 'region', label: 'Region', placeholder: 'us-east-1', defaultValue: 'us-east-1' },
  ],
  'aws-ecs-stop-task': [
    { key: 'cluster', label: 'ECS Cluster', placeholder: 'my-cluster' },
    { key: 'taskId', label: 'Task ID', placeholder: 'arn:aws:ecs:...' },
    { key: 'region', label: 'Region', placeholder: 'us-east-1', defaultValue: 'us-east-1' },
  ],
  'aws-az-failure': [
    { key: 'az', label: 'Availability Zone', placeholder: 'us-east-1a' },
    { key: 'region', label: 'Region', placeholder: 'us-east-1', defaultValue: 'us-east-1' },
  ],
  'azure-vm-stop': [
    { key: 'resourceGroup', label: 'Resource Group', placeholder: 'my-rg' },
    { key: 'vmName', label: 'VM Name', placeholder: 'my-vm' },
  ],
  'azure-aks-scale': [
    { key: 'resourceGroup', label: 'Resource Group', placeholder: 'my-rg' },
    { key: 'clusterName', label: 'Cluster Name', placeholder: 'my-aks' },
    { key: 'nodeCount', label: 'Node Count', placeholder: '1', defaultValue: '1' },
  ],
  'azure-cosmosdb-failover': [
    { key: 'resourceGroup', label: 'Resource Group', placeholder: 'my-rg' },
    { key: 'accountName', label: 'Account Name', placeholder: 'my-cosmos' },
    { key: 'failoverRegion', label: 'Failover Region', placeholder: 'eastus2' },
  ],
  'gcp-gke-scale': [
    { key: 'project', label: 'Project ID', placeholder: 'my-project' },
    { key: 'cluster', label: 'Cluster Name', placeholder: 'my-cluster' },
    { key: 'nodePool', label: 'Node Pool', placeholder: 'default-pool' },
    { key: 'nodeCount', label: 'Node Count', placeholder: '1', defaultValue: '1' },
  ],
  'gcp-cloudsql-failover': [
    { key: 'project', label: 'Project ID', placeholder: 'my-project' },
    { key: 'instance', label: 'Instance Name', placeholder: 'my-sql-instance' },
  ],
  'gcp-cloudrun-stop': [
    { key: 'project', label: 'Project ID', placeholder: 'my-project' },
    { key: 'service', label: 'Service Name', placeholder: 'my-service' },
    { key: 'region', label: 'Region', placeholder: 'us-central1', defaultValue: 'us-central1' },
  ],
  'vm-cpu-stress': [
    { key: 'workers', label: 'Workers', placeholder: '1', defaultValue: '1' },
    { key: 'load', label: 'CPU Load (%)', placeholder: '50', defaultValue: '50' },
  ],
  'vm-memory-stress': [
    { key: 'workers', label: 'Workers', placeholder: '1', defaultValue: '1' },
    { key: 'size', label: 'Memory Size', placeholder: '256MB', defaultValue: '256MB' },
  ],
  'vm-disk-stress': [
    { key: 'workers', label: 'Workers', placeholder: '1', defaultValue: '1' },
    { key: 'path', label: 'Disk Path', placeholder: '/tmp', defaultValue: '/tmp' },
    { key: 'size', label: 'Write Size', placeholder: '1GB', defaultValue: '1GB' },
  ],
  'vm-network-delay': [
    { key: 'latency', label: 'Latency', placeholder: '100ms', defaultValue: '100ms' },
    { key: 'interface', label: 'Network Interface', placeholder: 'eth0', defaultValue: 'eth0' },
  ],
  'vm-process-kill': [
    { key: 'processName', label: 'Process Name', placeholder: 'nginx' },
    { key: 'signal', label: 'Signal', placeholder: 'SIGKILL', defaultValue: 'SIGKILL' },
  ],
  'vm-process-suspend': [{ key: 'processName', label: 'Process Name', placeholder: 'nginx' }],
};

export const MODE_OPTIONS = ['one', 'all', 'fixed', 'fixed-percent', 'random-max-percent'] as const;
export type TargetMode = (typeof MODE_OPTIONS)[number];

export function paramDefsFor(type: ActionType): ParamDef[] {
  return ACTION_PARAMS[type] ?? [];
}

export function defaultParamsFor(type: ActionType): Record<string, string> {
  const defaults: Record<string, string> = {};
  for (const def of paramDefsFor(type)) {
    if (def.defaultValue) defaults[def.key] = def.defaultValue;
  }
  return defaults;
}

export function requiredParamKeys(type: ActionType): string[] {
  return paramDefsFor(type)
    .filter((d) => d.defaultValue === undefined)
    .map((d) => d.key);
}

// Local fallback catalog built from the same tables the API mirrors, so the
// palette works before the fault-catalog route is reachable. Shape matches
// FaultCatalogGroup so consumers treat catalog and fallback identically.
export const LOCAL_FAULT_CATALOG: FaultCatalogGroup[] = Object.entries(ACTION_TYPE_GROUPS).map(
  ([group, types]) => ({
    group,
    types: types.map((type) => ({
      group,
      type,
      params: paramDefsFor(type).map((d) => ({ key: d.key, required: d.defaultValue === undefined })),
    })),
  }),
);
