'use client';

import { Grid, Column, Tile, SkeletonText, Tag, Link, InlineNotification, Button } from '@carbon/react';
import { useExperiments } from '@/lib/hooks/use-experiments';
import { useOnboarding } from '@/lib/hooks/use-onboarding';
import { useResilienceScore } from '@/lib/hooks/use-resilience';
import { useVulnerabilities } from '@/lib/hooks/use-vulnerabilities';
import { useSuggestions } from '@/lib/hooks/use-suggestions';
import { StatusTag } from '@/components/experiments/status-tag';
import styles from '@/components/experiments/experiments.module.scss';
import type { Experiment, ResilienceGrade, VulnerabilitySeverity } from '@/lib/types';
import NextLink from 'next/link';

function formatTime(iso?: string) {
  if (!iso) return '—';
  return new Date(iso).toLocaleString();
}

function calcSuccessRate(experiments: Experiment[]) {
  const terminal = experiments.filter((e) =>
    ['Completed', 'Failed', 'Aborted'].includes(e.status.phase),
  );
  if (!terminal.length) return '—';
  const completed = terminal.filter((e) => e.status.phase === 'Completed').length;
  return `${Math.round((completed / terminal.length) * 100)}%`;
}

function calcAvgDuration(experiments: Experiment[]) {
  const withDuration = experiments.filter(
    (e) => e.status.startTime && e.status.completionTime,
  );
  if (!withDuration.length) return '—';
  const avg =
    withDuration.reduce((sum, e) => {
      const diff =
        new Date(e.status.completionTime!).getTime() -
        new Date(e.status.startTime!).getTime();
      return sum + diff;
    }, 0) / withDuration.length;
  const mins = Math.floor(avg / 60000);
  const secs = Math.floor((avg % 60000) / 1000);
  return mins > 0 ? `${mins}m ${secs}s` : `${secs}s`;
}

const GRADE_COLORS: Record<ResilienceGrade, string> = {
  A: 'var(--cds-support-success)',
  B: 'var(--cds-support-success)',
  C: 'var(--cds-support-warning)',
  D: 'var(--cds-support-error)',
  F: 'var(--cds-support-error)',
};

const SEVERITY_TYPE: Record<VulnerabilitySeverity, 'red' | 'purple' | 'gray' | 'blue' | 'cyan'> = {
  critical: 'red',
  high: 'red',
  medium: 'purple',
  low: 'gray',
  info: 'blue',
};

