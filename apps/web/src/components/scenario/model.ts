import type { ActionType, ScenarioStep, CreateScenarioRequest } from '@/lib/types';
import { defaultParamsFor, requiredParamKeys, type TargetMode } from '@/lib/fault-params';

// A step as held in builder state. `id` is a stable client-only handle for
// dnd-kit; `name` is the user-facing identifier sent to the API and referenced
// by other steps' dependsOn, so the two are kept separate.
export interface BuilderStep {
  id: string;
  name: string;
  type: ActionType;
  group: string;
  params: Record<string, string>;
  targetNamespace: string;
  labelSelector: string;
  mode: TargetMode;
  duration: string;
  dependsOn: string[];
}

let stepSeq = 0;

function uniqueName(base: string, taken: Set<string>): string {
  if (!taken.has(base)) return base;
  let i = 2;
  while (taken.has(`${base}-${i}`)) i += 1;
  return `${base}-${i}`;
}

export function createStep(
  type: ActionType,
  group: string,
  existing: BuilderStep[],
): BuilderStep {
  stepSeq += 1;
  const taken = new Set(existing.map((s) => s.name));
  return {
    id: `step-${Date.now()}-${stepSeq}`,
    name: uniqueName(type, taken),
    type,
    group,
    params: defaultParamsFor(type),
    targetNamespace: 'default',
    labelSelector: '',
    mode: 'one',
    duration: '30s',
    dependsOn: [],
  };
}

export function parseLabelSelector(raw: string): Record<string, string> {
  return raw
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean)
    .reduce<Record<string, string>>((acc, kv) => {
      const [k, v] = kv.split('=');
      if (k) acc[k.trim()] = v?.trim() ?? '';
      return acc;
    }, {});
}

export interface StepValidation {
  missingParams: string[];
  missingName: boolean;
  duplicateName: boolean;
}

export function validateStep(step: BuilderStep, all: BuilderStep[]): StepValidation {
  const missingParams = requiredParamKeys(step.type).filter(
    (key) => !step.params[key] || step.params[key].trim() === '',
  );
  const duplicateName =
    step.name.trim() !== '' &&
    all.some((s) => s.id !== step.id && s.name.trim() === step.name.trim());
  return {
    missingParams,
    missingName: step.name.trim() === '',
    duplicateName,
  };
}

export function isStepValid(v: StepValidation): boolean {
  return v.missingParams.length === 0 && !v.missingName && !v.duplicateName;
}

// Walks the dependsOn graph from `targetId` to find every step that (directly or
// transitively) depends on it. Adding any of these as a dependency of targetId
// would close a cycle, so the dependency picker excludes them up front.
export function descendantsOf(targetId: string, steps: BuilderStep[]): Set<string> {
  const byName = new Map(steps.map((s) => [s.name, s]));
  const target = steps.find((s) => s.id === targetId);
  if (!target) return new Set();

  const dependents = new Map<string, string[]>();
  for (const s of steps) {
    for (const dep of s.dependsOn) {
      const parent = byName.get(dep);
      if (parent) {
        const list = dependents.get(parent.id) ?? [];
        list.push(s.id);
        dependents.set(parent.id, list);
      }
    }
  }

  const seen = new Set<string>();
  const stack = [targetId];
  while (stack.length) {
    const current = stack.pop() as string;
    for (const child of dependents.get(current) ?? []) {
      if (!seen.has(child)) {
        seen.add(child);
        stack.push(child);
      }
    }
  }
  return seen;
}

export function eligibleDependencies(step: BuilderStep, steps: BuilderStep[]): BuilderStep[] {
  const blocked = descendantsOf(step.id, steps);
  return steps.filter((s) => s.id !== step.id && !blocked.has(s.id) && s.name.trim() !== '');
}

export function toScenarioRequest(name: string, steps: BuilderStep[]): CreateScenarioRequest {
  const apiSteps: ScenarioStep[] = steps.map((s) => {
    const labels = parseLabelSelector(s.labelSelector);
    const params = Object.fromEntries(
      Object.entries(s.params).filter(([, v]) => v !== ''),
    );
    return {
      name: s.name.trim(),
      dependsOn: s.dependsOn.length ? s.dependsOn : undefined,
      action: {
        type: s.type,
        parameters: Object.keys(params).length ? params : undefined,
      },
      target: {
        namespace: s.targetNamespace,
        labelSelector: Object.keys(labels).length ? labels : undefined,
        mode: s.mode,
      },
      duration: s.duration || undefined,
    };
  });
  return { name: name.trim(), steps: apiSteps };
}
