import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { onboardingApi, agentsApi } from '@/lib/api';
import type { OnboardingUpdateRequest } from '@/lib/types';

export function useOnboarding() {
  return useQuery({
    queryKey: ['onboarding'],
    queryFn: () => onboardingApi.get(),
    staleTime: 0,
  });
}

export function useUpdateOnboarding() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: OnboardingUpdateRequest) => onboardingApi.update(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['onboarding'] }),
  });
}

export function useSkipOnboarding() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => onboardingApi.skip(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['onboarding'] }),
  });
}

export function useCompleteOnboarding() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => onboardingApi.complete(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['onboarding'] }),
  });
}

export function useQuickSetup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => onboardingApi.quickSetup(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['onboarding'] }),
  });
}

export function useTestAgentConnection() {
  return useMutation({
    mutationFn: () => agentsApi.testConnection(),
  });
}