export default function DashboardPage() {
  const { data, isLoading } = useExperiments({ limit: 100 });
  const { data: onboarding } = useOnboarding();
  const { data: resilience, isLoading: resilienceLoading } = useResilienceScore();
  const { data: vulnData, isLoading: vulnLoading } = useVulnerabilities({ limit: 5 });
  const { data: suggestionsData, isLoading: suggestionsLoading } = useSuggestions({ limit: 3 });

  const experiments = data?.experiments ?? [];
  const total = data?.total ?? 0;
  const running = experiments.filter((e) => e.status.phase === 'Running').length;
  const recent = experiments.slice(0, 5);
  const vulnerabilities = vulnData?.items ?? [];
  const suggestions = suggestionsData?.items ?? [];

  const showOnboarding = onboarding && !onboarding.completed && !onboarding.skipped;
  const completedSteps = onboarding?.steps.filter((s) => s.completed).length ?? 0;
  const totalSteps = onboarding?.steps.length ?? 0;

  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div className={styles.pageHeader}>
          <h2 className={styles.pageTitle}>Dashboard</h2>
          <p className={styles.pageSubtitle}>Monitor and manage your chaos experiments.</p>
        </div>
      </Column>

      {showOnboarding && (
        <Column lg={16} md={8} sm={4} style={{ marginBottom: 'var(--cds-spacing-05)' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--cds-spacing-04)' }}>
            <InlineNotification
              kind="info"
              title={`Getting started — ${completedSteps}/${totalSteps} steps completed`}
              subtitle="Complete setup to unlock the full platform."
              lowContrast
              style={{ flex: 1 }}
            />
            <Button kind="tertiary" size="sm" as={NextLink} href="/onboarding">
              Continue setup
            </Button>
          </div>
        </Column>
      )}

      <Column lg={4} md={4} sm={4}>
        <Tile className={styles.statTile}>
          <p className={styles.statLabel}>Total Experiments</p>
          {isLoading ? <SkeletonText /> : <p className={styles.statValue}>{total}</p>}
        </Tile>
      </Column>

      <Column lg={4} md={4} sm={4}>
        <Tile className={styles.statTile}>
          <p className={styles.statLabel}>Running</p>
          {isLoading ? <SkeletonText /> : <p className={styles.statValue}>{running}</p>}
        </Tile>
      </Column>

      <Column lg={4} md={4} sm={4}>
        <Tile className={styles.statTile}>
          <p className={styles.statLabel}>Success Rate</p>
          {isLoading ? <SkeletonText /> : <p className={styles.statValue}>{calcSuccessRate(experiments)}</p>}
        </Tile>
      </Column>

      <Column lg={4} md={4} sm={4}>
        <Tile className={styles.statTile}>
          <p className={styles.statLabel}>Avg Duration</p>
          {isLoading ? <SkeletonText /> : <p className={styles.statValue}>{calcAvgDuration(experiments)}</p>}
        </Tile>
      </Column>

      <Column lg={4} md={4} sm={4} style={{ marginTop: 'var(--cds-spacing-05)' }}>
        <Tile className={styles.statTile}>
          <p className={styles.statLabel}>Resilience Score</p>
          {resilienceLoading ? (
            <SkeletonText />
          ) : resilience?.current ? (
            <div style={{ display: 'flex', alignItems: 'baseline', gap: 'var(--cds-spacing-03)' }}>
              <p
                className={styles.statValue}
                style={{ color: GRADE_COLORS[resilience.current.overallGrade] }}
              >
                {resilience.current.overallGrade}
              </p>
              <span style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)' }}>
                {resilience.current.overallScore}/100
              </span>
            </div>
          ) : (
            <p className={styles.statValue} style={{ color: 'var(--cds-text-secondary)' }}>—</p>
          )}
        </Tile>
      </Column>

      <Column lg={6} md={8} sm={4} style={{ marginTop: 'var(--cds-spacing-05)' }}>
        <Tile style={{ height: '100%' }}>
          <h3 className={styles.sectionTitle}>Recent Vulnerabilities</h3>
          {vulnLoading ? (
            <SkeletonText paragraph lineCount={4} />
          ) : vulnerabilities.length === 0 ? (
            <p style={{ color: 'var(--cds-text-secondary)' }}>No vulnerabilities detected.</p>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--cds-spacing-03)' }}>
              {vulnerabilities.map((v) => (
                <div
                  key={v.id}
                  style={{
                    display: 'flex',
                    alignItems: 'flex-start',
                    justifyContent: 'space-between',
                    gap: 'var(--cds-spacing-03)',
                    padding: 'var(--cds-spacing-03)',
                    background: 'var(--cds-layer-02)',
                  }}
                >
                  <span style={{ fontSize: 'var(--cds-body-short-01-font-size)', color: 'var(--cds-text-primary)', flex: 1 }}>
                    {v.title}
                  </span>
                  <Tag type={SEVERITY_TYPE[v.severity]} size="sm">
                    {v.severity}
                  </Tag>
                </div>
              ))}
            </div>
          )}
        </Tile>
      </Column>

      <Column lg={6} md={8} sm={4} style={{ marginTop: 'var(--cds-spacing-05)' }}>
        <Tile style={{ height: '100%' }}>
          <h3 className={styles.sectionTitle}>AI Suggestions</h3>
          {suggestionsLoading ? (
            <SkeletonText paragraph lineCount={4} />
          ) : suggestions.length === 0 ? (
            <p style={{ color: 'var(--cds-text-secondary)' }}>No suggestions available.</p>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--cds-spacing-04)' }}>
              {suggestions.map((s) => (
                <div
                  key={s.id}
                  style={{
                    padding: 'var(--cds-spacing-04)',
                    background: 'var(--cds-layer-02)',
                    borderLeft: '3px solid var(--cds-interactive)',
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--cds-spacing-03)', marginBottom: 'var(--cds-spacing-02)' }}>
                    <span style={{ fontSize: 'var(--cds-body-short-01-font-size)', fontWeight: 600, color: 'var(--cds-text-primary)' }}>
                      {s.title}
                    </span>
                    <Tag type="blue" size="sm">
                      {Math.round(s.confidence * 100)}% confidence
                    </Tag>
                  </div>
                  <p style={{ margin: 0, fontSize: 'var(--cds-label-01-font-size)', color: 'var(--cds-text-secondary)' }}>
                    {s.description}
                  </p>
                </div>
              ))}
            </div>
          )}
        </Tile>
      </Column>

      <Column lg={16} md={8} sm={4} style={{ marginTop: 'var(--cds-spacing-05)' }}>
        <Tile>
          <h3 className={styles.sectionTitle}>Recent Experiments</h3>
          {isLoading ? (
            <SkeletonText paragraph lineCount={5} />
          ) : recent.length === 0 ? (
            <p style={{ color: 'var(--cds-text-secondary)', textAlign: 'center', padding: 'var(--cds-spacing-07) 0' }}>
              No experiments yet.
            </p>
          ) : (
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr>
                  {['Name', 'Action', 'Status', 'Started'].map((h) => (
                    <th
                      key={h}
                      style={{
                        textAlign: 'left',
                        padding: 'var(--cds-spacing-03)',
                        fontSize: 'var(--cds-label-01-font-size)',
                        color: 'var(--cds-text-secondary)',
                        borderBottom: '1px solid var(--cds-border-subtle)',
                      }}
                    >
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {recent.map((exp) => (
                  <tr key={exp.name}>
                    <td style={{ padding: 'var(--cds-spacing-03)', borderBottom: '1px solid var(--cds-border-subtle)' }}>
                      <Link as={NextLink} href={`/experiments/${exp.name}`}>{exp.name}</Link>
                    </td>
                    <td style={{ padding: 'var(--cds-spacing-03)', fontSize: 'var(--cds-label-01-font-size)', color: 'var(--cds-text-secondary)', borderBottom: '1px solid var(--cds-border-subtle)' }}>
                      {exp.action.type}
                    </td>
                    <td style={{ padding: 'var(--cds-spacing-03)', borderBottom: '1px solid var(--cds-border-subtle)' }}>
                      <StatusTag phase={exp.status.phase} size="sm" />
                    </td>
                    <td style={{ padding: 'var(--cds-spacing-03)', fontSize: 'var(--cds-label-01-font-size)', color: 'var(--cds-text-secondary)', borderBottom: '1px solid var(--cds-border-subtle)' }}>
                      {formatTime(exp.status.startTime)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </Tile>
      </Column>
    </Grid>
  );
}
