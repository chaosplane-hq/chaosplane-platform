'use client';

import { useState, useEffect } from 'react';
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
import { useResilienceScore, useCalculateResilienceScore } from '@/lib/hooks/use-resilience';
import { useEnvironmentsList } from '@/lib/hooks/use-environments';
import type { ResilienceGrade, ResilienceScore } from '@/lib/types';
import {
  ResilienceDimensionPanel,
  ResilienceTrendPanel,
  ResilienceComparisonPanel,
} from '@/components/metrics/metrics-panels';
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

export default function ResiliencePage() {
  const { data: environments = [], isLoading: envsLoading } = useEnvironmentsList();
  const [environmentId, setEnvironmentId] = useState('');

  useEffect(() => {
    if (!environmentId && environments.length > 0) {
      setEnvironmentId(environments[0].id);
    }
  }, [environments, environmentId]);

  const params = environmentId ? { environmentId } : undefined;
  const { data, isLoading: scoreLoading } = useResilienceScore(params);
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

  const score = data?.current ?? null;
  const history = data?.history ?? [];

  const historyRows = history.map((s: ResilienceScore, i: number) => ({
    id: String(i),
    grade: s.overallGrade,
    score: s.overallScore,
    calculatedAt: formatDate(s.calculatedAt),
  }));

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
              {envsLoading ? (
                <SelectItem value="" text="Loading…" />
              ) : environments.length === 0 ? (
                <SelectItem value="" text="No environments" />
              ) : (
                environments.map((env) => (
                  <SelectItem key={env.id} value={env.id} text={env.name} />
                ))
              )}
            </Select>
            <Button
              renderIcon={Renew}
              onClick={handleCalculate}
              disabled={!environmentId || calculate.isPending}
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
                {score.overallGrade}
              </p>
              <Tag type={GRADE_COLOR[score.overallGrade]} size="lg">
                Score: {score.overallScore}
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
          <ResilienceDimensionPanel score={score} isLoading={scoreLoading} />
        </Tile>
      </Column>

      <Column lg={16} md={8} sm={4} style={{ marginTop: 'var(--cds-spacing-06)' }}>
        <Tile>
          <ResilienceTrendPanel history={history} isLoading={scoreLoading} />
        </Tile>
      </Column>

      <Column lg={16} md={8} sm={4} style={{ marginTop: 'var(--cds-spacing-06)' }}>
        <Tile>
          <ResilienceComparisonPanel history={history} isLoading={scoreLoading} />
        </Tile>
      </Column>

      <Column lg={16} md={8} sm={4} style={{ marginTop: 'var(--cds-spacing-06)' }}>
        <Tile>
          <h3 className={styles.sectionTitle}>Score History</h3>
          {scoreLoading ? (
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
                        const original = history[Number(row.id)];
                        return (
                          <TableRow key={key} {...rowProps}>
                            <TableCell>
                              <Tag type={GRADE_COLOR[original?.overallGrade ?? 'F']} size="sm">
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
