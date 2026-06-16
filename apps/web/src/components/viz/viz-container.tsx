'use client';

import { useEffect, useRef, useState, type ReactNode } from 'react';
import { SkeletonPlaceholder } from '@carbon/react';
import styles from './viz-container.module.scss';

export interface VizSize {
  width: number;
  height: number;
}

export interface VizContainerProps {
  height?: number;
  isLoading?: boolean;
  isEmpty?: boolean;
  emptyLabel?: string;
  label: string;
  children: (size: VizSize) => ReactNode;
}

export function VizContainer({
  height = 320,
  isLoading = false,
  isEmpty = false,
  emptyLabel = 'No data to display.',
  label,
  children,
}: VizContainerProps) {
  const ref = useRef<HTMLDivElement>(null);
  const [size, setSize] = useState<VizSize>({ width: 0, height });

  useEffect(() => {
    const el = ref.current;
    if (!el) return;

    // ResizeObserver drives responsive sizing so D3/Recharts get real pixel
    // dimensions instead of guessing from CSS percentages.
    const observer = new ResizeObserver((entries) => {
      const entry = entries[0];
      if (entry) {
        setSize({ width: entry.contentRect.width, height });
      }
    });
    observer.observe(el);
    return () => observer.disconnect();
  }, [height]);

  const showChildren = !isLoading && !isEmpty && size.width > 0;

  return (
    <div
      ref={ref}
      className={styles.container}
      style={{ height }}
      role="img"
      aria-label={label}
    >
      {isLoading ? (
        <SkeletonPlaceholder style={{ width: '100%', height: '100%' }} />
      ) : isEmpty ? (
        <div className={styles.state} style={{ height }}>
          {emptyLabel}
        </div>
      ) : showChildren ? (
        <div className={styles.canvas}>{children(size)}</div>
      ) : null}
    </div>
  );
}
