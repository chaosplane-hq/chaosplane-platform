'use client';

import { Grid, Column, Tile, SkeletonText, Link } from '@carbon/react';
import { useExperiments } from '@/lib/hooks/use-experiments';
import { StatusTag } from '@/components/experiments/status-tag';
import styles from '@/components/experiments/experiments.module.scss';
import type { Experiment } from '@/lib/types';
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

export default function DashboardPage() {
  const { data, isLoading } = useExperiments({ limit: 100 });
  const experiments = data?.experiments ?? [];
  const total = data?.total ?? 0;
  const running = experiments.filter((e) => e.status.phase === 'Running').length;
  const recent = experiments.slice(0, 5);

  const actionDist = experiments.reduce<Record<string, number>>((acc, e) => {
    acc[e.action.type] = (acc[e.action.type] ?? 0) + 1;
    return acc;
  }, {});

  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div className={styles.pageHeader}>
          <h2 className={styles.pageTitle}>Dashboard</h2>
          <p className={styles.pageSubtitle}>
            Monitor and manage your chaos experiments.
          </p>
        </div>
      </Column>

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

      <Column lg={10} md={8} sm={4} style={{ marginTop: 'var(--cds-spacing-05)' }}>
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
                    <th key={h} style={{ textAlign: 'left', padding: 'var(--cds-spacing-03)', fontSize: 'var(--cds-label-01-font-size)', color: 'var(--cds-text-secondary)', borderBottom: '1px solid var(--cds-border-subtle)' }}>{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {recent.map((exp) => (
                  <tr key={exp.name}>
                    <td style={{ padding: 'var(--cds-spacing-03)', borderBottom: '1px solid var(--cds-border-subtle)' }}>
                      <Link as={NextLink} href={`/experiments/${exp.name}`}>{exp.name}</Link>
                    </td>
                    <td style={{ padding: 'var(--cds-spacing-03)', fontSize: 'var(--cds-label-01-font-size)', color: 'var(--cds-text-secondary)', borderBottom: '1px solid var(--cds-border-subtle)' }}>{exp.action.type}</td>
                    <td style={{ padding: 'var(--cds-spacing-03)', borderBottom: '1px solid var(--cds-border-subtle)' }}>
                      <StatusTag phase={exp.status.phase} size="sm" />
                    </td>
                    <td style={{ padding: 'var(--cds-spacing-03)', fontSize: 'var(--cds-label-01-font-size)', color: 'var(--cds-text-secondary)', borderBottom: '1px solid var(--cds-border-subtle)' }}>{formatTime(exp.status.startTime)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </Tile>
      </Column>

      <Column lg={6} md={8} sm={4} style={{ marginTop: 'var(--cds-spacing-05)' }}>
        <Tile>
          <h3 className={styles.sectionTitle}>Action Distribution</h3>
          {isLoading ? (
            <SkeletonText paragraph lineCount={4} />
          ) : Object.keys(actionDist).length === 0 ? (
            <p style={{ color: 'var(--cds-text-secondary)' }}>No data yet.</p>
          ) : (
            <div className={styles.distributionRow}>
              {Object.entries(actionDist)
                .sort(([, a], [, b]) => b - a)
                .map(([type, count]) => (
                  <div key={type} className={styles.distributionItem}>
                    <span className={styles.distributionCount}>{count}</span>
                    <span>{type}</span>
                  </div>
                ))}
            </div>
          )}
        </Tile>
      </Column>
    </Grid>
  );
}
