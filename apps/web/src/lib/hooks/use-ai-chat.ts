import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { aiChatApi } from '@/lib/api';
import type { SendMessageRequest } from '@/lib/types';

export function useChatSessions() {
  return useQuery({
    queryKey: ['chat-sessions'],
    queryFn: () => aiChatApi.listSessions(),
  });
}

export function useCreateChatSession() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => aiChatApi.createSession(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['chat-sessions'] }),
  });
}

export function useDeleteChatSession() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => aiChatApi.deleteSession(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['chat-sessions'] }),
  });
}

export function useChatMessages(sessionId: string) {
  return useQuery({
    queryKey: ['chat-messages', sessionId],
    queryFn: () => aiChatApi.listMessages(sessionId),
    enabled: !!sessionId,
  });
}

export function useSendMessage(sessionId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: SendMessageRequest) => aiChatApi.sendMessage(sessionId, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['chat-messages', sessionId] }),
  });
}
