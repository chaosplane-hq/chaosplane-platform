'use client';

import { useMemo, useState } from 'react';
import { useDraggable } from '@dnd-kit/core';
import { Search, Add, Draggable as DraggableIcon } from '@carbon/icons-react';
import { Search as SearchInput } from '@carbon/react';
import { categoricalColor } from '@/components/viz';
import type { ActionType, FaultCatalogGroup } from '@/lib/types';
import styles from './scenario.module.scss';

export const PALETTE_PREFIX = 'palette:';

export interface PaletteDragData {
  source: 'palette';
  type: ActionType;
  group: string;
}

interface PaletteItemProps {
  type: ActionType;
  group: string;
  accent: string;
  onAdd: (type: ActionType, group: string) => void;
}

function PaletteItem({ type, group, accent, onAdd }: PaletteItemProps) {
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id: `${PALETTE_PREFIX}${type}`,
    data: { source: 'palette', type, group } satisfies PaletteDragData,
  });

  return (
    <div
      ref={setNodeRef}
      className={`${styles.paletteItem} ${isDragging ? styles.paletteItemDragging : ''}`}
      style={{ ['--group-accent' as string]: accent }}
      {...attributes}
      {...listeners}
    >
      <DraggableIcon className={styles.paletteHandle} size={16} aria-hidden />
      <span className={styles.paletteItemLabel}>{type}</span>
      <button
        type="button"
        className={styles.paletteAdd}
        aria-label={`Add ${type} step`}
        // Stopping pointer-down keeps the click from being swallowed by the drag sensor.
        onPointerDown={(e) => e.stopPropagation()}
        onClick={() => onAdd(type, group)}
      >
        <Add size={16} />
      </button>
    </div>
  );
}

interface FaultPaletteProps {
  groups: FaultCatalogGroup[];
  onAdd: (type: ActionType, group: string) => void;
}

export function FaultPalette({ groups, onAdd }: FaultPaletteProps) {
  const [query, setQuery] = useState('');

  const groupAccents = useMemo(() => {
    const map = new Map<string, string>();
    groups.forEach((g, i) => map.set(g.group, categoricalColor(i)));
    return map;
  }, [groups]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return groups;
    return groups
      .map((g) => ({
        ...g,
        types: g.types.filter(
          (t) => t.type.toLowerCase().includes(q) || g.group.toLowerCase().includes(q),
        ),
      }))
      .filter((g) => g.types.length > 0);
  }, [groups, query]);

  return (
    <div className={styles.palette}>
      <SearchInput
        size="lg"
        labelText="Search fault types"
        placeholder="Search faults"
        value={query}
        renderIcon={Search}
        onChange={(e) => setQuery(e.target.value)}
      />
      <div className={styles.paletteScroll}>
        {filtered.length === 0 ? (
          <p className={styles.paletteEmpty}>No fault types match “{query}”.</p>
        ) : (
          filtered.map((g) => (
            <div key={g.group} className={styles.paletteGroup}>
              <p className={styles.paletteGroupLabel}>
                {g.group}
                <span className={styles.paletteGroupRule} aria-hidden />
              </p>
              {g.types.map((t) => (
                <PaletteItem
                  key={t.type}
                  type={t.type as ActionType}
                  group={g.group}
                  accent={groupAccents.get(g.group) ?? 'var(--cds-border-strong)'}
                  onAdd={onAdd}
                />
              ))}
            </div>
          ))
        )}
      </div>
    </div>
  );
}
