'use client';

export const dynamic = 'force-dynamic';

import { useState, useMemo } from 'react';
import {
  Grid,
  Column,
  Form,
  FormGroup,
  TextInput,
  Dropdown,
  Button,
  Tile,
  InlineNotification,
  Select,
  SelectItem,
} from '@carbon/react';
import { ArrowLeft, Send } from '@carbon/icons-react';
import { useRouter } from 'next/navigation';
import { useCreateExperiment } from '@/lib/hooks/use-experiments';
import styles from '@/components/experiments/experiments.module.scss';
import type { ActionType } from '@/lib/types';
import { ACTION_TYPE_GROUPS } from '@/lib/types';

interface ParamDef {
  key: string;
  label: string;
  placeholder?: string;
  defaultValue?: string;
}

const ACTION_PARAMS: Partial<Record<ActionType, ParamDef[]>> = {
  'container-kill': [{ key: 'containerName', label: 'Container Name', placeholder: 'app' }],
  'pod-cpu-stress': [
    { key: 'workers', label: 'Workers', placeholder: '1', defaultValue: '1' },
    { key: 'load', label: 'CPU Load (%)', placeholder: '50', defaultValue: '50' },
  ],
  'pod-memory-stress': [
    { key: 'workers', label: 'Workers', placeholder: '1', defaultValue: '1' },
    { key: 'size', label: 'Memory Size', placeholder: '256MB', defaultValue: '256MB' },
  ],
  'pod-io-stress': [
    { key: 'workers', label: 'Workers', placeholder: '1', defaultValue: '1' },
    { key: 'volumePath', label: 'Volume Path', placeholder: '/data' },
  ],
  'pod-dns-error': [{ key: 'patterns', label: 'DNS Patterns (comma-separated)', placeholder: 'example.com' }],
  'pod-http-abort': [
    { key: 'port', label: 'Port', placeholder: '8080', defaultValue: '8080' },
    { key: 'path', label: 'Path', placeholder: '/', defaultValue: '/' },
    { key: 'method', label: 'Method', placeholder: 'GET', defaultValue: 'GET' },
  ],
  'pod-http-delay': [
    { key: 'port', label: 'Port', placeholder: '8080', defaultValue: '8080' },
    { key: 'path', label: 'Path', placeholder: '/', defaultValue: '/' },
    { key: 'delay', label: 'Delay (ms)', placeholder: '1000', defaultValue: '1000' },
    { key: 'method', label: 'Method', placeholder: 'GET', defaultValue: 'GET' },
  ],
  'network-delay': [
    { key: 'latency', label: 'Latency', placeholder: '100ms', defaultValue: '100ms' },
    { key: 'jitter', label: 'Jitter', placeholder: '10ms', defaultValue: '10ms' },
    { key: 'correlation', label: 'Correlation (%)', placeholder: '0', defaultValue: '0' },
  ],
  'network-loss': [
    { key: 'loss', label: 'Loss (%)', placeholder: '10', defaultValue: '10' },
    { key: 'correlation', label: 'Correlation (%)', placeholder: '0', defaultValue: '0' },
  ],
  'network-corrupt': [
    { key: 'corrupt', label: 'Corrupt (%)', placeholder: '10', defaultValue: '10' },
    { key: 'correlation', label: 'Correlation (%)', placeholder: '0', defaultValue: '0' },
  ],
  'network-duplicate': [
    { key: 'duplicate', label: 'Duplicate (%)', placeholder: '10', defaultValue: '10' },
    { key: 'correlation', label: 'Correlation (%)', placeholder: '0', defaultValue: '0' },
  ],
  'network-partition': [{ key: 'direction', label: 'Direction', placeholder: 'both', defaultValue: 'both' }],
  'network-bandwidth': [
    { key: 'rate', label: 'Rate', placeholder: '1mbps', defaultValue: '1mbps' },
    { key: 'limit', label: 'Limit (bytes)', placeholder: '10000', defaultValue: '10000' },
    { key: 'buffer', label: 'Buffer (bytes)', placeholder: '10000', defaultValue: '10000' },
  ],
  'node-taint': [
    { key: 'key', label: 'Taint Key', placeholder: 'chaos' },
    { key: 'value', label: 'Taint Value', placeholder: 'true' },
    { key: 'effect', label: 'Effect', placeholder: 'NoSchedule', defaultValue: 'NoSchedule' },
  ],
  'node-cpu-stress': [
    { key: 'workers', label: 'Workers', placeholder: '1', defaultValue: '1' },
    { key: 'load', label: 'CPU Load (%)', placeholder: '50', defaultValue: '50' },
  ],
  'stress-cpu': [
    { key: 'workers', label: 'Workers', placeholder: '1', defaultValue: '1' },
    { key: 'load', label: 'CPU Load (%)', placeholder: '50', defaultValue: '50' },
  ],
  'stress-memory': [
    { key: 'workers', label: 'Workers', placeholder: '1', defaultValue: '1' },
    { key: 'size', label: 'Memory Size', placeholder: '256MB', defaultValue: '256MB' },
  ],
};

