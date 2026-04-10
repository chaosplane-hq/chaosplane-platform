import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { billingApi } from '@/lib/api';

export function useBilling() {
  return useQuery({
    queryKey: ['billing'],
    queryFn: () => billingApi.get(),
  });
}

export function useUpgradePlan() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (plan: string) => billingApi.upgrade(plan),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['billing'] }),
  });
}

export function useCancelPlan() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => billingApi.cancel(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['billing'] }),
  });
}
