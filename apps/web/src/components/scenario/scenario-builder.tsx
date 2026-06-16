'use client';

import { useMemo, useState } from 'react';
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  KeyboardSensor,
  useSensor,
  useSensors,
  closestCenter,
  type DragStartEvent,
  type DragEndEvent,
} from '@dnd-kit/core';
import {
  arrayMove,
  sortableKeyboardCoordinates,
} from '@dnd-kit/sortable';
import { Grid, Column, Tile, TextInput, Button, InlineNotification } from '@carbon/react';
import { ArrowLeft, Send, Draggable as DraggableIcon } from '@carbon/icons-react';
import { useRouter } from 'next/navigation';
import { useCreateScenario, useFaultCatalog } from '@/lib/hooks/use-experiments';
import type { ActionType } from '@/lib/types';
import { FaultPalette, PALETTE_PREFIX, type PaletteDragData } from './fault-palette';
import { ScenarioCanvas, CANVAS_DROPPABLE_ID } from './scenario-canvas';
import { StepConfigPanel } from './step-config-panel';
import {
  type BuilderStep,
  type StepValidation,
  createStep,
  validateStep,
  isStepValid,
  toScenarioRequest,
} from './model';
import styles from './scenario.module.scss';

interface ActiveDrag {
  kind: 'palette' | 'step';
  label: string;
}

export function ScenarioBuilder() {
  const router = useRouter();
  const { data: catalog } = useFaultCatalog();
  const createScenario = useCreateScenario();

  const [name, setName] = useState('');
  const [steps, setSteps] = useState<BuilderStep[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [activeDrag, setActiveDrag] = useState<ActiveDrag | null>(null);

  const sensors = useSensors(
    // 6px activation distance lets card clicks (select) and button taps through
    // while still starting a drag once the pointer actually moves.
    useSensor(PointerSensor, { activationConstraint: { distance: 6 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );

  const validations = useMemo(() => {
    const map = new Map<string, StepValidation>();
    for (const s of steps) map.set(s.id, validateStep(s, steps));
    return map;
  }, [steps]);

  const selectedStep = steps.find((s) => s.id === selectedId) ?? null;
  const allValid = steps.length > 0 && steps.every((s) => isStepValid(validations.get(s.id)!));
  const nameValid = name.trim() !== '';

  const invalidStepNames = steps
    .filter((s) => !isStepValid(validations.get(s.id)!))
    .map((s) => s.name || 'unnamed step');

  function addStep(type: ActionType, group: string) {
    setSteps((prev) => {
      const next = createStep(type, group, prev);
      setSelectedId(next.id);
      return [...prev, next];
    });
  }

  function patchStep(id: string, patch: Partial<BuilderStep>) {
    setSteps((prev) => prev.map((s) => (s.id === id ? { ...s, ...patch } : s)));
  }

  function removeStep(id: string) {
    setSteps((prev) => {
      const goneName = prev.find((s) => s.id === id)?.name ?? '';
      return prev
        .filter((s) => s.id !== id)
        .map((s) => ({ ...s, dependsOn: s.dependsOn.filter((d) => d !== goneName) }));
    });
    setSelectedId((cur) => (cur === id ? null : cur));
  }

  function handleDragStart(event: DragStartEvent) {
    const data = event.active.data.current as PaletteDragData | undefined;
    if (data?.source === 'palette') {
      setActiveDrag({ kind: 'palette', label: data.type });
    } else {
      const step = steps.find((s) => s.id === event.active.id);
      setActiveDrag({ kind: 'step', label: step?.name ?? '' });
    }
  }

  function handleDragEnd(event: DragEndEvent) {
    setActiveDrag(null);
    const { active, over } = event;
    if (!over) return;

    const activeId = String(active.id);

    if (activeId.startsWith(PALETTE_PREFIX)) {
      const data = active.data.current as PaletteDragData;
      addStep(data.type, data.group);
      return;
    }

    // Reorder: only when dropped onto another step (not the empty canvas zone).
    if (activeId !== String(over.id) && over.id !== CANVAS_DROPPABLE_ID) {
      setSteps((prev) => {
        const oldIndex = prev.findIndex((s) => s.id === activeId);
        const newIndex = prev.findIndex((s) => s.id === String(over.id));
        if (oldIndex === -1 || newIndex === -1) return prev;
        return arrayMove(prev, oldIndex, newIndex);
      });
    }
  }

  function handleSubmit() {
    if (!nameValid || !allValid) return;
    createScenario.mutate(toScenarioRequest(name, steps), {
      onSuccess: (exp) => router.push(`/experiments/${exp.id}`),
    });
  }

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
      onDragCancel={() => setActiveDrag(null)}
    >
      <Grid fullWidth>
        <Column lg={16} md={8} sm={4}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--cds-spacing-04)', marginBottom: 'var(--cds-spacing-05)' }}>
            <Button kind="ghost" size="sm" renderIcon={ArrowLeft} onClick={() => router.push('/experiments')}>
              Back
            </Button>
          </div>
          <div className={styles.pageHeader}>
            <h2 className={styles.pageTitle}>Scenario Builder</h2>
            <p className={styles.pageSubtitle}>
              Drag faults onto the timeline to compose a multi-step chaos experiment, then define
              ordering and dependencies.
            </p>
          </div>
        </Column>

        <Column lg={16} md={8} sm={4}>
          <Tile style={{ marginBottom: 'var(--cds-spacing-05)' }}>
            <div className={styles.summaryRow}>
              <div className={styles.summaryStat}>
                <span className={styles.summaryValue}>{steps.length}</span>
                <span className={styles.summaryLabel}>Steps</span>
              </div>
              <TextInput
                id="scenario-name"
                labelText="Scenario name"
                placeholder="checkout-resilience-drill"
                value={name}
                invalid={!nameValid && steps.length > 0}
                invalidText="Scenario name is required."
                onChange={(e) => setName(e.target.value)}
                style={{ flex: 1, minWidth: 240 }}
              />
              <Button
                renderIcon={Send}
                disabled={!nameValid || !allValid || createScenario.isPending}
                onClick={handleSubmit}
              >
                {createScenario.isPending ? 'Creating…' : 'Create scenario'}
              </Button>
            </div>

            {createScenario.isError && (
              <InlineNotification
                kind="error"
                title="Failed to create scenario"
                subtitle={createScenario.error?.message}
                lowContrast
              />
            )}
            {steps.length > 0 && invalidStepNames.length > 0 && (
              <InlineNotification
                kind="warning"
                title="Some steps need attention"
                subtitle={`Complete required fields for: ${invalidStepNames.join(', ')}`}
                lowContrast
                hideCloseButton
              />
            )}
          </Tile>
        </Column>

        <Column lg={16} md={8} sm={4}>
          <div className={styles.layout}>
            <FaultPalette groups={catalog ?? []} onAdd={addStep} />

            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--cds-spacing-05)' }}>
              <ScenarioCanvas
                steps={steps}
                selectedId={selectedId}
                validations={validations}
                onSelect={setSelectedId}
                onRemove={removeStep}
              />
              <Tile>
                <StepConfigPanel
                  step={selectedStep}
                  allSteps={steps}
                  validation={selectedStep ? validations.get(selectedStep.id) ?? null : null}
                  onPatch={patchStep}
                />
              </Tile>
            </div>
          </div>
        </Column>
      </Grid>

      <DragOverlay>
        {activeDrag ? (
          <div className={styles.dragOverlayCard}>
            <DraggableIcon size={16} aria-hidden />
            {activeDrag.label}
          </div>
        ) : null}
      </DragOverlay>
    </DndContext>
  );
}
