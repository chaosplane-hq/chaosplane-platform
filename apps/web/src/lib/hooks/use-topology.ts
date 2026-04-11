import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { topologyApi } from '@/lib/api';

export function useTopologyDependencies() {
  return useQuery({
    queryKey: ['topology', 'dependencies'],
    queryFn: () => topologyApi.dependencies(),
    refetchInterval: 30_000,
  });
}

export function useTopologyDrifts() {
  return useQuery({
    queryKey: ['topology', 'drifts'],
    queryFn: () => topologyApi.drifts(),
    refetchInterval: 30_000,
  });
}

export function useTopologyMetrics() {
  return useQuery({
    queryKey: ['topology', 'metrics'],
    queryFn: () => topologyApi.metrics(),
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
