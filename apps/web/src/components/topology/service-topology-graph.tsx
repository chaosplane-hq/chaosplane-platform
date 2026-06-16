'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import {
  drag as d3drag,
  forceCenter,
  forceCollide,
  forceLink,
  forceManyBody,
  forceSimulation,
  select,
  zoom as d3zoom,
  type D3DragEvent,
  type Selection,
  type Simulation,
  type SimulationLinkDatum,
  type SimulationNodeDatum,
} from 'd3';
import { useVizTheme, categoricalColor } from '@/components/viz/theme';
import type { VizSize } from '@/components/viz/viz-container';
import type { ServiceDependency } from '@/lib/types';
import styles from './service-topology-graph.module.scss';

interface GraphNode extends SimulationNodeDatum {
  id: string;
  name: string;
  namespace: string;
  kind: string;
  degree: number;
}

interface GraphLink extends SimulationLinkDatum<GraphNode> {
  source: string | GraphNode;
  target: string | GraphNode;
  protocol?: string;
  port?: number;
}

interface TooltipState {
  x: number;
  y: number;
  title: string;
  rows: { label: string; value: string }[];
}

function nodeId(kind: string, namespace: string, name: string): string {
  return `${kind}/${namespace}/${name}`;
}

// Node radius scales with degree (dependency count) so hubs read as bigger.
function nodeRadius(degree: number): number {
  return Math.min(26, 9 + Math.sqrt(degree) * 3.5);
}

function asNode(end: string | GraphNode): GraphNode {
  return end as GraphNode;
}

