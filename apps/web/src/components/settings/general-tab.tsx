'use client';

import { Tile, TextInput, Button, Form, FormGroup } from '@carbon/react';

export function GeneralTab() {
  return (
    <div style={{ paddingTop: 'var(--cds-spacing-06)', maxWidth: '640px' }}>
      <Tile>
        <h3 style={{ fontSize: 'var(--cds-heading-02-font-size)', fontWeight: 'var(--cds-heading-02-font-weight)', color: 'var(--cds-text-primary)', margin: '0 0 var(--cds-spacing-06)' }}>
          Workspace
        </h3>
        <Form>
          <FormGroup legendText="">
            <TextInput
              id="workspace-name"
              labelText="Workspace name"
              defaultValue="My Workspace"
              style={{ marginBottom: 'var(--cds-spacing-05)' }}
            />
            <TextInput
              id="workspace-slug"
              labelText="Slug"
              defaultValue="my-workspace"
              helperText="Used in URLs and API references."
              style={{ marginBottom: 'var(--cds-spacing-06)' }}
            />
            <Button kind="primary">Save changes</Button>
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
        <Button kind="danger">Delete workspace</Button>
      </Tile>
    </div>
  );
}
