'use client';

import { useState } from 'react';
import {
  Grid,
  Column,
  Button,
  Tag,
  Modal,
  TextInput,
  TextArea,
  DatePicker,
  DatePickerInput,
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
} from '@carbon/react';
import { Add } from '@carbon/icons-react';
import { useRouter } from 'next/navigation';
import { useGameDays, useCreateGameDay } from '@/lib/hooks/use-gamedays';
import { useDefaultEnvironmentId } from '@/lib/hooks/use-environments';
import type { GameDayStatus, GameDay } from '@/lib/types';
import styles from '@/components/experiments/experiments.module.scss';

const STATUS_TAG: Record<GameDayStatus, 'blue' | 'green' | 'gray' | 'red'> = {
  planned: 'blue',
  in_progress: 'green',
  completed: 'gray',
  cancelled: 'red',
};

const headers = [
  { key: 'title', header: 'Title' },
  { key: 'status', header: 'Status' },
  { key: 'scheduledAt', header: 'Scheduled' },
  { key: 'createdAt', header: 'Created' },
];

function formatDate(iso?: string) {
  if (!iso) return '—';
  return new Date(iso).toLocaleDateString();
}

export default function GameDaysPage() {
  const router = useRouter();
  const { data, isLoading, isError: queryError, error: queryErrorObj } = useGameDays();
  const createGameDay = useCreateGameDay();
  const { environmentId } = useDefaultEnvironmentId();

  const [createOpen, setCreateOpen] = useState(false);
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [scheduledAt, setScheduledAt] = useState('');
  const [error, setError] = useState('');

  const gameDays = data?.items ?? [];

  const rows = gameDays.map((g: GameDay) => ({
    id: g.id,
    title: g.title,
    status: g.status,
    scheduledAt: formatDate(g.scheduledAt),
    createdAt: formatDate(g.createdAt),
  }));

  async function handleCreate() {
    if (!title.trim() || !scheduledAt) return;
    if (!environmentId) {
      setError('No environment available. Create an environment first.');
      return;
    }
    setError('');
    try {
      await createGameDay.mutateAsync({
        environmentId,
        title: title.trim(),
        description: description.trim() || undefined,
        scheduledAt,
      });
      setCreateOpen(false);
      setTitle('');
      setDescription('');
      setScheduledAt('');
    } catch {
      setError('Failed to create game day.');
    }
  }

  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div className={styles.pageHeader}>
          <h2 className={styles.pageTitle}>Game Days</h2>
          <p className={styles.pageSubtitle}>Plan, run, and review chaos game days.</p>
        </div>
      </Column>

      <Column lg={16} md={8} sm={4}>
        {queryError && (
          <InlineNotification
            kind="error"
            title="Failed to load game days"
            subtitle={(queryErrorObj as Error)?.message ?? ''}
            style={{ marginBottom: 'var(--cds-spacing-05)' }}
          />
        )}

        {error && (
          <InlineNotification
            kind="error"
            title={error}
            onCloseButtonClick={() => setError('')}
            style={{ marginBottom: 'var(--cds-spacing-05)' }}
          />
        )}

        <Modal
          open={createOpen}
          modalHeading="Create Game Day"
          primaryButtonText="Create"
          secondaryButtonText="Cancel"
          onRequestSubmit={handleCreate}
          onRequestClose={() => { setCreateOpen(false); setTitle(''); setDescription(''); setScheduledAt(''); }}
          primaryButtonDisabled={!title.trim() || !scheduledAt || createGameDay.isPending}
        >
          <TextInput
            id="gd-title"
            labelText="Title"
            placeholder="e.g. Q2 Resilience GameDay"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            style={{ marginBottom: 'var(--cds-spacing-05)' }}
          />
          <TextArea
            id="gd-description"
            labelText="Description (optional)"
            placeholder="What are the goals of this game day?"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={3}
            style={{ marginBottom: 'var(--cds-spacing-05)' }}
          />
          <DatePicker
            datePickerType="single"
            onChange={(dates) => setScheduledAt(dates[0]?.toISOString() ?? '')}
          >
            <DatePickerInput
              id="gd-scheduled"
              labelText="Scheduled date"
              placeholder="mm/dd/yyyy"
            />
          </DatePicker>
        </Modal>

        <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 'var(--cds-spacing-05)' }}>
          <Button renderIcon={Add} onClick={() => setCreateOpen(true)}>
            Create Game Day
          </Button>
        </div>

        {isLoading ? (
          <SkeletonText paragraph lineCount={5} />
        ) : (
          <DataTable rows={rows} headers={headers}>
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
                        <TableCell colSpan={headers.length} style={{ textAlign: 'center', color: 'var(--cds-text-secondary)' }}>
                          No game days yet.
                        </TableCell>
                      </TableRow>
                    ) : tableRows.map((row) => {
                      const { key, ...rowProps } = getRowProps({ row });
                      const original = gameDays.find((g) => g.id === row.id);
                      return (
                        <TableRow
                          key={key}
                          {...rowProps}
                          style={{ cursor: 'pointer' }}
                          onClick={() => router.push(`/gamedays/${row.id}`)}
                        >
                          <TableCell>{row.cells[0].value}</TableCell>
                          <TableCell>
                            <Tag type={STATUS_TAG[original?.status ?? 'planned']} size="sm">
                              {row.cells[1].value}
                            </Tag>
                          </TableCell>
                          <TableCell>{row.cells[2].value}</TableCell>
                          <TableCell>{row.cells[3].value}</TableCell>
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
