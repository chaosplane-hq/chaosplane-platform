import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { resultAnalysisApi } from '@/lib/api';
import type { TriggerAnalysisRequest } from '@/lib/types';

export function useResultAnalyses() {
  return useQuery({
    queryKey: ['result-analysis'],
    queryFn: () => resultAnalysisApi.list(),
    refetchInterval: 30_000,
  });
}

export function useResultAnalysis(id: string) {
  return useQuery({
    queryKey: ['result-analysis', id],
    queryFn: () => resultAnalysisApi.get(id),
    enabled: !!id,
    refetchInterval: 10_000,
  });
}

export function useTriggerAnalysis() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: TriggerAnalysisRequest) => resultAnalysisApi.trigger(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['result-analysis'] }),
  });
}
