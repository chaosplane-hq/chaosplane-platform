import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { auditApi } from '@/lib/api';
import type { AuditLogListParams } from '@/lib/types';

export function useAuditLogs(params?: AuditLogListParams) {
  return useQuery({
    queryKey: ['audit-logs', params],
    queryFn: () => auditApi.list(params),
  });
}

export function useAuditExports() {
  return useQuery({
    queryKey: ['audit-exports'],
    queryFn: () => auditApi.listExports(),
  });
}

export function useCreateAuditExport() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => auditApi.createExport(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['audit-exports'] }),
  });
}
