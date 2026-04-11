import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { hierarchyApi } from '@/lib/api';
import type {
  CreateOrganizationRequest,
  CreateWorkspaceRequest,
  CreateProjectRequest,
  CreateEnvironmentRequest,
  PatchOrganizationRequest,
  PatchWorkspaceRequest,
  PatchProjectRequest,
  PatchEnvironmentRequest,
} from '@/lib/types';

export function useHierarchy() {
  return useQuery({
    queryKey: ['hierarchy'],
    queryFn: () => hierarchyApi.list(),
    staleTime: 30_000,
  });
}

export function useCreateOrganization() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateOrganizationRequest) => hierarchyApi.createOrg(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hierarchy'] }),
  });
}

export function usePatchOrganization() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: PatchOrganizationRequest }) =>
      hierarchyApi.patchOrg(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hierarchy'] }),
  });
}

export function useCreateWorkspace() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateWorkspaceRequest) => hierarchyApi.createWorkspace(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hierarchy'] }),
  });
}

export function usePatchWorkspace() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: PatchWorkspaceRequest }) =>
      hierarchyApi.patchWorkspace(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hierarchy'] }),
  });
}

export function useCreateProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateProjectRequest) => hierarchyApi.createProject(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hierarchy'] }),
  });
}

export function usePatchProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: PatchProjectRequest }) =>
      hierarchyApi.patchProject(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hierarchy'] }),
  });
}

export function useCreateEnvironment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateEnvironmentRequest) => hierarchyApi.createEnvironment(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hierarchy'] }),
  });
}

export function usePatchEnvironment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: PatchEnvironmentRequest }) =>
      hierarchyApi.patchEnvironment(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hierarchy'] }),
  });
}
