'use client';

import { useState } from 'react';
import {
  Tile,
  Button,
  TextInput,
  Select,
  SelectItem,
  Tag,
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
} from '@carbon/react';
import { Add, Send, TrashCan } from '@carbon/icons-react';
import {
  useTeamMembers,
  useInvitations,
  useCreateInvitation,
  useResendInvitation,
  useRevokeInvitation,
} from '@/lib/hooks/use-invitations';
import { useHierarchy } from '@/lib/hooks/use-hierarchy';
import type { MemberRole, InvitationStatus } from '@/lib/types';

const ROLE_OPTIONS: MemberRole[] = ['owner', 'admin', 'member', 'viewer'];

const memberHeaders = [
  { key: 'name', header: 'Name' },
  { key: 'email', header: 'Email' },
  { key: 'role', header: 'Role' },
  { key: 'joinedAt', header: 'Joined' },
];

const inviteHeaders = [
  { key: 'email', header: 'Email' },
  { key: 'role', header: 'Role' },
  { key: 'status', header: 'Status' },
  { key: 'expiresAt', header: 'Expires' },
  { key: 'actions', header: '' },
];

function statusKind(status: InvitationStatus): 'blue' | 'green' | 'red' | 'gray' {
  if (status === 'pending') return 'blue';
  if (status === 'accepted') return 'green';
  if (status === 'expired') return 'gray';
  return 'red';
}

function formatDate(iso?: string) {
  if (!iso) return '—';
  return new Date(iso).toLocaleDateString();
}

export function MembersTab() {
  const [email, setEmail] = useState('');
  const [role, setRole] = useState<MemberRole>('member');
  const [successMsg, setSuccessMsg] = useState('');
  const [errorMsg, setErrorMsg] = useState('');

  const { data: membersData, isLoading: membersLoading } = useTeamMembers();
  const { data: invitesData, isLoading: invitesLoading } = useInvitations();
  const { data: hierarchy } = useHierarchy();
  const createInvitation = useCreateInvitation();
  const resendInvitation = useResendInvitation();
  const revokeInvitation = useRevokeInvitation();

  const members = membersData?.items ?? [];
  const invitations = invitesData?.items ?? [];
  const organizationId = hierarchy?.organizations[0]?.id;

  async function handleInvite() {
    if (!email) return;
    if (!organizationId) {
      setErrorMsg('No organization available. Create an organization first.');
      return;
    }
    setSuccessMsg('');
    setErrorMsg('');
    try {
      await createInvitation.mutateAsync({ email, organizationId, role });
      setSuccessMsg(`Invitation sent to ${email}`);
      setEmail('');
    } catch {
      setErrorMsg('Failed to send invitation. Please try again.');
    }
  }

  const memberRows = members.map((m) => ({
    id: m.id,
    name: m.name,
    email: m.email,
    role: m.role,
    joinedAt: formatDate(m.joinedAt),
  }));

  const inviteRows = invitations.map((inv) => ({
    id: inv.id,
    email: inv.email,
    role: inv.role,
    status: inv.status,
    expiresAt: formatDate(inv.expiresAt),
  }));

  return (
    <div style={{ paddingTop: 'var(--cds-spacing-06)' }}>
      <Tile style={{ marginBottom: 'var(--cds-spacing-05)' }}>
        <h3 style={{ fontSize: 'var(--cds-heading-02-font-size)', fontWeight: 'var(--cds-heading-02-font-weight)', color: 'var(--cds-text-primary)', margin: '0 0 var(--cds-spacing-05)' }}>
          Invite member
        </h3>
        {successMsg && (
          <InlineNotification
            kind="success"
            title={successMsg}
            onCloseButtonClick={() => setSuccessMsg('')}
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
        <div style={{ display: 'flex', gap: 'var(--cds-spacing-04)', alignItems: 'flex-end', flexWrap: 'wrap' }}>
          <div style={{ flex: '1 1 260px' }}>
            <TextInput
              id="invite-email"
              labelText="Email address"
              placeholder="colleague@company.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Enter') handleInvite(); }}
            />
          </div>
          <div style={{ minWidth: '160px' }}>
            <Select
              id="invite-role"
              labelText="Role"
              value={role}
              onChange={(e) => setRole(e.target.value as MemberRole)}
            >
              {ROLE_OPTIONS.map((r) => (
                <SelectItem key={r} value={r} text={r.charAt(0).toUpperCase() + r.slice(1)} />
              ))}
            </Select>
          </div>
          <Button
            renderIcon={Add}
            onClick={handleInvite}
            disabled={!email || createInvitation.isPending}
          >
            Send invite
          </Button>
        </div>
      </Tile>

      <Tile style={{ marginBottom: 'var(--cds-spacing-05)' }}>
        <h3 style={{ fontSize: 'var(--cds-heading-02-font-size)', fontWeight: 'var(--cds-heading-02-font-weight)', color: 'var(--cds-text-primary)', margin: '0 0 var(--cds-spacing-05)' }}>
          Team members
        </h3>
        {membersLoading ? (
          <SkeletonText paragraph lineCount={4} />
        ) : (
          <DataTable rows={memberRows} headers={memberHeaders}>
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
                        <TableCell colSpan={4} style={{ textAlign: 'center', color: 'var(--cds-text-secondary)' }}>
                          No members yet.
                        </TableCell>
                      </TableRow>
                    ) : rows.map((row) => {
                      const { key, ...rowProps } = getRowProps({ row });
                      return (
                        <TableRow key={key} {...rowProps}>
                          {row.cells.map((cell) => (
                            <TableCell key={cell.id}>{cell.value}</TableCell>
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
      </Tile>

      <Tile>
        <h3 style={{ fontSize: 'var(--cds-heading-02-font-size)', fontWeight: 'var(--cds-heading-02-font-weight)', color: 'var(--cds-text-primary)', margin: '0 0 var(--cds-spacing-05)' }}>
          Pending invitations
        </h3>
        {invitesLoading ? (
          <SkeletonText paragraph lineCount={3} />
        ) : inviteRows.length === 0 ? (
          <p style={{ color: 'var(--cds-text-secondary)' }}>No pending invitations.</p>
        ) : (
          <DataTable rows={inviteRows} headers={inviteHeaders}>
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
                    {rows.map((row) => {
                      const inv = invitations.find((i) => i.id === row.id);
                      const { key, ...rowProps } = getRowProps({ row });
                      return (
                        <TableRow key={key} {...rowProps}>
                          <TableCell>{row.cells[0].value}</TableCell>
                          <TableCell>{row.cells[1].value}</TableCell>
                          <TableCell>
                            <Tag type={statusKind(row.cells[2].value as InvitationStatus)} size="sm">
                              {row.cells[2].value}
                            </Tag>
                          </TableCell>
                          <TableCell>{row.cells[3].value}</TableCell>
                          <TableCell>
                            <div style={{ display: 'flex', gap: 'var(--cds-spacing-02)' }}>
                              {inv?.status === 'pending' && (
                                <Button
                                  kind="ghost"
                                  size="sm"
                                  renderIcon={Send}
                                  iconDescription="Resend"
                                  hasIconOnly
                                  onClick={() => resendInvitation.mutate(row.id)}
                                  disabled={resendInvitation.isPending}
                                />
                              )}
                              <Button
                                kind="danger--ghost"
                                size="sm"
                                renderIcon={TrashCan}
                                iconDescription="Revoke"
                                hasIconOnly
                                onClick={() => revokeInvitation.mutate(row.id)}
                                disabled={revokeInvitation.isPending}
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
