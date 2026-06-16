import type { ResilienceScore } from '@/lib/types';

// The backend's only real numeric metric source is the resilience score
// (overall + 3 weighted dimensions, persisted per recalculation). result-analysis
// `metricsImpact` is written as an empty `{}` blob, so there is no per-experiment
// latency/error series to chart — we chart what is genuinely measured.

export interface DimensionDatum {
  key: string;
  label: string;
  value: number;
}

export interface TrendDatum {
  label: string;
  timestamp: string;
  overall: number;
  availability: number;
  faultTolerance: number;
  recoverability: number;
}

export interface ComparisonDatum {
  key: string;
  label: string;
  before: number;
  after: number;
  delta: number;
}

export const DIMENSION_META: ReadonlyArray<{ key: keyof DimensionSource; label: string }> = [
  { key: 'availability', label: 'Availability' },
  { key: 'faultTolerance', label: 'Fault Tolerance' },
  { key: 'recoverability', label: 'Recoverability' },
  { key: 'overallScore', label: 'Overall' },
];

type DimensionSource = Pick<
  ResilienceScore,
  'availability' | 'faultTolerance' | 'recoverability' | 'overallScore'
>;

function round(n: number): number {
  return Math.round(n * 10) / 10;
}

function shortLabel(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '—';
  return d.toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function toDimensionData(score: ResilienceScore): DimensionDatum[] {
  return [
    { key: 'availability', label: 'Availability', value: round(score.availability) },
    { key: 'faultTolerance', label: 'Fault Tolerance', value: round(score.faultTolerance) },
    { key: 'recoverability', label: 'Recoverability', value: round(score.recoverability) },
    { key: 'overall', label: 'Overall', value: round(score.overallScore) },
  ];
}

/**
 * History arrives newest-first; reverse to oldest-first so the time axis reads
 * left-to-right and the trend shows resilience evolving across recalculations.
 */
export function toTrendData(history: ResilienceScore[]): TrendDatum[] {
  return [...history].reverse().map((s) => ({
    label: shortLabel(s.calculatedAt),
    timestamp: s.calculatedAt,
    overall: round(s.overallScore),
    availability: round(s.availability),
    faultTolerance: round(s.faultTolerance),
    recoverability: round(s.recoverability),
  }));
}

/**
 * Before/after comparison from the two most recent snapshots: "before" is the
 * prior recalculation, "after" is the latest. Returns null when there is no
 * prior snapshot to compare against.
 */
export function toComparisonData(history: ResilienceScore[]): ComparisonDatum[] | null {
  if (history.length < 2) return null;
  const after = history[0];
  const before = history[1];
  return DIMENSION_META.map(({ key, label }) => {
    const b = round(before[key]);
    const a = round(after[key]);
    return { key: String(key), label, before: b, after: a, delta: round(a - b) };
  });
}
