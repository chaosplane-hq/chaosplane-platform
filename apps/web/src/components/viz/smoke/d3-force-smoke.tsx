'use client';

import { useEffect, useRef } from 'react';
import {
  forceCenter,
  forceLink,
  forceManyBody,
  forceSimulation,
  select,
  type Simulation,
  type SimulationLinkDatum,
  type SimulationNodeDatum,
} from 'd3';
import { useVizTheme, categoricalColor } from '../theme';
import type { VizSize } from '../viz-container';

interface GraphNode extends SimulationNodeDatum {
  id: string;
}

const NODES: GraphNode[] = [
  { id: 'api' },
  { id: 'db' },
  { id: 'cache' },
];

const LINKS: SimulationLinkDatum<GraphNode>[] = [
  { source: 'api', target: 'db' },
  { source: 'api', target: 'cache' },
];

export function D3ForceSmoke({ size }: { size: VizSize }) {
  const svgRef = useRef<SVGSVGElement>(null);
  const tokens = useVizTheme();

  useEffect(() => {
    if (!svgRef.current || size.width === 0) return;

    const { width, height } = size;
    const svg = select(svgRef.current);
    svg.selectAll('*').remove();

    // d3-force mutates node/link objects in place, so clone to keep React state
    // and re-runs deterministic across re-renders.
    const nodes = NODES.map((n) => ({ ...n }));
    const links = LINKS.map((l) => ({ ...l }));

    const link = svg
      .append('g')
      .attr('stroke', tokens['border-strong'])
      .attr('stroke-width', 1.5)
      .selectAll('line')
      .data(links)
      .join('line');

    const node = svg
      .append('g')
      .selectAll('circle')
      .data(nodes)
      .join('circle')
      .attr('r', 14)
      .attr('fill', (_d, i) => categoricalColor(i))
      .attr('stroke', tokens.background)
      .attr('stroke-width', 2);

    const labels = svg
      .append('g')
      .selectAll('text')
      .data(nodes)
      .join('text')
      .text((d) => d.id)
      .attr('font-size', 11)
      .attr('fill', tokens['text-primary'])
      .attr('text-anchor', 'middle')
      .attr('dy', -20);

    const simulation: Simulation<GraphNode, SimulationLinkDatum<GraphNode>> =
      forceSimulation(nodes)
        .force('link', forceLink<GraphNode, SimulationLinkDatum<GraphNode>>(links).id((d) => d.id).distance(90))
        .force('charge', forceManyBody().strength(-220))
        .force('center', forceCenter(width / 2, height / 2));

    simulation.on('tick', () => {
      link
        .attr('x1', (d) => (d.source as GraphNode).x ?? 0)
        .attr('y1', (d) => (d.source as GraphNode).y ?? 0)
        .attr('x2', (d) => (d.target as GraphNode).x ?? 0)
        .attr('y2', (d) => (d.target as GraphNode).y ?? 0);
      node.attr('cx', (d) => d.x ?? 0).attr('cy', (d) => d.y ?? 0);
      labels.attr('x', (d) => d.x ?? 0).attr('y', (d) => d.y ?? 0);
    });

    return () => {
      simulation.stop();
    };
  }, [size, tokens]);

  return <svg ref={svgRef} width={size.width} height={size.height} role="presentation" />;
}
