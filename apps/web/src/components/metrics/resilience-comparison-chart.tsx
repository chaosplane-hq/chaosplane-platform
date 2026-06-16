'use client';

import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { useVizTheme, categoricalColor } from '@/components/viz';
import type { VizSize } from '@/components/viz';
import type { ComparisonDatum } from './transforms';

interface Props {
  size: VizSize;
  data: ComparisonDatum[];
}

export function ResilienceComparisonChart({ size, data }: Props) {
  const tokens = useVizTheme();

  return (
    <BarChart width={size.width} height={size.height} data={data} margin={{ top: 16, right: 24, bottom: 8, left: 0 }} barGap={4}>
      <CartesianGrid stroke={tokens['border-subtle']} strokeDasharray="3 3" vertical={false} />
      <XAxis dataKey="label" stroke={tokens['text-secondary']} fontSize={12} />
      <YAxis domain={[0, 100]} stroke={tokens['text-secondary']} fontSize={12} unit="%" width={48} />
      <Tooltip
        cursor={{ fill: tokens['layer-02'], opacity: 0.4 }}
        contentStyle={{
          background: tokens['layer-02'],
          border: `1px solid ${tokens['border-subtle']}`,
          color: tokens['text-primary'],
        }}
        formatter={(value) => `${value as number}%`}
      />
      <Legend wrapperStyle={{ fontSize: 12, color: tokens['text-secondary'] }} />
      <Bar dataKey="before" name="Before (previous)" fill={categoricalColor(7)} radius={[3, 3, 0, 0]} />
      <Bar dataKey="after" name="After (latest)" fill={categoricalColor(0)} radius={[3, 3, 0, 0]} />
    </BarChart>
  );
}
