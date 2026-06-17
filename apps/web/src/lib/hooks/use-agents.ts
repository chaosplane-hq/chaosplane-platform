import { useMutation } from '@tanstack/react-query';
import { agentsApi } from '@/lib/api';

export function useCreateAgentToken() {
  return useMutation({
    mutationFn: (data: { environmentId: string; name: string }) =>
      agentsApi.createToken(data),
  });
}
