'use client';

import {
  Grid,
  Column,
  Button,
  Tag,
  SkeletonText,
  InlineNotification,
  Tile,
  OverflowMenu,
  OverflowMenuItem,
} from '@carbon/react';
import { Play } from '@carbon/icons-react';
import { usePredictions, useRunPredictions, usePatchPredictionStatus } from '@/lib/hooks/use-predictions';
import type { PredictionSeverity, PredictionStatus, Prediction } from '@/lib/types';
import styles from '@/components/experiments/experiments.module.scss';

const SEVERITY_TAG: Record<PredictionSeverity, 'red' | 'magenta' | 'teal' | 'blue'> = {
  critical: 'red',
  high: 'magenta',
  medium: 'teal',
  low: 'blue',
};

const STATUS_TAG: Record<PredictionStatus, 'red' | 'gray' | 'green' | 'purple'> = {
  active: 'red',
  acknowledged: 'purple',
  resolved: 'green',
  dismissed: 'gray',
};

export default function PredictionsPage() {
  const { data, isLoading, isError, error } = usePredictions();
  const runMutation = useRunPredictions();
  const patchMutation = usePatchPredictionStatus();

  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end' }}>
          <div className={styles.pageHeader}>
            <h2 className={styles.pageTitle}>Predictions</h2>
            <p className={styles.pageSubtitle}>AI-powered failure predictions and recommended actions.</p>
          </div>
          <Button
            renderIcon={Play}
            disabled={runMutation.isPending}
            onClick={() => runMutation.mutate()}
            style={{ marginBottom: 'var(--cds-spacing-06)' }}
          >
            {runMutation.isPending ? 'Running…' : 'Run Analysis'}
          </Button>
        </div>
      </Column>

      <Column lg={16} md={8} sm={4}>
        {isError && (
          <InlineNotification
            kind="error"
            title="Failed to load predictions"
            subtitle={(error as Error)?.message}
            style={{ marginBottom: 'var(--cds-spacing-05)' }}
          />
        )}
        {runMutation.isSuccess && (
          <InlineNotification
            kind="success"
            title="Analysis started"
            subtitle="Predictions will update shortly."
            style={{ marginBottom: 'var(--cds-spacing-05)' }}
          />
        )}

        {isLoading && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--cds-spacing-04)' }}>
            {Array.from({ length: 4 }).map((_, i) => (
              <Tile key={i}><SkeletonText paragraph lineCount={3} /></Tile>
            ))}
          </div>
        )}

        {!isLoading && data && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--cds-spacing-04)' }}>
            {(data.items ?? []).map((p: Prediction) => (
              <Tile key={p.id}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                  <div style={{ flex: 1 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--cds-spacing-03)', marginBottom: 'var(--cds-spacing-03)' }}>
                      <Tag type={SEVERITY_TAG[p.severity]} size="sm">{p.severity}</Tag>
                      <Tag type={STATUS_TAG[p.status]} size="sm">{p.status}</Tag>
                      <span style={{ fontSize: 'var(--cds-label-01-font-size)', color: 'var(--cds-text-secondary)' }}>
                        {Math.round(p.confidence * 100)}% confidence
                      </span>
                    </div>
                    <strong style={{ display: 'block', marginBottom: 'var(--cds-spacing-02)' }}>{p.title}</strong>
                    <p style={{ fontSize: 'var(--cds-body-short-01-font-size)', color: 'var(--cds-text-secondary)', margin: '0 0 var(--cds-spacing-03)' }}>
                      {p.description}
                    </p>
                    <div style={{ fontSize: 'var(--cds-label-01-font-size)', color: 'var(--cds-text-primary)' }}>
                      <strong>Recommended:</strong> {p.recommendedAction}
                    </div>
                  </div>
                  {p.status === 'active' && (
                    <OverflowMenu flipped>
                      <OverflowMenuItem
                        itemText="Acknowledge"
                        onClick={() => patchMutation.mutate({ id: p.id, data: { status: 'acknowledged' } })}
                      />
                      <OverflowMenuItem
                        itemText="Resolve"
                        onClick={() => patchMutation.mutate({ id: p.id, data: { status: 'resolved' } })}
                      />
                      <OverflowMenuItem
                        itemText="Dismiss"
                        isDelete
                        onClick={() => patchMutation.mutate({ id: p.id, data: { status: 'dismissed' } })}
                      />
                    </OverflowMenu>
                  )}
                </div>
              </Tile>
            ))}
            {(data.items ?? []).length === 0 && (
              <p style={{ color: 'var(--cds-text-secondary)' }}>No predictions available. Run an analysis to get started.</p>
            )}
          </div>
        )}
      </Column>
    </Grid>
  );
}
