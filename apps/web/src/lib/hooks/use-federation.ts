import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { federationApi } from '@/lib/api';
import type { RegisterClusterRequest } from '@/lib/types';

export function useFederatedClusters() {
  return useQuery({
    queryKey: ['federation', 'clusters'],
    queryFn: () => federationApi.list(),
  });
}

export function useRegisterCluster() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: RegisterClusterRequest) => federationApi.register(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['federation'] }),
  });
}

export function useRemoveCluster() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => federationApi.remove(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['federation'] }),
  });
}
