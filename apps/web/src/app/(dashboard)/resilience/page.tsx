'use client';

import { useState } from 'react';
import {
  Grid,
  Column,
  Tile,
  Button,
  Tag,
  Select,
  SelectItem,
  SkeletonText,
  InlineNotification,
  DataTable,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
  TableContainer,
} from '@carbon/react';
import { Renew } from '@carbon/icons-react';
import { useResilienceScore, useResilienceHistory, useCalculateResilienceScore } from '@/lib/hooks/use-resilience';
import type { ResilienceGrade } from '@/lib/types';
import styles from '@/components/experiments/experiments.module.scss';

const GRADE_COLOR: Record<ResilienceGrade, 'green' | 'teal' | 'blue' | 'red' | 'purple'> = {
  A: 'green',
  B: 'teal',
  C: 'blue',
  D: 'red',
  F: 'purple',
};

const historyHeaders = [
  { key: 'grade', header: 'Grade' },
  { key: 'score', header: 'Score' },
  { key: 'calculatedAt', header: 'Calculated At' },
];

function formatDate(iso?: string) {
  if (!iso) return '—';
  return new Date(iso).toLocaleString();
}

const DIMENSION_LABELS: Record<string, string> = {
  availability: 'Availability',
  fault_tolerance: 'Fault Tolerance',
  recoverability: 'Recoverability',
};

