'use client';

import { useState } from 'react';
import {
  Grid,
  Column,
  Tile,
  SkeletonText,
  Tag,
  Button,
  Modal,
  TextInput,
  InlineLoading,
} from '@carbon/react';
import { Add, Edit } from '@carbon/icons-react';
import {
  useHierarchy,
  useCreateOrganization,
  usePatchOrganization,
  useCreateWorkspace,
  usePatchWorkspace,
  useCreateProject,
  usePatchProject,
  useCreateEnvironment,
  usePatchEnvironment,
} from '@/lib/hooks/use-hierarchy';
import styles from '@/components/experiments/experiments.module.scss';
import type { AgentStatus, Organization, Workspace, Project, Environment } from '@/lib/types';

const AGENT_STATUS_TYPE: Record<AgentStatus, 'green' | 'red' | 'purple'> = {
  connected: 'green',
  disconnected: 'red',
  degraded: 'purple',
};

type ModalMode =
  | { kind: 'create-org' }
  | { kind: 'edit-org'; org: Organization }
  | { kind: 'create-workspace'; orgId: string }
  | { kind: 'edit-workspace'; workspace: Workspace }
  | { kind: 'create-project'; workspaceId: string }
  | { kind: 'edit-project'; project: Project }
  | { kind: 'create-environment'; projectId: string }
  | { kind: 'edit-environment'; environment: Environment };

function modalTitle(mode: ModalMode): string {
  switch (mode.kind) {
    case 'create-org': return 'New Organization';
    case 'edit-org': return `Edit Organization: ${mode.org.name}`;
    case 'create-workspace': return 'New Workspace';
    case 'edit-workspace': return `Edit Workspace: ${mode.workspace.name}`;
    case 'create-project': return 'New Project';
    case 'edit-project': return `Edit Project: ${mode.project.name}`;
    case 'create-environment': return 'New Environment';
    case 'edit-environment': return `Edit Environment: ${mode.environment.name}`;
  }
}

function initialName(mode: ModalMode): string {
  switch (mode.kind) {
    case 'edit-org': return mode.org.name;
    case 'edit-workspace': return mode.workspace.name;
    case 'edit-project': return mode.project.name;
    case 'edit-environment': return mode.environment.name;
    default: return '';
  }
}

