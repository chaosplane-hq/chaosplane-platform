import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { marketplaceApi } from '@/lib/api';
import type { MarketplaceListParams } from '@/lib/types';

export function useMarketplace(params?: MarketplaceListParams) {
  return useQuery({
    queryKey: ['marketplace', params],
    queryFn: () => marketplaceApi.list(params),
  });
}

export function useInstallPlugin() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => marketplaceApi.install(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['marketplace'] }),
  });
}

export function useUninstallPlugin() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => marketplaceApi.uninstall(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['marketplace'] }),
  });
}
