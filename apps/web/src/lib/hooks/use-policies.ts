import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { policiesApi } from '@/lib/api';
import type { CreatePolicyRequest } from '@/lib/types';

export function usePolicies() {
  return useQuery({
    queryKey: ['policies'],
    queryFn: () => policiesApi.list(),
  });
}

export function usePolicy(id: string) {
  return useQuery({
    queryKey: ['policies', id],
    queryFn: () => policiesApi.get(id),
    enabled: !!id,
  });
}

export function useCreatePolicy() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreatePolicyRequest) => policiesApi.create(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['policies'] }),
  });
}

export function useDeletePolicy() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => policiesApi.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['policies'] }),
  });
}
