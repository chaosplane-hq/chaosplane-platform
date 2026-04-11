import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { predictionsApi } from '@/lib/api';
import type { PatchPredictionStatusRequest } from '@/lib/types';

export function usePredictions() {
  return useQuery({
    queryKey: ['predictions'],
    queryFn: () => predictionsApi.list(),
    refetchInterval: 60_000,
  });
}

export function useRunPredictions() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => predictionsApi.run(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['predictions'] }),
  });
}

export function usePatchPredictionStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: PatchPredictionStatusRequest }) =>
      predictionsApi.patchStatus(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['predictions'] }),
  });
}
