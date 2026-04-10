'use client';

import { Tag } from '@carbon/react';
import type { ExperimentPhase } from '@/lib/types';

const phaseToTagType: Record<ExperimentPhase, 'gray' | 'blue' | 'green' | 'red' | 'high-contrast'> = {
  Pending: 'gray',
  Running: 'blue',
  Completed: 'green',
  Failed: 'red',
  Aborted: 'high-contrast',
};

interface StatusTagProps {
  phase: ExperimentPhase;
  size?: 'sm' | 'md' | 'lg';
}

export function StatusTag({ phase, size = 'md' }: StatusTagProps) {
  return (
    <Tag type={phaseToTagType[phase]} size={size}>
      {phase}
    </Tag>
  );
}
