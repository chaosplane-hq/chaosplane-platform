'use client';

import dynamic from 'next/dynamic';
import { Grid, Column, Tile } from '@carbon/react';
import { VizContainer } from '@/components/viz';
import { DndKitSmoke } from '@/components/viz/smoke/dnd-kit-smoke';
import { FramerSmoke } from '@/components/viz/smoke/framer-smoke';

// d3-force and recharts touch the DOM/measure layout, so load them client-only
// (ssr:false) to keep them out of the server bundle and avoid hydration drift.
const D3ForceSmoke = dynamic(
  () => import('@/components/viz/smoke/d3-force-smoke').then((m) => m.D3ForceSmoke),
  { ssr: false },
);
const RechartsSmoke = dynamic(
  () => import('@/components/viz/smoke/recharts-smoke').then((m) => m.RechartsSmoke),
  { ssr: false },
);

const panelTitle: React.CSSProperties = {
  fontSize: '1rem',
  fontWeight: 600,
  margin: '0 0 var(--cds-spacing-04)',
};

export default function VizSmokePage() {
  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div style={{ padding: '2rem 0 1rem' }}>
          <h2 style={{ fontSize: '1.75rem', fontWeight: 600, marginBottom: '0.25rem' }}>
            Viz smoke
          </h2>
          <p style={{ color: 'var(--cds-text-secondary)' }}>
            Proves dnd-kit, d3, framer-motion and recharts render against the Carbon theme.
          </p>
        </div>
      </Column>

      <Column lg={8} md={4} sm={4} style={{ marginBottom: '1rem' }}>
        <Tile>
          <h3 style={panelTitle}>dnd-kit — sortable</h3>
          <DndKitSmoke />
        </Tile>
      </Column>

      <Column lg={8} md={4} sm={4} style={{ marginBottom: '1rem' }}>
        <Tile>
          <h3 style={panelTitle}>framer-motion — animation</h3>
          <FramerSmoke />
        </Tile>
      </Column>

      <Column lg={8} md={4} sm={4} style={{ marginBottom: '1rem' }}>
        <Tile>
          <h3 style={panelTitle}>d3 — force graph</h3>
          <VizContainer label="D3 force-directed graph of three services" height={280}>
            {(size) => <D3ForceSmoke size={size} />}
          </VizContainer>
        </Tile>
      </Column>

      <Column lg={8} md={4} sm={4} style={{ marginBottom: '1rem' }}>
        <Tile>
          <h3 style={panelTitle}>recharts — line chart</h3>
          <VizContainer label="Recharts line chart of baseline vs chaos availability" height={280}>
            {(size) => <RechartsSmoke size={size} />}
          </VizContainer>
        </Tile>
      </Column>
    </Grid>
  );
}
