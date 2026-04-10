'use client';

import { useRouter } from 'next/navigation';
import {
  Grid,
  Column,
  ProgressIndicator,
  ProgressStep,
  Button,
  InlineNotification,
  SkeletonText,
  Tile,
} from '@carbon/react';
import { Rocket } from '@carbon/icons-react';
import {
  useOnboarding,
  useUpdateOnboarding,
  useSkipOnboarding,
  useCompleteOnboarding,
} from '@/lib/hooks/use-onboarding';
import {
  STEP_META,
  StepOrg,
  StepWorkspace,
  StepTeam,
  StepInviteMember,
  StepConnectCluster,
  StepFirstExperiment,
  StepViewResults,
  QuickSetupPanel,
} from '@/components/onboarding/steps';
import type { OnboardingStepId } from '@/lib/types';
import styles from './onboarding.module.scss';

const STEP_ORDER: OnboardingStepId[] = [
  'org',
  'workspace',
  'team',
  'invite_member',
  'connect_cluster',
  'first_experiment',
  'view_results',
];

function getStepIndex(stepId: OnboardingStepId): number {
  return STEP_ORDER.indexOf(stepId);
}

function StepContent({
  stepId,
  onNext,
  onSkip,
}: {
  stepId: OnboardingStepId;
  onNext: () => void;
  onSkip: () => void;
}) {
  const props = { onNext, onSkip };
  switch (stepId) {
    case 'org': return <StepOrg {...props} />;
    case 'workspace': return <StepWorkspace {...props} />;
    case 'team': return <StepTeam {...props} />;
    case 'invite_member': return <StepInviteMember {...props} />;
    case 'connect_cluster': return <StepConnectCluster {...props} />;
    case 'first_experiment': return <StepFirstExperiment {...props} />;
    case 'view_results': return <StepViewResults onNext={onNext} />;
  }
}

export default function OnboardingPage() {
  const router = useRouter();
  const { data, isLoading, error } = useOnboarding();
  const { mutate: updateOnboarding } = useUpdateOnboarding();
  const { mutate: skipOnboarding, isPending: isSkipping } = useSkipOnboarding();
  const { mutate: completeOnboarding, isPending: isCompleting } = useCompleteOnboarding();

  if (isLoading) {
    return (
      <Grid fullWidth>
        <Column lg={10} md={8} sm={4}>
          <div className={styles.pageHeader}>
            <SkeletonText heading width="40%" />
            <SkeletonText width="60%" />
          </div>
          <SkeletonText paragraph lineCount={6} />
        </Column>
      </Grid>
    );
  }

  if (error) {
    return (
      <Grid fullWidth>
        <Column lg={10} md={8} sm={4}>
          <InlineNotification
            kind="error"
            title="Failed to load onboarding"
            subtitle={(error as Error).message}
            lowContrast
          />
        </Column>
      </Grid>
    );
  }

  if (data?.completed || data?.skipped) {
    router.replace('/');
    return null;
  }

  const currentStepId = data?.currentStep ?? 'org';
  const currentIndex = getStepIndex(currentStepId);
  const steps = data?.steps ?? [];

  function getStepStatus(stepId: OnboardingStepId): 'complete' | 'current' | 'incomplete' {
    const step = steps.find((s) => s.id === stepId);
    if (step?.completed) return 'complete';
    if (stepId === currentStepId) return 'current';
    return 'incomplete';
  }

  function handleNext() {
    const nextIndex = currentIndex + 1;
    if (nextIndex >= STEP_ORDER.length) {
      completeOnboarding(undefined, { onSuccess: () => router.push('/') });
      return;
    }
    const nextStep = STEP_ORDER[nextIndex];
    updateOnboarding({
      stepId: currentStepId,
      stepCompleted: true,
      currentStep: nextStep,
    });
  }

  function handleSkipStep() {
    const nextIndex = currentIndex + 1;
    if (nextIndex >= STEP_ORDER.length) {
      completeOnboarding(undefined, { onSuccess: () => router.push('/') });
      return;
    }
    updateOnboarding({ currentStep: STEP_ORDER[nextIndex] });
  }

  function handleSkipAll() {
    skipOnboarding(undefined, { onSuccess: () => router.push('/') });
  }

  const isLastStep = currentIndex === STEP_ORDER.length - 1;
  const meta = STEP_META[currentStepId];
  const StepIcon = meta.icon;

  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div className={styles.pageHeader}>
          <div className={styles.pageTitleRow}>
            <Rocket size={24} className={styles.pageTitleIcon} />
            <h2 className={styles.pageTitle}>Get Started with ChaosPlane</h2>
          </div>
          <p className={styles.pageSubtitle}>
            Complete these steps to set up your chaos engineering environment.
          </p>
        </div>
      </Column>

      <Column lg={4} md={8} sm={4}>
        <div className={styles.progressPanel}>
          <ProgressIndicator vertical currentIndex={currentIndex} spaceEqually={false}>
            {STEP_ORDER.map((stepId) => {
              const status = getStepStatus(stepId);
              return (
                <ProgressStep
                  key={stepId}
                  label={STEP_META[stepId].label}
                  description={STEP_META[stepId].description}
                  complete={status === 'complete'}
                  current={status === 'current'}
                  invalid={false}
                />
              );
            })}
          </ProgressIndicator>

          <div className={styles.progressActions}>
            <Button
              kind="ghost"
              size="sm"
              onClick={handleSkipAll}
              disabled={isSkipping || isCompleting}
            >
              {isSkipping ? 'Skipping…' : 'Skip all steps'}
            </Button>
          </div>
        </div>
      </Column>

      <Column lg={8} md={8} sm={4}>
        <Tile className={styles.stepTile}>
          <div className={styles.stepHeader}>
            <StepIcon size={24} className={styles.stepIcon} />
            <div>
              <h3 className={styles.stepTitle}>{meta.label}</h3>
              <p className={styles.stepSubtitle}>{meta.description}</p>
            </div>
          </div>

          <div className={styles.stepBody}>
            <StepContent
              stepId={currentStepId}
              onNext={handleNext}
              onSkip={handleSkipStep}
            />
          </div>

          <div className={styles.stepFooter}>
            <span className={styles.stepCount}>
              Step {currentIndex + 1} of {STEP_ORDER.length}
            </span>
            {isLastStep && (
              <Button
                kind="primary"
                onClick={() => completeOnboarding(undefined, { onSuccess: () => router.push('/') })}
                disabled={isCompleting}
              >
                {isCompleting ? 'Finishing…' : 'Complete Setup'}
              </Button>
            )}
          </div>
        </Tile>
      </Column>

      <Column lg={4} md={8} sm={4}>
        <QuickSetupPanel onDone={() => router.push('/')} />
      </Column>
    </Grid>
  );
}
