import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiKeysApi } from '@/lib/api';
import type { CreateAPIKeyRequest } from '@/lib/types';

export function useAPIKeys() {
  return useQuery({
    queryKey: ['api-keys'],
    queryFn: () => apiKeysApi.list(),
  });
}

export function useCreateAPIKey() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateAPIKeyRequest) => apiKeysApi.create(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['api-keys'] }),
  });
}

export function useRotateAPIKey() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiKeysApi.rotate(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['api-keys'] }),
  });
}

export function useRevokeAPIKey() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiKeysApi.revoke(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['api-keys'] }),
  });
}
