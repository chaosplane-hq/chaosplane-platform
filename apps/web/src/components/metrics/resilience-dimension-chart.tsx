'use client';

import {
  RadialBar,
  RadialBarChart,
  PolarAngleAxis,
  Tooltip,
} from 'recharts';
import { useVizTheme, categoricalColor } from '@/components/viz';
import type { VizSize } from '@/components/viz';
import type { DimensionDatum } from './transforms';

interface Props {
  size: VizSize;
  data: DimensionDatum[];
}

function scoreColor(value: number): number {
  if (value >= 80) return 6;
  if (value >= 60) return 9;
  return 4;
}

export function ResilienceDimensionChart({ size, data }: Props) {
  const tokens = useVizTheme();

  const chartData = [...data]
    .sort((a, b) => a.value - b.value)
    .map((d) => ({ ...d, fill: categoricalColor(scoreColor(d.value)) }));

  return (
    <RadialBarChart
      width={size.width}
      height={size.height}
      cx="50%"
      cy="50%"
      innerRadius="30%"
      outerRadius="100%"
      barSize={14}
      data={chartData}
      startAngle={90}
      endAngle={-270}
    >
      <PolarAngleAxis type="number" domain={[0, 100]} tick={false} />
      <RadialBar background={{ fill: tokens['layer-02'] }} dataKey="value" cornerRadius={7} />
      <Tooltip
        cursor={false}
        contentStyle={{
          background: tokens['layer-02'],
          border: `1px solid ${tokens['border-subtle']}`,
          color: tokens['text-primary'],
        }}
        formatter={(value, _name, item) => [
          `${value as number}%`,
          (item?.payload as DimensionDatum | undefined)?.label ?? '',
        ]}
      />
    </RadialBarChart>
  );
}
