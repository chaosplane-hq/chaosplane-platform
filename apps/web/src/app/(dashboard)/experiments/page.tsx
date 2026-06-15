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
  TableToolbarSearch,
  Button,
  Dropdown,
  SkeletonText,
  Link,
} from '@carbon/react';
import { Add } from '@carbon/icons-react';
import NextLink from 'next/link';
import { useRouter } from 'next/navigation';
import { useExperiments } from '@/lib/hooks/use-experiments';
import { StatusTag } from '@/components/experiments/status-tag';
import styles from '@/components/experiments/experiments.module.scss';
import type { ExperimentPhase, ActionType } from '@/lib/types';
import { ACTION_TYPES } from '@/lib/types';

const STATUS_OPTIONS: { id: string; label: string }[] = [
  { id: '', label: 'All statuses' },
  { id: 'Pending', label: 'Pending' },
  { id: 'Running', label: 'Running' },
  { id: 'Completed', label: 'Completed' },
  { id: 'Failed', label: 'Failed' },
  { id: 'Aborted', label: 'Aborted' },
];

const ACTION_OPTIONS: { id: string; label: string }[] = [
  { id: '', label: 'All actions' },
  ...ACTION_TYPES.map((t) => ({ id: t, label: t })),
];

function formatTime(iso?: string) {
  if (!iso) return '—';
  return new Date(iso).toLocaleString();
}

function calcDuration(start?: string, end?: string) {
  if (!start) return '—';
  const diff = (end ? new Date(end).getTime() : Date.now()) - new Date(start).getTime();
  const mins = Math.floor(diff / 60000);
  const secs = Math.floor((diff % 60000) / 1000);
  return mins > 0 ? `${mins}m ${secs}s` : `${secs}s`;
}

const headers = [
  { key: 'name', header: 'Name' },
  { key: 'namespace', header: 'Namespace' },
  { key: 'action', header: 'Action Type' },
  { key: 'status', header: 'Status' },
  { key: 'startTime', header: 'Started' },
  { key: 'duration', header: 'Duration' },
];

export default function ExperimentsPage() {
  const router = useRouter();
  const [statusFilter, setStatusFilter] = useState<ExperimentPhase | ''>('');
  const [actionFilter, setActionFilter] = useState<ActionType | ''>('');
  const [search, setSearch] = useState('');

  const { data, isLoading } = useExperiments({
    limit: 100,
    status: statusFilter || undefined,
    action: actionFilter || undefined,
  });

  const experiments = (data?.experiments ?? [])
    .filter((e) =>
      search
        ? e.name.toLowerCase().includes(search.toLowerCase()) ||
          e.namespace.toLowerCase().includes(search.toLowerCase())
        : true,
    )
    .sort((a, b) => {
      const ta = a.status.startTime ? new Date(a.status.startTime).getTime() : 0;
      const tb = b.status.startTime ? new Date(b.status.startTime).getTime() : 0;
      return tb - ta;
    });

  const rows = experiments.map((e) => ({
    id: e.id,
    name: e.name,
    namespace: e.namespace,
    action: e.action.type,
    status: e.status.phase,
    startTime: e.status.startTime,
    completionTime: e.status.completionTime,
  }));

  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div className={styles.pageHeader}>
          <h2 className={styles.pageTitle}>Experiments</h2>
          <p className={styles.pageSubtitle}>
            View and manage all chaos experiments.
          </p>
        </div>
      </Column>

      <Column lg={16} md={8} sm={4}>
        <div className={styles.filterRow}>
          <div className={styles.filterItem}>
            <Dropdown
              id="status-filter"
              titleText="Status"
              label="All statuses"
              items={STATUS_OPTIONS}
              itemToString={(item) => item?.label ?? ''}
              selectedItem={STATUS_OPTIONS.find((o) => o.id === statusFilter) ?? STATUS_OPTIONS[0]}
              onChange={({ selectedItem }) =>
                setStatusFilter((selectedItem?.id as ExperimentPhase | '') ?? '')
              }
            />
          </div>
          <div className={styles.filterItem}>
            <Dropdown
              id="action-filter"
              titleText="Action Type"
              label="All actions"
              items={ACTION_OPTIONS}
              itemToString={(item) => item?.label ?? ''}
              selectedItem={ACTION_OPTIONS.find((o) => o.id === actionFilter) ?? ACTION_OPTIONS[0]}
              onChange={({ selectedItem }) =>
                setActionFilter((selectedItem?.id as ActionType | '') ?? '')
              }
            />
          </div>
        </div>
      </Column>

      <Column lg={16} md={8} sm={4}>
        {isLoading ? (
          <SkeletonText paragraph lineCount={8} />
        ) : (
          <DataTable rows={rows} headers={headers}>
            {({ rows: tableRows, headers: tableHeaders, getTableProps, getHeaderProps, getRowProps }) => (
              <TableContainer>
                <TableToolbar>
                  <TableToolbarContent>
                    <TableToolbarSearch
                      value={search}
                      onChange={(e) => {
                        if (e && typeof e === 'object' && 'target' in e) {
                          setSearch((e as React.ChangeEvent<HTMLInputElement>).target.value);
                        }
                      }}
                      placeholder="Search by name or namespace"
                    />
                    <Button
                      renderIcon={Add}
                      onClick={() => router.push('/experiments/create')}
                    >
                      Create experiment
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
                          No experiments found.
                        </TableCell>
                      </TableRow>
                    ) : (
                      tableRows.map((row) => {
                        const exp = experiments.find((e) => e.id === row.id);
                        const { key, ...rowProps } = getRowProps({ row });
                        return (
                          <TableRow key={key} {...rowProps}>
                            <TableCell>
                              <Link as={NextLink} href={`/experiments/${row.id}`}>
                                {row.cells[0].value}
                              </Link>
                            </TableCell>
                            <TableCell>{row.cells[1].value}</TableCell>
                            <TableCell>{row.cells[2].value}</TableCell>
                            <TableCell>
                              <StatusTag phase={row.cells[3].value as ExperimentPhase} size="sm" />
                            </TableCell>
                            <TableCell>{formatTime(row.cells[4].value as string)}</TableCell>
                            <TableCell>
                              {calcDuration(exp?.status.startTime, exp?.status.completionTime)}
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
