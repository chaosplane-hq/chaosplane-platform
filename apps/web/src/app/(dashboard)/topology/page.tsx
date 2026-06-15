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
import type { ServiceDependency, TopologyDrift, TopologyMetric } from '@/lib/types';

const depHeaders = [
  { key: 'source', header: 'Source' },
  { key: 'target', header: 'Target' },
  { key: 'protocol', header: 'Protocol' },
  { key: 'port', header: 'Port' },
  { key: 'lastSeenAt', header: 'Last Seen' },
];

const driftHeaders = [
  { key: 'resource', header: 'Resource' },
  { key: 'driftType', header: 'Type' },
  { key: 'namespace', header: 'Namespace' },
  { key: 'detectedAt', header: 'Detected' },
  { key: 'actions', header: '' },
];

const metricHeaders = [
  { key: 'metricName', header: 'Metric' },
  { key: 'metricValue', header: 'Value' },
  { key: 'collectedAt', header: 'Collected' },
];

export default function TopologyPage() {
  const deps = useTopologyDependencies();
  const drifts = useTopologyDrifts();
  const metrics = useTopologyMetrics();
  const acknowledge = useAcknowledgeDrift();

  const depRows = (deps.data?.dependencies ?? []).map((d: ServiceDependency) => ({
    id: d.id,
    source: `${d.sourceName} (${d.sourceNamespace})`,
    target: `${d.targetName} (${d.targetNamespace})`,
    protocol: d.protocol ?? '—',
    port: d.port ?? '—',
    lastSeenAt: d.lastSeenAt ? new Date(d.lastSeenAt).toLocaleString() : '—',
  }));

  const driftRows = (drifts.data?.items ?? []).map((d: TopologyDrift) => ({
    id: d.id,
    resource: `${d.resourceKind}/${d.resourceName}`,
    driftType: d.driftType,
    namespace: d.resourceNamespace,
    detectedAt: new Date(d.detectedAt).toLocaleString(),
    acknowledged: !!d.acknowledgedAt,
  }));

  const metricRows = (metrics.data?.items ?? []).map((m: TopologyMetric, i: number) => ({
    id: String(i),
    metricName: m.metricName,
    metricValue: m.metricValue,
    collectedAt: m.collectedAt ? new Date(m.collectedAt).toLocaleString() : '—',
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
                                <TableCell>{row.cells[3].value as string}</TableCell>
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
