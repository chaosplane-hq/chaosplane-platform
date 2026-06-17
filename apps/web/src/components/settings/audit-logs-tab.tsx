'use client';

import { useState, useMemo } from 'react';
import {
  Tile,
  Button,
  Tag,
  TextInput,
  Select,
  SelectItem,
  DataTable,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
  TableContainer,
  SkeletonText,
  InlineNotification,
  StructuredListWrapper,
  StructuredListBody,
  StructuredListRow,
  StructuredListCell,
} from '@carbon/react';
import { Download } from '@carbon/icons-react';
import { useAuditLogs, useAuditExports, useCreateAuditExport } from '@/lib/hooks/use-audit';
import { useQuery } from '@tanstack/react-query';
import { invitationsApi } from '@/lib/api';
import type { AuditLogListParams, TeamMember } from '@/lib/types';

const logHeaders = [
  { key: 'action', header: 'Action' },
  { key: 'resource', header: 'Resource' },
  { key: 'userEmail', header: 'User' },
  { key: 'createdAt', header: 'Timestamp' },
];

function formatDate(iso?: string) {
  if (!iso) return '—';
  return new Date(iso).toLocaleString();
}

export function AuditLogsTab() {
  const [filters, setFilters] = useState<AuditLogListParams>({});
  const [actionFilter, setActionFilter] = useState('');
  const [resourceFilter, setResourceFilter] = useState('');

  const { data, isLoading } = useAuditLogs(filters);
  const { data: exportsData, isLoading: exportsLoading } = useAuditExports();
  const createExport = useCreateAuditExport();
  const { data: membersData } = useQuery({
    queryKey: ['members'],
    queryFn: () => invitationsApi.listMembers(),
    staleTime: 60_000,
  });

  const memberMap = useMemo(() => {
    const map = new Map<string, string>();
    for (const m of membersData?.items ?? []) {
      map.set(m.id, m.name || m.email);
    }
    return map;
  }, [membersData]);

  const [exportError, setExportError] = useState('');

  const logs = data?.items ?? [];
  const exports = exportsData?.items ?? [];

  const rows = logs.map((l) => ({
    id: l.id,
    action: l.action,
    resource: l.resourceId ? `${l.resourceType}/${l.resourceId}` : l.resourceType,
    userEmail: l.userId ? (memberMap.get(l.userId) ?? l.userId) : '—',
    createdAt: formatDate(l.createdAt),
  }));

  function applyFilters() {
    setFilters({
      action: actionFilter.trim() || undefined,
      resource: resourceFilter.trim() || undefined,
    });
  }

  async function handleExport() {
    setExportError('');
    try {
      await createExport.mutateAsync();
    } catch {
      setExportError('Failed to create export.');
    }
  }

  return (
    <div style={{ paddingTop: 'var(--cds-spacing-06)' }}>
      {exportError && (
        <InlineNotification
          kind="error"
          title={exportError}
          onCloseButtonClick={() => setExportError('')}
          style={{ marginBottom: 'var(--cds-spacing-05)' }}
        />
      )}

      <Tile style={{ marginBottom: 'var(--cds-spacing-06)' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--cds-spacing-05)' }}>
          <h3 style={{ fontSize: 'var(--cds-heading-02-font-size)', fontWeight: 'var(--cds-heading-02-font-weight)', color: 'var(--cds-text-primary)', margin: 0 }}>
            Audit Logs
          </h3>
          <Button renderIcon={Download} kind="secondary" onClick={handleExport} disabled={createExport.isPending}>
            Export
          </Button>
        </div>

        <div style={{ display: 'flex', gap: 'var(--cds-spacing-04)', marginBottom: 'var(--cds-spacing-05)', flexWrap: 'wrap' }}>
          <TextInput
            id="audit-action-filter"
            labelText="Filter by action"
            placeholder="e.g. experiment.create"
            value={actionFilter}
            onChange={(e) => setActionFilter(e.target.value)}
            style={{ minWidth: 200 }}
          />
          <TextInput
            id="audit-resource-filter"
            labelText="Filter by resource"
            placeholder="e.g. experiment"
            value={resourceFilter}
            onChange={(e) => setResourceFilter(e.target.value)}
            style={{ minWidth: 200 }}
          />
          <div style={{ display: 'flex', alignItems: 'flex-end' }}>
            <Button kind="secondary" size="md" onClick={applyFilters}>
              Apply
            </Button>
          </div>
        </div>

        {isLoading ? (
          <SkeletonText paragraph lineCount={5} />
        ) : (
          <DataTable rows={rows} headers={logHeaders}>
            {({ rows: tableRows, headers, getTableProps, getHeaderProps, getRowProps }) => (
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
                    {tableRows.length === 0 ? (
                      <TableRow>
                        <TableCell colSpan={logHeaders.length} style={{ textAlign: 'center', color: 'var(--cds-text-secondary)' }}>
                          No audit logs found.
                        </TableCell>
                      </TableRow>
                    ) : tableRows.map((row) => {
                      const { key, ...rowProps } = getRowProps({ row });
                      return (
                        <TableRow key={key} {...rowProps}>
                          <TableCell>
                            <Tag type="gray" size="sm">{row.cells[0].value}</Tag>
                          </TableCell>
                          <TableCell style={{ fontSize: 'var(--cds-label-01-font-size)', fontFamily: 'IBM Plex Mono, monospace' }}>
                            {row.cells[1].value}
                          </TableCell>
                          <TableCell>{row.cells[2].value}</TableCell>
                          <TableCell style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)' }}>
                            {row.cells[3].value}
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
      </Tile>

      <Tile>
        <h3 style={{ fontSize: 'var(--cds-heading-02-font-size)', fontWeight: 'var(--cds-heading-02-font-weight)', color: 'var(--cds-text-primary)', margin: '0 0 var(--cds-spacing-05)' }}>
          Export History
        </h3>
        {exportsLoading ? (
          <SkeletonText paragraph lineCount={3} />
        ) : exports.length === 0 ? (
          <p style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)' }}>No exports yet.</p>
        ) : (
          <StructuredListWrapper>
            <StructuredListBody>
              {exports.map((exp) => (
                <StructuredListRow key={exp.id}>
                  <StructuredListCell>
                    <Tag
                      type={exp.status === 'ready' || exp.status === 'completed' ? 'green' : exp.status === 'failed' ? 'red' : 'blue'}
                      size="sm"
                    >
                      {exp.status}
                    </Tag>
                  </StructuredListCell>
                  <StructuredListCell style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)' }}>
                    {formatDate(exp.createdAt)}
                  </StructuredListCell>
                  <StructuredListCell>
                    <span style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)' }}>
                      {exp.destination}
                    </span>
                  </StructuredListCell>
                </StructuredListRow>
              ))}
            </StructuredListBody>
          </StructuredListWrapper>
        )}
      </Tile>
    </div>
  );
}
