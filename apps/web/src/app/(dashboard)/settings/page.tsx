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
} from '@carbon/react';
import styles from '@/components/experiments/experiments.module.scss';
import { MembersTab } from '@/components/settings/members-tab';
import { APIKeysTab } from '@/components/settings/api-keys-tab';
import { BillingTab } from '@/components/settings/billing-tab';
import { NotificationsTab } from '@/components/settings/notifications-tab';
import { GeneralTab } from '@/components/settings/general-tab';
import { CICDTab } from '@/components/settings/cicd-tab';
import { AuditLogsTab } from '@/components/settings/audit-logs-tab';
import { SecurityTab } from '@/components/settings/security-tab';

export default function SettingsPage() {
  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div className={styles.pageHeader}>
          <h2 className={styles.pageTitle}>Settings</h2>
          <p className={styles.pageSubtitle}>
            Manage your workspace, team, and integrations.
          </p>
        </div>
      </Column>
      <Column lg={16} md={8} sm={4}>
        <Tabs>
          <TabList aria-label="Settings tabs">
            <Tab>General</Tab>
            <Tab>Members</Tab>
            <Tab>API Keys</Tab>
            <Tab>Billing</Tab>
            <Tab>Notifications</Tab>
            <Tab>CI/CD</Tab>
            <Tab>Audit Logs</Tab>
            <Tab>Security</Tab>
          </TabList>
          <TabPanels>
            <TabPanel>
              <GeneralTab />
            </TabPanel>
            <TabPanel>
              <MembersTab />
            </TabPanel>
            <TabPanel>
              <APIKeysTab />
            </TabPanel>
            <TabPanel>
              <BillingTab />
            </TabPanel>
            <TabPanel>
              <NotificationsTab />
            </TabPanel>
            <TabPanel>
              <CICDTab />
            </TabPanel>
            <TabPanel>
              <AuditLogsTab />
            </TabPanel>
            <TabPanel>
              <SecurityTab />
            </TabPanel>
          </TabPanels>
        </Tabs>
      </Column>
    </Grid>
  );
}
