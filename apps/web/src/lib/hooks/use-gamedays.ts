import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { gameDaysApi } from '@/lib/api';
import type {
  GameDay,
  GameDayDetailResponse,
  GameDayPostmortem,
  CreateGameDayRequest,
  CreateGameDayEventRequest,
  CreatePostmortemRequest,
} from '@/lib/types';

function actionItemsToText(raw: unknown): string {
  if (raw == null) return '';
  if (typeof raw === 'string') return raw;
  if (Array.isArray(raw)) return raw.map((v) => (typeof v === 'string' ? v : JSON.stringify(v))).join('\n');
  return JSON.stringify(raw);
}

function adaptGameDayDetail(detail: GameDayDetailResponse): GameDay {
  const postmortem: GameDayPostmortem | undefined = detail.postmortem
    ? {
        id: detail.postmortem.id,
        summary: detail.postmortem.summary,
        whatWentWell: detail.postmortem.whatWentWell,
        whatWentWrong: detail.postmortem.whatWentWrong,
        lessonsLearned: detail.postmortem.whatWentWell,
        actionItems: actionItemsToText(detail.postmortem.actionItems),
        createdBy: detail.postmortem.createdBy,
        createdAt: detail.postmortem.createdAt,
      }
    : undefined;

  return {
    ...detail.gameday,
    completedAt: detail.gameday.endedAt,
    events: (detail.events ?? []).map((e) => ({
      ...e,
      occurredAt: e.createdAt,
    })),
    postmortem,
  };
}

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
    select: adaptGameDayDetail,
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
