'use client';

import { useState } from 'react';
import {
  Button,
  TextInput,
  InlineNotification,
  Tile,
} from '@carbon/react';
import {
  Building,
  Workspace,
  Group,
  Email,
  Connect,
  Chemistry,
  ChartLine,
  Checkmark,
  Warning,
} from '@carbon/icons-react';
import { useTestAgentConnection, useQuickSetup } from '@/lib/hooks/use-onboarding';
import { useHierarchy, usePatchOrganization, usePatchWorkspace } from '@/lib/hooks/use-hierarchy';
import { useCreateInvitation } from '@/lib/hooks/use-invitations';
import { useCreateAgentToken } from '@/lib/hooks/use-agents';
import { useDefaultEnvironmentId } from '@/lib/hooks/use-environments';
import type { OnboardingStepId } from '@/lib/types';
import styles from './steps.module.scss';

interface StepProps {
  onNext: () => void;
  onSkip?: () => void;
}

export const STEP_META: Record<OnboardingStepId, { label: string; description: string; icon: React.ComponentType<{ size?: number }> }> = {
  org: {
    label: 'Create Organization',
    description: 'Set up your organization to get started',
    icon: Building,
  },
  workspace: {
    label: 'Create Workspace',
    description: 'Create a workspace to organize your experiments',
    icon: Workspace,
  },
  team: {
    label: 'Set Up Team',
    description: 'Configure your team settings',
    icon: Group,
  },
  invite_member: {
    label: 'Invite Members',
    description: 'Invite your teammates to collaborate',
    icon: Email,
  },
  connect_cluster: {
    label: 'Connect Cluster',
    description: 'Connect your Kubernetes cluster to ChaosPlane',
    icon: Connect,
  },
  first_experiment: {
    label: 'First Experiment',
    description: 'Run your first chaos experiment',
    icon: Chemistry,
  },
  view_results: {
    label: 'View Results',
    description: 'Explore experiment results and insights',
    icon: ChartLine,
  },
};

export function StepOrg({ onNext, onSkip }: StepProps) {
  const [name, setName] = useState('');
  const { data } = useHierarchy();
  const patchOrg = usePatchOrganization();

  const org = data?.organizations?.[0];

  const handleContinue = () => {
    if (!org) {
      onNext();
      return;
    }
    patchOrg.mutate(
      { id: org.id, data: { name: name.trim() } },
      { onSuccess: () => onNext() },
    );
  };

  return (
    <div className={styles.step}>
      <p className={styles.desc}>
        Your organization is the top-level container for all your workspaces, teams, and experiments.
      </p>
      <TextInput
        id="org-name"
        labelText="Organization name"
        placeholder="e.g. Acme Corp"
        value={name}
        onChange={(e) => setName(e.target.value)}
      />
      {patchOrg.isError && (
        <InlineNotification
          kind="error"
          title="Failed to update organization"
          subtitle={(patchOrg.error as Error).message}
          lowContrast
        />
      )}
      <div className={styles.actions}>
        <Button onClick={handleContinue} disabled={!name.trim() || patchOrg.isPending}>
          {patchOrg.isPending ? 'Saving…' : 'Continue'}
        </Button>
        {onSkip && <Button kind="ghost" onClick={onSkip}>Skip</Button>}
      </div>
    </div>
  );
}

export function StepWorkspace({ onNext, onSkip }: StepProps) {
  const [name, setName] = useState('');
  const { data } = useHierarchy();
  const patchWorkspace = usePatchWorkspace();

  const workspace = data?.organizations?.[0]?.workspaces?.[0];

  const handleContinue = () => {
    if (!workspace) {
      onNext();
      return;
    }
    patchWorkspace.mutate(
      { id: workspace.id, data: { name: name.trim() } },
      { onSuccess: () => onNext() },
    );
  };

  return (
    <div className={styles.step}>
      <p className={styles.desc}>
        Workspaces let you group experiments by environment, team, or project.
      </p>
      <TextInput
        id="workspace-name"
        labelText="Workspace name"
        placeholder="e.g. production"
        value={name}
        onChange={(e) => setName(e.target.value)}
      />
      {patchWorkspace.isError && (
        <InlineNotification
          kind="error"
          title="Failed to update workspace"
          subtitle={(patchWorkspace.error as Error).message}
          lowContrast
        />
      )}
      <div className={styles.actions}>
        <Button onClick={handleContinue} disabled={!name.trim() || patchWorkspace.isPending}>
          {patchWorkspace.isPending ? 'Saving…' : 'Continue'}
        </Button>
        {onSkip && <Button kind="ghost" onClick={onSkip}>Skip</Button>}
      </div>
    </div>
  );
}

