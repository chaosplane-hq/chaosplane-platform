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
  Dropdown,
  SkeletonText,
  InlineNotification,
} from '@carbon/react';
import { Scan, Checkmark, Close, View } from '@carbon/icons-react';
import { useVulnerabilities, useUpdateVulnerabilityStatus, useScanVulnerabilities } from '@/lib/hooks/use-vulnerabilities';
import { useDefaultEnvironmentId } from '@/lib/hooks/use-environments';
import type { VulnerabilitySeverity, VulnerabilityStatus, Vulnerability } from '@/lib/types';

function severityTagType(s: VulnerabilitySeverity): 'red' | 'magenta' | 'teal' | 'blue' {
  if (s === 'critical') return 'red';
  if (s === 'high') return 'magenta';
  if (s === 'medium') return 'teal';
  return 'blue';
}

const SEVERITY_OPTIONS = [
  { id: '', label: 'All severities' },
  { id: 'critical', label: 'Critical' },
  { id: 'high', label: 'High' },
  { id: 'medium', label: 'Medium' },
  { id: 'low', label: 'Low' },
];

const STATUS_OPTIONS = [
  { id: '', label: 'All statuses' },
  { id: 'open', label: 'Open' },
  { id: 'acknowledged', label: 'Acknowledged' },
  { id: 'resolved', label: 'Resolved' },
  { id: 'false_positive', label: 'False Positive' },
];

const headers = [
  { key: 'title', header: 'Title' },
  { key: 'severity', header: 'Severity' },
  { key: 'status', header: 'Status' },
  { key: 'affectedResource', header: 'Affected Resource' },
  { key: 'detectedAt', header: 'Detected' },
  { key: 'actions', header: '' },
];

