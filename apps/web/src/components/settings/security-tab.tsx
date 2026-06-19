'use client';

import { useState } from 'react';
import {
  Tile,
  Button,
  Tag,
  TextInput,
  Select,
  SelectItem,
  TextArea,
  Modal,
  SkeletonText,
  InlineNotification,
  CodeSnippet,
  StructuredListWrapper,
  StructuredListBody,
  StructuredListRow,
  StructuredListCell,
  DataTable,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
  TableContainer,
} from '@carbon/react';
import { Add, TrashCan } from '@carbon/icons-react';
import {
  useSSOProviders,
  useCreateSSOProvider,
  useDeleteSSOProvider,
  useABACPolicies,
  useCreateABACPolicy,
  useDeleteABACPolicy,
  useMFARecoveryCodes,
  useGenerateMFACodes,
  useActiveSessions,
  useRevokeSession,
  useRevokeAllSessions,
  useRequestEmailChange,
  useRequestAccountDeletion,
} from '@/lib/hooks/use-security';
import type { CreateSSOProviderRequest, CreateABACPolicyRequest } from '@/lib/types';

function formatDate(iso?: string) {
  if (!iso) return '—';
  return new Date(iso).toLocaleString();
}

const sessionHeaders = [
  { key: 'ipAddress', header: 'IP Address' },
  { key: 'userAgent', header: 'User Agent' },
  { key: 'lastActiveAt', header: 'Last Active' },
  { key: 'actions', header: '' },
];

