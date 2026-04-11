'use client';

import {
  Grid,
  Column,
  Tile,
  Button,
  Tag,
  SkeletonText,
  InlineNotification,
} from '@carbon/react';
import { Add, TrashCan, Launch } from '@carbon/icons-react';
import { useRouter } from 'next/navigation';
import { useSuggestions, useGenerateSuggestions, useDeleteSuggestion } from '@/lib/hooks/use-suggestions';
import type { SuggestionWithConfidence } from '@/lib/types';

function confidenceColor(score: number): string {
  if (score >= 0.8) return 'var(--cds-support-success)';
  if (score >= 0.5) return 'var(--cds-support-warning)';
  return 'var(--cds-support-error)';
}

function SuggestionCard({ suggestion, onDelete }: { suggestion: SuggestionWithConfidence; onDelete: (id: string) => void }) {
  const router = useRouter();

  function handleCreateExperiment() {
    const params = new URLSearchParams();
    if (suggestion.targetService) params.set('target', suggestion.targetService);
    if (suggestion.experimentParams) {
      for (const [k, v] of Object.entries(suggestion.experimentParams)) {
        params.set(k, String(v));
      }
    }
    router.push(`/experiments/create?${params.toString()}`);
  }

  return (
    <Tile style={{ padding: '1.5rem', display: 'flex', flexDirection: 'column', gap: '0.75rem', height: '100%' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: '1rem' }}>
        <h4 style={{ fontSize: '1rem', fontWeight: 600, margin: 0 }}>{suggestion.title}</h4>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexShrink: 0 }}>
          <span style={{ fontSize: '0.75rem', color: 'var(--cds-text-secondary)' }}>Confidence</span>
          <span style={{ fontWeight: 700, color: confidenceColor(suggestion.confidence) }}>
            {Math.round(suggestion.confidence * 100)}%
          </span>
        </div>
      </div>

      <p style={{ color: 'var(--cds-text-secondary)', fontSize: '0.875rem', margin: 0, flexGrow: 1 }}>
        {suggestion.description}
      </p>

      <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
        {suggestion.category && <Tag type="blue" size="sm">{suggestion.category}</Tag>}
        {suggestion.targetService && <Tag type="gray" size="sm">{suggestion.targetService}</Tag>}
      </div>

      <div style={{ display: 'flex', gap: '0.5rem', marginTop: '0.25rem' }}>
        <Button
          kind="primary"
          size="sm"
          renderIcon={Launch}
          onClick={handleCreateExperiment}
        >
          Create Experiment
        </Button>
        <Button
          kind="danger--ghost"
          size="sm"
          renderIcon={TrashCan}
          onClick={() => onDelete(suggestion.id)}
        >
          Delete
        </Button>
      </div>
    </Tile>
  );
}

export default function SuggestionsPage() {
  const { data, isLoading, isError } = useSuggestions({ limit: 50 });
  const generate = useGenerateSuggestions();
  const deleteSuggestion = useDeleteSuggestion();

  const suggestions = (data?.suggestions ?? []) as SuggestionWithConfidence[];

  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div style={{ padding: '2rem 0 1rem', display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end' }}>
          <div>
            <h2 style={{ fontSize: '1.75rem', fontWeight: 600, marginBottom: '0.25rem' }}>Suggestions</h2>
            <p style={{ color: 'var(--cds-text-secondary)' }}>AI-generated experiment suggestions based on your topology.</p>
          </div>
          <Button
            renderIcon={Add}
            kind="primary"
            disabled={generate.isPending}
            onClick={() => generate.mutate()}
          >
            {generate.isPending ? 'Generating…' : 'Generate Suggestions'}
          </Button>
        </div>
      </Column>

      <Column lg={16} md={8} sm={4}>
        {isLoading ? (
          <SkeletonText paragraph lineCount={6} />
        ) : isError ? (
          <InlineNotification kind="error" title="Failed to load suggestions" subtitle="" />
        ) : suggestions.length === 0 ? (
          <Tile style={{ textAlign: 'center', padding: '3rem' }}>
            <p style={{ color: 'var(--cds-text-secondary)' }}>No suggestions yet. Click "Generate Suggestions" to get started.</p>
          </Tile>
        ) : (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(360px, 1fr))', gap: '1rem' }}>
            {suggestions.map((s) => (
              <SuggestionCard
                key={s.id}
                suggestion={s}
                onDelete={(id) => deleteSuggestion.mutate(id)}
              />
            ))}
          </div>
        )}
      </Column>
    </Grid>
  );
}
