import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { workflowsApi } from '@/lib/api';
import type { CreateWorkflowTemplateRequest } from '@/lib/types';

export function useWorkflowTemplates() {
  return useQuery({
    queryKey: ['workflow-templates'],
    queryFn: () => workflowsApi.list(),
  });
}

export function useWorkflowTemplate(id: string) {
  return useQuery({
    queryKey: ['workflow-templates', id],
    queryFn: () => workflowsApi.get(id),
    enabled: !!id,
  });
}

export function useCreateWorkflowTemplate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateWorkflowTemplateRequest) => workflowsApi.create(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['workflow-templates'] }),
  });
}

export function useDeleteWorkflowTemplate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => workflowsApi.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['workflow-templates'] }),
  });
}
