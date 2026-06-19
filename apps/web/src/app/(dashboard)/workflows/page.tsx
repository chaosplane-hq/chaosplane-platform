'use client';

import { useState } from 'react';
import {
  Grid,
  Column,
  Tile,
  Button,
  Tag,
  Modal,
  TextInput,
  TextArea,
  Select,
  SelectItem,
  Toggle,
  SkeletonText,
  InlineNotification,
  CodeSnippet,
} from '@carbon/react';
import { Add, TrashCan } from '@carbon/icons-react';
import { useWorkflowTemplates, useCreateWorkflowTemplate, useDeleteWorkflowTemplate } from '@/lib/hooks/use-workflows';
import type { WorkflowCategory } from '@/lib/types';
import styles from '@/components/experiments/experiments.module.scss';

const CATEGORY_TAG: Record<WorkflowCategory, 'blue' | 'green' | 'red' | 'gray'> = {
  chaos: 'red',
  load: 'blue',
  security: 'green',
  custom: 'gray',
};

const DEFAULT_SPEC = `{
  "steps": [
    {
      "name": "step-1",
      "action": "pod-kill",
      "target": { "namespace": "default" }
    }
  ]
}`;

export default function WorkflowsPage() {
  const { data, isLoading, isError: queryError, error: queryErrorObj } = useWorkflowTemplates();
  const createTemplate = useCreateWorkflowTemplate();
  const deleteTemplate = useDeleteWorkflowTemplate();

  const [createOpen, setCreateOpen] = useState(false);
  const [previewId, setPreviewId] = useState<string | null>(null);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [category, setCategory] = useState<WorkflowCategory>('chaos');
  const [isPublic, setIsPublic] = useState(false);
  const [specText, setSpecText] = useState(DEFAULT_SPEC);
  const [specError, setSpecError] = useState('');
  const [error, setError] = useState('');

  const templates = data?.items ?? [];
  const previewTemplate = templates.find((t) => t.id === previewId);

  async function handleCreate() {
    setSpecError('');
    setError('');
    let spec: Record<string, unknown>;
    try {
      spec = JSON.parse(specText);
    } catch {
      setSpecError('Invalid JSON spec.');
      return;
    }
    try {
      await createTemplate.mutateAsync({
        name: name.trim(),
        description: description.trim() || undefined,
        category,
        isPublic,
        spec,
      });
      setCreateOpen(false);
      setName('');
      setDescription('');
      setCategory('chaos');
      setIsPublic(false);
      setSpecText(DEFAULT_SPEC);
    } catch {
      setError('Failed to create workflow template.');
    }
  }

  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div className={styles.pageHeader} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <h2 className={styles.pageTitle}>Workflow Templates</h2>
            <p className={styles.pageSubtitle}>Reusable chaos workflow definitions.</p>
          </div>
          <Button renderIcon={Add} onClick={() => setCreateOpen(true)}>
            Create Template
          </Button>
        </div>
      </Column>

      {queryError && (
        <Column lg={16} md={8} sm={4}>
          <InlineNotification
            kind="error"
            title="Failed to load workflow templates"
            subtitle={(queryErrorObj as Error)?.message ?? ''}
            style={{ marginBottom: 'var(--cds-spacing-05)' }}
          />
        </Column>
      )}

      {error && (
        <Column lg={16} md={8} sm={4}>
          <InlineNotification
            kind="error"
            title={error}
            onCloseButtonClick={() => setError('')}
            style={{ marginBottom: 'var(--cds-spacing-05)' }}
          />
        </Column>
      )}

      {deleteTemplate.isError && (
        <Column lg={16} md={8} sm={4}>
          <InlineNotification
            kind="error"
            title="Operation failed"
            subtitle={(deleteTemplate.error as Error)?.message ?? ''}
            style={{ marginBottom: 'var(--cds-spacing-05)' }}
          />
        </Column>
      )}

      <Column lg={16} md={8} sm={4}>
        {isLoading ? (
          <SkeletonText paragraph lineCount={6} />
        ) : templates.length === 0 ? (
          <Tile>
            <p style={{ color: 'var(--cds-text-secondary)', textAlign: 'center' }}>No workflow templates yet.</p>
          </Tile>
        ) : (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))', gap: 'var(--cds-spacing-05)' }}>
            {templates.map((t) => (
              <Tile key={t.id} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--cds-spacing-03)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                  <h4 style={{ margin: 0, fontSize: 'var(--cds-heading-02-font-size)', fontWeight: 'var(--cds-heading-02-font-weight)', color: 'var(--cds-text-primary)' }}>
                    {t.name}
                  </h4>
                  <Button
                    kind="danger--ghost"
                    size="sm"
                    renderIcon={TrashCan}
                    iconDescription="Delete"
                    hasIconOnly
                    onClick={() => deleteTemplate.mutate(t.id)}
                    disabled={deleteTemplate.isPending}
                  />
                </div>
                <div style={{ display: 'flex', gap: 'var(--cds-spacing-02)', flexWrap: 'wrap' }}>
                  <Tag type={CATEGORY_TAG[(t.category as WorkflowCategory) ?? 'custom']} size="sm">{t.category ?? 'custom'}</Tag>
                  {t.isPublic && <Tag type="teal" size="sm">public</Tag>}
                </div>
                {t.description && (
                  <p style={{ margin: 0, fontSize: 'var(--cds-label-01-font-size)', color: 'var(--cds-text-secondary)' }}>
                    {t.description}
                  </p>
                )}
                <Button
                  kind="ghost"
                  size="sm"
                  onClick={() => setPreviewId(t.id)}
                  style={{ alignSelf: 'flex-start', padding: 0 }}
                >
                  View spec
                </Button>
              </Tile>
            ))}
          </div>
        )}
      </Column>

      <Modal
        open={createOpen}
        modalHeading="Create Workflow Template"
        primaryButtonText="Create"
        secondaryButtonText="Cancel"
        onRequestSubmit={handleCreate}
        onRequestClose={() => { setCreateOpen(false); setSpecError(''); setError(''); }}
        primaryButtonDisabled={!name.trim() || createTemplate.isPending}
        size="lg"
      >
        <TextInput
          id="wf-name"
          labelText="Name"
          placeholder="e.g. Pod Kill Workflow"
          value={name}
          onChange={(e) => setName(e.target.value)}
          style={{ marginBottom: 'var(--cds-spacing-05)' }}
        />
        <TextArea
          id="wf-desc"
          labelText="Description (optional)"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={2}
          style={{ marginBottom: 'var(--cds-spacing-05)' }}
        />
        <Select
          id="wf-category"
          labelText="Category"
          value={category}
          onChange={(e) => setCategory(e.target.value as WorkflowCategory)}
          style={{ marginBottom: 'var(--cds-spacing-05)' }}
        >
          <SelectItem value="chaos" text="Chaos" />
          <SelectItem value="load" text="Load" />
          <SelectItem value="security" text="Security" />
          <SelectItem value="custom" text="Custom" />
        </Select>
        <Toggle
          id="wf-public"
          labelText="Public template"
          labelA="Private"
          labelB="Public"
          toggled={isPublic}
          onToggle={(v) => setIsPublic(v)}
          style={{ marginBottom: 'var(--cds-spacing-05)' }}
        />
        <TextArea
          id="wf-spec"
          labelText="Spec (JSON)"
          value={specText}
          onChange={(e) => setSpecText(e.target.value)}
          rows={10}
          invalid={!!specError}
          invalidText={specError}
        />
      </Modal>

      <Modal
        open={!!previewId}
        modalHeading={previewTemplate ? `Spec: ${previewTemplate.name}` : 'Spec'}
        primaryButtonText="Close"
        onRequestSubmit={() => setPreviewId(null)}
        onRequestClose={() => setPreviewId(null)}
        size="lg"
      >
        {previewTemplate && (
          <CodeSnippet type="multi" feedback="Copied!">
            {JSON.stringify(previewTemplate.spec, null, 2)}
          </CodeSnippet>
        )}
      </Modal>
    </Grid>
  );
}
