'use client';

import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { Draggable as DraggableIcon, TrashCan, WarningFilled, Time, Network_3 } from '@carbon/icons-react';
import { IconButton } from '@carbon/react';
import { categoricalColor } from '@/components/viz';
import type { BuilderStep } from './model';
import styles from './scenario.module.scss';

interface StepCardProps {
  step: BuilderStep;
  index: number;
  accentIndex: number;
  selected: boolean;
  valid: boolean;
  onSelect: (id: string) => void;
  onRemove: (id: string) => void;
}

export function StepCard({
  step,
  index,
  accentIndex,
  selected,
  valid,
  onSelect,
  onRemove,
}: StepCardProps) {
  const { attributes, listeners, setNodeRef, setActivatorNodeRef, transform, transition, isDragging } =
    useSortable({ id: step.id });

  const accent = categoricalColor(accentIndex);

  return (
    <li
      ref={setNodeRef}
      className={[
        styles.stepCard,
        selected ? styles.stepCardSelected : '',
        !valid ? styles.stepCardInvalid : '',
        isDragging ? styles.stepCardDragging : '',
      ]
        .filter(Boolean)
        .join(' ')}
      style={{
        transform: CSS.Transform.toString(transform),
        transition,
        ['--group-accent' as string]: accent,
      }}
    >
      <button
        ref={setActivatorNodeRef}
        type="button"
        className={styles.stepHandle}
        aria-label={`Reorder step ${step.name}`}
        {...attributes}
        {...listeners}
      >
        <DraggableIcon size={20} aria-hidden />
      </button>

      <button
        type="button"
        className={styles.stepBody}
        onClick={() => onSelect(step.id)}
        aria-pressed={selected}
      >
        <div className={styles.stepTopRow}>
          <span className={styles.stepIndex}>{index + 1}</span>
          <span className={styles.stepName}>{step.name || 'unnamed step'}</span>
          {!valid && (
            <WarningFilled size={16} style={{ color: 'var(--cds-support-error)' }} aria-label="Incomplete" />
          )}
        </div>
        <span className={styles.stepType}>{step.type}</span>
        <div className={styles.stepMeta}>
          <span className={styles.stepMetaItem}>
            <Network_3 size={14} aria-hidden /> {step.targetNamespace}/{step.mode}
          </span>
          <span className={styles.stepMetaItem}>
            <Time size={14} aria-hidden /> {step.duration || '—'}
          </span>
        </div>
        {step.dependsOn.length > 0 && (
          <div className={styles.stepDepends}>runs after: {step.dependsOn.join(', ')}</div>
        )}
      </button>

      <div className={styles.stepActions}>
        <IconButton label="Remove step" kind="ghost" size="sm" align="left" onClick={() => onRemove(step.id)}>
          <TrashCan size={16} />
        </IconButton>
      </div>
    </li>
  );
}