export function ServiceTopologyGraph({
  size,
  dependencies,
}: {
  size: VizSize;
  dependencies: ServiceDependency[];
}) {
  const svgRef = useRef<SVGSVGElement>(null);
  const tokens = useVizTheme();
  const [tooltip, setTooltip] = useState<TooltipState | null>(null);
  const [focusedId, setFocusedId] = useState<string | null>(null);

  // Selections are stored in refs so the focus-highlight effect can restyle
  // existing DOM without tearing down and rebuilding the whole simulation.
  const nodeSelRef = useRef<Selection<SVGCircleElement, GraphNode, SVGGElement, unknown> | null>(null);
  const linkSelRef = useRef<Selection<SVGLineElement, GraphLink, SVGGElement, unknown> | null>(null);
  const labelSelRef = useRef<Selection<SVGTextElement, GraphNode, SVGGElement, unknown> | null>(null);
  const adjacencyRef = useRef<Map<string, Set<string>>>(new Map());

  const { nodes, links, namespaces } = useMemo(() => {
    const nodeMap = new Map<string, GraphNode>();
    const ensure = (kind: string, name: string, namespace: string): GraphNode => {
      const id = nodeId(kind, namespace, name);
      let n = nodeMap.get(id);
      if (!n) {
        n = { id, name, namespace, kind, degree: 0 };
        nodeMap.set(id, n);
      }
      return n;
    };

    const builtLinks: GraphLink[] = dependencies.map((d) => {
      const s = ensure(d.sourceKind, d.sourceName, d.sourceNamespace);
      const t = ensure(d.targetKind, d.targetName, d.targetNamespace);
      s.degree += 1;
      t.degree += 1;
      return { source: s.id, target: t.id, protocol: d.protocol, port: d.port };
    });

    const builtNodes = [...nodeMap.values()];
    const ns = [...new Set(builtNodes.map((n) => n.namespace))].sort();
    return { nodes: builtNodes, links: builtLinks, namespaces: ns };
  }, [dependencies]);

  const namespaceColor = useMemo(() => {
    const map = new Map<string, string>();
    namespaces.forEach((ns, i) => map.set(ns, categoricalColor(i)));
    return map;
  }, [namespaces]);

  useEffect(() => {
    if (!svgRef.current || size.width === 0 || nodes.length === 0) return;

    const { width, height } = size;
    const svg = select(svgRef.current);
    svg.selectAll('*').remove();

    // d3-force mutates node/link objects in place; clone so re-runs stay
    // deterministic and React's memoized arrays are never written to.
    const simNodes: GraphNode[] = nodes.map((n) => ({ ...n }));
    const simLinks: GraphLink[] = links.map((l) => ({ ...l }));

    const adjacency = new Map<string, Set<string>>();
    for (const n of simNodes) adjacency.set(n.id, new Set());
    for (const l of simLinks) {
      const s = typeof l.source === 'string' ? l.source : l.source.id;
      const t = typeof l.target === 'string' ? l.target : l.target.id;
      adjacency.get(s)?.add(t);
      adjacency.get(t)?.add(s);
    }
    adjacencyRef.current = adjacency;

    const defs = svg.append('defs');
    defs
      .append('marker')
      .attr('id', 'topology-arrow')
      .attr('viewBox', '0 -5 10 10')
      .attr('refX', 9)
      .attr('refY', 0)
      .attr('markerWidth', 6)
      .attr('markerHeight', 6)
      .attr('orient', 'auto')
      .append('path')
      .attr('d', 'M0,-5L10,0L0,5')
      .attr('fill', tokens['border-strong']);

    const zoomG = svg.append('g');

    const linkSel = zoomG
      .append('g')
      .attr('stroke', tokens['border-strong'])
      .attr('stroke-opacity', 0.7)
      .attr('stroke-width', 1.5)
      .selectAll<SVGLineElement, GraphLink>('line')
      .data(simLinks)
      .join('line')
      .attr('marker-end', 'url(#topology-arrow)')
      .style('cursor', 'pointer')
      .on('mouseover', function (event: MouseEvent, d) {
        const s = asNode(d.source);
        const t = asNode(d.target);
        setTooltip({
          x: event.offsetX,
          y: event.offsetY,
          title: `${s.name} → ${t.name}`,
          rows: [
            { label: 'Protocol', value: d.protocol ?? '—' },
            { label: 'Port', value: d.port != null ? String(d.port) : '—' },
          ],
        });
      })
      .on('mousemove', (event: MouseEvent) =>
        setTooltip((prev) => (prev ? { ...prev, x: event.offsetX, y: event.offsetY } : prev)),
      )
      .on('mouseout', () => setTooltip(null));

    const nodeSel = zoomG
      .append('g')
      .attr('stroke', tokens.background)
      .attr('stroke-width', 2)
      .selectAll<SVGCircleElement, GraphNode>('circle')
      .data(simNodes)
      .join('circle')
      .attr('r', (d) => nodeRadius(d.degree))
      .attr('fill', (d) => namespaceColor.get(d.namespace) ?? tokens.interactive)
      .style('cursor', 'pointer')
      .on('mouseover', function (event: MouseEvent, d) {
        setTooltip({
          x: event.offsetX,
          y: event.offsetY,
          title: d.name,
          rows: [
            { label: 'Namespace', value: d.namespace },
            { label: 'Kind', value: d.kind },
            { label: 'Dependencies', value: String(d.degree) },
          ],
        });
      })
      .on('mousemove', (event: MouseEvent) =>
        setTooltip((prev) => (prev ? { ...prev, x: event.offsetX, y: event.offsetY } : prev)),
      )
      .on('mouseout', () => setTooltip(null))
      .on('click', (event: MouseEvent, d) => {
        event.stopPropagation();
        setFocusedId((prev) => (prev === d.id ? null : d.id));
      });

    const labelSel = zoomG
      .append('g')
      .selectAll<SVGTextElement, GraphNode>('text')
      .data(simNodes)
      .join('text')
      .text((d) => d.name)
      .attr('font-size', 11)
      .attr('font-weight', 500)
      .attr('fill', tokens['text-primary'])
      .attr('text-anchor', 'middle')
      .attr('pointer-events', 'none')
      .attr('paint-order', 'stroke')
      .attr('stroke', tokens.background)
      .attr('stroke-width', 3)
      .attr('stroke-linejoin', 'round');

    nodeSelRef.current = nodeSel;
    linkSelRef.current = linkSel;
    labelSelRef.current = labelSel;

    const simulation: Simulation<GraphNode, GraphLink> = forceSimulation(simNodes)
      .force(
        'link',
        forceLink<GraphNode, GraphLink>(simLinks)
          .id((d) => d.id)
          .distance(110)
          .strength(0.6),
      )
      .force('charge', forceManyBody<GraphNode>().strength(-380))
      .force('center', forceCenter(width / 2, height / 2))
      .force('collide', forceCollide<GraphNode>((d) => nodeRadius(d.degree) + 14));

    const dragBehavior = d3drag<SVGCircleElement, GraphNode>()
      .on('start', (event: D3DragEvent<SVGCircleElement, GraphNode, GraphNode>, d) => {
        if (!event.active) simulation.alphaTarget(0.3).restart();
        d.fx = d.x;
        d.fy = d.y;
      })
      .on('drag', (event: D3DragEvent<SVGCircleElement, GraphNode, GraphNode>, d) => {
        d.fx = event.x;
        d.fy = event.y;
      })
      .on('end', (event: D3DragEvent<SVGCircleElement, GraphNode, GraphNode>, d) => {
        if (!event.active) simulation.alphaTarget(0);
        d.fx = null;
        d.fy = null;
      });
    nodeSel.call(dragBehavior);

    const zoomBehavior = d3zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.25, 4])
      .on('zoom', (event) => zoomG.attr('transform', event.transform.toString()));
    svg.call(zoomBehavior);
    svg.on('click', () => setFocusedId(null));

    // Tick updates the DOM directly (no React state per frame) to keep the
    // simulation smooth; links are shortened to the target radius so the
    // arrowhead sits at the node edge instead of under the circle.
    simulation.on('tick', () => {
      linkSel
        .attr('x1', (d) => asNode(d.source).x ?? 0)
        .attr('y1', (d) => asNode(d.source).y ?? 0)
        .attr('x2', (d) => {
          const s = asNode(d.source);
          const t = asNode(d.target);
          const dx = (t.x ?? 0) - (s.x ?? 0);
          const dy = (t.y ?? 0) - (s.y ?? 0);
          const dist = Math.hypot(dx, dy) || 1;
          return (t.x ?? 0) - (dx / dist) * (nodeRadius(t.degree) + 4);
        })
        .attr('y2', (d) => {
          const s = asNode(d.source);
          const t = asNode(d.target);
          const dx = (t.x ?? 0) - (s.x ?? 0);
          const dy = (t.y ?? 0) - (s.y ?? 0);
          const dist = Math.hypot(dx, dy) || 1;
          return (t.y ?? 0) - (dy / dist) * (nodeRadius(t.degree) + 4);
        });
      nodeSel.attr('cx', (d) => d.x ?? 0).attr('cy', (d) => d.y ?? 0);
      labelSel.attr('x', (d) => d.x ?? 0).attr('y', (d) => (d.y ?? 0) - nodeRadius(d.degree) - 6);
    });

    return () => {
      simulation.stop();
    };
  }, [nodes, links, size, tokens, namespaceColor]);

  // Highlight the focused node, its neighbors, and connecting edges; dim the
  // rest. Runs on click (infrequent), so restyling via d3 selections is cheap.
  useEffect(() => {
    const nodeSel = nodeSelRef.current;
    const linkSel = linkSelRef.current;
    const labelSel = labelSelRef.current;
    if (!nodeSel || !linkSel || !labelSel) return;

    if (!focusedId) {
      nodeSel.attr('opacity', 1).attr('stroke', tokens.background).attr('stroke-width', 2);
      linkSel.attr('stroke-opacity', 0.7).attr('stroke', tokens['border-strong']);
      labelSel.attr('opacity', 1);
      return;
    }

    const neighbors = adjacencyRef.current.get(focusedId) ?? new Set<string>();
    const isActive = (id: string) => id === focusedId || neighbors.has(id);

    nodeSel
      .attr('opacity', (d) => (isActive(d.id) ? 1 : 0.15))
      .attr('stroke', (d) => (d.id === focusedId ? tokens['link-primary'] : tokens.background))
      .attr('stroke-width', (d) => (d.id === focusedId ? 3.5 : 2));
    labelSel.attr('opacity', (d) => (isActive(d.id) ? 1 : 0.15));
    linkSel
      .attr('stroke', (d) => {
        const s = asNode(d.source).id;
        const t = asNode(d.target).id;
        return s === focusedId || t === focusedId ? tokens['link-primary'] : tokens['border-strong'];
      })
      .attr('stroke-opacity', (d) => {
        const s = asNode(d.source).id;
        const t = asNode(d.target).id;
        return s === focusedId || t === focusedId ? 0.95 : 0.08;
      });
  }, [focusedId, tokens]);

  return (
    <div className={styles.wrap}>
      <svg ref={svgRef} width={size.width} height={size.height} role="presentation" />

      {namespaces.length > 0 && (
        <ul className={styles.legend} aria-label="Namespace legend">
          {namespaces.map((ns) => (
            <li key={ns} className={styles.legendItem}>
              <span
                className={styles.legendSwatch}
                style={{ background: namespaceColor.get(ns) }}
                aria-hidden="true"
              />
              <span>{ns}</span>
            </li>
          ))}
        </ul>
      )}

      <div className={styles.hint}>Drag to reposition · scroll to zoom · click a node to focus</div>

      {tooltip && (
        <div
          className={styles.tooltip}
          style={{ left: tooltip.x + 12, top: tooltip.y + 12 }}
          role="presentation"
        >
          <div className={styles.tooltipTitle}>{tooltip.title}</div>
          {tooltip.rows.map((r) => (
            <div key={r.label} className={styles.tooltipRow}>
              <span className={styles.tooltipLabel}>{r.label}</span>
              <span>{r.value}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
