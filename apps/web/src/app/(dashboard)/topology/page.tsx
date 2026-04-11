'use client';

import { useState } from 'react';
import {
  Grid,
  Column,
  Tabs,
  TabList,
  Tab,
  TabPanels,
  TabPanel,
  DataTable,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
  TableContainer,
  Button,
  Tag,
  SkeletonText,
  InlineNotification,
} from '@carbon/react';
import { Checkmark } from '@carbon/icons-react';
import {
  useTopologyDependencies,
  useTopologyDrifts,
  useTopologyMetrics,
  useAcknowledgeDrift,
} from '@/lib/hooks/use-topology';
import type { DriftSeverity } from '@/lib/types';

function severityType(s: DriftSeverity): 'red' | 'magenta' | 'teal' | 'blue' {
  if (s === 'critical') return 'red';
  if (s === 'high') return 'magenta';
  if (s === 'medium') return 'teal';
  return 'blue';
}

const depHeaders = [
  { key: 'source', header: 'Source' },
  { key: 'target', header: 'Target' },
  { key: 'protocol', header: 'Protocol' },
  { key: 'latencyP99', header: 'Latency P99 (ms)' },
  { key: 'errorRate', header: 'Error Rate (%)' },
  { key: 'requestsPerSecond', header: 'RPS' },
];

const driftHeaders = [
  { key: 'service', header: 'Service' },
  { key: 'type', header: 'Type' },
  { key: 'description', header: 'Description' },
  { key: 'severity', header: 'Severity' },
  { key: 'detectedAt', header: 'Detected' },
  { key: 'actions', header: '' },
];

const metricHeaders = [
  { key: 'service', header: 'Service' },
  { key: 'metric', header: 'Metric' },
  { key: 'value', header: 'Value' },
  { key: 'unit', header: 'Unit' },
  { key: 'trend', header: 'Trend' },
  { key: 'timestamp', header: 'Timestamp' },
];

