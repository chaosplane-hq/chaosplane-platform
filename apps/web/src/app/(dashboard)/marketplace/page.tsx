'use client';

import { useState } from 'react';
import {
  Grid,
  Column,
  Tabs,
  Tab,
  TabList,
  TabPanels,
  TabPanel,
  Tag,
  Button,
  SkeletonText,
  InlineNotification,
  Tile,
} from '@carbon/react';
import { Star, Download, Checkmark, Add, TrashCan } from '@carbon/icons-react';
import { useMarketplace, useInstallPlugin, useUninstallPlugin } from '@/lib/hooks/use-marketplace';
import type { PluginCategory, MarketplacePlugin } from '@/lib/types';
import styles from '@/components/experiments/experiments.module.scss';

const CATEGORIES: { id: PluginCategory | 'all'; label: string }[] = [
  { id: 'all', label: 'All' },
  { id: 'chaos_action', label: 'Chaos Actions' },
  { id: 'workflow_template', label: 'Workflow Templates' },
  { id: 'integration', label: 'Integrations' },
  { id: 'monitoring', label: 'Monitoring' },
];

export default function MarketplacePage() {
  const [activeCategory, setActiveCategory] = useState<PluginCategory | 'all'>('all');
  const { data, isLoading, isError, error } = useMarketplace(
    activeCategory !== 'all' ? { category: activeCategory } : undefined,
  );
  const installMutation = useInstallPlugin();
  const uninstallMutation = useUninstallPlugin();

  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div className={styles.pageHeader}>
          <h2 className={styles.pageTitle}>Marketplace</h2>
          <p className={styles.pageSubtitle}>Discover and install plugins, templates, and integrations.</p>
        </div>
      </Column>

      <Column lg={16} md={8} sm={4}>
        <Tabs onChange={({ selectedIndex }) => setActiveCategory(CATEGORIES[selectedIndex].id)}>
          <TabList aria-label="Plugin categories">
            {CATEGORIES.map((c) => <Tab key={c.id}>{c.label}</Tab>)}
          </TabList>
          <TabPanels>
            {CATEGORIES.map((c) => (
              <TabPanel key={c.id}>
                {isError && (
                  <InlineNotification
                    kind="error"
                    title="Failed to load plugins"
                    subtitle={(error as Error)?.message}
                    style={{ marginBottom: 'var(--cds-spacing-05)' }}
                  />
                )}
                {isLoading && (
                  <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))', gap: 'var(--cds-spacing-05)', marginTop: 'var(--cds-spacing-05)' }}>
                    {Array.from({ length: 6 }).map((_, i) => (
                      <Tile key={i}><SkeletonText paragraph lineCount={4} /></Tile>
                    ))}
                  </div>
                )}
                {!isLoading && data && (
                  <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))', gap: 'var(--cds-spacing-05)', marginTop: 'var(--cds-spacing-05)' }}>
                    {(data.items ?? []).map((plugin: MarketplacePlugin) => (
                      <Tile key={plugin.id} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--cds-spacing-03)' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                          <div>
                            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--cds-spacing-02)', marginBottom: 'var(--cds-spacing-02)' }}>
                              <strong>{plugin.name}</strong>
                              {plugin.verified && (
                                <Tag type="green" size="sm">
                                  <Checkmark size={12} /> Verified
                                </Tag>
                              )}
                            </div>
                            <p style={{ fontSize: 'var(--cds-body-short-01-font-size)', color: 'var(--cds-text-secondary)', margin: 0 }}>
                              by {plugin.author} · v{plugin.version}
                            </p>
                          </div>
                          <Tag type="blue" size="sm">{plugin.category.replace('_', ' ')}</Tag>
                        </div>

                        <p style={{ fontSize: 'var(--cds-body-short-01-font-size)', color: 'var(--cds-text-primary)', margin: 0, flexGrow: 1 }}>
                          {plugin.description}
                        </p>

                        <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--cds-spacing-05)', fontSize: 'var(--cds-label-01-font-size)', color: 'var(--cds-text-secondary)' }}>
                          <span style={{ display: 'flex', alignItems: 'center', gap: 'var(--cds-spacing-02)' }}>
                            <Star size={14} /> {plugin.rating.toFixed(1)}
                          </span>
                          <span style={{ display: 'flex', alignItems: 'center', gap: 'var(--cds-spacing-02)' }}>
                            <Download size={14} /> {plugin.downloads.toLocaleString()}
                          </span>
                        </div>

                        <div style={{ marginTop: 'var(--cds-spacing-03)' }}>
                          {plugin.installed ? (
                            <Button
                              kind="danger--ghost"
                              size="sm"
                              renderIcon={TrashCan}
                              disabled={uninstallMutation.isPending}
                              onClick={() => uninstallMutation.mutate(plugin.id)}
                            >
                              Uninstall
                            </Button>
                          ) : (
                            <Button
                              kind="primary"
                              size="sm"
                              renderIcon={Add}
                              disabled={installMutation.isPending}
                              onClick={() => installMutation.mutate(plugin.id)}
                            >
                              Install
                            </Button>
                          )}
                        </div>
                      </Tile>
                    ))}
                  </div>
                )}
                {!isLoading && data?.items.length === 0 && (
                  <p style={{ color: 'var(--cds-text-secondary)', marginTop: 'var(--cds-spacing-07)' }}>
                    No plugins found in this category.
                  </p>
                )}
              </TabPanel>
            ))}
          </TabPanels>
        </Tabs>
      </Column>
    </Grid>
  );
}
