'use client';

import { useState } from 'react';
import {
  Button,
  DataTable,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
  TableToolbar,
  TableToolbarContent,
  Tag,
  Modal,
  TextInput,
  Select,
  SelectItem,
  TextArea,
  SkeletonText,
  InlineNotification,
} from '@carbon/react';
import { Add, TrashCan } from '@carbon/icons-react';
import { useCICDIntegrations, useCreateCICDIntegration, useDeleteCICDIntegration } from '@/lib/hooks/use-cicd';
import type { CICDProvider, CICDIntegration } from '@/lib/types';

const PROVIDERS: { id: CICDProvider; label: string }[] = [
  { id: 'github_actions', label: 'GitHub Actions' },
  { id: 'gitlab_ci', label: 'GitLab CI' },
  { id: 'jenkins', label: 'Jenkins' },
];

const headers = [
  { key: 'provider', header: 'Provider' },
  { key: 'name', header: 'Name' },
  { key: 'enabled', header: 'Status' },
  { key: 'createdAt', header: 'Created' },
  { key: 'actions', header: '' },
];

export function CICDTab() {
  const { data, isLoading, isError, error } = useCICDIntegrations();
  const createMutation = useCreateCICDIntegration();
  const deleteMutation = useDeleteCICDIntegration();

  const [modalOpen, setModalOpen] = useState(false);
  const [name, setName] = useState('');
  const [provider, setProvider] = useState<CICDProvider>('github_actions');
  const [configJson, setConfigJson] = useState('{}');

  function handleCreate() {
    let config: Record<string, string> = {};
    try { config = JSON.parse(configJson); } catch { config = {}; }
    createMutation.mutate(
      { name, provider, config },
      {
        onSuccess: () => {
          setModalOpen(false);
          setName(''); setConfigJson('{}');
        },
      },
    );
  }

  const rows = (data?.items ?? []).map((i: CICDIntegration) => ({ ...i, id: i.id }));

  return (
    <div style={{ marginTop: 'var(--cds-spacing-05)' }}>
      {isError && (
        <InlineNotification
          kind="error"
          title="Failed to load integrations"
          subtitle={(error as Error)?.message}
          style={{ marginBottom: 'var(--cds-spacing-05)' }}
        />
      )}

      {isLoading ? (
        <SkeletonText paragraph />
      ) : (
        <DataTable rows={rows} headers={headers}>
          {({ rows: tableRows, headers: tableHeaders, getTableProps, getHeaderProps, getRowProps }) => (
            <>
              <TableToolbar>
                <TableToolbarContent>
                  <Button renderIcon={Add} onClick={() => setModalOpen(true)}>
                    Add Integration
                  </Button>
                </TableToolbarContent>
              </TableToolbar>
              <Table {...getTableProps()}>
                <TableHead>
                  <TableRow>
                     {tableHeaders.map((h) => {
                        const { key, ...headerProps } = getHeaderProps({ header: h });
                        return <TableHeader key={h.key} {...headerProps}>{h.header}</TableHeader>;
                      })}
                  </TableRow>
                </TableHead>
                <TableBody>
                    {tableRows.map((row) => {
                      const integration = (data?.items ?? []).find((i: CICDIntegration) => i.id === row.id);
                      const { key: rowKey, ...rowProps } = getRowProps({ row });
                      return (
                        <TableRow key={row.id} {...rowProps}>
                        {row.cells.map((cell) => {
                          if (cell.info.header === 'provider') {
                            const label = PROVIDERS.find((p) => p.id === cell.value)?.label ?? cell.value;
                            return <TableCell key={cell.id}>{label}</TableCell>;
                          }
                          if (cell.info.header === 'enabled') {
                            return (
                              <TableCell key={cell.id}>
                                <Tag type={cell.value ? 'green' : 'gray'} size="sm">
                                  {cell.value ? 'Enabled' : 'Disabled'}
                                </Tag>
                              </TableCell>
                            );
                          }
                          if (cell.info.header === 'createdAt') {
                            return <TableCell key={cell.id}>{new Date(cell.value).toLocaleDateString()}</TableCell>;
                          }
                          if (cell.info.header === 'actions') {
                            return (
                              <TableCell key={cell.id}>
                                <Button
                                  kind="danger--ghost"
                                  size="sm"
                                  renderIcon={TrashCan}
                                  iconDescription="Delete"
                                  hasIconOnly
                                  disabled={deleteMutation.isPending}
                                  onClick={() => integration && deleteMutation.mutate(integration.id)}
                                />
                              </TableCell>
                            );
                          }
                          return <TableCell key={cell.id}>{cell.value}</TableCell>;
                        })}
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </>
          )}
        </DataTable>
      )}

      <Modal
        open={modalOpen}
        modalHeading="Add CI/CD Integration"
        primaryButtonText={createMutation.isPending ? 'Creating…' : 'Create'}
        secondaryButtonText="Cancel"
        onRequestClose={() => setModalOpen(false)}
        onRequestSubmit={handleCreate}
        primaryButtonDisabled={!name || createMutation.isPending}
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--cds-spacing-05)' }}>
          <Select id="cicd-provider" labelText="Provider" value={provider} onChange={(e) => setProvider(e.target.value as CICDProvider)}>
            {PROVIDERS.map((p) => <SelectItem key={p.id} value={p.id} text={p.label} />)}
          </Select>
          <TextInput id="cicd-name" labelText="Name" value={name} onChange={(e) => setName(e.target.value)} placeholder="my-github-integration" />
          <TextArea
            id="cicd-config"
            labelText="Config (JSON)"
            value={configJson}
            onChange={(e) => setConfigJson(e.target.value)}
            rows={5}
            placeholder='{"token": "...", "webhook_secret": "..."}'
          />
          {createMutation.isError && (
            <InlineNotification kind="error" title="Failed to create" subtitle={(createMutation.error as Error)?.message} />
          )}
        </div>
      </Modal>
    </div>
  );
}
