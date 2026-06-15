import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { hierarchyApi } from '@/lib/api';
import type {
  HierarchyResponse,
  HierarchyTree,
  Organization,
  Workspace,
  Project,
  Environment,
  CreateOrganizationRequest,
  CreateWorkspaceRequest,
  CreateProjectRequest,
  CreateEnvironmentRequest,
  PatchOrganizationRequest,
  PatchWorkspaceRequest,
  PatchProjectRequest,
  PatchEnvironmentRequest,
} from '@/lib/types';

function buildTree(raw: HierarchyResponse): HierarchyTree {
  const environmentsByProject = new Map<string, Environment[]>();
  for (const env of raw.environments ?? []) {
    const list = environmentsByProject.get(env.projectId) ?? [];
    list.push({
      id: env.id,
      tenantId: env.tenantId,
      name: env.name,
      slug: env.slug,
      type: env.type,
      projectId: env.projectId,
      agentStatus: env.agentStatus,
    });
    environmentsByProject.set(env.projectId, list);
  }

  const projectsByWorkspace = new Map<string, Project[]>();
  for (const proj of raw.projects ?? []) {
    const list = projectsByWorkspace.get(proj.workspaceId) ?? [];
    list.push({
      id: proj.id,
      tenantId: proj.tenantId,
      name: proj.name,
      slug: proj.slug,
      workspaceId: proj.workspaceId,
      environments: environmentsByProject.get(proj.id) ?? [],
    });
    projectsByWorkspace.set(proj.workspaceId, list);
  }

  const workspacesByOrg = new Map<string, Workspace[]>();
  for (const ws of raw.workspaces ?? []) {
    const list = workspacesByOrg.get(ws.organizationId) ?? [];
    list.push({
      id: ws.id,
      tenantId: ws.tenantId,
      name: ws.name,
      slug: ws.slug,
      organizationId: ws.organizationId,
      projects: projectsByWorkspace.get(ws.id) ?? [],
    });
    workspacesByOrg.set(ws.organizationId, list);
  }

  const organizations: Organization[] = (raw.organizations ?? []).map((org) => ({
    id: org.id,
    tenantId: org.tenantId,
    name: org.name,
    slug: org.slug,
    workspaces: workspacesByOrg.get(org.id) ?? [],
  }));

  return { organizations };
}

export function useHierarchy() {
  return useQuery({
    queryKey: ['hierarchy'],
    queryFn: () => hierarchyApi.list(),
    select: buildTree,
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
