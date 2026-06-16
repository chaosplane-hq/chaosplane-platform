import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { experimentsApi } from '@/lib/api';
import type {
  ExperimentListParams,
  CreateExperimentRequest,
  CreateScenarioRequest,
  FaultCatalogGroup,
} from '@/lib/types';
import { LOCAL_FAULT_CATALOG } from '@/lib/fault-params';

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

export function useCreateScenario() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateScenarioRequest) => experimentsApi.createScenario(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['experiments'] }),
  });
}

// Sources the fault palette from the API, falling back to the locally-mirrored
// catalog so the builder is fully usable before the fault-catalog route ships.
export function useFaultCatalog() {
  return useQuery<FaultCatalogGroup[]>({
    queryKey: ['fault-catalog'],
    queryFn: async () => {
      const res = await experimentsApi.faultCatalog();
      return res.groups.length ? res.groups : LOCAL_FAULT_CATALOG;
    },
    placeholderData: LOCAL_FAULT_CATALOG,
    staleTime: 5 * 60_000,
    retry: false,
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
