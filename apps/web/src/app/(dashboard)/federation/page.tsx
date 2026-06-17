'use client';

import { useState } from 'react';
import {
  Grid,
  Column,
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
  SkeletonText,
  InlineNotification,
} from '@carbon/react';
import { Add, TrashCan } from '@carbon/icons-react';
import { useFederatedClusters, useRegisterCluster, useRemoveCluster } from '@/lib/hooks/use-federation';
import type { ClusterProvider, ClusterStatus, FederatedCluster } from '@/lib/types';
import styles from '@/components/experiments/experiments.module.scss';

const STATUS_TAG: Record<ClusterStatus, 'green' | 'red' | 'warm-gray' | 'gray'> = {
  connected: 'green',
  disconnected: 'red',
  error: 'red',
  pending: 'warm-gray',
};

const PROVIDERS: ClusterProvider[] = ['aws', 'gcp', 'azure', 'on-premise', 'other'];

const headers = [
  { key: 'name', header: 'Name' },
  { key: 'region', header: 'Region' },
  { key: 'provider', header: 'Provider' },
  { key: 'status', header: 'Status' },
  { key: 'apiEndpoint', header: 'API Endpoint' },
  { key: 'actions', header: '' },
];

export default function FederationPage() {
  const { data, isLoading, isError, error } = useFederatedClusters();
  const registerMutation = useRegisterCluster();
  const removeMutation = useRemoveCluster();

  const [modalOpen, setModalOpen] = useState(false);
  const [name, setName] = useState('');
  const [region, setRegion] = useState('');
  const [provider, setProvider] = useState<ClusterProvider>('aws');
  const [apiEndpoint, setApiEndpoint] = useState('');

  function handleRegister() {
    if (!name || !region || !apiEndpoint) return;
    registerMutation.mutate(
      { name, region, provider, apiEndpoint },
      {
        onSuccess: () => {
          setModalOpen(false);
          setName(''); setRegion(''); setApiEndpoint('');
        },
      },
    );
  }

  const rows = (data?.items ?? []).map((c: FederatedCluster) => ({ ...c, id: c.id }));

  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div className={styles.pageHeader}>
          <h2 className={styles.pageTitle}>Federation</h2>
          <p className={styles.pageSubtitle}>Manage federated clusters across regions and providers.</p>
        </div>
      </Column>

      <Column lg={16} md={8} sm={4}>
        {isError && (
          <InlineNotification
            kind="error"
            title="Failed to load clusters"
            subtitle={(error as Error)?.message}
            style={{ marginBottom: 'var(--cds-spacing-05)' }}
          />
        )}

        {isLoading ? (
          <SkeletonText paragraph lineCount={6} />
        ) : (
          <DataTable rows={rows} headers={headers}>
            {({ rows: tableRows, headers: tableHeaders, getTableProps, getHeaderProps, getRowProps }) => (
              <>
                <TableToolbar>
                  <TableToolbarContent>
                    <Button renderIcon={Add} onClick={() => setModalOpen(true)}>
                      Register Cluster
                    </Button>
                  </TableToolbarContent>
                </TableToolbar>
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
                    {tableRows.map((row) => {
                      const cluster = (data?.items ?? []).find((c: FederatedCluster) => c.id === row.id);
                      const { key: rowKey, ...rowProps } = getRowProps({ row });
                      return (
                        <TableRow key={rowKey} {...rowProps}>
                          {row.cells.map((cell) => {
                            if (cell.info.header === 'status') {
                              return (
                                <TableCell key={cell.id}>
                                  <Tag type={STATUS_TAG[cell.value as ClusterStatus]} size="sm">
                                    {cell.value}
                                  </Tag>
                                </TableCell>
                              );
                            }
                            if (cell.info.header === 'actions') {
                              return (
                                <TableCell key={cell.id}>
                                  <Button
                                    kind="danger--ghost"
                                    size="sm"
                                    renderIcon={TrashCan}
                                    iconDescription="Remove cluster"
                                    hasIconOnly
                                    disabled={removeMutation.isPending}
                                    onClick={() => cluster && removeMutation.mutate(cluster.id)}
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
      </Column>

      <Modal
        open={modalOpen}
        modalHeading="Register Cluster"
        primaryButtonText={registerMutation.isPending ? 'Registering…' : 'Register'}
        secondaryButtonText="Cancel"
        onRequestClose={() => setModalOpen(false)}
        onRequestSubmit={handleRegister}
        primaryButtonDisabled={!name || !region || !apiEndpoint || registerMutation.isPending}
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--cds-spacing-05)' }}>
          <TextInput id="cluster-name" labelText="Cluster Name" value={name} onChange={(e) => setName(e.target.value)} placeholder="prod-us-east-1" />
          <TextInput id="cluster-region" labelText="Region" value={region} onChange={(e) => setRegion(e.target.value)} placeholder="us-east-1" />
          <Select id="cluster-provider" labelText="Provider" value={provider} onChange={(e) => setProvider(e.target.value as ClusterProvider)}>
            {PROVIDERS.map((p) => <SelectItem key={p} value={p} text={p} />)}
          </Select>
          <TextInput id="cluster-endpoint" labelText="API Endpoint" value={apiEndpoint} onChange={(e) => setApiEndpoint(e.target.value)} placeholder="https://k8s.example.com" />
          {registerMutation.isError && (
            <InlineNotification kind="error" title="Registration failed" subtitle={(registerMutation.error as Error)?.message} />
          )}
        </div>
      </Modal>
    </Grid>
  );
}
