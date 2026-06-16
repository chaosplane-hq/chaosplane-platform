'use client';

import {
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { useVizTheme, categoricalColor } from '../theme';
import type { VizSize } from '../viz-container';

const DATA = [
  { run: 'R1', baseline: 99.9, chaos: 98.2 },
  { run: 'R2', baseline: 99.8, chaos: 96.5 },
  { run: 'R3', baseline: 99.9, chaos: 97.8 },
  { run: 'R4', baseline: 99.7, chaos: 94.1 },
  { run: 'R5', baseline: 99.9, chaos: 98.9 },
];

export function RechartsSmoke({ size }: { size: VizSize }) {
  const tokens = useVizTheme();

  return (
    <ResponsiveContainer width={size.width} height={size.height}>
      <LineChart data={DATA} margin={{ top: 16, right: 16, bottom: 8, left: 0 }}>
        <CartesianGrid stroke={tokens['border-subtle']} strokeDasharray="3 3" />
        <XAxis dataKey="run" stroke={tokens['text-secondary']} fontSize={12} />
        <YAxis domain={[90, 100]} stroke={tokens['text-secondary']} fontSize={12} />
        <Tooltip
          contentStyle={{
            background: tokens['layer-02'],
            border: `1px solid ${tokens['border-subtle']}`,
            color: tokens['text-primary'],
          }}
        />
        <Line
          type="monotone"
          dataKey="baseline"
          stroke={categoricalColor(2)}
          strokeWidth={2}
          dot={false}
        />
        <Line
          type="monotone"
          dataKey="chaos"
          stroke={categoricalColor(4)}
          strokeWidth={2}
          dot={false}
        />
      </LineChart>
    </ResponsiveContainer>
  );
}
