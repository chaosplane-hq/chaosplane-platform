import {
  Grid,
  Column,
  Tile,
  Tag,
} from '@carbon/react';

export default function DashboardPage() {
  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div style={{ marginBottom: 'var(--cds-spacing-07)' }}>
          <h2 style={{
            fontSize: 'var(--cds-heading-04-font-size)',
            fontWeight: 'var(--cds-heading-04-font-weight)',
            color: 'var(--cds-text-primary)',
            margin: '0 0 var(--cds-spacing-02)',
          }}>
            Dashboard
          </h2>
          <p style={{ color: 'var(--cds-text-secondary)', margin: 0 }}>
            Welcome to ChaosPlane. Monitor and manage your chaos experiments.
          </p>
        </div>
      </Column>

      <Column lg={4} md={4} sm={4}>
        <Tile>
          <p style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)', marginBottom: 'var(--cds-spacing-03)' }}>
            Total Experiments
          </p>
          <p style={{ fontSize: '2rem', fontWeight: 600, color: 'var(--cds-text-primary)', margin: 0 }}>0</p>
        </Tile>
      </Column>

      <Column lg={4} md={4} sm={4}>
        <Tile>
          <p style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)', marginBottom: 'var(--cds-spacing-03)' }}>
            Running
          </p>
          <p style={{ fontSize: '2rem', fontWeight: 600, color: 'var(--cds-text-primary)', margin: 0 }}>0</p>
        </Tile>
      </Column>

      <Column lg={4} md={4} sm={4}>
        <Tile>
          <p style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)', marginBottom: 'var(--cds-spacing-03)' }}>
            Environments
          </p>
          <p style={{ fontSize: '2rem', fontWeight: 600, color: 'var(--cds-text-primary)', margin: 0 }}>0</p>
        </Tile>
      </Column>

      <Column lg={4} md={4} sm={4}>
        <Tile>
          <p style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)', marginBottom: 'var(--cds-spacing-03)' }}>
            Status
          </p>
          <Tag type="green" size="md">Healthy</Tag>
        </Tile>
      </Column>

      <Column lg={16} md={8} sm={4}>
        <Tile style={{ marginTop: 'var(--cds-spacing-05)' }}>
          <p style={{ color: 'var(--cds-text-secondary)', textAlign: 'center', padding: 'var(--cds-spacing-09) 0' }}>
            No experiments yet. Create your first experiment to get started.
          </p>
        </Tile>
      </Column>
    </Grid>
  );
}
