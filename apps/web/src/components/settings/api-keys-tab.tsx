'use client';

import { useState } from 'react';
import {
  Tile,
  Button,
  TextInput,
  Modal,
  InlineNotification,
  SkeletonText,
  DataTable,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
  TableContainer,
  CodeSnippet,
} from '@carbon/react';
import { Add, Renew, TrashCan } from '@carbon/icons-react';
import { useAPIKeys, useCreateAPIKey, useRotateAPIKey, useRevokeAPIKey } from '@/lib/hooks/use-api-keys';
import type { CreateAPIKeyResponse } from '@/lib/types';

const keyHeaders = [
  { key: 'name', header: 'Name' },
  { key: 'createdAt', header: 'Created' },
  { key: 'lastUsedAt', header: 'Last used' },
  { key: 'expiresAt', header: 'Expires' },
  { key: 'actions', header: '' },
];

function formatDate(iso?: string) {
  if (!iso) return '—';
  return new Date(iso).toLocaleDateString();
}

export function APIKeysTab() {
  const [createOpen, setCreateOpen] = useState(false);
  const [revealData, setRevealData] = useState<CreateAPIKeyResponse | null>(null);
  const [newKeyName, setNewKeyName] = useState('');
  const [errorMsg, setErrorMsg] = useState('');

  const { data, isLoading, isError: queryError, error: queryErrorObj } = useAPIKeys();
  const createKey = useCreateAPIKey();
  const rotateKey = useRotateAPIKey();
  const revokeKey = useRevokeAPIKey();

  const keys = data?.items ?? [];

  async function handleCreate() {
    if (!newKeyName.trim()) return;
    setErrorMsg('');
    try {
      const result = await createKey.mutateAsync({ name: newKeyName.trim() });
      setRevealData(result);
      setCreateOpen(false);
      setNewKeyName('');
    } catch {
      setErrorMsg('Failed to create API key.');
    }
  }

  async function handleRotate(id: string) {
    setErrorMsg('');
    try {
      const result = await rotateKey.mutateAsync(id);
      setRevealData(result);
    } catch {
      setErrorMsg('Failed to rotate API key.');
    }
  }

  const rows = keys.map((k) => ({
    id: k.id,
    name: k.name,
    createdAt: formatDate(k.createdAt),
    lastUsedAt: formatDate(k.lastUsedAt),
    expiresAt: formatDate(k.expiresAt),
  }));

  return (
    <div style={{ paddingTop: 'var(--cds-spacing-06)' }}>
      {queryError && (
        <InlineNotification
          kind="error"
          title="Failed to load API keys"
          subtitle={(queryErrorObj as Error)?.message ?? ''}
          style={{ marginBottom: 'var(--cds-spacing-05)' }}
        />
      )}

      {errorMsg && (
        <InlineNotification
          kind="error"
          title={errorMsg}
          onCloseButtonClick={() => setErrorMsg('')}
          style={{ marginBottom: 'var(--cds-spacing-05)' }}
        />
      )}

      <Modal
        open={createOpen}
        modalHeading="Create API key"
        primaryButtonText="Create"
        secondaryButtonText="Cancel"
        onRequestSubmit={handleCreate}
        onRequestClose={() => { setCreateOpen(false); setNewKeyName(''); }}
        primaryButtonDisabled={!newKeyName.trim() || createKey.isPending}
      >
        <p style={{ color: 'var(--cds-text-secondary)', marginBottom: 'var(--cds-spacing-05)' }}>
          The key will be shown once after creation. Store it securely.
        </p>
        <TextInput
          id="new-key-name"
          labelText="Key name"
          placeholder="e.g. CI pipeline"
          value={newKeyName}
          onChange={(e) => setNewKeyName(e.target.value)}
          onKeyDown={(e) => { if (e.key === 'Enter') handleCreate(); }}
        />
      </Modal>

      <Modal
        open={!!revealData}
        modalHeading="Your new API key"
        primaryButtonText="I've copied it"
        onRequestSubmit={() => setRevealData(null)}
        onRequestClose={() => setRevealData(null)}
        preventCloseOnClickOutside
      >
        <InlineNotification
          kind="warning"
          title="This key will not be shown again."
          subtitle="Copy it now and store it somewhere safe."
          style={{ marginBottom: 'var(--cds-spacing-05)' }}
          hideCloseButton
        />
        <p style={{ color: 'var(--cds-text-secondary)', marginBottom: 'var(--cds-spacing-03)', fontSize: 'var(--cds-label-01-font-size)' }}>
          Key name: <strong style={{ color: 'var(--cds-text-primary)' }}>{revealData?.name}</strong>
        </p>
        <CodeSnippet type="single" feedback="Copied!">
          {revealData?.plaintext ?? ''}
        </CodeSnippet>
      </Modal>

      <Tile>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--cds-spacing-05)' }}>
          <h3 style={{ fontSize: 'var(--cds-heading-02-font-size)', fontWeight: 'var(--cds-heading-02-font-weight)', color: 'var(--cds-text-primary)', margin: 0 }}>
            API keys
          </h3>
          <Button renderIcon={Add} onClick={() => setCreateOpen(true)}>
            Create key
          </Button>
        </div>

        {isLoading ? (
          <SkeletonText paragraph lineCount={4} />
        ) : (
          <DataTable rows={rows} headers={keyHeaders}>
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
                        <TableCell colSpan={keyHeaders.length} style={{ textAlign: 'center', color: 'var(--cds-text-secondary)' }}>
                          No API keys yet.
                        </TableCell>
                      </TableRow>
                    ) : tableRows.map((row) => {
                      const { key, ...rowProps } = getRowProps({ row });
                      return (
                        <TableRow key={key} {...rowProps}>
                          <TableCell>{row.cells[0].value}</TableCell>
                          <TableCell>{row.cells[1].value}</TableCell>
                          <TableCell>{row.cells[2].value}</TableCell>
                          <TableCell>{row.cells[3].value}</TableCell>
                          <TableCell>
                            <div style={{ display: 'flex', gap: 'var(--cds-spacing-02)' }}>
                              <Button
                                kind="ghost"
                                size="sm"
                                renderIcon={Renew}
                                iconDescription="Rotate"
                                hasIconOnly
                                onClick={() => handleRotate(row.id)}
                                disabled={rotateKey.isPending}
                              />
                              <Button
                                kind="danger--ghost"
                                size="sm"
                                renderIcon={TrashCan}
                                iconDescription="Revoke"
                                hasIconOnly
                                onClick={() => revokeKey.mutate(row.id)}
                                disabled={revokeKey.isPending}
                              />
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
      </Tile>
    </div>
  );
}
