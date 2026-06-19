'use client';

import dynamic from 'next/dynamic';
import { useMemo, useState } from 'react';
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
  Tile,
  Dropdown,
  Tag,
} from '@carbon/react';
import { Checkmark } from '@carbon/icons-react';
import { VizContainer } from '@/components/viz';
import { useDefaultEnvironmentId } from '@/lib/hooks/use-environments';
import { useExperiments } from '@/lib/hooks/use-experiments';
import {
  useTopologyDependencies,
  useTopologyDrifts,
  useTopologyMetrics,
  useAcknowledgeDrift,
} from '@/lib/hooks/use-topology';
import { deriveFaultModel } from '@/lib/topology/fault-model';
import type {
  Experiment,
  ServiceDependency,
  TopologyDrift,
  TopologyMetric,
} from '@/lib/types';

// d3-force measures the DOM, so load it client-only to keep it out of the
// server bundle and avoid hydration drift (matches the viz-smoke pattern).
const ServiceTopologyGraph = dynamic(
  () =>
    import('@/components/topology/service-topology-graph').then(
      (m) => m.ServiceTopologyGraph,
    ),
  { ssr: false },
);

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
  const { environmentId } = useDefaultEnvironmentId();
  const deps = useTopologyDependencies(environmentId);
  const drifts = useTopologyDrifts(environmentId);
  const metrics = useTopologyMetrics(environmentId);
  const acknowledge = useAcknowledgeDrift();
  const experiments = useExperiments({ limit: 50 });

  const dependencies = (deps.data?.dependencies ?? []) as ServiceDependency[];

  const experimentList = useMemo<Experiment[]>(
    () => experiments.data?.experiments ?? [],
    [experiments.data],
  );

  const [selectedExperimentName, setSelectedExperimentName] = useState<string | null>(null);

  // Default to a live experiment so the cascade is visible on first load without
  // a manual pick; fall back to the most recent one otherwise.
  const activeExperiment = useMemo<Experiment | null>(() => {
    if (experimentList.length === 0) return null;
    if (selectedExperimentName) {
      return experimentList.find((e) => e.name === selectedExperimentName) ?? null;
    }
    return (
      experimentList.find((e) => e.status.phase === 'Running') ??
      experimentList.find((e) => e.status.phase === 'Pending') ??
      experimentList[0]
    );
  }, [experimentList, selectedExperimentName]);

  const faultModel = useMemo(
    () => deriveFaultModel(activeExperiment, dependencies),
    [activeExperiment, dependencies],
  );

  const impactedCount = faultModel ? faultModel.nodes.size : 0;

  const depRows = dependencies.map((d) => ({
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

  const graphSummary =
    dependencies.length > 0
      ? `Force-directed service topology graph with ${
          deps.data?.nodeCount ?? '—'
        } services and ${dependencies.length} dependencies. Use the Dependencies tab for an accessible table view.`
      : 'Service topology graph. No dependencies to display.';

  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div style={{ padding: '2rem 0 1rem' }}>
          <h2 style={{ fontSize: '1.75rem', fontWeight: 600, marginBottom: '0.25rem' }}>Topology</h2>
          <p style={{ color: 'var(--cds-text-secondary)' }}>Service dependencies, configuration drifts, and metrics.</p>
        </div>
      </Column>

      <Column lg={16} md={8} sm={4} style={{ marginBottom: '1rem' }}>
        {!environmentId ? (
          <Tile style={{ textAlign: 'center', padding: '3rem' }}>
            <p style={{ color: 'var(--cds-text-secondary)' }}>
              No environment available. Connect an environment to map your service topology.
            </p>
          </Tile>
        ) : deps.isError ? (
          <InlineNotification kind="error" title="Failed to load topology" subtitle="" />
        ) : (
          <>
            {dependencies.length > 0 && experimentList.length > 0 && (
              <div
                style={{
                  display: 'flex',
                  alignItems: 'flex-end',
                  flexWrap: 'wrap',
                  gap: 'var(--cds-spacing-05)',
                  marginBottom: 'var(--cds-spacing-04)',
                }}
              >
                <div style={{ minWidth: '20rem' }}>
                  <Dropdown
                    id="fault-experiment"
                    titleText="Fault propagation"
                    label="Select an experiment"
                    items={experimentList}
                    selectedItem={activeExperiment}
                    itemToString={(e) =>
                      e ? `${e.name} · ${e.action.type} (${e.status.phase})` : ''
                    }
                    onChange={({ selectedItem }) =>
                      setSelectedExperimentName(selectedItem?.name ?? null)
                    }
                  />
                </div>
                {faultModel && (
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 'var(--cds-spacing-03)',
                      paddingBottom: '2px',
                    }}
                  >
                    <Tag type={faultModel.recovered ? 'green' : 'red'} size="sm">
                      {faultModel.recovered ? 'Recovered' : 'Cascading'}
                    </Tag>
                    <span style={{ color: 'var(--cds-text-secondary)', fontSize: '0.75rem' }}>
                      {faultModel.sourceIds.length} source ·{' '}
                      {Math.max(0, impactedCount - faultModel.sourceIds.length)} downstream impacted
                    </span>
                  </div>
                )}
              </div>
            )}
            <VizContainer
              label={graphSummary}
              height={560}
              isLoading={deps.isLoading}
              isEmpty={!deps.isLoading && dependencies.length === 0}
              emptyLabel="No service dependencies discovered yet for this environment."
            >
              {(size) => (
                <ServiceTopologyGraph
                  size={size}
                  dependencies={dependencies}
                  faultModel={faultModel}
                />
              )}
            </VizContainer>
          </>
        )}
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
                <>
                {acknowledge.isError && (
                  <InlineNotification
                    kind="error"
                    title="Operation failed"
                    subtitle={(acknowledge.error as Error)?.message ?? ''}
                    style={{ marginBottom: 'var(--cds-spacing-05)' }}
                  />
                )}
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
                </>
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
