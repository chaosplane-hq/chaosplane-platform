'use client';

export const dynamic = 'force-dynamic';

import { useState } from 'react';
import {
  Grid,
  Column,
  Form,
  FormGroup,
  TextInput,
  TextArea,
  NumberInput,
  Select,
  SelectItem,
  Button,
  Tile,
  InlineNotification,
} from '@carbon/react';
import { ArrowLeft, Send } from '@carbon/icons-react';
import { useRouter } from 'next/navigation';
import { useCreatePolicy } from '@/lib/hooks/use-policies';
import styles from '@/components/experiments/experiments.module.scss';
import type { PolicyEnforcement } from '@/lib/types';

const ENFORCEMENT_OPTIONS: PolicyEnforcement[] = ['enforce', 'audit', 'disabled'];

function parseNamespaces(value: string): string[] {
  return value
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean);
}

export default function CreatePolicyPage() {
  const router = useRouter();
  const createMutation = useCreatePolicy();

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [enforcement, setEnforcement] = useState<PolicyEnforcement>('audit');
  const [maxConcurrent, setMaxConcurrent] = useState('');
  const [maxTargets, setMaxTargets] = useState('');
  const [allowedNamespaces, setAllowedNamespaces] = useState('');
  const [blockedNamespaces, setBlockedNamespaces] = useState('');

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!name) return;

    const allowed = parseNamespaces(allowedNamespaces);
    const blocked = parseNamespaces(blockedNamespaces);

    createMutation.mutate(
      {
        name,
        description: description || undefined,
        enforcement,
        maxConcurrent: maxConcurrent !== '' ? Number(maxConcurrent) : undefined,
        maxTargets: maxTargets !== '' ? Number(maxTargets) : undefined,
        allowedNamespaces: allowed.length ? allowed : undefined,
        blockedNamespaces: blocked.length ? blocked : undefined,
      },
      {
        onSuccess: () => router.push('/policies'),
      },
    );
  }

  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--cds-spacing-04)', marginBottom: 'var(--cds-spacing-05)' }}>
          <Button kind="ghost" size="sm" renderIcon={ArrowLeft} onClick={() => router.push('/policies')}>
            Back
          </Button>
        </div>
        <div className={styles.pageHeader}>
          <h2 className={styles.pageTitle}>Create Policy</h2>
          <p className={styles.pageSubtitle}>Define blast radius guardrails for chaos experiments.</p>
        </div>
      </Column>

      <Column lg={10} md={8} sm={4}>
        <Form onSubmit={handleSubmit}>
          <Tile style={{ marginBottom: 'var(--cds-spacing-05)' }}>
            <h3 className={styles.sectionTitle}>Basic Info</h3>
            <FormGroup legendText="">
              <TextInput
                id="policy-name"
                labelText="Policy Name"
                placeholder="production-guardrails"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />
            </FormGroup>
            <FormGroup legendText="" style={{ marginTop: 'var(--cds-spacing-05)' }}>
              <TextArea
                id="policy-description"
                labelText="Description"
                placeholder="Limits blast radius for production experiments"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={3}
              />
            </FormGroup>
            <FormGroup legendText="" style={{ marginTop: 'var(--cds-spacing-05)' }}>
              <Select
                id="policy-enforcement"
                labelText="Enforcement"
                value={enforcement}
                onChange={(e) => setEnforcement(e.target.value as PolicyEnforcement)}
              >
                {ENFORCEMENT_OPTIONS.map((opt) => (
                  <SelectItem key={opt} value={opt} text={opt} />
                ))}
              </Select>
            </FormGroup>
          </Tile>

          <Tile style={{ marginBottom: 'var(--cds-spacing-05)' }}>
            <h3 className={styles.sectionTitle}>Limits</h3>
            <FormGroup legendText="">
              <NumberInput
                id="policy-max-concurrent"
                label="Max Concurrent"
                placeholder="3"
                min={0}
                value={maxConcurrent}
                onChange={(_e, { value }) => setMaxConcurrent(value === '' ? '' : String(value))}
                helperText="Maximum concurrent experiments (optional)"
                allowEmpty
              />
            </FormGroup>
            <FormGroup legendText="" style={{ marginTop: 'var(--cds-spacing-05)' }}>
              <NumberInput
                id="policy-max-targets"
                label="Max Targets"
                placeholder="10"
                min={0}
                value={maxTargets}
                onChange={(_e, { value }) => setMaxTargets(value === '' ? '' : String(value))}
                helperText="Maximum targets per experiment (optional)"
                allowEmpty
              />
            </FormGroup>
          </Tile>

          <Tile style={{ marginBottom: 'var(--cds-spacing-05)' }}>
            <h3 className={styles.sectionTitle}>Namespaces</h3>
            <FormGroup legendText="">
              <TextInput
                id="policy-allowed-namespaces"
                labelText="Allowed Namespaces"
                placeholder="default, staging"
                value={allowedNamespaces}
                onChange={(e) => setAllowedNamespaces(e.target.value)}
                helperText="Comma-separated namespace names"
              />
            </FormGroup>
            <FormGroup legendText="" style={{ marginTop: 'var(--cds-spacing-05)' }}>
              <TextInput
                id="policy-blocked-namespaces"
                labelText="Blocked Namespaces"
                placeholder="kube-system, production"
                value={blockedNamespaces}
                onChange={(e) => setBlockedNamespaces(e.target.value)}
                helperText="Comma-separated namespace names"
              />
            </FormGroup>
          </Tile>

          {createMutation.isError && (
            <InlineNotification
              kind="error"
              title="Failed to create policy"
              subtitle={createMutation.error?.message}
              style={{ marginBottom: 'var(--cds-spacing-05)' }}
            />
          )}

          <Button
            type="submit"
            renderIcon={Send}
            disabled={!name || createMutation.isPending}
          >
            {createMutation.isPending ? 'Creating…' : 'Create Policy'}
          </Button>
        </Form>
      </Column>
    </Grid>
  );
}