export default function ResiliencePage() {
  const [environmentId, setEnvironmentId] = useState('');

  const params = environmentId ? { environmentId } : undefined;
  const { data: score, isLoading: scoreLoading } = useResilienceScore(params);
  const { data: history, isLoading: historyLoading } = useResilienceHistory(params);
  const calculate = useCalculateResilienceScore();

  const [calcError, setCalcError] = useState('');

  async function handleCalculate() {
    setCalcError('');
    try {
      await calculate.mutateAsync(environmentId);
    } catch {
      setCalcError('Failed to recalculate score.');
    }
  }

  const historyRows = (history?.scores ?? []).map((s, i) => ({
    id: String(i),
    grade: s.grade,
    score: s.score,
    calculatedAt: formatDate(s.calculatedAt),
  }));

  const breakdown = score?.breakdown ?? {};
  const dimensionKeys = Object.keys(breakdown).filter((k) => k in DIMENSION_LABELS || true);

  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div className={styles.pageHeader} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <h2 className={styles.pageTitle}>Resilience Score</h2>
            <p className={styles.pageSubtitle}>Track your system resilience across environments.</p>
          </div>
          <div style={{ display: 'flex', gap: 'var(--cds-spacing-04)', alignItems: 'flex-end' }}>
            <Select
              id="env-select"
              labelText="Environment"
              value={environmentId}
              onChange={(e) => setEnvironmentId(e.target.value)}
              style={{ minWidth: 200 }}
            >
              <SelectItem value="" text="All environments" />
              <SelectItem value="production" text="Production" />
              <SelectItem value="staging" text="Staging" />
            </Select>
            <Button
              renderIcon={Renew}
              onClick={handleCalculate}
              disabled={calculate.isPending}
              kind="secondary"
            >
              Recalculate
            </Button>
          </div>
        </div>
      </Column>

      {calcError && (
        <Column lg={16} md={8} sm={4}>
          <InlineNotification
            kind="error"
            title={calcError}
            onCloseButtonClick={() => setCalcError('')}
            style={{ marginBottom: 'var(--cds-spacing-05)' }}
          />
        </Column>
      )}

      <Column lg={4} md={4} sm={4}>
        <Tile className={styles.statTile} style={{ textAlign: 'center', padding: 'var(--cds-spacing-07)' }}>
          {scoreLoading ? (
            <SkeletonText paragraph lineCount={2} />
          ) : score ? (
            <>
              <p
                style={{
                  fontSize: '6rem',
                  fontWeight: 700,
                  lineHeight: 1,
                  margin: '0 0 var(--cds-spacing-04)',
                  color: 'var(--cds-text-primary)',
                }}
              >
                {score.grade}
              </p>
              <Tag type={GRADE_COLOR[score.grade]} size="lg">
                Score: {score.score}
              </Tag>
              <p style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)', marginTop: 'var(--cds-spacing-03)' }}>
                Last calculated: {formatDate(score.calculatedAt)}
              </p>
            </>
          ) : (
            <p style={{ color: 'var(--cds-text-secondary)' }}>No score available.</p>
          )}
        </Tile>
      </Column>

      <Column lg={12} md={4} sm={4}>
        <Tile style={{ height: '100%' }}>
          <h3 className={styles.sectionTitle}>Dimensional Scores</h3>
          {scoreLoading ? (
            <SkeletonText paragraph lineCount={4} />
          ) : dimensionKeys.length === 0 ? (
            <p style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)' }}>
              No dimensional data available.
            </p>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--cds-spacing-05)' }}>
              {dimensionKeys.map((key) => {
                const val = breakdown[key] ?? 0;
                const label = DIMENSION_LABELS[key] ?? key;
                return (
                  <div key={key}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 'var(--cds-spacing-02)' }}>
                      <span style={{ fontSize: 'var(--cds-label-01-font-size)', color: 'var(--cds-text-secondary)' }}>{label}</span>
                      <span style={{ fontSize: 'var(--cds-label-01-font-size)', fontWeight: 600, color: 'var(--cds-text-primary)' }}>
                        {val}%
                      </span>
                    </div>
                    <div style={{ height: 8, background: 'var(--cds-layer-02)', borderRadius: 4, overflow: 'hidden' }}>
                      <div
                        style={{
                          height: '100%',
                          width: `${Math.min(val, 100)}%`,
                          background: val >= 80 ? 'var(--cds-support-success)' : val >= 60 ? 'var(--cds-support-warning)' : 'var(--cds-support-error)',
                          borderRadius: 4,
                          transition: 'width 0.4s ease',
                        }}
                      />
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </Tile>
      </Column>

      <Column lg={16} md={8} sm={4} style={{ marginTop: 'var(--cds-spacing-06)' }}>
        <Tile>
          <h3 className={styles.sectionTitle}>Score History</h3>
          {historyLoading ? (
            <SkeletonText paragraph lineCount={5} />
          ) : (
            <DataTable rows={historyRows} headers={historyHeaders}>
              {({ rows: tableRows, headers: tableHeaders, getTableProps, getHeaderProps, getRowProps }) => (
                <TableContainer>
                  <Table {...getTableProps()}>
                    <TableHead>
                      <TableRow>
                        {tableHeaders.map((h) => {
                          const { key, ...props } = getHeaderProps({ header: h });
                          return <TableHeader key={key} {...props}>{h.header}</TableHeader>;
                        })}
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {tableRows.length === 0 ? (
                        <TableRow>
                          <TableCell colSpan={historyHeaders.length} style={{ textAlign: 'center', color: 'var(--cds-text-secondary)' }}>
                            No history yet.
                          </TableCell>
                        </TableRow>
                      ) : tableRows.map((row) => {
                        const { key, ...rowProps } = getRowProps({ row });
                        const original = history?.scores[Number(row.id)];
                        return (
                          <TableRow key={key} {...rowProps}>
                            <TableCell>
                              <Tag type={GRADE_COLOR[original?.grade ?? 'F']} size="sm">
                                {row.cells[0].value}
                              </Tag>
                            </TableCell>
                            <TableCell>{row.cells[1].value}</TableCell>
                            <TableCell>{row.cells[2].value}</TableCell>
                          </TableRow>
                        );
                      })}
                    </TableBody>
                  </Table>
                </TableContainer>
              )}
            </DataTable>
          )}
        </Tile>
      </Column>
    </Grid>
  );
}
