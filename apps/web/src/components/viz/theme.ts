'use client';

import { useEffect, useState } from 'react';

// Viz libs (D3, Recharts) need concrete color strings, not `var(--cds-*)`
// references, so we resolve Carbon's computed token values at runtime and
// re-resolve on theme change (light/dark) instead of hardcoding colors.
const TOKEN_NAMES = [
  'background',
  'layer-01',
  'layer-02',
  'border-subtle',
  'border-strong',
  'text-primary',
  'text-secondary',
  'text-on-color',
  'link-primary',
  'support-success',
  'support-error',
  'support-warning',
  'support-info',
  'interactive',
] as const;

export type VizTokenName = (typeof TOKEN_NAMES)[number];
export type VizTokens = Record<VizTokenName, string>;

// Carbon does not emit its categorical data-vis palette as `--cds-*` properties
// in this build config, so we ship the canonical 14-color IBM sequence.
export const CATEGORICAL_PALETTE = [
  '#8a3ffc',
  '#33b1ff',
  '#007d79',
  '#ff7eb6',
  '#fa4d56',
  '#fff1f1',
  '#6fdc8c',
  '#4589ff',
  '#d12771',
  '#d2a106',
  '#08bdba',
  '#bae6ff',
  '#ba4e00',
  '#d4bbff',
] as const;

// g100 (dark) values, used during SSR and before the first client read.
const FALLBACK_TOKENS: VizTokens = {
  background: '#161616',
  'layer-01': '#262626',
  'layer-02': '#393939',
  'border-subtle': '#393939',
  'border-strong': '#6f6f6f',
  'text-primary': '#f4f4f4',
  'text-secondary': '#c6c6c6',
  'text-on-color': '#ffffff',
  'link-primary': '#78a9ff',
  'support-success': '#42be65',
  'support-error': '#fa4d56',
  'support-warning': '#f1c21b',
  'support-info': '#4589ff',
  interactive: '#4589ff',
};

function readTokens(): VizTokens {
  if (typeof window === 'undefined') return FALLBACK_TOKENS;
  const style = getComputedStyle(document.documentElement);
  const resolved = {} as VizTokens;
  for (const name of TOKEN_NAMES) {
    const value = style.getPropertyValue(`--cds-${name}`).trim();
    resolved[name] = value || FALLBACK_TOKENS[name];
  }
  return resolved;
}

export function useVizTheme(): VizTokens {
  const [tokens, setTokens] = useState<VizTokens>(FALLBACK_TOKENS);

  useEffect(() => {
    setTokens(readTokens());

    // Carbon themes are applied via class / data attributes on the root; re-read
    // when those change so viz colors track light/dark without a remount.
    const observer = new MutationObserver(() => setTokens(readTokens()));
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class', 'data-carbon-theme', 'style'],
    });
    return () => observer.disconnect();
  }, []);

  return tokens;
}

export function categoricalColor(index: number): string {
  return CATEGORICAL_PALETTE[index % CATEGORICAL_PALETTE.length];
}
