'use client';

import { useState, useEffect } from 'react';
import {
  Grid,
  Column,
  Tile,
  Button,
  Tag,
  Tabs,
  TabList,
  Tab,
  TabPanels,
  TabPanel,
  SkeletonText,
  InlineNotification,
  Modal,
} from '@carbon/react';
import { ArrowLeft, StopFilled } from '@carbon/icons-react';
import { useRouter } from 'next/navigation';
import { useExperiment, useAbortExperiment, useDeleteExperiment } from '@/lib/hooks/use-experiments';
import { useExperimentWs } from '@/lib/hooks/use-experiment-ws';
import { StatusTag } from '@/components/experiments/status-tag';
import { ExperimentTimeline } from '@/components/experiments/timeline';
import { YamlViewer } from '@/components/experiments/yaml-viewer';
import styles from '@/components/experiments/experiments.module.scss';

interface PageProps {
  params: Promise<{ name: string }>;
}

export default function ExperimentDetailPage({ params }: PageProps) {
  const router = useRouter();
  const [name, setName] = useState('');
  const [abortOpen, setAbortOpen] = useState(false);

  useEffect(() => {
    params.then((p) => setName(p.name));
  }, [params]);

  const { data: experiment, isLoading, error, refetch } = useExperiment(name);
  const abortMutation = useAbortExperiment();
  const deleteMutation = useDeleteExperiment();
  const { lastMessage, connected } = useExperimentWs(name);

  useEffect(() => {
    if (lastMessage) refetch();
  }, [lastMessage, refetch]);

  const canAbort = experiment?.status.phase === 'Running' || experiment?.status.phase === 'Pending';

  if (!name || isLoading) {
    return (
      <Grid fullWidth>
        <Column lg={16} md={8} sm={4}>
          <SkeletonText paragraph lineCount={10} />
        </Column>
      </Grid>
    );
  }

  if (error) {
    return (
      <Grid fullWidth>
        <Column lg={16} md={8} sm={4}>
          <InlineNotification
            kind="error"
            title="Failed to load experiment"
            subtitle={error.message}
          />
        </Column>
      </Grid>
    );
  }

  if (!experiment) return null;

  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--cds-spacing-04)', marginBottom: 'var(--cds-spacing-05)' }}>
          <Button
            kind="ghost"
            size="sm"
            renderIcon={ArrowLeft}
            onClick={() => router.push('/experiments')}
          >
            Back
          </Button>
        </div>
        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', flexWrap: 'wrap', gap: 'var(--cds-spacing-04)', marginBottom: 'var(--cds-spacing-06)' }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--cds-spacing-03)', marginBottom: 'var(--cds-spacing-02)' }}>
              <h2 className={styles.pageTitle} style={{ margin: 0 }}>{experiment.name}</h2>
              <StatusTag phase={experiment.status.phase} />
              <span className={`${styles.wsIndicator} ${connected ? styles.wsConnected : ''}`}>
                <span className={styles.wsDot} />
                {connected ? 'Live' : 'Offline'}
              </span>
            </div>
            <p className={styles.pageSubtitle}>
              {experiment.namespace} · {experiment.action.type}
            </p>
          </div>
          <div style={{ display: 'flex', gap: 'var(--cds-spacing-03)' }}>
            {canAbort && (
              <Button
                kind="danger"
                size="sm"
                renderIcon={StopFilled}
                onClick={() => setAbortOpen(true)}
                disabled={abortMutation.isPending}
              >
                Abort
              </Button>
            )}
          </div>
        </div>
      </Column>

      <Column lg={16} md={8} sm={4}>
        <Tabs>
          <TabList aria-label="Experiment details">
            <Tab>Overview</Tab>
            <Tab>YAML</Tab>
            <Tab>Resources</Tab>
          </TabList>
          <TabPanels>
            <TabPanel>
              <Grid fullWidth style={{ padding: 0 }}>
                <Column lg={8} md={4} sm={4} style={{ marginTop: 'var(--cds-spacing-05)' }}>
                  <Tile>
                    <h3 className={styles.sectionTitle}>Status Timeline</h3>
                    <ExperimentTimeline status={experiment.status} />
                  </Tile>
                </Column>
                <Column lg={8} md={4} sm={4} style={{ marginTop: 'var(--cds-spacing-05)' }}>
                  <Tile>
                    <h3 className={styles.sectionTitle}>Details</h3>
                    <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                      <tbody>
                        {[
                          ['Name', experiment.name],
                          ['Namespace', experiment.namespace],
                          ['Action Type', experiment.action.type],
                          ['Target Namespace', experiment.target.namespace],
                          ['Mode', experiment.target.mode ?? 'one'],
                          ['Duration', experiment.duration ?? '—'],
                        ].map(([label, value]) => (
                          <tr key={label}>
                            <td style={{ padding: 'var(--cds-spacing-03) var(--cds-spacing-03) var(--cds-spacing-03) 0', fontSize: 'var(--cds-label-01-font-size)', color: 'var(--cds-text-secondary)', width: '40%', borderBottom: '1px solid var(--cds-border-subtle)' }}>{label}</td>
                            <td style={{ padding: 'var(--cds-spacing-03)', fontSize: 'var(--cds-label-01-font-size)', color: 'var(--cds-text-primary)', borderBottom: '1px solid var(--cds-border-subtle)' }}>{value}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                    {experiment.status.message && (
                      <div style={{ marginTop: 'var(--cds-spacing-05)' }}>
                        <InlineNotification
                          kind={experiment.status.phase === 'Failed' ? 'error' : 'info'}
                          title="Message"
                          subtitle={experiment.status.message}
                          lowContrast
                        />
                      </div>
                    )}
                  </Tile>
                </Column>
              </Grid>
            </TabPanel>

            <TabPanel>
              <div style={{ marginTop: 'var(--cds-spacing-05)' }}>
                <YamlViewer experiment={experiment} />
              </div>
            </TabPanel>

            <TabPanel>
              <div style={{ marginTop: 'var(--cds-spacing-05)' }}>
                <Tile>
                  <h3 className={styles.sectionTitle}>Affected Resources</h3>
                  {!experiment.status.affectedResources?.length ? (
                    <p style={{ color: 'var(--cds-text-secondary)' }}>No affected resources recorded.</p>
                  ) : (
                    <ul className={styles.resourceList}>
                      {experiment.status.affectedResources.map((r) => (
                        <li key={r} className={styles.resourceItem}>{r}</li>
                      ))}
                    </ul>
                  )}
                </Tile>
              </div>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </Column>

      <Modal
        open={abortOpen}
        danger
        modalHeading="Abort experiment"
        primaryButtonText="Abort"
        secondaryButtonText="Cancel"
        onRequestClose={() => setAbortOpen(false)}
        onRequestSubmit={() => {
          abortMutation.mutate(name, {
            onSuccess: () => setAbortOpen(false),
          });
        }}
      >
        <p>Are you sure you want to abort <strong>{experiment.name}</strong>? This will stop the experiment immediately.</p>
      </Modal>
    </Grid>
  );
}
