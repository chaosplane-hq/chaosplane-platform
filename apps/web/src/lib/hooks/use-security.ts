import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { securityApi } from '@/lib/api';
import type { CreateSSOProviderRequest, CreateABACPolicyRequest } from '@/lib/types';

export function useSSOProviders() {
  return useQuery({ queryKey: ['sso-providers'], queryFn: () => securityApi.listSSO() });
}

export function useCreateSSOProvider() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateSSOProviderRequest) => securityApi.createSSO(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['sso-providers'] }),
  });
}

export function useDeleteSSOProvider() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => securityApi.deleteSSO(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['sso-providers'] }),
  });
}

export function useABACPolicies() {
  return useQuery({ queryKey: ['abac-policies'], queryFn: () => securityApi.listABACPolicies() });
}

export function useCreateABACPolicy() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateABACPolicyRequest) => securityApi.createABACPolicy(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['abac-policies'] }),
  });
}

export function useDeleteABACPolicy() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => securityApi.deleteABACPolicy(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['abac-policies'] }),
  });
}

export function useMFARecoveryCodes() {
  return useQuery({ queryKey: ['mfa-recovery-codes'], queryFn: () => securityApi.getMFACodes() });
}

export function useGenerateMFACodes() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => securityApi.generateMFACodes(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['mfa-recovery-codes'] }),
  });
}

export function useActiveSessions() {
  return useQuery({ queryKey: ['active-sessions'], queryFn: () => securityApi.listSessions() });
}

export function useRevokeSession() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => securityApi.revokeSession(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['active-sessions'] }),
  });
}

export function useRevokeAllSessions() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => securityApi.revokeAllSessions(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['active-sessions'] }),
  });
}

export function useRequestEmailChange() {
  return useMutation({
    mutationFn: (newEmail: string) => securityApi.requestEmailChange(newEmail),
  });
}

export function useRequestAccountDeletion() {
  return useMutation({
    mutationFn: () => securityApi.requestAccountDeletion(),
  });
}