const ACTION_ITEMS = Object.entries(ACTION_TYPE_GROUPS).flatMap(([group, types]) =>
  types.map((t) => ({ id: t, label: `${t}`, group })),
);

const MODE_OPTIONS = ['one', 'all', 'fixed', 'fixed-percent', 'random-max-percent'];

function buildYaml(fields: {
  name: string;
  namespace: string;
  actionType: ActionType | '';
  targetNamespace: string;
  labelSelector: string;
  mode: string;
  duration: string;
  params: Record<string, string>;
}): string {
  const labels = fields.labelSelector
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean)
    .map((kv) => {
      const [k, v] = kv.split('=');
      return `        ${k?.trim() ?? ''}: ${v?.trim() ?? ''}`;
    })
    .join('\n');

  const paramLines = Object.entries(fields.params)
    .filter(([, v]) => v)
    .map(([k, v]) => `    ${k}: ${v}`)
    .join('\n');

  return `apiVersion: chaos.chaosplane.io/v1alpha1
kind: Experiment
metadata:
  name: ${fields.name || '<name>'}
  namespace: ${fields.namespace || '<namespace>'}
spec:
  action:
    type: ${fields.actionType || '<action-type>'}
${paramLines ? `    parameters:\n${paramLines}` : ''}
  target:
    namespace: ${fields.targetNamespace || '<target-namespace>'}
    mode: ${fields.mode}
    labelSelector:
${labels || '        {}'}
  duration: ${fields.duration || '30s'}`.trim();
}

