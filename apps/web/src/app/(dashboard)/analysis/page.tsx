'use client';

import { useState } from 'react';
import {
  Grid,
  Column,
  DataTable,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
  TableContainer,
  TableToolbar,
  TableToolbarContent,
  Button,
  Tag,
  Tile,
  TextInput,
  Modal,
  SkeletonText,
  InlineNotification,
} from '@carbon/react';
import { Add, ChevronRight } from '@carbon/icons-react';
import { useResultAnalyses, useResultAnalysis, useTriggerAnalysis } from '@/lib/hooks/use-result-analysis';
import { useResilienceScore } from '@/lib/hooks/use-resilience';
import { ResilienceComparisonPanel } from '@/components/metrics/metrics-panels';
import type { ResultAnalysis } from '@/lib/types';

function severityTagType(s?: string): 'blue' | 'green' | 'red' | 'gray' {
  const v = (s ?? '').toLowerCase();
  if (v.includes('critical') || v.includes('high')) return 'red';
  if (v.includes('medium') || v.includes('warning')) return 'blue';
  if (v.includes('low') || v.includes('healthy') || v.includes('ok')) return 'green';
  return 'gray';
}

const headers = [
  { key: 'experimentName', header: 'Experiment' },
  { key: 'severity', header: 'Severity' },
  { key: 'analyzedAt', header: 'Analyzed' },
  { key: 'actions', header: '' },
];

function AnalysisDetail({ id }: { id: string }) {
  const { data, isLoading, isError } = useResultAnalysis(id);
  const envParams = data?.environmentId ? { environmentId: data.environmentId } : undefined;
  const { data: resilience, isLoading: resilienceLoading } = useResilienceScore(envParams);

  if (isLoading) return <SkeletonText paragraph lineCount={6} />;
  if (isError || !data) return <InlineNotification kind="error" title="Failed to load analysis detail" subtitle="" />;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      <div>
        <p style={{ fontWeight: 600, marginBottom: '0.5rem' }}>Summary</p>
        <p style={{ color: 'var(--cds-text-secondary)' }}>{data.summary ?? 'No summary available.'}</p>
      </div>
      <div>
        <p style={{ fontWeight: 600, marginBottom: '0.5rem' }}>Impact</p>
        <p style={{ color: 'var(--cds-text-secondary)' }}>{data.impactAnalysis ?? 'No impact data available.'}</p>
      </div>
      {data.recommendations && (
        <div>
          <p style={{ fontWeight: 600, marginBottom: '0.5rem' }}>Recommendations</p>
          <p style={{ color: 'var(--cds-text-secondary)', whiteSpace: 'pre-wrap' }}>{data.recommendations}</p>
        </div>
      )}
      {data.environmentId && (
        <ResilienceComparisonPanel
          history={resilience?.history ?? []}
          isLoading={resilienceLoading}
        />
      )}
    </div>
  );
}

export default function AnalysisPage() {
  const { data, isLoading, isError } = useResultAnalyses();
  const trigger = useTriggerAnalysis();
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [triggerOpen, setTriggerOpen] = useState(false);
  const [experimentName, setExperimentName] = useState('');

  const analyses = data?.items ?? [];

  const rows = analyses.map((a: ResultAnalysis) => ({
    id: a.id,
    experimentName: a.experimentName,
    severity: a.severityAssessment ?? '—',
    analyzedAt: a.analyzedAt ? new Date(a.analyzedAt).toLocaleString() : '—',
  }));

  function handleTrigger() {
    if (!experimentName.trim()) return;
    trigger.mutate({ experimentName: experimentName.trim() }, {
      onSuccess: () => {
        setTriggerOpen(false);
        setExperimentName('');
      },
    });
  }

  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div style={{ padding: '2rem 0 1rem' }}>
          <h2 style={{ fontSize: '1.75rem', fontWeight: 600, marginBottom: '0.25rem' }}>Analysis</h2>
          <p style={{ color: 'var(--cds-text-secondary)' }}>AI-powered result analysis for your chaos experiments.</p>
        </div>
      </Column>

      <Column lg={selectedId ? 10 : 16} md={8} sm={4}>
        {isLoading ? (
          <SkeletonText paragraph lineCount={8} />
        ) : isError ? (
          <InlineNotification kind="error" title="Failed to load analyses" subtitle="" />
        ) : (
          <DataTable rows={rows} headers={headers}>
            {({ rows: tableRows, headers: tableHeaders, getTableProps, getHeaderProps, getRowProps }) => (
              <TableContainer>
                <TableToolbar>
                  <TableToolbarContent>
                    <Button renderIcon={Add} kind="primary" onClick={() => setTriggerOpen(true)}>
                      Trigger Analysis
                    </Button>
                  </TableToolbarContent>
                </TableToolbar>
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
                        <TableCell colSpan={headers.length} style={{ textAlign: 'center', color: 'var(--cds-text-secondary)' }}>
                          No analyses found.
                        </TableCell>
                      </TableRow>
                    ) : tableRows.map((row) => {
                      const { key, ...rowProps } = getRowProps({ row });
                      return (
                        <TableRow key={key} {...rowProps} style={{ cursor: 'pointer', background: selectedId === row.id ? 'var(--cds-layer-selected)' : undefined }}>
                          <TableCell>{row.cells[0].value as string}</TableCell>
                          <TableCell>
                            <Tag type={severityTagType(row.cells[1].value as string)} size="sm">
                              {row.cells[1].value as string}
                            </Tag>
                          </TableCell>
                          <TableCell>{row.cells[2].value as string}</TableCell>
                          <TableCell>
                            <Button
                              kind="ghost"
                              size="sm"
                              renderIcon={ChevronRight}
                              onClick={() => setSelectedId(selectedId === row.id ? null : row.id)}
                            >
                              {selectedId === row.id ? 'Close' : 'View'}
                            </Button>
                          </TableCell>
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
              </TableContainer>
            )}
          </DataTable>
        )}
      </Column>

      {selectedId && (
        <Column lg={6} md={8} sm={4}>
          <Tile style={{ padding: '1.5rem', height: '100%' }}>
            <h4 style={{ fontWeight: 600, marginBottom: '1rem' }}>Analysis Detail</h4>
            <AnalysisDetail id={selectedId} />
          </Tile>
        </Column>
      )}

      <Modal
        open={triggerOpen}
        modalHeading="Trigger Analysis"
        primaryButtonText={trigger.isPending ? 'Triggering…' : 'Trigger'}
        secondaryButtonText="Cancel"
        onRequestClose={() => setTriggerOpen(false)}
        onRequestSubmit={handleTrigger}
        primaryButtonDisabled={!experimentName.trim() || trigger.isPending}
      >
        <TextInput
          id="experiment-name"
          labelText="Experiment Name"
          placeholder="e.g. pod-kill-frontend"
          value={experimentName}
          onChange={(e) => setExperimentName(e.target.value)}
        />
      </Modal>
    </Grid>
  );
}
