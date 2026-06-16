'use client';

import dynamic from 'next/dynamic';
import { VizContainer } from '@/components/viz';
import type { ResilienceScore } from '@/lib/types';
import {
  toDimensionData,
  toTrendData,
  toComparisonData,
  type DimensionDatum,
  type TrendDatum,
  type ComparisonDatum,
} from './transforms';
import styles from './metrics.module.scss';

// Recharts measures the DOM, so load each chart client-only (ssr:false) to keep
// it out of the server bundle and avoid hydration mismatch (per viz-smoke).
const ResilienceDimensionChart = dynamic(
  () => import('./resilience-dimension-chart').then((m) => m.ResilienceDimensionChart),
  { ssr: false },
);
const ResilienceTrendChart = dynamic(
  () => import('./resilience-trend-chart').then((m) => m.ResilienceTrendChart),
  { ssr: false },
);
const ResilienceComparisonChart = dynamic(
  () => import('./resilience-comparison-chart').then((m) => m.ResilienceComparisonChart),
  { ssr: false },
);

function ChartHeading({ title, caption }: { title: string; caption: string }) {
  return (
    <>
      <h3 className={styles.chartTitle}>{title}</h3>
      <p className={styles.chartCaption}>{caption}</p>
    </>
  );
}

function DimensionTable({ data }: { data: DimensionDatum[] }) {
  return (
    <table className={styles.srOnly}>
      <caption>Resilience dimension scores</caption>
      <thead>
        <tr>
          <th scope="col">Dimension</th>
          <th scope="col">Score</th>
        </tr>
      </thead>
      <tbody>
        {data.map((d) => (
          <tr key={d.key}>
            <th scope="row">{d.label}</th>
            <td>{d.value}%</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

function TrendTable({ data }: { data: TrendDatum[] }) {
  return (
    <table className={styles.srOnly}>
      <caption>Resilience score history over time</caption>
      <thead>
        <tr>
          <th scope="col">Calculated at</th>
          <th scope="col">Overall</th>
          <th scope="col">Availability</th>
          <th scope="col">Fault Tolerance</th>
          <th scope="col">Recoverability</th>
        </tr>
      </thead>
      <tbody>
        {data.map((d) => (
          <tr key={d.timestamp}>
            <th scope="row">{new Date(d.timestamp).toLocaleString()}</th>
            <td>{d.overall}%</td>
            <td>{d.availability}%</td>
            <td>{d.faultTolerance}%</td>
            <td>{d.recoverability}%</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

function ComparisonTable({ data }: { data: ComparisonDatum[] }) {
  return (
    <table className={styles.srOnly}>
      <caption>Resilience scores before (previous) versus after (latest) recalculation</caption>
      <thead>
        <tr>
          <th scope="col">Dimension</th>
          <th scope="col">Before</th>
          <th scope="col">After</th>
          <th scope="col">Change</th>
        </tr>
      </thead>
      <tbody>
        {data.map((d) => (
          <tr key={d.key}>
            <th scope="row">{d.label}</th>
            <td>{d.before}%</td>
            <td>{d.after}%</td>
            <td>{d.delta >= 0 ? `+${d.delta}` : d.delta}%</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

export function ResilienceDimensionPanel({ score, isLoading }: { score: ResilienceScore | null; isLoading?: boolean }) {
  const data = score ? toDimensionData(score) : [];
  const label = score
    ? `Radial bar chart of resilience dimension scores. Availability ${data[0]?.value}%, fault tolerance ${data[1]?.value}%, recoverability ${data[2]?.value}%, overall ${data[3]?.value}%.`
    : 'Resilience dimension scores';

  return (
    <div>
      <ChartHeading title="Dimensional Scores" caption="Current scores across the three weighted resilience dimensions, with the overall ring." />
      <VizContainer
        label={label}
        height={300}
        isLoading={isLoading}
        isEmpty={!score}
        emptyLabel="No dimensional data available."
      >
        {(size) => <ResilienceDimensionChart size={size} data={data} />}
      </VizContainer>
      {score && <DimensionTable data={data} />}
    </div>
  );
}

export function ResilienceTrendPanel({ history, isLoading }: { history: ResilienceScore[]; isLoading?: boolean }) {
  const data = toTrendData(history);
  return (
    <div>
      <ChartHeading title="Resilience Trend" caption="How overall and per-dimension scores evolved across recalculations." />
      <VizContainer
        label={`Line chart of resilience scores over ${data.length} recalculations.`}
        height={320}
        isLoading={isLoading}
        isEmpty={data.length < 2}
        emptyLabel="Not enough history to plot a trend. Recalculate again to build a series."
      >
        {(size) => <ResilienceTrendChart size={size} data={data} />}
      </VizContainer>
      {data.length >= 2 && <TrendTable data={data} />}
    </div>
  );
}

export function ResilienceComparisonPanel({ history, isLoading }: { history: ResilienceScore[]; isLoading?: boolean }) {
  const data = toComparisonData(history);
  return (
    <div>
      <ChartHeading title="Before / After Comparison" caption="Latest scores versus the previous recalculation, per dimension." />
      <VizContainer
        label="Grouped bar chart comparing resilience scores before (previous recalculation) and after (latest)."
        height={300}
        isLoading={isLoading}
        isEmpty={!data}
        emptyLabel="Need at least two recalculations to compare before and after."
      >
        {(size) => <ResilienceComparisonChart size={size} data={data ?? []} />}
      </VizContainer>
      {data && <ComparisonTable data={data} />}
    </div>
  );
}
