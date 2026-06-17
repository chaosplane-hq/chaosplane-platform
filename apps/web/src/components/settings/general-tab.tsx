'use client';

import { useState } from 'react';
import { Tile, TextInput, Button, Form, FormGroup, InlineNotification } from '@carbon/react';
import { useHierarchy, usePatchWorkspace } from '@/lib/hooks/use-hierarchy';

export function GeneralTab() {
  const { data } = useHierarchy();
  const patchWorkspace = usePatchWorkspace();

  const workspace = data?.organizations?.[0]?.workspaces?.[0];
  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');
  const [initialized, setInitialized] = useState(false);

  // Initialize form values from server data once loaded
  if (workspace && !initialized) {
    setName(workspace.name);
    setSlug(workspace.slug ?? '');
    setInitialized(true);
  }

  const handleSave = () => {
    if (!workspace) return;
    patchWorkspace.mutate({ id: workspace.id, data: { name, slug } });
  };

  const handleDelete = () => {
    if (!confirm('Are you sure you want to delete this workspace? This action cannot be undone.')) return;
  };

  return (
    <div style={{ paddingTop: 'var(--cds-spacing-06)', maxWidth: '640px' }}>
      <Tile>
        <h3 style={{ fontSize: 'var(--cds-heading-02-font-size)', fontWeight: 'var(--cds-heading-02-font-weight)', color: 'var(--cds-text-primary)', margin: '0 0 var(--cds-spacing-06)' }}>
          Workspace
        </h3>
        {patchWorkspace.isSuccess && (
          <InlineNotification
            kind="success"
            title="Saved"
            subtitle="Workspace settings updated."
            lowContrast
            style={{ marginBottom: 'var(--cds-spacing-05)' }}
          />
        )}
        {patchWorkspace.isError && (
          <InlineNotification
            kind="error"
            title="Save failed"
            subtitle={(patchWorkspace.error as Error).message}
            lowContrast
            style={{ marginBottom: 'var(--cds-spacing-05)' }}
          />
        )}
        <Form onSubmit={(e: React.FormEvent) => { e.preventDefault(); handleSave(); }}>
          <FormGroup legendText="">
            <TextInput
              id="workspace-name"
              labelText="Workspace name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              style={{ marginBottom: 'var(--cds-spacing-05)' }}
            />
            <TextInput
              id="workspace-slug"
              labelText="Slug"
              value={slug}
              onChange={(e) => setSlug(e.target.value)}
              helperText="Used in URLs and API references."
              style={{ marginBottom: 'var(--cds-spacing-06)' }}
            />
            <Button kind="primary" type="submit" disabled={patchWorkspace.isPending || !name.trim()}>
              {patchWorkspace.isPending ? 'Saving…' : 'Save changes'}
            </Button>
          </FormGroup>
        </Form>
      </Tile>

      <Tile style={{ marginTop: 'var(--cds-spacing-05)', borderLeft: '3px solid var(--cds-support-error)' }}>
        <h3 style={{ fontSize: 'var(--cds-heading-02-font-size)', fontWeight: 'var(--cds-heading-02-font-weight)', color: 'var(--cds-text-primary)', margin: '0 0 var(--cds-spacing-03)' }}>
          Danger zone
        </h3>
        <p style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-body-short-01-font-size)', margin: '0 0 var(--cds-spacing-05)' }}>
          Deleting your workspace is permanent and cannot be undone.
        </p>
        <Button kind="danger" onClick={handleDelete}>Delete workspace</Button>
      </Tile>
    </div>
  );
}
