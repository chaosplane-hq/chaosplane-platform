'use client';

import { useDroppable } from '@dnd-kit/core';
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable';
import { FlowConnection } from '@carbon/icons-react';
import type { BuilderStep, StepValidation } from './model';
import { isStepValid } from './model';
import { StepCard } from './step-card';
import styles from './scenario.module.scss';

export const CANVAS_DROPPABLE_ID = 'scenario-canvas';

interface ScenarioCanvasProps {
  steps: BuilderStep[];
  selectedId: string | null;
  validations: Map<string, StepValidation>;
  onSelect: (id: string) => void;
  onRemove: (id: string) => void;
}

export function ScenarioCanvas({
  steps,
  selectedId,
  validations,
  onSelect,
  onRemove,
}: ScenarioCanvasProps) {
  const { setNodeRef, isOver } = useDroppable({ id: CANVAS_DROPPABLE_ID });

  return (
    <div
      ref={setNodeRef}
      className={`${styles.canvas} ${isOver ? styles.canvasOver : ''}`}
      aria-label="Scenario timeline"
    >
      {steps.length === 0 ? (
        <div className={styles.canvasEmpty}>
          <FlowConnection size={32} className={styles.canvasEmptyIcon} aria-hidden />
          <p className={styles.canvasEmptyTitle}>Drag a fault here to start your scenario</p>
          <p>Each fault becomes an ordered step. Reorder by dragging, then set dependencies.</p>
        </div>
      ) : (
        <SortableContext items={steps.map((s) => s.id)} strategy={verticalListSortingStrategy}>
          <ol className={styles.stepList}>
            {steps.map((step, index) => {
              const v = validations.get(step.id);
              return (
                <StepCard
                  key={step.id}
                  step={step}
                  index={index}
                  accentIndex={index}
                  selected={step.id === selectedId}
                  valid={v ? isStepValid(v) : true}
                  onSelect={onSelect}
                  onRemove={onRemove}
                />
              );
            })}
          </ol>
        </SortableContext>
      )}
    </div>
  );
}
