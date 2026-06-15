import { useQuery } from '@tanstack/react-query';
import { hierarchyApi } from '@/lib/api';
import type { HierarchyResponse, RawEnvironment } from '@/lib/types';

export function useEnvironmentsList() {
  return useQuery({
    queryKey: ['hierarchy'],
    queryFn: () => hierarchyApi.list(),
    select: (raw: HierarchyResponse): RawEnvironment[] => raw.environments ?? [],
    staleTime: 30_000,
  });
}

export function useDefaultEnvironmentId() {
  const { data, isLoading } = useEnvironmentsList();
  return {
    environments: data ?? [],
    environmentId: data?.[0]?.id,
    isLoading,
  };
}
