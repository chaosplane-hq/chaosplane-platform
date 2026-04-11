import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { gameDaysApi } from '@/lib/api';
import type { CreateGameDayRequest, CreateGameDayEventRequest, CreatePostmortemRequest } from '@/lib/types';

export function useGameDays() {
  return useQuery({
    queryKey: ['gamedays'],
    queryFn: () => gameDaysApi.list(),
  });
}

export function useGameDay(id: string) {
  return useQuery({
    queryKey: ['gamedays', id],
    queryFn: () => gameDaysApi.get(id),
    enabled: !!id,
  });
}

export function useCreateGameDay() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateGameDayRequest) => gameDaysApi.create(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['gamedays'] }),
  });
}

export function useUpdateGameDayStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      gameDaysApi.updateStatus(id, status),
    onSuccess: (_data, { id }) => {
      qc.invalidateQueries({ queryKey: ['gamedays', id] });
      qc.invalidateQueries({ queryKey: ['gamedays'] });
    },
  });
}

export function useAddGameDayEvent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: CreateGameDayEventRequest }) =>
      gameDaysApi.addEvent(id, data),
    onSuccess: (_data, { id }) => qc.invalidateQueries({ queryKey: ['gamedays', id] }),
  });
}

export function useCreatePostmortem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: CreatePostmortemRequest }) =>
      gameDaysApi.createPostmortem(id, data),
    onSuccess: (_data, { id }) => qc.invalidateQueries({ queryKey: ['gamedays', id] }),
  });
}

export function useUpdatePostmortem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: CreatePostmortemRequest }) =>
      gameDaysApi.updatePostmortem(id, data),
    onSuccess: (_data, { id }) => qc.invalidateQueries({ queryKey: ['gamedays', id] }),
  });
}
