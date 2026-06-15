import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { topologyApi } from '@/lib/api';

export function useTopologyDependencies(environmentId?: string) {
  return useQuery({
    queryKey: ['topology', 'dependencies', environmentId],
    queryFn: () => topologyApi.dependencies(environmentId!),
    enabled: !!environmentId,
    refetchInterval: 30_000,
  });
}

export function useTopologyDrifts(environmentId?: string) {
  return useQuery({
    queryKey: ['topology', 'drifts', environmentId],
    queryFn: () => topologyApi.drifts(environmentId!),
    enabled: !!environmentId,
    refetchInterval: 30_000,
  });
}

export function useTopologyMetrics(environmentId?: string) {
  return useQuery({
    queryKey: ['topology', 'metrics', environmentId],
    queryFn: () => topologyApi.metrics(environmentId!),
    enabled: !!environmentId,
    refetchInterval: 30_000,
  });
}

export function useAcknowledgeDrift() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => topologyApi.acknowledgeDrift(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['topology', 'drifts'] }),
  });
}
