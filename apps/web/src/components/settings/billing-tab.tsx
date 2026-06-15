'use client';

import {
  Tile,
  Button,
  Tag,
  SkeletonText,
  InlineNotification,
  Modal,
} from '@carbon/react';
import { useState } from 'react';
import { useBilling, useUpgradePlan, useCancelPlan } from '@/lib/hooks/use-billing';
import type { BillingPlan } from '@/lib/types';

const PLAN_LABELS: Record<BillingPlan, string> = {
  free: 'Free',
  pro: 'Pro',
  enterprise: 'Enterprise',
};

const PLAN_COLORS: Record<BillingPlan, 'gray' | 'blue' | 'purple'> = {
  free: 'gray',
  pro: 'blue',
  enterprise: 'purple',
};

function UsageBar({ used, limit, label }: { used: number; limit: number; label: string }) {
  const pct = limit > 0 ? Math.min((used / limit) * 100, 100) : 0;
  const color = pct >= 90 ? 'var(--cds-support-error)' : pct >= 70 ? 'var(--cds-support-warning)' : 'var(--cds-interactive)';
  return (
    <div style={{ marginBottom: 'var(--cds-spacing-05)' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 'var(--cds-spacing-02)' }}>
        <span style={{ fontSize: 'var(--cds-label-01-font-size)', color: 'var(--cds-text-secondary)' }}>{label}</span>
        <span style={{ fontSize: 'var(--cds-label-01-font-size)', color: 'var(--cds-text-primary)', fontVariantNumeric: 'tabular-nums' }}>
          {used.toLocaleString()} / {limit === -1 ? '∞' : limit.toLocaleString()}
        </span>
      </div>
      <div style={{ height: '6px', background: 'var(--cds-layer-02)', borderRadius: '3px', overflow: 'hidden' }}>
        <div style={{ height: '100%', width: `${pct}%`, background: color, borderRadius: '3px', transition: 'width 0.3s ease' }} />
      </div>
    </div>
  );
}

export function BillingTab() {
  const [cancelOpen, setCancelOpen] = useState(false);
  const [upgradeOpen, setUpgradeOpen] = useState(false);
  const [errorMsg, setErrorMsg] = useState('');

  const { data, isLoading } = useBilling();
  const upgradePlan = useUpgradePlan();
  const cancelPlan = useCancelPlan();

  const subscription = data?.subscription ?? null;
  const usage = data?.usage ?? null;
  const limits = data?.limits ?? null;
  const plan: BillingPlan = subscription?.plan ?? 'free';

  async function handleUpgrade() {
    setErrorMsg('');
    try {
      await upgradePlan.mutateAsync('pro');
      setUpgradeOpen(false);
    } catch {
      setErrorMsg('Failed to upgrade plan.');
    }
  }

  async function handleCancel() {
    setErrorMsg('');
    try {
      await cancelPlan.mutateAsync();
      setCancelOpen(false);
    } catch {
      setErrorMsg('Failed to cancel plan.');
    }
  }

  return (
    <div style={{ paddingTop: 'var(--cds-spacing-06)', maxWidth: '720px' }}>
      {errorMsg && (
        <InlineNotification
          kind="error"
          title={errorMsg}
          onCloseButtonClick={() => setErrorMsg('')}
          style={{ marginBottom: 'var(--cds-spacing-05)' }}
        />
      )}

      <Modal
        open={upgradeOpen}
        modalHeading="Upgrade to Pro"
        primaryButtonText="Confirm upgrade"
        secondaryButtonText="Cancel"
        onRequestSubmit={handleUpgrade}
        onRequestClose={() => setUpgradeOpen(false)}
        primaryButtonDisabled={upgradePlan.isPending}
      >
        <p style={{ color: 'var(--cds-text-secondary)' }}>
          You'll be upgraded to the Pro plan. Billing will start at the next cycle.
        </p>
      </Modal>

      <Modal
        open={cancelOpen}
        danger
        modalHeading="Cancel subscription"
        primaryButtonText="Yes, cancel"
        secondaryButtonText="Keep plan"
        onRequestSubmit={handleCancel}
        onRequestClose={() => setCancelOpen(false)}
        primaryButtonDisabled={cancelPlan.isPending}
      >
        <p style={{ color: 'var(--cds-text-secondary)' }}>
          Your plan will be downgraded to Free at the end of the current billing period.
        </p>
      </Modal>

      <Tile style={{ marginBottom: 'var(--cds-spacing-05)' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: 'var(--cds-spacing-04)' }}>
          <div>
            <p style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)', margin: '0 0 var(--cds-spacing-02)' }}>
              Current plan
            </p>
            {isLoading ? (
              <SkeletonText width="120px" />
            ) : (
              <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--cds-spacing-03)' }}>
                <span style={{ fontSize: 'var(--cds-heading-04-font-size)', fontWeight: 600, color: 'var(--cds-text-primary)' }}>
                  {PLAN_LABELS[plan]}
                </span>
                <Tag type={PLAN_COLORS[plan]} size="sm">
                  {subscription?.status ?? 'active'}
                </Tag>
              </div>
            )}
            {subscription?.currentPeriodEnd && (
              <p style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)', margin: 'var(--cds-spacing-02) 0 0' }}>
                Renews {new Date(subscription.currentPeriodEnd).toLocaleDateString()}
              </p>
            )}
          </div>
          {!isLoading && (
            <div style={{ display: 'flex', gap: 'var(--cds-spacing-03)' }}>
              {plan !== 'enterprise' && (
                <Button kind="primary" onClick={() => setUpgradeOpen(true)}>
                  Upgrade
                </Button>
              )}
              {plan !== 'free' && (
                <Button kind="ghost" onClick={() => setCancelOpen(true)}>
                  Cancel plan
                </Button>
              )}
            </div>
          )}
        </div>
      </Tile>

      <Tile>
        <h3 style={{ fontSize: 'var(--cds-heading-02-font-size)', fontWeight: 'var(--cds-heading-02-font-weight)', color: 'var(--cds-text-primary)', margin: '0 0 var(--cds-spacing-06)' }}>
          Usage
        </h3>
        {isLoading ? (
          <SkeletonText paragraph lineCount={6} />
        ) : usage ? (
          <>
            <UsageBar
              label="Experiments run"
              used={usage.experiments ?? 0}
              limit={limits?.maxExperiments ?? 0}
            />
            <UsageBar
              label="Agents"
              used={usage.agents ?? 0}
              limit={limits?.maxAgents ?? 0}
            />
            <UsageBar
              label="API calls"
              used={usage.apiCalls ?? 0}
              limit={limits?.maxApiCalls ?? 0}
            />
          </>
        ) : (
          <p style={{ color: 'var(--cds-text-secondary)' }}>No usage data available.</p>
        )}
      </Tile>
    </div>
  );
}
