'use client';

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
  SkeletonText,
  Tag,
  InlineNotification,
} from '@carbon/react';
import { Add, TrashCan } from '@carbon/icons-react';
import { useRouter } from 'next/navigation';
import { usePolicies, useDeletePolicy } from '@/lib/hooks/use-policies';
import styles from '@/components/experiments/experiments.module.scss';
import type { PolicyEnforcement } from '@/lib/types';

const enforcementToTagType: Record<PolicyEnforcement, 'green' | 'teal' | 'gray' | 'red'> = {
  enforce: 'green',
  audit: 'gray',
  disabled: 'red',
};

function formatTime(iso?: string) {
  if (!iso) return '—';
  return new Date(iso).toLocaleString();
}

const headers = [
  { key: 'name', header: 'Name' },
  { key: 'enforcement', header: 'Enforcement' },
  { key: 'maxConcurrent', header: 'Max Concurrent' },
  { key: 'maxTargets', header: 'Max Targets' },
  { key: 'createdAt', header: 'Created' },
  { key: 'actions', header: '' },
];

export default function PoliciesPage() {
  const router = useRouter();
  const { data, isLoading, isError, error } = usePolicies();
  const deleteMutation = useDeletePolicy();

  const policies = data?.policies ?? [];

  const rows = policies.map((p) => ({
    id: p.id,
    name: p.name,
    enforcement: p.enforcement,
    maxConcurrent: p.maxConcurrent != null ? String(p.maxConcurrent) : '—',
    maxTargets: p.maxTargets != null ? String(p.maxTargets) : '—',
    createdAt: p.createdAt,
  }));

  function handleDelete(id: string, name: string) {
    if (window.confirm(`Delete policy "${name}"? This cannot be undone.`)) {
      deleteMutation.mutate(id);
    }
  }

  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div className={styles.pageHeader}>
          <h2 className={styles.pageTitle}>Policies</h2>
          <p className={styles.pageSubtitle}>
            Define blast radius guardrails for chaos experiments.
          </p>
        </div>
      </Column>

      {isError && (
        <Column lg={16} md={8} sm={4}>
          <InlineNotification
            kind="error"
            title="Failed to load policies"
            subtitle={(error as Error)?.message ?? ''}
            style={{ marginBottom: 'var(--cds-spacing-05)' }}
          />
        </Column>
      )}

      {deleteMutation.isError && (
        <Column lg={16} md={8} sm={4}>
          <InlineNotification
            kind="error"
            title="Operation failed"
            subtitle={(deleteMutation.error as Error)?.message ?? ''}
            style={{ marginBottom: 'var(--cds-spacing-05)' }}
          />
        </Column>
      )}

      <Column lg={16} md={8} sm={4}>
        {isLoading ? (
          <SkeletonText paragraph lineCount={8} />
        ) : (
          <DataTable rows={rows} headers={headers}>
            {({ rows: tableRows, headers: tableHeaders, getTableProps, getHeaderProps, getRowProps }) => (
              <TableContainer>
                <TableToolbar>
                  <TableToolbarContent>
                    <Button
                      renderIcon={Add}
                      onClick={() => router.push('/policies/create')}
                    >
                      Create policy
                    </Button>
                  </TableToolbarContent>
                </TableToolbar>
                <Table {...getTableProps()}>
                  <TableHead>
                    <TableRow>
                      {tableHeaders.map((header) => {
                        const { key, ...headerProps } = getHeaderProps({ header });
                        return (
                          <TableHeader key={key} {...headerProps}>
                            {header.header}
                          </TableHeader>
                        );
                      })}
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {tableRows.length === 0 ? (
                      <TableRow>
                        <TableCell colSpan={headers.length} style={{ textAlign: 'center', color: 'var(--cds-text-secondary)' }}>
                          No policies found.
                        </TableCell>
                      </TableRow>
                    ) : (
                      tableRows.map((row) => {
                        const { key, ...rowProps } = getRowProps({ row });
                        return (
                          <TableRow key={key} {...rowProps}>
                            <TableCell>{row.cells[0].value}</TableCell>
                            <TableCell>
                              <Tag type={enforcementToTagType[row.cells[1].value as PolicyEnforcement]} size="sm">
                                {row.cells[1].value}
                              </Tag>
                            </TableCell>
                            <TableCell>{row.cells[2].value}</TableCell>
                            <TableCell>{row.cells[3].value}</TableCell>
                            <TableCell>{formatTime(row.cells[4].value as string)}</TableCell>
                            <TableCell>
                              <Button
                                kind="ghost"
                                size="sm"
                                hasIconOnly
                                renderIcon={TrashCan}
                                iconDescription="Delete policy"
                                tooltipPosition="left"
                                onClick={() => handleDelete(row.id, row.cells[0].value as string)}
                              />
                            </TableCell>
                          </TableRow>
                        );
                      })
                    )}
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
