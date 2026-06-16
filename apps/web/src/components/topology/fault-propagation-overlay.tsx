'use client';

import { useEffect, useMemo, useRef, type RefObject } from 'react';
import { AnimatePresence, motion, useReducedMotion, type TargetAndTransition } from 'framer-motion';
import type { VizTokens } from '@/components/viz/theme';
import type { VizSize } from '@/components/viz/viz-container';
import type { FaultModel, FaultNodeState } from '@/lib/topology/fault-model';

export interface NodePosition {
  x: number;
  y: number;
  r: number;
}

export type PositionsRef = RefObject<Map<string, NodePosition>>;
export type TransformRef = RefObject<{ x: number; y: number; k: number }>;

const PULSE_DURATION_S = 1.5;
const HOP_STAGGER_S = 0.28;

function frac(value: number): number {
  const f = value - Math.floor(value);
  return f < 0 ? f + 1 : f;
}

function haloColor(state: FaultNodeState, tokens: VizTokens): string {
  if (state === 'degraded') return tokens['support-warning'];
  if (state === 'healthy') return tokens['border-strong'];
  return tokens['support-error'];
}

function haloAnimation(
  state: FaultNodeState,
  recovered: boolean,
  reduced: boolean,
): TargetAndTransition {
  if (recovered) return { scale: 0.7, opacity: 0 };
  if (reduced) return { scale: 1, opacity: state === 'source' ? 0.5 : 0.32 };
  if (state === 'source') return { scale: [1, 1.7, 1], opacity: [0.55, 0, 0.55] };
  if (state === 'degraded') return { scale: [1, 1.3, 1], opacity: [0.35, 0.1, 0.35] };
  return { scale: [1, 1.18, 1], opacity: [0.4, 0.15, 0.4] };
}

export function FaultPropagationOverlay({
  model,
  positionsRef,
  transformRef,
  size,
  tokens,
}: {
  model: FaultModel | null;
  positionsRef: PositionsRef;
  transformRef: TransformRef;
  size: VizSize;
  tokens: VizTokens;
}) {
  const reduced = useReducedMotion() ?? false;
  const rootRef = useRef<SVGGElement>(null);
  const nodeEls = useRef<Map<string, SVGGElement>>(new Map());
  const pulseEls = useRef<{ el: SVGGElement; from: string; to: string; hop: number }[]>([]);
  const startRef = useRef<number>(performance.now());

  const haloNodes = useMemo(
    () => (model ? [...model.nodes.values()] : []),
    [model],
  );
  const active = model != null;

  // A single rAF loop keeps the overlay locked to the live D3 layout: it mirrors
  // the zoom transform, repositions each halo from the shared positions ref, and
  // advances the traveling pulses — no React state is touched per frame.
  useEffect(() => {
    if (!active) return;
    startRef.current = performance.now();
    let raf = 0;

    const tick = () => {
      const tf = transformRef.current ?? { x: 0, y: 0, k: 1 };
      rootRef.current?.setAttribute('transform', `translate(${tf.x},${tf.y}) scale(${tf.k})`);

      const positions = positionsRef.current;
      if (positions) {
        for (const [id, el] of nodeEls.current) {
          const p = positions.get(id);
          if (p) el.setAttribute('transform', `translate(${p.x},${p.y})`);
        }

        if (!reduced) {
          const elapsed = (performance.now() - startRef.current) / 1000;
          for (const pulse of pulseEls.current) {
            const from = positions.get(pulse.from);
            const to = positions.get(pulse.to);
            if (!from || !to) {
              pulse.el.setAttribute('opacity', '0');
              continue;
            }
            const t = frac(elapsed / PULSE_DURATION_S - pulse.hop * (HOP_STAGGER_S / PULSE_DURATION_S));
            const x = from.x + (to.x - from.x) * t;
            const y = from.y + (to.y - from.y) * t;
            pulse.el.setAttribute('transform', `translate(${x},${y})`);
            pulse.el.setAttribute('opacity', String(Math.sin(t * Math.PI)));
          }
        }
      }
      raf = requestAnimationFrame(tick);
    };

    raf = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(raf);
  }, [active, reduced, positionsRef, transformRef]);

  if (!model) return null;

  const danger = tokens['support-error'];

  return (
    <svg
      className="fault-overlay"
      width={size.width}
      height={size.height}
      style={{ position: 'absolute', inset: 0, pointerEvents: 'none' }}
      aria-hidden="true"
    >
      <g ref={rootRef}>
        {!reduced &&
          model.edges.map((edge, i) => (
            <g
              key={`${edge.from}->${edge.to}`}
              ref={(el) => {
                if (el) pulseEls.current[i] = { el, from: edge.from, to: edge.to, hop: edge.hop };
              }}
              opacity={0}
            >
              <circle r={7} fill={danger} opacity={0.25} />
              <circle r={3.5} fill={danger} />
            </g>
          ))}

        <AnimatePresence>
          {haloNodes.map((node) => {
            const pos = positionsRef.current?.get(node.id);
            const radius = (pos?.r ?? 12) + 6;
            const color = haloColor(node.state, tokens);
            const anim = haloAnimation(node.state, model.recovered, reduced);
            return (
              <g
                key={node.id}
                ref={(el) => {
                  if (el) nodeEls.current.set(node.id, el);
                  else nodeEls.current.delete(node.id);
                }}
              >
                <motion.circle
                  r={radius}
                  fill="none"
                  stroke={color}
                  strokeWidth={2.5}
                  initial={{ scale: 0.6, opacity: 0 }}
                  animate={anim}
                  exit={{ scale: 0.7, opacity: 0 }}
                  transition={
                    reduced || model.recovered
                      ? { duration: 0.3 }
                      : {
                          duration: node.state === 'source' ? 1.4 : 1.8,
                          repeat: Infinity,
                          ease: 'easeInOut',
                          delay: node.hop * 0.15,
                        }
                  }
                  style={{ transformOrigin: 'center' }}
                />
              </g>
            );
          })}
        </AnimatePresence>
      </g>
    </svg>
  );
}
