'use client';

import {
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { useVizTheme, categoricalColor } from '@/components/viz';
import type { VizSize } from '@/components/viz';
import type { TrendDatum } from './transforms';

interface Props {
  size: VizSize;
  data: TrendDatum[];
}

const SERIES: ReadonlyArray<{ key: keyof TrendDatum; label: string; color: number }> = [
  { key: 'overall', label: 'Overall', color: 0 },
  { key: 'availability', label: 'Availability', color: 6 },
  { key: 'faultTolerance', label: 'Fault Tolerance', color: 1 },
  { key: 'recoverability', label: 'Recoverability', color: 3 },
];

export function ResilienceTrendChart({ size, data }: Props) {
  const tokens = useVizTheme();

  return (
    <LineChart width={size.width} height={size.height} data={data} margin={{ top: 16, right: 24, bottom: 8, left: 0 }}>
      <CartesianGrid stroke={tokens['border-subtle']} strokeDasharray="3 3" />
      <XAxis dataKey="label" stroke={tokens['text-secondary']} fontSize={12} minTickGap={24} />
      <YAxis domain={[0, 100]} stroke={tokens['text-secondary']} fontSize={12} unit="%" width={48} />
      <Tooltip
        contentStyle={{
          background: tokens['layer-02'],
          border: `1px solid ${tokens['border-subtle']}`,
          color: tokens['text-primary'],
        }}
        formatter={(value) => `${value as number}%`}
      />
      <Legend wrapperStyle={{ fontSize: 12, color: tokens['text-secondary'] }} />
      {SERIES.map((s) => (
        <Line
          key={s.key}
          type="monotone"
          dataKey={s.key}
          name={s.label}
          stroke={categoricalColor(s.color)}
          strokeWidth={s.key === 'overall' ? 3 : 1.5}
          dot={{ r: 2 }}
          activeDot={{ r: 4 }}
        />
      ))}
    </LineChart>
  );
}
