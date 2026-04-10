import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { invitationsApi } from '@/lib/api';
import type { CreateInvitationRequest } from '@/lib/types';

export function useTeamMembers() {
  return useQuery({
    queryKey: ['members'],
    queryFn: () => invitationsApi.listMembers(),
  });
}

export function useInvitations() {
  return useQuery({
    queryKey: ['invitations'],
    queryFn: () => invitationsApi.list(),
  });
}

export function useCreateInvitation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateInvitationRequest) => invitationsApi.create(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invitations'] }),
  });
}

export function useResendInvitation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => invitationsApi.resend(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invitations'] }),
  });
}

export function useRevokeInvitation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => invitationsApi.revoke(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invitations'] }),
  });
}
