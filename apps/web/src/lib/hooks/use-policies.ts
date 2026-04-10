import { useQuery } from '@tanstack/react-query';
import { policiesApi } from '@/lib/api';

export function usePolicies() {
  return useQuery({
    queryKey: ['policies'],
    queryFn: () => policiesApi.list(),
  });
}

export function usePolicy(name: string) {
  return useQuery({
    queryKey: ['policies', name],
    queryFn: () => policiesApi.get(name),
    enabled: !!name,
  });
}
