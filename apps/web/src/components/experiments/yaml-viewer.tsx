'use client';

import { CodeSnippet } from '@carbon/react';
import type { Experiment } from '@/lib/types';

interface YamlViewerProps {
  experiment: Experiment;
}

function toYaml(exp: Experiment): string {
  const labelSelector = exp.target.labelSelector
    ? Object.entries(exp.target.labelSelector)
        .map(([k, v]) => `        ${k}: ${v}`)
        .join('\n')
    : '        {}';

  const params = exp.action.parameters
    ? Object.entries(exp.action.parameters)
        .map(([k, v]) => `    ${k}: ${String(v)}`)
        .join('\n')
    : '';

  return `apiVersion: chaos.chaosplane.dev/v1alpha1
kind: Experiment
metadata:
  name: ${exp.name}
  namespace: ${exp.namespace}
spec:
  action:
    type: ${exp.action.type}
${params ? `    parameters:\n${params}` : ''}
  target:
    namespace: ${exp.target.namespace}
    mode: ${exp.target.mode ?? 'one'}
    labelSelector:
${labelSelector}
  duration: ${exp.duration ?? '30s'}
status:
  phase: ${exp.status.phase}
${exp.status.message ? `  message: ${exp.status.message}` : ''}`.trim();
}

export function YamlViewer({ experiment }: YamlViewerProps) {
  return (
    <CodeSnippet type="multi" feedback="Copied!" hideCopyButton={false}>
      {toYaml(experiment)}
    </CodeSnippet>
  );
}
