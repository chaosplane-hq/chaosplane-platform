'use client';

import type { ExperimentStatus } from '@/lib/types';
import styles from './experiments.module.scss';

interface TimelineProps {
  status: ExperimentStatus;
}

function formatTime(iso?: string) {
  if (!iso) return '—';
  return new Date(iso).toLocaleString();
}

function calcDuration(start?: string, end?: string) {
  if (!start) return null;
  const diff = (end ? new Date(end).getTime() : Date.now()) - new Date(start).getTime();
  const mins = Math.floor(diff / 60000);
  const secs = Math.floor((diff % 60000) / 1000);
  return mins > 0 ? `${mins}m ${secs}s` : `${secs}s`;
}

export function ExperimentTimeline({ status }: TimelineProps) {
  const terminalLabel =
    status.phase === 'Aborted' ? 'Aborted' : status.phase === 'Failed' ? 'Failed' : 'Completed';

  const events = [
    { label: 'Created', time: status.startTime, done: !!status.startTime, last: false },
    { label: 'Running', time: status.startTime, done: ['Running', 'Completed', 'Failed', 'Aborted'].includes(status.phase), last: false },
    { label: terminalLabel, time: status.completionTime, done: !!status.completionTime, last: true },
  ];

  return (
    <div className={styles.timeline}>
      {events.map((evt) => (
        <div key={evt.label} className={`${styles.timelineItem} ${evt.done ? styles.timelineDone : ''}`}>
          <div className={styles.timelineDot} />
          {!evt.last && <div className={styles.timelineLine} />}
          <div className={styles.timelineContent}>
            <p className={styles.timelineLabel}>{evt.label}</p>
            <p className={styles.timelineTime}>{formatTime(evt.time)}</p>
          </div>
        </div>
      ))}
      {status.startTime && (
        <p className={styles.timelineDuration}>
          Duration: {calcDuration(status.startTime, status.completionTime) ?? '—'}
        </p>
      )}
    </div>
  );
}