export function StepTeam({ onNext, onSkip }: StepProps) {
  const [name, setName] = useState('');
  return (
    <div className={styles.step}>
      <p className={styles.desc}>
        Teams help you manage access and permissions across your organization.
      </p>
      <TextInput
        id="team-name"
        labelText="Team name"
        placeholder="e.g. Platform Engineering"
        value={name}
        onChange={(e) => setName(e.target.value)}
      />
      <div className={styles.actions}>
        <Button onClick={onNext} disabled={!name.trim()}>Continue</Button>
        {onSkip && <Button kind="ghost" onClick={onSkip}>Skip</Button>}
      </div>
    </div>
  );
}

export function StepInviteMember({ onNext, onSkip }: StepProps) {
  const [email, setEmail] = useState('');
  const createInvitation = useCreateInvitation();
  const { data: hierarchy } = useHierarchy();

  const orgId = hierarchy?.organizations?.[0]?.id;

  const handleContinue = () => {
    if (!email.trim() || !orgId) {
      onNext();
      return;
    }
    createInvitation.mutate(
      { email: email.trim(), organizationId: orgId, role: 'member' },
      { onSuccess: () => onNext() },
    );
  };

  return (
    <div className={styles.step}>
      <p className={styles.desc}>
        Invite a teammate to collaborate on chaos experiments. You can invite more people later from Settings.
      </p>
      <TextInput
        id="invite-email"
        labelText="Email address"
        placeholder="colleague@company.com"
        type="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
      />
      {createInvitation.isError && (
        <InlineNotification
          kind="error"
          title="Failed to send invitation"
          subtitle={(createInvitation.error as Error).message}
          lowContrast
        />
      )}
      <div className={styles.actions}>
        <Button onClick={handleContinue} disabled={createInvitation.isPending}>
          {createInvitation.isPending ? 'Sending…' : email.trim() ? 'Send Invite & Continue' : 'Continue'}
        </Button>
        {onSkip && <Button kind="ghost" onClick={onSkip}>Skip</Button>}
      </div>
    </div>
  );
}

export function StepConnectCluster({ onNext, onSkip }: StepProps) {
  const { mutate: testConnection, isPending, data, error, reset } = useTestAgentConnection();
  const { environmentId } = useDefaultEnvironmentId();
  const createToken = useCreateAgentToken();

  const platformUrl = typeof window !== 'undefined' ? window.location.origin : 'https://app.chaosplane.dev';

  const handleGenerateToken = () => {
    if (!environmentId) return;
    createToken.mutate({ environmentId, name: 'onboarding-agent' });
  };

  const helmCommand = createToken.data?.plaintext
    ? `helm repo add chaosplane https://charts.chaosplane.dev
helm install chaosplane-agent chaosplane/agent \\
  --namespace chaosplane-system --create-namespace \\
  --set agent.token="${createToken.data.plaintext}" \\
  --set agent.platformUrl="${platformUrl}" \\
  --set agent.environmentId="${environmentId}"`
    : `helm repo add chaosplane https://charts.chaosplane.dev
helm install chaosplane-agent chaosplane/agent \\
  --namespace chaosplane-system --create-namespace`;

  return (
    <div className={styles.step}>
      <p className={styles.desc}>
        Install the ChaosPlane agent in your Kubernetes cluster to start running experiments.
      </p>
      {!createToken.data?.plaintext && (
        <div className={styles.actions} style={{ marginBottom: '1rem' }}>
          <Button
            kind="tertiary"
            onClick={handleGenerateToken}
            disabled={createToken.isPending || !environmentId}
          >
            {createToken.isPending ? 'Generating…' : 'Generate Agent Token'}
          </Button>
        </div>
      )}
      {createToken.isError && (
        <InlineNotification
          kind="error"
          title="Failed to generate token"
          subtitle={(createToken.error as Error).message}
          lowContrast
        />
      )}
      {createToken.data?.plaintext && (
        <InlineNotification
          kind="info"
          title="Token generated"
          subtitle="Copy the Helm command below. The token is shown only once."
          lowContrast
          hideCloseButton
          style={{ marginBottom: '1rem' }}
        />
      )}
      <Tile className={styles.codeTile}>
        <p className={styles.codeLabel}>Install via Helm</p>
        <pre className={styles.code}>
          {helmCommand}
        </pre>
      </Tile>
      {data && (
        <InlineNotification
          kind={data.connected ? 'success' : 'error'}
          title={data.connected ? 'Agent connected' : 'Connection failed'}
          subtitle={data.message ?? (data.connected ? `Version: ${data.agentVersion ?? 'unknown'}` : 'Check agent logs and try again.')}
          onClose={reset}
          lowContrast
        />
      )}
      {error && (
        <InlineNotification
          kind="error"
          title="Connection test failed"
          subtitle={(error as Error).message}
          onClose={reset}
          lowContrast
        />
      )}
      <div className={styles.actions}>
        <Button
          kind="secondary"
          onClick={() => testConnection()}
          disabled={isPending}
          renderIcon={isPending ? undefined : Connect}
        >
          {isPending ? 'Testing…' : 'Test Connection'}
        </Button>
        <Button onClick={onNext}>Continue</Button>
        {onSkip && <Button kind="ghost" onClick={onSkip}>Skip</Button>}
      </div>
    </div>
  );
}

