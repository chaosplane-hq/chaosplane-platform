import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { experimentsApi } from '@/lib/api';
import type { ExperimentListParams, CreateExperimentRequest } from '@/lib/types';

export function useExperiments(params?: ExperimentListParams) {
  return useQuery({
    queryKey: ['experiments', params],
    queryFn: () => experimentsApi.list(params),
    refetchInterval: 30_000,
  });
}

export function useExperiment(name: string) {
  return useQuery({
    queryKey: ['experiments', name],
    queryFn: () => experimentsApi.get(name),
    refetchInterval: 10_000,
    enabled: !!name,
  });
}

export function useCreateExperiment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateExperimentRequest) => experimentsApi.create(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['experiments'] }),
  });
}

export function useDeleteExperiment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (name: string) => experimentsApi.delete(name),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['experiments'] }),
  });
}

export function useAbortExperiment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (name: string) => experimentsApi.abort(name),
    onSuccess: (_data, name) => {
      qc.invalidateQueries({ queryKey: ['experiments', name] });
      qc.invalidateQueries({ queryKey: ['experiments'] });
    },
  });
}
