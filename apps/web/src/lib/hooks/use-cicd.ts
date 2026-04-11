import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { cicdApi } from '@/lib/api';
import type { CreateCICDIntegrationRequest } from '@/lib/types';

export function useCICDIntegrations() {
  return useQuery({
    queryKey: ['cicd-integrations'],
    queryFn: () => cicdApi.list(),
  });
}

export function useCreateCICDIntegration() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateCICDIntegrationRequest) => cicdApi.create(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['cicd-integrations'] }),
  });
}

export function useDeleteCICDIntegration() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => cicdApi.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['cicd-integrations'] }),
  });
}