export default function VulnerabilitiesPage() {
  const [severityFilter, setSeverityFilter] = useState<VulnerabilitySeverity | ''>('');
  const [statusFilter, setStatusFilter] = useState<VulnerabilityStatus | ''>('');

  const { environmentId } = useDefaultEnvironmentId();

  const { data, isLoading, isError } = useVulnerabilities({
    limit: 100,
    severity: severityFilter || undefined,
    status: statusFilter || undefined,
    environmentId,
  });

  const updateStatus = useUpdateVulnerabilityStatus();
  const scan = useScanVulnerabilities();

  const vulns = data?.items ?? [];
  const bySeverity = data?.bySeverity;

  const rows = vulns.map((v: Vulnerability) => ({
    id: v.id,
    title: v.title,
    severity: v.severity,
    status: v.status,
    affectedResource: `${v.resourceKind}/${v.resourceName}`,
    detectedAt: new Date(v.detectedAt).toLocaleString(),
  }));

  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div style={{ padding: '2rem 0 1rem' }}>
          <h2 style={{ fontSize: '1.75rem', fontWeight: 600, marginBottom: '0.25rem' }}>Vulnerabilities</h2>
          <p style={{ color: 'var(--cds-text-secondary)' }}>Security vulnerabilities detected across your environments.</p>
        </div>
      </Column>

      {bySeverity && (
        <>
          {(['critical', 'high', 'medium', 'low'] as const).map((sev) => (
            <Column key={sev} lg={4} md={2} sm={4}>
              <Tile style={{ textAlign: 'center', padding: '1.5rem' }}>
                <p style={{ fontSize: '2rem', fontWeight: 700, color: sev === 'critical' ? 'var(--cds-support-error)' : sev === 'high' ? 'var(--cds-support-warning)' : 'var(--cds-text-primary)' }}>
                  {bySeverity[sev] ?? 0}
                </p>
                <p style={{ textTransform: 'capitalize', color: 'var(--cds-text-secondary)', marginTop: '0.25rem' }}>{sev}</p>
              </Tile>
            </Column>
          ))}
        </>
      )}

      <Column lg={16} md={8} sm={4} style={{ marginTop: '1rem' }}>
        <div style={{ display: 'flex', gap: '1rem', marginBottom: '1rem', alignItems: 'flex-end' }}>
          <div style={{ width: '200px' }}>
            <Dropdown
              id="severity-filter"
              titleText="Severity"
              label="All severities"
              items={SEVERITY_OPTIONS}
              itemToString={(item) => item?.label ?? ''}
              selectedItem={SEVERITY_OPTIONS.find((o) => o.id === severityFilter) ?? SEVERITY_OPTIONS[0]}
              onChange={({ selectedItem }) => setSeverityFilter((selectedItem?.id as VulnerabilitySeverity | '') ?? '')}
            />
          </div>
          <div style={{ width: '200px' }}>
            <Dropdown
              id="status-filter"
              titleText="Status"
              label="All statuses"
              items={STATUS_OPTIONS}
              itemToString={(item) => item?.label ?? ''}
              selectedItem={STATUS_OPTIONS.find((o) => o.id === statusFilter) ?? STATUS_OPTIONS[0]}
              onChange={({ selectedItem }) => setStatusFilter((selectedItem?.id as VulnerabilityStatus | '') ?? '')}
            />
          </div>
        </div>

        {isLoading ? (
          <SkeletonText paragraph lineCount={8} />
        ) : isError ? (
          <InlineNotification kind="error" title="Failed to load vulnerabilities" subtitle="" />
        ) : (
          <DataTable rows={rows} headers={headers}>
            {({ rows: tableRows, headers: tableHeaders, getTableProps, getHeaderProps, getRowProps }) => (
              <TableContainer>
                <TableToolbar>
                  <TableToolbarContent>
                    <Button
                      renderIcon={Scan}
                      kind="primary"
                      disabled={scan.isPending || !environmentId}
                      onClick={() => environmentId && scan.mutate(environmentId)}
                    >
                      {scan.isPending ? 'Scanning…' : 'Trigger Scan'}
                    </Button>
                  </TableToolbarContent>
                </TableToolbar>
                {scan.isError && (
                  <InlineNotification
                    kind="error"
                    title="Operation failed"
                    subtitle={(scan.error as Error)?.message ?? ''}
                    style={{ marginBottom: 'var(--cds-spacing-05)' }}
                  />
                )}
                {updateStatus.isError && (
                  <InlineNotification
                    kind="error"
                    title="Operation failed"
                    subtitle={(updateStatus.error as Error)?.message ?? ''}
                    style={{ marginBottom: 'var(--cds-spacing-05)' }}
                  />
                )}
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
                          No vulnerabilities found.
                        </TableCell>
                      </TableRow>
                    ) : tableRows.map((row) => {
                      const { key, ...rowProps } = getRowProps({ row });
                      const vuln = vulns.find((v) => v.id === row.id);
                      return (
                        <TableRow key={key} {...rowProps}>
                          <TableCell>{row.cells[0].value as string}</TableCell>
                          <TableCell>
                            <Tag type={severityTagType(row.cells[1].value as VulnerabilitySeverity)} size="sm">
                              {row.cells[1].value as string}
                            </Tag>
                          </TableCell>
                          <TableCell>
                            <Tag type={row.cells[2].value === 'open' ? 'red' : row.cells[2].value === 'resolved' ? 'green' : 'gray'} size="sm">
                              {row.cells[2].value as string}
                            </Tag>
                          </TableCell>
                          <TableCell>{row.cells[3].value as string}</TableCell>
                          <TableCell>{row.cells[4].value as string}</TableCell>
                          <TableCell>
                            <div style={{ display: 'flex', gap: '0.25rem' }}>
                              {vuln?.status === 'open' && (
                                <Button
                                  kind="ghost"
                                  size="sm"
                                  renderIcon={View}
                                  disabled={updateStatus.isPending}
                                  onClick={() => updateStatus.mutate({ id: row.id, status: 'acknowledged' })}
                                >
                                  Acknowledge
                                </Button>
                              )}
                              {vuln?.status !== 'resolved' && (
                                <Button
                                  kind="ghost"
                                  size="sm"
                                  renderIcon={Checkmark}
                                  disabled={updateStatus.isPending}
                                  onClick={() => updateStatus.mutate({ id: row.id, status: 'resolved' })}
                                >
                                  Resolve
                                </Button>
                              )}
                              {vuln?.status === 'open' && (
                                <Button
                                  kind="ghost"
                                  size="sm"
                                  renderIcon={Close}
                                  disabled={updateStatus.isPending}
                                  onClick={() => updateStatus.mutate({ id: row.id, status: 'false_positive' })}
                                >
                                  Ignore
                                </Button>
                              )}
                            </div>
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
    </Grid>
  );
}
