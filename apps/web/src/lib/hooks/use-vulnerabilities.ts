import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { vulnerabilitiesApi } from '@/lib/api';
import type { VulnerabilityListParams, VulnerabilityStatus } from '@/lib/types';

export function useVulnerabilities(params?: VulnerabilityListParams) {
  return useQuery({
    queryKey: ['vulnerabilities', params],
    queryFn: () => vulnerabilitiesApi.list(params),
    staleTime: 60_000,
  });
}

export function useUpdateVulnerabilityStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: VulnerabilityStatus }) =>
      vulnerabilitiesApi.updateStatus(id, status),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['vulnerabilities'] }),
  });
}

export function useScanVulnerabilities() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (environmentId?: string) => vulnerabilitiesApi.scan(environmentId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['vulnerabilities'] }),
  });
}