export default function EnvironmentsPage() {
  const { data, isLoading } = useHierarchy();
  const [modal, setModal] = useState<ModalMode | null>(null);
  const [nameValue, setNameValue] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const createOrg = useCreateOrganization();
  const patchOrg = usePatchOrganization();
  const createWorkspace = useCreateWorkspace();
  const patchWorkspace = usePatchWorkspace();
  const createProject = useCreateProject();
  const patchProject = usePatchProject();
  const createEnv = useCreateEnvironment();
  const patchEnv = usePatchEnvironment();

  function openModal(mode: ModalMode) {
    setNameValue(initialName(mode));
    setModal(mode);
  }

  function closeModal() {
    setModal(null);
    setNameValue('');
    setSubmitting(false);
  }

  async function handleSubmit() {
    if (!modal || !nameValue.trim()) return;
    setSubmitting(true);
    try {
      switch (modal.kind) {
        case 'create-org':
          await createOrg.mutateAsync({ name: nameValue.trim() });
          break;
        case 'edit-org':
          await patchOrg.mutateAsync({ id: modal.org.id, data: { name: nameValue.trim() } });
          break;
        case 'create-workspace':
          await createWorkspace.mutateAsync({ name: nameValue.trim(), organizationId: modal.orgId });
          break;
        case 'edit-workspace':
          await patchWorkspace.mutateAsync({ id: modal.workspace.id, data: { name: nameValue.trim() } });
          break;
        case 'create-project':
          await createProject.mutateAsync({ name: nameValue.trim(), workspaceId: modal.workspaceId });
          break;
        case 'edit-project':
          await patchProject.mutateAsync({ id: modal.project.id, data: { name: nameValue.trim() } });
          break;
        case 'create-environment':
          await createEnv.mutateAsync({ name: nameValue.trim(), projectId: modal.projectId });
          break;
        case 'edit-environment':
          await patchEnv.mutateAsync({ id: modal.environment.id, data: { name: nameValue.trim() } });
          break;
      }
      closeModal();
    } catch {
      setSubmitting(false);
    }
  }

  const orgs: Organization[] = data?.organizations ?? [];

  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div className={styles.pageHeader}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div>
              <h2 className={styles.pageTitle}>Environments</h2>
              <p className={styles.pageSubtitle}>Manage your organization hierarchy and environments.</p>
            </div>
            <Button
              renderIcon={Add}
              size="md"
              onClick={() => openModal({ kind: 'create-org' })}
            >
              New Organization
            </Button>
          </div>
        </div>
      </Column>

      <Column lg={16} md={8} sm={4}>
        {isLoading ? (
          <Tile>
            <SkeletonText paragraph lineCount={8} />
          </Tile>
        ) : orgs.length === 0 ? (
          <Tile>
            <p style={{ color: 'var(--cds-text-secondary)', textAlign: 'center', padding: 'var(--cds-spacing-09) 0' }}>
              No organizations yet. Create one to get started.
            </p>
          </Tile>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--cds-spacing-05)' }}>
            {orgs.map((org) => (
              <Tile key={org.id}>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 'var(--cds-spacing-04)' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--cds-spacing-03)' }}>
                    <Tag type="blue" size="sm">Org</Tag>
                    <span style={{ fontSize: 'var(--cds-heading-02-font-size)', fontWeight: 'var(--cds-heading-02-font-weight)', color: 'var(--cds-text-primary)' }}>
                      {org.name}
                    </span>
                  </div>
                  <div style={{ display: 'flex', gap: 'var(--cds-spacing-03)' }}>
                    <Button
                      kind="ghost"
                      size="sm"
                      renderIcon={Edit}
                      iconDescription="Edit organization"
                      hasIconOnly
                      onClick={() => openModal({ kind: 'edit-org', org })}
                    />
                    <Button
                      kind="tertiary"
                      size="sm"
                      renderIcon={Add}
                      onClick={() => openModal({ kind: 'create-workspace', orgId: org.id })}
                    >
                      Add Workspace
                    </Button>
                  </div>
                </div>

                {org.workspaces.length === 0 ? (
                  <p style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)', paddingLeft: 'var(--cds-spacing-05)' }}>
                    No workspaces.
                  </p>
                ) : (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--cds-spacing-04)', paddingLeft: 'var(--cds-spacing-05)', borderLeft: '2px solid var(--cds-border-subtle)' }}>
                    {org.workspaces.map((ws) => (
                      <div key={ws.id}>
                        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 'var(--cds-spacing-03)' }}>
                          <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--cds-spacing-03)' }}>
                            <Tag type="teal" size="sm">Workspace</Tag>
                            <span style={{ fontSize: 'var(--cds-body-short-01-font-size)', fontWeight: 600, color: 'var(--cds-text-primary)' }}>
                              {ws.name}
                            </span>
                          </div>
                          <div style={{ display: 'flex', gap: 'var(--cds-spacing-03)' }}>
                            <Button
                              kind="ghost"
                              size="sm"
                              renderIcon={Edit}
                              iconDescription="Edit workspace"
                              hasIconOnly
                              onClick={() => openModal({ kind: 'edit-workspace', workspace: ws })}
                            />
                            <Button
                              kind="ghost"
                              size="sm"
                              renderIcon={Add}
                              onClick={() => openModal({ kind: 'create-project', workspaceId: ws.id })}
                            >
                              Add Project
                            </Button>
                          </div>
                        </div>

                        {ws.projects.length === 0 ? (
                          <p style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)', paddingLeft: 'var(--cds-spacing-05)' }}>
                            No projects.
                          </p>
                        ) : (
                          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--cds-spacing-03)', paddingLeft: 'var(--cds-spacing-05)', borderLeft: '2px solid var(--cds-border-subtle)' }}>
                            {ws.projects.map((proj) => (
                              <div key={proj.id}>
                                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 'var(--cds-spacing-03)' }}>
                                  <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--cds-spacing-03)' }}>
                                    <Tag type="purple" size="sm">Project</Tag>
                                    <span style={{ fontSize: 'var(--cds-body-short-01-font-size)', color: 'var(--cds-text-primary)' }}>
                                      {proj.name}
                                    </span>
                                  </div>
                                  <div style={{ display: 'flex', gap: 'var(--cds-spacing-03)' }}>
                                    <Button
                                      kind="ghost"
                                      size="sm"
                                      renderIcon={Edit}
                                      iconDescription="Edit project"
                                      hasIconOnly
                                      onClick={() => openModal({ kind: 'edit-project', project: proj })}
                                    />
                                    <Button
                                      kind="ghost"
                                      size="sm"
                                      renderIcon={Add}
                                      onClick={() => openModal({ kind: 'create-environment', projectId: proj.id })}
                                    >
                                      Add Environment
                                    </Button>
                                  </div>
                                </div>

                                {proj.environments.length === 0 ? (
                                  <p style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)', paddingLeft: 'var(--cds-spacing-05)' }}>
                                    No environments.
                                  </p>
                                ) : (
                                  <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--cds-spacing-02)', paddingLeft: 'var(--cds-spacing-05)', borderLeft: '2px solid var(--cds-border-subtle)' }}>
                                    {proj.environments.map((env) => (
                                      <div
                                        key={env.id}
                                        style={{
                                          display: 'flex',
                                          alignItems: 'center',
                                          justifyContent: 'space-between',
                                          padding: 'var(--cds-spacing-03) var(--cds-spacing-04)',
                                          background: 'var(--cds-layer-02)',
                                        }}
                                      >
                                        <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--cds-spacing-03)' }}>
                                          <span style={{ fontSize: 'var(--cds-body-short-01-font-size)', color: 'var(--cds-text-primary)' }}>
                                            {env.name}
                                          </span>
                                          <Tag type={AGENT_STATUS_TYPE[env.agentStatus]} size="sm">
                                            {env.agentStatus}
                                          </Tag>
                                        </div>
                                        <Button
                                          kind="ghost"
                                          size="sm"
                                          renderIcon={Edit}
                                          iconDescription="Edit environment"
                                          hasIconOnly
                                          onClick={() => openModal({ kind: 'edit-environment', environment: env })}
                                        />
                                      </div>
                                    ))}
                                  </div>
                                )}
                              </div>
                            ))}
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </Tile>
            ))}
          </div>
        )}
      </Column>

      {modal && (
        <Modal
          open
          modalHeading={modalTitle(modal)}
          primaryButtonText={submitting ? undefined : modal.kind.startsWith('edit') ? 'Save' : 'Create'}
          secondaryButtonText="Cancel"
          onRequestClose={closeModal}
          onRequestSubmit={handleSubmit}
          primaryButtonDisabled={submitting || !nameValue.trim()}
        >
          {submitting ? (
            <InlineLoading description="Saving..." />
          ) : (
            <TextInput
              id="entity-name"
              labelText="Name"
              value={nameValue}
              onChange={(e) => setNameValue(e.target.value)}
              autoFocus
            />
          )}
        </Modal>
      )}
    </Grid>
  );
}