export function SecurityTab() {
  const { data: ssoData, isLoading: ssoLoading, isError: ssoError, error: ssoErrorObj } = useSSOProviders();
  const createSSO = useCreateSSOProvider();
  const deleteSSO = useDeleteSSOProvider();

  const { data: abacData, isLoading: abacLoading, isError: abacError, error: abacErrorObj } = useABACPolicies();
  const createABAC = useCreateABACPolicy();
  const deleteABAC = useDeleteABACPolicy();

  const { data: mfaData, isLoading: mfaLoading, isError: mfaError, error: mfaErrorObj } = useMFARecoveryCodes();
  const generateMFA = useGenerateMFACodes();

  const { data: sessionsData, isLoading: sessionsLoading, isError: sessionsError, error: sessionsErrorObj } = useActiveSessions();

  const hasQueryError = ssoError || abacError || mfaError || sessionsError;
  const queryErrorMessage = (ssoErrorObj as Error)?.message ?? (abacErrorObj as Error)?.message ?? (mfaErrorObj as Error)?.message ?? (sessionsErrorObj as Error)?.message ?? '';
  const revokeSession = useRevokeSession();
  const revokeAll = useRevokeAllSessions();

  const requestEmailChange = useRequestEmailChange();
  const requestDeletion = useRequestAccountDeletion();

  const [ssoOpen, setSsoOpen] = useState(false);
  const [ssoName, setSsoName] = useState('');
  const [ssoType, setSsoType] = useState<'saml' | 'oidc'>('saml');
  const [ssoEntityId, setSsoEntityId] = useState('');
  const [ssoUrl, setSsoUrl] = useState('');

  const [abacOpen, setAbacOpen] = useState(false);
  const [abacName, setAbacName] = useState('');
  const [abacDesc, setAbacDesc] = useState('');
  const [abacEffect, setAbacEffect] = useState<'allow' | 'deny'>('allow');
  const [abacActions, setAbacActions] = useState('');
  const [abacResources, setAbacResources] = useState('');

  const [newEmail, setNewEmail] = useState('');
  const [deleteConfirm, setDeleteConfirm] = useState(false);

  const [error, setError] = useState('');
  const [successMsg, setSuccessMsg] = useState('');

  const providers = ssoData?.items ?? [];
  const policies = abacData?.items ?? [];
  const sessions = sessionsData?.items ?? [];

  const sessionRows = sessions.map((s) => ({
    id: s.id,
    ipAddress: s.ipAddress ?? '—',
    userAgent: (s.userAgent ?? '').slice(0, 60) + ((s.userAgent ?? '').length > 60 ? '…' : ''),
    lastActiveAt: formatDate(s.lastActivity),
    isCurrent: s.isCurrent,
  }));

  async function handleCreateSSO() {
    setError('');
    try {
      const data: CreateSSOProviderRequest = {
        name: ssoName.trim(),
        type: ssoType,
        entityId: ssoEntityId.trim(),
        ssoUrl: ssoUrl.trim(),
      };
      await createSSO.mutateAsync(data);
      setSsoOpen(false);
      setSsoName(''); setSsoEntityId(''); setSsoUrl('');
    } catch {
      setError('Failed to create SSO provider.');
    }
  }

  async function handleCreateABAC() {
    setError('');
    try {
      const data: CreateABACPolicyRequest = {
        name: abacName.trim(),
        description: abacDesc.trim() || undefined,
        effect: abacEffect,
        actions: abacActions.split(',').map((a) => a.trim()).filter(Boolean),
        resources: abacResources.split(',').map((r) => r.trim()).filter(Boolean),
      };
      await createABAC.mutateAsync(data);
      setAbacOpen(false);
      setAbacName(''); setAbacDesc(''); setAbacActions(''); setAbacResources('');
    } catch {
      setError('Failed to create ABAC policy.');
    }
  }

  async function handleEmailChange() {
    setError(''); setSuccessMsg('');
    try {
      await requestEmailChange.mutateAsync(newEmail.trim());
      setSuccessMsg('Email change request sent. Check your inbox.');
      setNewEmail('');
    } catch {
      setError('Failed to request email change.');
    }
  }

  async function handleAccountDeletion() {
    setError(''); setSuccessMsg('');
    try {
      await requestDeletion.mutateAsync();
      setSuccessMsg('Account deletion request submitted.');
      setDeleteConfirm(false);
    } catch {
      setError('Failed to request account deletion.');
    }
  }

  return (
    <div style={{ paddingTop: 'var(--cds-spacing-06)', display: 'flex', flexDirection: 'column', gap: 'var(--cds-spacing-06)' }}>
      {hasQueryError && (
        <InlineNotification
          kind="error"
          title="Failed to load security settings"
          subtitle={queryErrorMessage}
          style={{ marginBottom: 'var(--cds-spacing-05)' }}
        />
      )}

      {error && (
        <InlineNotification kind="error" title={error} onCloseButtonClick={() => setError('')} />
      )}
      {successMsg && (
        <InlineNotification kind="success" title={successMsg} onCloseButtonClick={() => setSuccessMsg('')} />
      )}

      <Tile>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--cds-spacing-05)' }}>
          <h3 style={{ fontSize: 'var(--cds-heading-02-font-size)', fontWeight: 'var(--cds-heading-02-font-weight)', color: 'var(--cds-text-primary)', margin: 0 }}>
            SSO / SAML Providers
          </h3>
          <Button renderIcon={Add} size="sm" onClick={() => setSsoOpen(true)}>Add Provider</Button>
        </div>
        {ssoLoading ? <SkeletonText paragraph lineCount={3} /> : providers.length === 0 ? (
          <p style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)' }}>No SSO providers configured.</p>
        ) : (
          <StructuredListWrapper>
            <StructuredListBody>
              {providers.map((p) => (
                <StructuredListRow key={p.id}>
                  <StructuredListCell>{p.name}</StructuredListCell>
                  <StructuredListCell><Tag type="blue" size="sm">{p.type.toUpperCase()}</Tag></StructuredListCell>
                  <StructuredListCell style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)' }}>
                    {p.ssoUrl ?? '—'}
                  </StructuredListCell>
                  <StructuredListCell>
                    <Tag type={p.enabled ? 'green' : 'gray'} size="sm">{p.enabled ? 'enabled' : 'disabled'}</Tag>
                  </StructuredListCell>
                  <StructuredListCell>
                    <Button kind="danger--ghost" size="sm" renderIcon={TrashCan} iconDescription="Delete" hasIconOnly
                      onClick={() => deleteSSO.mutate(p.id)} disabled={deleteSSO.isPending} />
                  </StructuredListCell>
                </StructuredListRow>
              ))}
            </StructuredListBody>
          </StructuredListWrapper>
        )}
      </Tile>

      <Tile>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--cds-spacing-05)' }}>
          <h3 style={{ fontSize: 'var(--cds-heading-02-font-size)', fontWeight: 'var(--cds-heading-02-font-weight)', color: 'var(--cds-text-primary)', margin: 0 }}>
            ABAC Policies
          </h3>
          <Button renderIcon={Add} size="sm" onClick={() => setAbacOpen(true)}>Add Policy</Button>
        </div>
        {abacLoading ? <SkeletonText paragraph lineCount={3} /> : policies.length === 0 ? (
          <p style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)' }}>No ABAC policies defined.</p>
        ) : (
          <StructuredListWrapper>
            <StructuredListBody>
              {policies.map((p) => (
                <StructuredListRow key={p.id}>
                  <StructuredListCell>{p.name}</StructuredListCell>
                  <StructuredListCell>
                    <Tag type={p.effect === 'allow' ? 'green' : 'red'} size="sm">{p.effect}</Tag>
                  </StructuredListCell>
                  <StructuredListCell style={{ fontSize: 'var(--cds-label-01-font-size)', color: 'var(--cds-text-secondary)' }}>
                    {Array.isArray(p.actions) ? p.actions.join(', ') : String(p.actions ?? '')}
                  </StructuredListCell>
                  <StructuredListCell style={{ fontSize: 'var(--cds-label-01-font-size)', color: 'var(--cds-text-secondary)' }}>
                    {Array.isArray(p.resources) ? p.resources.join(', ') : String(p.resources ?? '')}
                  </StructuredListCell>
                  <StructuredListCell>
                    <Button kind="danger--ghost" size="sm" renderIcon={TrashCan} iconDescription="Delete" hasIconOnly
                      onClick={() => deleteABAC.mutate(p.id)} disabled={deleteABAC.isPending} />
                  </StructuredListCell>
                </StructuredListRow>
              ))}
            </StructuredListBody>
          </StructuredListWrapper>
        )}
      </Tile>

      <Tile>
        <h3 style={{ fontSize: 'var(--cds-heading-02-font-size)', fontWeight: 'var(--cds-heading-02-font-weight)', color: 'var(--cds-text-primary)', margin: '0 0 var(--cds-spacing-05)' }}>
          MFA Recovery Codes
        </h3>
        {mfaLoading ? <SkeletonText paragraph lineCount={2} /> : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--cds-spacing-04)' }}>
            <p style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)', margin: 0 }}>
              Remaining codes: <strong style={{ color: 'var(--cds-text-primary)' }}>{mfaData?.remaining ?? 0}</strong>
            </p>
            {mfaData?.codes && mfaData.codes.length > 0 && (
              <CodeSnippet type="multi" feedback="Copied!">
                {mfaData.codes.join('\n')}
              </CodeSnippet>
            )}
            <div>
              <Button kind="secondary" size="sm" onClick={() => generateMFA.mutate()} disabled={generateMFA.isPending}>
                Generate New Codes
              </Button>
            </div>
          </div>
        )}
      </Tile>

      <Tile>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--cds-spacing-05)' }}>
          <h3 style={{ fontSize: 'var(--cds-heading-02-font-size)', fontWeight: 'var(--cds-heading-02-font-weight)', color: 'var(--cds-text-primary)', margin: 0 }}>
            Active Sessions
          </h3>
          <Button kind="danger--ghost" size="sm" onClick={() => revokeAll.mutate()} disabled={revokeAll.isPending}>
            Revoke All
          </Button>
        </div>
        {sessionsLoading ? <SkeletonText paragraph lineCount={4} /> : (
          <DataTable rows={sessionRows} headers={sessionHeaders}>
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
                        <TableCell colSpan={sessionHeaders.length} style={{ textAlign: 'center', color: 'var(--cds-text-secondary)' }}>
                          No active sessions.
                        </TableCell>
                      </TableRow>
                    ) : tableRows.map((row) => {
                      const { key, ...rowProps } = getRowProps({ row });
                      const original = sessions.find((s) => s.id === row.id);
                      return (
                        <TableRow key={key} {...rowProps}>
                          <TableCell>
                            {row.cells[0].value}
                            {original?.isCurrent && <Tag type="green" size="sm" style={{ marginLeft: 'var(--cds-spacing-02)' }}>current</Tag>}
                          </TableCell>
                          <TableCell style={{ fontSize: 'var(--cds-label-01-font-size)', color: 'var(--cds-text-secondary)' }}>
                            {row.cells[1].value}
                          </TableCell>
                          <TableCell style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)' }}>
                            {row.cells[2].value}
                          </TableCell>
                          <TableCell>
                            {!original?.isCurrent && (
                              <Button kind="danger--ghost" size="sm" renderIcon={TrashCan} iconDescription="Revoke" hasIconOnly
                                onClick={() => revokeSession.mutate(row.id)} disabled={revokeSession.isPending} />
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
      </Tile>

      <Tile>
        <h3 style={{ fontSize: 'var(--cds-heading-02-font-size)', fontWeight: 'var(--cds-heading-02-font-weight)', color: 'var(--cds-text-primary)', margin: '0 0 var(--cds-spacing-05)' }}>
          Change Email
        </h3>
        <div style={{ display: 'flex', gap: 'var(--cds-spacing-04)', alignItems: 'flex-end' }}>
          <TextInput
            id="new-email"
            labelText="New email address"
            placeholder="you@example.com"
            value={newEmail}
            onChange={(e) => setNewEmail(e.target.value)}
            style={{ minWidth: 280 }}
          />
          <Button kind="secondary" size="md" onClick={handleEmailChange} disabled={!newEmail.trim() || requestEmailChange.isPending}>
            Request Change
          </Button>
        </div>
      </Tile>

      <Tile>
        <h3 style={{ fontSize: 'var(--cds-heading-02-font-size)', fontWeight: 'var(--cds-heading-02-font-weight)', color: 'var(--cds-text-primary)', margin: '0 0 var(--cds-spacing-03)' }}>
          Delete Account
        </h3>
        <p style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)', margin: '0 0 var(--cds-spacing-05)' }}>
          This will submit a deletion request. Your account will be scheduled for removal.
        </p>
        <Button kind="danger" size="sm" onClick={() => setDeleteConfirm(true)}>
          Request Account Deletion
        </Button>
      </Tile>

      <Modal
        open={ssoOpen}
        modalHeading="Add SSO Provider"
        primaryButtonText="Add"
        secondaryButtonText="Cancel"
        onRequestSubmit={handleCreateSSO}
        onRequestClose={() => { setSsoOpen(false); setSsoName(''); setSsoEntityId(''); setSsoUrl(''); }}
        primaryButtonDisabled={!ssoName.trim() || createSSO.isPending}
      >
        <TextInput id="sso-name" labelText="Name" value={ssoName} onChange={(e) => setSsoName(e.target.value)}
          style={{ marginBottom: 'var(--cds-spacing-05)' }} />
        <Select id="sso-type" labelText="Type" value={ssoType} onChange={(e) => setSsoType(e.target.value as 'saml' | 'oidc')}
          style={{ marginBottom: 'var(--cds-spacing-05)' }}>
          <SelectItem value="saml" text="SAML" />
          <SelectItem value="oidc" text="OIDC" />
        </Select>
        <TextInput id="sso-entity" labelText="Entity ID (optional)" value={ssoEntityId} onChange={(e) => setSsoEntityId(e.target.value)}
          style={{ marginBottom: 'var(--cds-spacing-05)' }} />
        <TextInput id="sso-url" labelText="SSO URL (optional)" value={ssoUrl} onChange={(e) => setSsoUrl(e.target.value)} />
      </Modal>

      <Modal
        open={abacOpen}
        modalHeading="Add ABAC Policy"
        primaryButtonText="Add"
        secondaryButtonText="Cancel"
        onRequestSubmit={handleCreateABAC}
        onRequestClose={() => { setAbacOpen(false); setAbacName(''); setAbacDesc(''); setAbacActions(''); setAbacResources(''); }}
        primaryButtonDisabled={!abacName.trim() || createABAC.isPending}
      >
        <TextInput id="abac-name" labelText="Name" value={abacName} onChange={(e) => setAbacName(e.target.value)}
          style={{ marginBottom: 'var(--cds-spacing-05)' }} />
        <TextArea id="abac-desc" labelText="Description (optional)" value={abacDesc} onChange={(e) => setAbacDesc(e.target.value)}
          rows={2} style={{ marginBottom: 'var(--cds-spacing-05)' }} />
        <Select id="abac-effect" labelText="Effect" value={abacEffect} onChange={(e) => setAbacEffect(e.target.value as 'allow' | 'deny')}
          style={{ marginBottom: 'var(--cds-spacing-05)' }}>
          <SelectItem value="allow" text="Allow" />
          <SelectItem value="deny" text="Deny" />
        </Select>
        <TextInput id="abac-actions" labelText="Actions (comma-separated)" placeholder="e.g. experiment:create, experiment:delete"
          value={abacActions} onChange={(e) => setAbacActions(e.target.value)} style={{ marginBottom: 'var(--cds-spacing-05)' }} />
        <TextInput id="abac-resources" labelText="Resources (comma-separated)" placeholder="e.g. experiments/*, policies/*"
          value={abacResources} onChange={(e) => setAbacResources(e.target.value)} />
      </Modal>

      <Modal
        open={deleteConfirm}
        danger
        modalHeading="Request Account Deletion"
        primaryButtonText="Confirm"
        secondaryButtonText="Cancel"
        onRequestSubmit={handleAccountDeletion}
        onRequestClose={() => setDeleteConfirm(false)}
        primaryButtonDisabled={requestDeletion.isPending}
      >
        <p style={{ color: 'var(--cds-text-secondary)' }}>
          Are you sure you want to request account deletion? This action cannot be undone.
        </p>
      </Modal>
    </div>
  );
}
