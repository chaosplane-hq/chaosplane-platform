import type { Experiment, ServiceDependency } from '@/lib/types';

export type FaultNodeState = 'healthy' | 'source' | 'degraded' | 'failed';

export interface FaultNode {
  id: string;
  hop: number;
  state: FaultNodeState;
}

export interface FaultEdge {
  from: string;
  to: string;
  hop: number;
}

export interface FaultModel {
  sourceIds: string[];
  nodes: Map<string, FaultNode>;
  edges: FaultEdge[];
  maxHop: number;
  recovered: boolean;
}

function nodeId(kind: string, namespace: string, name: string): string {
  return `${kind}/${namespace}/${name}`;
}

// affectedResources refs arrive in mixed shapes across backends (bare name,
// kind/name, namespace/name, full ref), so match loosely on id and name tail.
function affectedMatches(
  affected: string,
  node: { id: string; name: string; namespace: string },
): boolean {
  const ref = affected.trim().toLowerCase();
  if (!ref) return false;
  if (ref === node.id.toLowerCase()) return true;
  if (ref === node.name.toLowerCase()) return true;
  const tail = ref.split('/').pop() ?? ref;
  return tail === node.name.toLowerCase();
}

interface GraphShape {
  nodes: { id: string; name: string; namespace: string }[];
  dependents: Map<string, string[]>;
}

function buildGraph(dependencies: ServiceDependency[]): GraphShape {
  const nodes = new Map<string, { id: string; name: string; namespace: string }>();
  const dependents = new Map<string, string[]>();

  const ensure = (kind: string, name: string, namespace: string): string => {
    const id = nodeId(kind, namespace, name);
    if (!nodes.has(id)) nodes.set(id, { id, name, namespace });
    if (!dependents.has(id)) dependents.set(id, []);
    return id;
  };

  for (const d of dependencies) {
    const sourceId = ensure(d.sourceKind, d.sourceName, d.sourceNamespace);
    const targetId = ensure(d.targetKind, d.targetName, d.targetNamespace);
    // Edge A->B means A depends on B; a fault travels the reverse direction, so
    // register the dependent (source) against the thing it depends on (target).
    dependents.get(targetId)!.push(sourceId);
  }

  return { nodes: [...nodes.values()], dependents };
}

function resolveSources(experiment: Experiment, graph: GraphShape): string[] {
  const affected = experiment.status.affectedResources ?? [];
  const targetNs = experiment.target.namespace;

  const matched = graph.nodes.filter((n) =>
    affected.some((a) => affectedMatches(a, n)),
  );
  if (matched.length > 0) return matched.map((n) => n.id);

  // Without a resource match we can't pin the exact pod, so the targeted
  // namespace's services stand in as the blast origin (namespace target is real).
  return graph.nodes.filter((n) => n.namespace === targetNs).map((n) => n.id);
}

export function deriveFaultModel(
  experiment: Experiment | null | undefined,
  dependencies: ServiceDependency[],
): FaultModel | null {
  if (!experiment || dependencies.length === 0) return null;

  const graph = buildGraph(dependencies);
  const sourceIds = resolveSources(experiment, graph);
  if (sourceIds.length === 0) return null;

  const recovered =
    experiment.status.phase === 'Completed' ||
    experiment.status.phase === 'Failed' ||
    experiment.status.phase === 'Aborted';

  const nodes = new Map<string, FaultNode>();
  const edges: FaultEdge[] = [];
  const queue: string[] = [];

  for (const id of sourceIds) {
    if (nodes.has(id)) continue;
    nodes.set(id, { id, hop: 0, state: 'source' });
    queue.push(id);
  }

  let maxHop = 0;
  while (queue.length > 0) {
    const current = queue.shift()!;
    const currentHop = nodes.get(current)!.hop;
    for (const dependent of graph.dependents.get(current) ?? []) {
      if (nodes.has(dependent)) continue;
      const hop = currentHop + 1;
      maxHop = Math.max(maxHop, hop);
      // First-hop dependents degrade; anything deeper is treated as failed.
      nodes.set(dependent, {
        id: dependent,
        hop,
        state: hop === 1 ? 'degraded' : 'failed',
      });
      queue.push(dependent);
      edges.push({ from: current, to: dependent, hop });
    }
  }

  return { sourceIds, nodes, edges, maxHop, recovered };
}
