import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { notificationsApi } from '@/lib/api';
import type { CreateNotificationChannelRequest, CreateNotificationRuleRequest } from '@/lib/types';

export function useNotificationChannels() {
  return useQuery({
    queryKey: ['notification-channels'],
    queryFn: () => notificationsApi.listChannels(),
  });
}

export function useCreateNotificationChannel() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateNotificationChannelRequest) => notificationsApi.createChannel(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['notification-channels'] }),
  });
}

export function useDeleteNotificationChannel() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => notificationsApi.deleteChannel(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['notification-channels'] }),
  });
}

export function useNotificationRules() {
  return useQuery({
    queryKey: ['notification-rules'],
    queryFn: () => notificationsApi.listRules(),
  });
}

export function useCreateNotificationRule() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateNotificationRuleRequest) => notificationsApi.createRule(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['notification-rules'] });
    },
  });
}

export function useDeleteNotificationRule() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => notificationsApi.deleteRule(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['notification-rules'] }),
  });
}
