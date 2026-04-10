'use client';

import { useState } from 'react';
import {
  Tile,
  Button,
  TextInput,
  Select,
  SelectItem,
  Tag,
  InlineNotification,
  SkeletonText,
  Modal,
  Checkbox,
} from '@carbon/react';
import { Add, TrashCan } from '@carbon/icons-react';
import {
  useNotificationChannels,
  useCreateNotificationChannel,
  useDeleteNotificationChannel,
  useNotificationRules,
  useCreateNotificationRule,
  useDeleteNotificationRule,
} from '@/lib/hooks/use-notifications';
import type { NotificationChannelType, NotificationEvent } from '@/lib/types';

const CHANNEL_TYPES: NotificationChannelType[] = ['slack', 'webhook', 'email', 'pagerduty'];

const ALL_EVENTS: NotificationEvent[] = [
  'experiment.started',
  'experiment.completed',
  'experiment.failed',
  'experiment.aborted',
];

const CHANNEL_TYPE_COLORS: Record<NotificationChannelType, 'blue' | 'green' | 'purple' | 'red'> = {
  slack: 'green',
  webhook: 'blue',
  email: 'purple',
  pagerduty: 'red',
};

export function NotificationsTab() {
  const [channelOpen, setChannelOpen] = useState(false);
  const [ruleOpen, setRuleOpen] = useState(false);
  const [channelName, setChannelName] = useState('');
  const [channelType, setChannelType] = useState<NotificationChannelType>('slack');
  const [channelUrl, setChannelUrl] = useState('');
  const [ruleChannelId, setRuleChannelId] = useState('');
  const [ruleEvents, setRuleEvents] = useState<NotificationEvent[]>([]);
  const [ruleNamespace, setRuleNamespace] = useState('');
  const [errorMsg, setErrorMsg] = useState('');

  const { data: channelsData, isLoading: channelsLoading } = useNotificationChannels();
  const { data: rulesData, isLoading: rulesLoading } = useNotificationRules();
  const createChannel = useCreateNotificationChannel();
  const deleteChannel = useDeleteNotificationChannel();
  const createRule = useCreateNotificationRule();
  const deleteRule = useDeleteNotificationRule();

  const channels = channelsData?.channels ?? [];
  const rules = rulesData?.rules ?? [];

  function configForType(type: NotificationChannelType, url: string): Record<string, string> {
    if (type === 'slack') return { webhookUrl: url };
    if (type === 'webhook') return { url };
    if (type === 'email') return { address: url };
    return { integrationKey: url };
  }

  async function handleCreateChannel() {
    if (!channelName.trim() || !channelUrl.trim()) return;
    setErrorMsg('');
    try {
      await createChannel.mutateAsync({
        name: channelName.trim(),
        type: channelType,
        config: configForType(channelType, channelUrl.trim()),
      });
      setChannelOpen(false);
      setChannelName('');
      setChannelUrl('');
    } catch {
      setErrorMsg('Failed to create channel.');
    }
  }

  async function handleCreateRule() {
    if (!ruleChannelId || ruleEvents.length === 0) return;
    setErrorMsg('');
    try {
      await createRule.mutateAsync({
        channelId: ruleChannelId,
        events: ruleEvents,
        namespaceFilter: ruleNamespace.trim() || undefined,
      });
      setRuleOpen(false);
      setRuleChannelId('');
      setRuleEvents([]);
      setRuleNamespace('');
    } catch {
      setErrorMsg('Failed to create rule.');
    }
  }

  function toggleEvent(event: NotificationEvent) {
    setRuleEvents((prev) =>
      prev.includes(event) ? prev.filter((e) => e !== event) : [...prev, event],
    );
  }

  function channelLabel(id: string) {
    return channels.find((c) => c.id === id)?.name ?? id;
  }

  const urlLabel: Record<NotificationChannelType, string> = {
    slack: 'Slack webhook URL',
    webhook: 'Webhook URL',
    email: 'Email address',
    pagerduty: 'Integration key',
  };

  return (
    <div style={{ paddingTop: 'var(--cds-spacing-06)' }}>
      {errorMsg && (
        <InlineNotification
          kind="error"
          title={errorMsg}
          onCloseButtonClick={() => setErrorMsg('')}
          style={{ marginBottom: 'var(--cds-spacing-05)' }}
        />
      )}

      <Modal
        open={channelOpen}
        modalHeading="Add notification channel"
        primaryButtonText="Add channel"
        secondaryButtonText="Cancel"
        onRequestSubmit={handleCreateChannel}
        onRequestClose={() => { setChannelOpen(false); setChannelName(''); setChannelUrl(''); }}
        primaryButtonDisabled={!channelName.trim() || !channelUrl.trim() || createChannel.isPending}
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--cds-spacing-05)' }}>
          <TextInput
            id="channel-name"
            labelText="Channel name"
            placeholder="e.g. #chaos-alerts"
            value={channelName}
            onChange={(e) => setChannelName(e.target.value)}
          />
          <Select
            id="channel-type"
            labelText="Type"
            value={channelType}
            onChange={(e) => setChannelType(e.target.value as NotificationChannelType)}
          >
            {CHANNEL_TYPES.map((t) => (
              <SelectItem key={t} value={t} text={t.charAt(0).toUpperCase() + t.slice(1)} />
            ))}
          </Select>
          <TextInput
            id="channel-url"
            labelText={urlLabel[channelType]}
            placeholder={channelType === 'email' ? 'alerts@company.com' : 'https://...'}
            value={channelUrl}
            onChange={(e) => setChannelUrl(e.target.value)}
          />
        </div>
      </Modal>

      <Modal
        open={ruleOpen}
        modalHeading="Add notification rule"
        primaryButtonText="Add rule"
        secondaryButtonText="Cancel"
        onRequestSubmit={handleCreateRule}
        onRequestClose={() => { setRuleOpen(false); setRuleChannelId(''); setRuleEvents([]); setRuleNamespace(''); }}
        primaryButtonDisabled={!ruleChannelId || ruleEvents.length === 0 || createRule.isPending}
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--cds-spacing-05)' }}>
          <Select
            id="rule-channel"
            labelText="Channel"
            value={ruleChannelId}
            onChange={(e) => setRuleChannelId(e.target.value)}
          >
            <SelectItem value="" text="Select a channel" />
            {channels.map((c) => (
              <SelectItem key={c.id} value={c.id} text={c.name} />
            ))}
          </Select>
          <fieldset style={{ border: 'none', padding: 0, margin: 0 }}>
            <legend style={{ fontSize: 'var(--cds-label-01-font-size)', color: 'var(--cds-text-secondary)', marginBottom: 'var(--cds-spacing-03)' }}>
              Events
            </legend>
            {ALL_EVENTS.map((ev) => (
              <Checkbox
                key={ev}
                id={`event-${ev}`}
                labelText={ev}
                checked={ruleEvents.includes(ev)}
                onChange={() => toggleEvent(ev)}
              />
            ))}
          </fieldset>
          <TextInput
            id="rule-namespace"
            labelText="Namespace filter (optional)"
            placeholder="e.g. production"
            value={ruleNamespace}
            onChange={(e) => setRuleNamespace(e.target.value)}
          />
        </div>
      </Modal>

      <Tile style={{ marginBottom: 'var(--cds-spacing-05)' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--cds-spacing-05)' }}>
          <h3 style={{ fontSize: 'var(--cds-heading-02-font-size)', fontWeight: 'var(--cds-heading-02-font-weight)', color: 'var(--cds-text-primary)', margin: 0 }}>
            Channels
          </h3>
          <Button renderIcon={Add} onClick={() => setChannelOpen(true)}>
            Add channel
          </Button>
        </div>
        {channelsLoading ? (
          <SkeletonText paragraph lineCount={3} />
        ) : channels.length === 0 ? (
          <p style={{ color: 'var(--cds-text-secondary)' }}>No channels configured.</p>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--cds-spacing-03)' }}>
            {channels.map((ch) => (
              <div
                key={ch.id}
                style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: 'var(--cds-spacing-04)', background: 'var(--cds-layer-02)', borderLeft: '3px solid var(--cds-interactive)' }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--cds-spacing-03)' }}>
                  <Tag type={CHANNEL_TYPE_COLORS[ch.type]} size="sm">{ch.type}</Tag>
                  <span style={{ color: 'var(--cds-text-primary)', fontWeight: 500 }}>{ch.name}</span>
                </div>
                <Button
                  kind="danger--ghost"
                  size="sm"
                  renderIcon={TrashCan}
                  iconDescription="Delete"
                  hasIconOnly
                  onClick={() => deleteChannel.mutate(ch.id)}
                  disabled={deleteChannel.isPending}
                />
              </div>
            ))}
          </div>
        )}
      </Tile>

      <Tile>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--cds-spacing-05)' }}>
          <h3 style={{ fontSize: 'var(--cds-heading-02-font-size)', fontWeight: 'var(--cds-heading-02-font-weight)', color: 'var(--cds-text-primary)', margin: 0 }}>
            Rules
          </h3>
          <Button renderIcon={Add} onClick={() => setRuleOpen(true)} disabled={channels.length === 0}>
            Add rule
          </Button>
        </div>
        {rulesLoading ? (
          <SkeletonText paragraph lineCount={3} />
        ) : rules.length === 0 ? (
          <p style={{ color: 'var(--cds-text-secondary)' }}>No rules configured.</p>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--cds-spacing-03)' }}>
            {rules.map((rule) => (
              <div
                key={rule.id}
                style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', padding: 'var(--cds-spacing-04)', background: 'var(--cds-layer-02)', borderLeft: '3px solid var(--cds-border-subtle)' }}
              >
                <div>
                  <p style={{ margin: '0 0 var(--cds-spacing-02)', color: 'var(--cds-text-primary)', fontWeight: 500 }}>
                    → {channelLabel(rule.channelId)}
                  </p>
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: 'var(--cds-spacing-02)' }}>
                    {rule.events.map((ev) => (
                      <Tag key={ev} type="gray" size="sm">{ev}</Tag>
                    ))}
                    {rule.namespaceFilter && (
                      <Tag type="blue" size="sm">ns: {rule.namespaceFilter}</Tag>
                    )}
                  </div>
                </div>
                <Button
                  kind="danger--ghost"
                  size="sm"
                  renderIcon={TrashCan}
                  iconDescription="Delete"
                  hasIconOnly
                  onClick={() => deleteRule.mutate(rule.id)}
                  disabled={deleteRule.isPending}
                />
              </div>
            ))}
          </div>
        )}
      </Tile>
    </div>
  );
}
