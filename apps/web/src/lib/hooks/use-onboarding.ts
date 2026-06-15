import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { onboardingApi, agentsApi } from '@/lib/api';
import type {
  OnboardingProgressResponse,
  OnboardingState,
  OnboardingStepId,
  OnboardingStep,
  OnboardingUpdateRequest,
  OnboardingPatchRequest,
} from '@/lib/types';

const STEP_ORDER: OnboardingStepId[] = [
  'org',
  'workspace',
  'team',
  'invite_member',
  'connect_cluster',
  'first_experiment',
  'view_results',
];

const STEP_FLAGS: Record<OnboardingStepId, keyof OnboardingPatchRequest> = {
  org: 'stepOrgCreated',
  workspace: 'stepWorkspaceCreated',
  team: 'stepTeamCreated',
  invite_member: 'stepMemberInvited',
  connect_cluster: 'stepClusterConnected',
  first_experiment: 'stepFirstExperiment',
  view_results: 'stepResultViewed',
};

function adaptOnboarding(raw: OnboardingProgressResponse): OnboardingState {
  const steps: OnboardingStep[] = STEP_ORDER.map((id) => ({
    id,
    completed: Boolean(raw[STEP_FLAGS[id]]),
  }));
  const currentStep = steps.find((s) => !s.completed)?.id ?? STEP_ORDER[STEP_ORDER.length - 1];
  return {
    completed: Boolean(raw.completedAt),
    skipped: Boolean(raw.skippedAt),
    currentStep,
    steps,
  };
}

export function useOnboarding() {
  return useQuery({
    queryKey: ['onboarding'],
    queryFn: () => onboardingApi.get(),
    select: adaptOnboarding,
    staleTime: 0,
  });
}

export function useUpdateOnboarding() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: OnboardingUpdateRequest) => {
      const patch: OnboardingPatchRequest = {};
      if (data.stepId && data.stepCompleted) {
        patch[STEP_FLAGS[data.stepId]] = true;
      }
      return onboardingApi.update(patch);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['onboarding'] }),
  });
}

export function useSkipOnboarding() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => onboardingApi.skip(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['onboarding'] }),
  });
}

export function useCompleteOnboarding() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => onboardingApi.complete(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['onboarding'] }),
  });
}

export function useQuickSetup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => onboardingApi.quickSetup(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['onboarding'] }),
  });
}

export function useTestAgentConnection() {
  return useMutation({
    mutationFn: () => agentsApi.testConnection(),
  });
}