export function StepFirstExperiment({ onNext, onSkip }: StepProps) {
  return (
    <div className={styles.step}>
      <p className={styles.desc}>
        Run your first chaos experiment to validate your system&apos;s resilience. Start with a simple pod-kill to see how your services recover.
      </p>
      <div className={styles.experimentCards}>
        {[
          { type: 'pod-kill', label: 'Pod Kill', desc: 'Terminate a random pod and observe recovery' },
          { type: 'network-delay', label: 'Network Delay', desc: 'Inject latency between services' },
          { type: 'pod-cpu-stress', label: 'CPU Stress', desc: 'Stress CPU on target pods' },
        ].map((exp) => (
          <Tile key={exp.type} className={styles.expCard}>
            <p className={styles.expType}>{exp.label}</p>
            <p className={styles.expDesc}>{exp.desc}</p>
          </Tile>
        ))}
      </div>
      <div className={styles.actions}>
        <Button renderIcon={Chemistry} href="/experiments/create">Create Experiment</Button>
        <Button kind="secondary" onClick={onNext}>I&apos;ll do this later</Button>
        {onSkip && <Button kind="ghost" onClick={onSkip}>Skip</Button>}
      </div>
    </div>
  );
}

export function StepViewResults({ onNext }: StepProps) {
  return (
    <div className={styles.step}>
      <p className={styles.desc}>
        Once experiments run, you can view detailed results, timelines, and affected resources from the Experiments dashboard.
      </p>
      <div className={styles.checklist}>
        {[
          'Real-time experiment status via WebSocket',
          'Affected resources and pod logs',
          'Success rate and duration metrics',
          'Export results as YAML or JSON',
        ].map((item) => (
          <div key={item} className={styles.checkItem}>
            <Checkmark size={16} className={styles.checkIcon} />
            <span>{item}</span>
          </div>
        ))}
      </div>
      <div className={styles.actions}>
        <Button renderIcon={ChartLine} href="/experiments">Go to Experiments</Button>
        <Button kind="secondary" onClick={onNext}>Finish Setup</Button>
      </div>
    </div>
  );
}

export function QuickSetupPanel({ onDone }: { onDone: () => void }) {
  const { mutate: quickSetup, isPending, data, error } = useQuickSetup();

  return (
    <Tile className={styles.quickSetup}>
      <div className={styles.quickSetupHeader}>
        <Warning size={20} />
        <p className={styles.quickSetupTitle}>Quick Setup</p>
      </div>
      <p className={styles.quickSetupDesc}>
        Skip the wizard and auto-configure a default organization, workspace, and team in one click.
      </p>
      {data?.success && (
        <InlineNotification
          kind="success"
          title="Quick setup complete"
          subtitle={data.message ?? 'Your environment is ready.'}
          lowContrast
        />
      )}
      {error && (
        <InlineNotification
          kind="error"
          title="Quick setup failed"
          subtitle={(error as Error).message}
          lowContrast
        />
      )}
      <div className={styles.quickSetupActions}>
        <Button
          kind="secondary"
          size="sm"
          onClick={() => quickSetup(undefined, { onSuccess: (d) => d.success && onDone() })}
          disabled={isPending}
        >
          {isPending ? 'Setting up…' : 'Run Quick Setup'}
        </Button>
      </div>
    </Tile>
  );
}