export default function TopologyPage() {
  const deps = useTopologyDependencies();
  const drifts = useTopologyDrifts();
  const metrics = useTopologyMetrics();
  const acknowledge = useAcknowledgeDrift();

  const depRows = (deps.data?.dependencies ?? []).map((d) => ({
    id: d.id,
    source: d.source,
    target: d.target,
    protocol: d.protocol,
    latencyP99: d.latencyP99 ?? '—',
    errorRate: d.errorRate != null ? `${d.errorRate.toFixed(2)}` : '—',
    requestsPerSecond: d.requestsPerSecond ?? '—',
  }));

  const driftRows = (drifts.data?.drifts ?? []).map((d) => ({
    id: d.id,
    service: d.service,
    type: d.type,
    description: d.description,
    severity: d.severity,
    detectedAt: new Date(d.detectedAt).toLocaleString(),
    acknowledged: !!d.acknowledgedAt,
  }));

  const metricRows = (metrics.data?.metrics ?? []).map((m) => ({
    id: m.id,
    service: m.service,
    metric: m.metric,
    value: m.value,
    unit: m.unit,
    trend: m.trend ?? '—',
    timestamp: new Date(m.timestamp).toLocaleString(),
  }));

  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div style={{ padding: '2rem 0 1rem' }}>
          <h2 style={{ fontSize: '1.75rem', fontWeight: 600, marginBottom: '0.25rem' }}>Topology</h2>
          <p style={{ color: 'var(--cds-text-secondary)' }}>Service dependencies, configuration drifts, and metrics.</p>
        </div>
      </Column>

      <Column lg={16} md={8} sm={4}>
        <Tabs>
          <TabList aria-label="Topology tabs">
            <Tab>Dependencies</Tab>
            <Tab>Drifts</Tab>
            <Tab>Metrics</Tab>
          </TabList>
          <TabPanels>
            <TabPanel>
              {deps.isLoading ? (
                <SkeletonText paragraph lineCount={6} />
              ) : deps.isError ? (
                <InlineNotification kind="error" title="Failed to load dependencies" subtitle="" />
              ) : (
                <DataTable rows={depRows} headers={depHeaders}>
                  {({ rows, headers, getTableProps, getHeaderProps, getRowProps }) => (
                    <TableContainer>
                      <Table {...getTableProps()}>
                        <TableHead>
                          <TableRow>
                            {headers.map((h) => {
                              const { key, ...props } = getHeaderProps({ header: h });
                              return <TableHeader key={key} {...props}>{h.header}</TableHeader>;
                            })}
                          </TableRow>
                        </TableHead>
                        <TableBody>
                          {rows.length === 0 ? (
                            <TableRow>
                              <TableCell colSpan={depHeaders.length} style={{ textAlign: 'center', color: 'var(--cds-text-secondary)' }}>
                                No dependencies found.
                              </TableCell>
                            </TableRow>
                          ) : rows.map((row) => {
                            const { key, ...rowProps } = getRowProps({ row });
                            return (
                              <TableRow key={key} {...rowProps}>
                                {row.cells.map((cell) => (
                                  <TableCell key={cell.id}>{cell.value as string}</TableCell>
                                ))}
                              </TableRow>
                            );
                          })}
                        </TableBody>
                      </Table>
                    </TableContainer>
                  )}
                </DataTable>
              )}
            </TabPanel>

            <TabPanel>
              {drifts.isLoading ? (
                <SkeletonText paragraph lineCount={6} />
              ) : drifts.isError ? (
                <InlineNotification kind="error" title="Failed to load drifts" subtitle="" />
              ) : (
                <DataTable rows={driftRows} headers={driftHeaders}>
                  {({ rows, headers, getTableProps, getHeaderProps, getRowProps }) => (
                    <TableContainer>
                      <Table {...getTableProps()}>
                        <TableHead>
                          <TableRow>
                            {headers.map((h) => {
                              const { key, ...props } = getHeaderProps({ header: h });
                              return <TableHeader key={key} {...props}>{h.header}</TableHeader>;
                            })}
                          </TableRow>
                        </TableHead>
                        <TableBody>
                          {rows.length === 0 ? (
                            <TableRow>
                              <TableCell colSpan={driftHeaders.length} style={{ textAlign: 'center', color: 'var(--cds-text-secondary)' }}>
                                No drifts detected.
                              </TableCell>
                            </TableRow>
                          ) : rows.map((row) => {
                            const drift = driftRows.find((d) => d.id === row.id);
                            const { key, ...rowProps } = getRowProps({ row });
                            return (
                              <TableRow key={key} {...rowProps}>
                                <TableCell>{row.cells[0].value as string}</TableCell>
                                <TableCell>{row.cells[1].value as string}</TableCell>
                                <TableCell>{row.cells[2].value as string}</TableCell>
                                <TableCell>
                                  <Tag type={severityType(row.cells[3].value as DriftSeverity)} size="sm">
                                    {row.cells[3].value as string}
                                  </Tag>
                                </TableCell>
                                <TableCell>{row.cells[4].value as string}</TableCell>
                                <TableCell>
                                  {!drift?.acknowledged && (
                                    <Button
                                      kind="ghost"
                                      size="sm"
                                      renderIcon={Checkmark}
                                      disabled={acknowledge.isPending}
                                      onClick={() => acknowledge.mutate(row.id)}
                                    >
                                      Acknowledge
                                    </Button>
                                  )}
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
            </TabPanel>

            <TabPanel>
              {metrics.isLoading ? (
                <SkeletonText paragraph lineCount={6} />
              ) : metrics.isError ? (
                <InlineNotification kind="error" title="Failed to load metrics" subtitle="" />
              ) : (
                <DataTable rows={metricRows} headers={metricHeaders}>
                  {({ rows, headers, getTableProps, getHeaderProps, getRowProps }) => (
                    <TableContainer>
                      <Table {...getTableProps()}>
                        <TableHead>
                          <TableRow>
                            {headers.map((h) => {
                              const { key, ...props } = getHeaderProps({ header: h });
                              return <TableHeader key={key} {...props}>{h.header}</TableHeader>;
                            })}
                          </TableRow>
                        </TableHead>
                        <TableBody>
                          {rows.length === 0 ? (
                            <TableRow>
                              <TableCell colSpan={metricHeaders.length} style={{ textAlign: 'center', color: 'var(--cds-text-secondary)' }}>
                                No metrics available.
                              </TableCell>
                            </TableRow>
                          ) : rows.map((row) => {
                            const { key, ...rowProps } = getRowProps({ row });
                            return (
                              <TableRow key={key} {...rowProps}>
                                {row.cells.map((cell) => (
                                  <TableCell key={cell.id}>{cell.value as string}</TableCell>
                                ))}
                              </TableRow>
                            );
                          })}
                        </TableBody>
                      </Table>
                    </TableContainer>
                  )}
                </DataTable>
              )}
            </TabPanel>
          </TabPanels>
        </Tabs>
      </Column>
    </Grid>
  );
}