export default function CreateExperimentPage() {
  const router = useRouter();
  const createMutation = useCreateExperiment();

  const [name, setName] = useState('');
  const [namespace, setNamespace] = useState('default');
  const [actionType, setActionType] = useState<ActionType | ''>('');
  const [targetNamespace, setTargetNamespace] = useState('default');
  const [labelSelector, setLabelSelector] = useState('');
  const [mode, setMode] = useState('one');
  const [duration, setDuration] = useState('30s');
  const [params, setParams] = useState<Record<string, string>>({});

  const paramDefs = actionType ? (ACTION_PARAMS[actionType] ?? []) : [];

  const setParam = (key: string, value: string) =>
    setParams((prev) => ({ ...prev, [key]: value }));

  const yamlPreview = useMemo(
    () => buildYaml({ name, namespace, actionType, targetNamespace, labelSelector, mode, duration, params }),
    [name, namespace, actionType, targetNamespace, labelSelector, mode, duration, params],
  );

  function handleActionChange(type: ActionType) {
    setActionType(type);
    const defs = ACTION_PARAMS[type] ?? [];
    const defaults: Record<string, string> = {};
    defs.forEach((d) => {
      if (d.defaultValue) defaults[d.key] = d.defaultValue;
    });
    setParams(defaults);
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!actionType || !name) return;

    const labelObj = labelSelector
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
      .reduce<Record<string, string>>((acc, kv) => {
        const [k, v] = kv.split('=');
        if (k) acc[k.trim()] = v?.trim() ?? '';
        return acc;
      }, {});

    createMutation.mutate(
      {
        name,
        namespace,
        action: {
          type: actionType,
          parameters: Object.fromEntries(
            Object.entries(params).filter(([, v]) => v !== ''),
          ),
        },
        target: {
          namespace: targetNamespace,
          labelSelector: Object.keys(labelObj).length ? labelObj : undefined,
          mode: mode as 'one' | 'all' | 'fixed' | 'fixed-percent' | 'random-max-percent',
        },
        duration,
      },
      {
        onSuccess: (exp) => router.push(`/experiments/${exp.name}`),
      },
    );
  }

  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--cds-spacing-04)', marginBottom: 'var(--cds-spacing-05)' }}>
          <Button kind="ghost" size="sm" renderIcon={ArrowLeft} onClick={() => router.push('/experiments')}>
            Back
          </Button>
        </div>
        <div className={styles.pageHeader}>
          <h2 className={styles.pageTitle}>Create Experiment</h2>
          <p className={styles.pageSubtitle}>Configure and launch a new chaos experiment.</p>
        </div>
      </Column>

      <Column lg={10} md={8} sm={4}>
        <Form onSubmit={handleSubmit}>
          <Tile style={{ marginBottom: 'var(--cds-spacing-05)' }}>
            <h3 className={styles.sectionTitle}>Basic Info</h3>
            <FormGroup legendText="">
              <TextInput
                id="exp-name"
                labelText="Experiment Name"
                placeholder="my-pod-kill"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />
            </FormGroup>
            <FormGroup legendText="" style={{ marginTop: 'var(--cds-spacing-05)' }}>
              <TextInput
                id="exp-namespace"
                labelText="Experiment Namespace"
                placeholder="default"
                value={namespace}
                onChange={(e) => setNamespace(e.target.value)}
                required
              />
            </FormGroup>
          </Tile>

          <Tile style={{ marginBottom: 'var(--cds-spacing-05)' }}>
            <h3 className={styles.sectionTitle}>Action</h3>
            <FormGroup legendText="">
              <Dropdown
                id="action-type"
                titleText="Action Type"
                label="Select action type"
                items={ACTION_ITEMS}
                itemToString={(item) => item ? `[${item.group}] ${item.label}` : ''}
                selectedItem={ACTION_ITEMS.find((i) => i.id === actionType) ?? null}
                onChange={({ selectedItem }) => {
                  if (selectedItem) handleActionChange(selectedItem.id as ActionType);
                }}
              />
            </FormGroup>

            {paramDefs.length > 0 && (
              <div style={{ marginTop: 'var(--cds-spacing-05)' }}>
                {paramDefs.map((def) => (
                  <FormGroup key={def.key} legendText="" style={{ marginBottom: 'var(--cds-spacing-04)' }}>
                    <TextInput
                      id={`param-${def.key}`}
                      labelText={def.label}
                      placeholder={def.placeholder}
                      value={params[def.key] ?? ''}
                      onChange={(e) => setParam(def.key, e.target.value)}
                    />
                  </FormGroup>
                ))}
              </div>
            )}

            <FormGroup legendText="" style={{ marginTop: 'var(--cds-spacing-05)' }}>
              <TextInput
                id="exp-duration"
                labelText="Duration"
                placeholder="30s"
                value={duration}
                onChange={(e) => setDuration(e.target.value)}
                helperText="e.g. 30s, 5m, 1h"
              />
            </FormGroup>
          </Tile>

          <Tile style={{ marginBottom: 'var(--cds-spacing-05)' }}>
            <h3 className={styles.sectionTitle}>Target</h3>
            <FormGroup legendText="">
              <TextInput
                id="target-namespace"
                labelText="Target Namespace"
                placeholder="default"
                value={targetNamespace}
                onChange={(e) => setTargetNamespace(e.target.value)}
                required
              />
            </FormGroup>
            <FormGroup legendText="" style={{ marginTop: 'var(--cds-spacing-05)' }}>
              <TextInput
                id="label-selector"
                labelText="Label Selector"
                placeholder="app=nginx, tier=frontend"
                value={labelSelector}
                onChange={(e) => setLabelSelector(e.target.value)}
                helperText="Comma-separated key=value pairs"
              />
            </FormGroup>
            <FormGroup legendText="" style={{ marginTop: 'var(--cds-spacing-05)' }}>
              <Select
                id="target-mode"
                labelText="Mode"
                value={mode}
                onChange={(e) => setMode(e.target.value)}
              >
                {MODE_OPTIONS.map((m) => (
                  <SelectItem key={m} value={m} text={m} />
                ))}
              </Select>
            </FormGroup>
          </Tile>

          {createMutation.isError && (
            <InlineNotification
              kind="error"
              title="Failed to create experiment"
              subtitle={createMutation.error?.message}
              style={{ marginBottom: 'var(--cds-spacing-05)' }}
            />
          )}

          <Button
            type="submit"
            renderIcon={Send}
            disabled={!name || !actionType || createMutation.isPending}
          >
            {createMutation.isPending ? 'Creating…' : 'Create Experiment'}
          </Button>
        </Form>
      </Column>

      <Column lg={6} md={8} sm={4}>
        <Tile style={{ position: 'sticky', top: 'var(--cds-spacing-09)' }}>
          <h3 className={styles.sectionTitle}>YAML Preview</h3>
          <pre className={styles.yamlPreview}>{yamlPreview}</pre>
        </Tile>
      </Column>
    </Grid>
  );
}
