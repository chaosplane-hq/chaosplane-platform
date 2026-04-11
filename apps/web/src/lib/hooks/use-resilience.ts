import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { resilienceApi, resilienceHistoryApi } from '@/lib/api';
import type { ResilienceScoreParams } from '@/lib/types';

export function useResilienceScore(params?: ResilienceScoreParams) {
  return useQuery({
    queryKey: ['resilience-score', params],
    queryFn: () => resilienceApi.get(params),
    staleTime: 60_000,
  });
}

export function useResilienceHistory(params?: ResilienceScoreParams) {
  return useQuery({
    queryKey: ['resilience-history', params],
    queryFn: () => resilienceHistoryApi.history(params),
  });
}

export function useCalculateResilienceScore() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (environmentId: string) => resilienceApi.calculate(environmentId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['resilience-score'] });
      qc.invalidateQueries({ queryKey: ['resilience-history'] });
    },
  });
}
