import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { suggestionsApi } from '@/lib/api';
import type { SuggestionListParams } from '@/lib/types';

export function useSuggestions(params?: SuggestionListParams) {
  return useQuery({
    queryKey: ['suggestions', params],
    queryFn: () => suggestionsApi.list(params),
    enabled: !!params?.environmentId,
    staleTime: 60_000,
  });
}

export function useGenerateSuggestions() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (environmentId: string) => suggestionsApi.generate(environmentId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['suggestions'] }),
  });
}

export function useDeleteSuggestion() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => suggestionsApi.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['suggestions'] }),
  });
}
