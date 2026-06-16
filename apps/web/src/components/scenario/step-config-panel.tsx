'use client';

import { useEffect, useRef } from 'react';
import {
  TextInput,
  Select,
  SelectItem,
  FormGroup,
  FilterableMultiSelect,
} from '@carbon/react';
import { paramDefsFor, MODE_OPTIONS, type TargetMode } from '@/lib/fault-params';
import type { BuilderStep, StepValidation } from './model';
import { eligibleDependencies } from './model';
import styles from './scenario.module.scss';

interface StepConfigPanelProps {
  step: BuilderStep | null;
  allSteps: BuilderStep[];
  validation: StepValidation | null;
  onPatch: (id: string, patch: Partial<BuilderStep>) => void;
}

export function StepConfigPanel({ step, allSteps, validation, onPatch }: StepConfigPanelProps) {
  const nameRef = useRef<HTMLInputElement>(null);
  const selectedId = step?.id ?? null;

  useEffect(() => {
    // Move focus into the panel whenever a different step opens so keyboard
    // users land on the first field after selecting a card.
    if (selectedId) nameRef.current?.focus();
  }, [selectedId]);

  if (!step) {
    return (
      <p className={styles.configEmpty}>
        Select a step to configure its parameters, target, and dependencies.
      </p>
    );
  }

  const paramDefs = paramDefsFor(step.type);
  const deps = eligibleDependencies(step, allSteps);
  const selectedDeps = deps.filter((d) => step.dependsOn.includes(d.name));

  return (
    <div className={styles.configPanel}>
      <div className={styles.configSection}>
        <h4 className={styles.configSectionTitle}>Step</h4>
        <TextInput
          ref={nameRef}
          id={`step-name-${step.id}`}
          labelText="Step name"
          value={step.name}
          invalid={validation?.missingName || validation?.duplicateName}
          invalidText={validation?.duplicateName ? 'Step names must be unique.' : 'Name is required.'}
          onChange={(e) => onPatch(step.id, { name: e.target.value })}
        />
        <TextInput
          id={`step-duration-${step.id}`}
          labelText="Duration"
          helperText="e.g. 30s, 5m, 1h"
          value={step.duration}
          onChange={(e) => onPatch(step.id, { duration: e.target.value })}
        />
      </div>

      {paramDefs.length > 0 && (
        <div className={styles.configSection}>
          <h4 className={styles.configSectionTitle}>Parameters</h4>
          {paramDefs.map((def) => {
            const required = def.defaultValue === undefined;
            const missing = validation?.missingParams.includes(def.key) ?? false;
            return (
              <FormGroup key={def.key} legendText="">
                <TextInput
                  id={`param-${step.id}-${def.key}`}
                  labelText={`${def.label}${required ? ' *' : ''}`}
                  placeholder={def.placeholder}
                  value={step.params[def.key] ?? ''}
                  invalid={missing}
                  invalidText="This parameter is required."
                  onChange={(e) =>
                    onPatch(step.id, { params: { ...step.params, [def.key]: e.target.value } })
                  }
                />
              </FormGroup>
            );
          })}
        </div>
      )}

      <div className={styles.configSection}>
        <h4 className={styles.configSectionTitle}>Target</h4>
        <TextInput
          id={`target-ns-${step.id}`}
          labelText="Target namespace"
          value={step.targetNamespace}
          onChange={(e) => onPatch(step.id, { targetNamespace: e.target.value })}
        />
        <TextInput
          id={`target-labels-${step.id}`}
          labelText="Label selector"
          placeholder="app=nginx, tier=frontend"
          helperText="Comma-separated key=value pairs"
          value={step.labelSelector}
          onChange={(e) => onPatch(step.id, { labelSelector: e.target.value })}
        />
        <Select
          id={`target-mode-${step.id}`}
          labelText="Mode"
          value={step.mode}
          onChange={(e) => onPatch(step.id, { mode: e.target.value as TargetMode })}
        >
          {MODE_OPTIONS.map((m) => (
            <SelectItem key={m} value={m} text={m} />
          ))}
        </Select>
      </div>

      <div className={styles.configSection}>
        <h4 className={styles.configSectionTitle}>Dependencies</h4>
        {deps.length === 0 ? (
          <p className={styles.paletteEmpty}>
            No earlier steps available. Add more steps to define ordering.
          </p>
        ) : (
          <FilterableMultiSelect
            id={`step-deps-${step.id}`}
            titleText="Runs after"
            placeholder="Select prerequisite steps"
            items={deps}
            itemToString={(item) => (item ? item.name : '')}
            initialSelectedItems={selectedDeps}
            selectionFeedback="top-after-reopen"
            onChange={({ selectedItems }) =>
              onPatch(step.id, { dependsOn: (selectedItems ?? []).map((s) => s.name) })
            }
          />
        )}
      </div>
    </div>
  );
}
