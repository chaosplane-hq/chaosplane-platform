'use client';

import { useState } from 'react';
import { use } from 'react';
import {
  Grid,
  Column,
  Tile,
  Button,
  Tag,
  Modal,
  TextInput,
  TextArea,
  SkeletonText,
  InlineNotification,
  StructuredListWrapper,
  StructuredListHead,
  StructuredListRow,
  StructuredListCell,
  StructuredListBody,
} from '@carbon/react';
import { Add, ChevronLeft } from '@carbon/icons-react';
import { useRouter } from 'next/navigation';
import {
  useGameDay,
  useUpdateGameDayStatus,
  useAddGameDayEvent,
  useCreatePostmortem,
  useUpdatePostmortem,
} from '@/lib/hooks/use-gamedays';
import type { GameDayStatus } from '@/lib/types';
import styles from '@/components/experiments/experiments.module.scss';

const STATUS_TAG: Record<GameDayStatus, 'blue' | 'green' | 'gray' | 'red'> = {
  planned: 'blue',
  in_progress: 'green',
  completed: 'gray',
  cancelled: 'red',
};

const STATUS_TRANSITIONS: Record<GameDayStatus, GameDayStatus[]> = {
  planned: ['in_progress', 'cancelled'],
  in_progress: ['completed', 'cancelled'],
  completed: [],
  cancelled: [],
};

function formatDateTime(iso?: string) {
  if (!iso) return '—';
  return new Date(iso).toLocaleString();
}

export default function GameDayDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const router = useRouter();
  const { data: gameDay, isLoading, error } = useGameDay(id);
  const updateStatus = useUpdateGameDayStatus();
  const addEvent = useAddGameDayEvent();
  const createPostmortem = useCreatePostmortem();
  const updatePostmortem = useUpdatePostmortem();

  const [eventOpen, setEventOpen] = useState(false);
  const [eventTitle, setEventTitle] = useState('');
  const [eventDesc, setEventDesc] = useState('');
  const [eventOccurredAt, setEventOccurredAt] = useState('');

  const [postmortemOpen, setPostmortemOpen] = useState(false);
  const [pmSummary, setPmSummary] = useState('');
  const [pmLessons, setPmLessons] = useState('');
  const [pmActions, setPmActions] = useState('');

  const [actionError, setActionError] = useState('');

  async function handleAddEvent() {
    if (!eventTitle.trim()) return;
    setActionError('');
    try {
      await addEvent.mutateAsync({
        id,
        data: {
          title: eventTitle.trim(),
          description: eventDesc.trim() || '',
          occurredAt: eventOccurredAt || new Date().toISOString(),
        },
      });
      setEventOpen(false);
      setEventTitle('');
      setEventDesc('');
      setEventOccurredAt('');
    } catch {
      setActionError('Failed to add event.');
    }
  }

  async function handleSavePostmortem() {
    if (!pmSummary.trim()) return;
    setActionError('');
    const data = { summary: pmSummary, lessonsLearned: pmLessons, actionItems: pmActions };
    try {
      if (gameDay?.postmortem) {
        await updatePostmortem.mutateAsync({ id, data });
      } else {
        await createPostmortem.mutateAsync({ id, data });
      }
      setPostmortemOpen(false);
    } catch {
      setActionError('Failed to save postmortem.');
    }
  }

  function openPostmortem() {
    if (gameDay?.postmortem) {
      setPmSummary(gameDay.postmortem.summary);
      setPmLessons(gameDay.postmortem.lessonsLearned ?? '');
      setPmActions(gameDay.postmortem.actionItems ?? '');
    }
    setPostmortemOpen(true);
  }

  if (isLoading) {
    return (
      <Grid fullWidth>
        <Column lg={16} md={8} sm={4}>
          <SkeletonText paragraph lineCount={8} />
        </Column>
      </Grid>
    );
  }

  if (error || !gameDay) {
    return (
      <Grid fullWidth>
        <Column lg={16} md={8} sm={4}>
          <InlineNotification kind="error" title="Failed to load game day." hideCloseButton />
        </Column>
      </Grid>
    );
  }

  const transitions = STATUS_TRANSITIONS[gameDay.status];

  return (
    <Grid fullWidth>
      <Column lg={16} md={8} sm={4}>
        <div style={{ marginBottom: 'var(--cds-spacing-05)' }}>
          <Button kind="ghost" renderIcon={ChevronLeft} onClick={() => router.push('/gamedays')} size="sm">
            Back to Game Days
          </Button>
        </div>
        <div className={styles.pageHeader} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <h2 className={styles.pageTitle}>{gameDay.title}</h2>
            {gameDay.description && <p className={styles.pageSubtitle}>{gameDay.description}</p>}
          </div>
          <div style={{ display: 'flex', gap: 'var(--cds-spacing-03)', alignItems: 'center' }}>
            <Tag type={STATUS_TAG[gameDay.status]}>{gameDay.status.replace('_', ' ')}</Tag>
            {transitions.map((s) => (
              <Button
                key={s}
                kind={s === 'cancelled' ? 'danger--ghost' : 'secondary'}
                size="sm"
                onClick={() => updateStatus.mutate({ id, status: s })}
                disabled={updateStatus.isPending}
              >
                Mark {s.replace('_', ' ')}
              </Button>
            ))}
          </div>
        </div>
      </Column>

      {actionError && (
        <Column lg={16} md={8} sm={4}>
          <InlineNotification
            kind="error"
            title={actionError}
            onCloseButtonClick={() => setActionError('')}
            style={{ marginBottom: 'var(--cds-spacing-05)' }}
          />
        </Column>
      )}

      <Column lg={8} md={8} sm={4}>
        <Tile style={{ marginBottom: 'var(--cds-spacing-06)' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--cds-spacing-05)' }}>
            <h3 className={styles.sectionTitle} style={{ margin: 0 }}>Event Timeline</h3>
            <Button renderIcon={Add} size="sm" onClick={() => setEventOpen(true)}>
              Add Event
            </Button>
          </div>

          {(!gameDay.events || gameDay.events.length === 0) ? (
            <p style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)' }}>
              No events recorded yet.
            </p>
          ) : (
            <div className={styles.timeline}>
              {gameDay.events.map((evt, i) => (
                <div key={evt.id} className={`${styles.timelineItem} ${styles.timelineDone}`}>
                  <div className={styles.timelineDot} />
                  {i < (gameDay.events?.length ?? 0) - 1 && <div className={styles.timelineLine} />}
                  <div className={styles.timelineContent}>
                    <p className={styles.timelineLabel}>{evt.title}</p>
                    {evt.description && (
                      <p style={{ margin: '2px 0', fontSize: 'var(--cds-label-01-font-size)', color: 'var(--cds-text-secondary)' }}>
                        {evt.description}
                      </p>
                    )}
                    <p className={styles.timelineTime}>{formatDateTime(evt.occurredAt)}</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </Tile>
      </Column>

      <Column lg={8} md={8} sm={4}>
        <Tile style={{ marginBottom: 'var(--cds-spacing-06)' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--cds-spacing-05)' }}>
            <h3 className={styles.sectionTitle} style={{ margin: 0 }}>Postmortem</h3>
            <Button size="sm" kind={gameDay.postmortem ? 'secondary' : 'primary'} onClick={openPostmortem}>
              {gameDay.postmortem ? 'Edit Postmortem' : 'Write Postmortem'}
            </Button>
          </div>

          {!gameDay.postmortem ? (
            <p style={{ color: 'var(--cds-text-secondary)', fontSize: 'var(--cds-label-01-font-size)' }}>
              No postmortem written yet.
            </p>
          ) : (
            <StructuredListWrapper>
              <StructuredListHead>
                <StructuredListRow head>
                  <StructuredListCell head>Section</StructuredListCell>
                  <StructuredListCell head>Content</StructuredListCell>
                </StructuredListRow>
              </StructuredListHead>
              <StructuredListBody>
                <StructuredListRow>
                  <StructuredListCell noWrap>Summary</StructuredListCell>
                  <StructuredListCell>{gameDay.postmortem.summary}</StructuredListCell>
                </StructuredListRow>
                <StructuredListRow>
                  <StructuredListCell noWrap>Lessons Learned</StructuredListCell>
                  <StructuredListCell>{gameDay.postmortem.lessonsLearned}</StructuredListCell>
                </StructuredListRow>
                <StructuredListRow>
                  <StructuredListCell noWrap>Action Items</StructuredListCell>
                  <StructuredListCell>{gameDay.postmortem.actionItems}</StructuredListCell>
                </StructuredListRow>
              </StructuredListBody>
            </StructuredListWrapper>
          )}
        </Tile>

        <Tile>
          <h3 className={styles.sectionTitle}>Details</h3>
          <StructuredListWrapper>
            <StructuredListBody>
              <StructuredListRow>
                <StructuredListCell noWrap>Scheduled</StructuredListCell>
                <StructuredListCell>{formatDateTime(gameDay.scheduledAt)}</StructuredListCell>
              </StructuredListRow>
              <StructuredListRow>
                <StructuredListCell noWrap>Created</StructuredListCell>
                <StructuredListCell>{formatDateTime(gameDay.createdAt)}</StructuredListCell>
              </StructuredListRow>
              {gameDay.completedAt && (
                <StructuredListRow>
                  <StructuredListCell noWrap>Completed</StructuredListCell>
                  <StructuredListCell>{formatDateTime(gameDay.completedAt)}</StructuredListCell>
                </StructuredListRow>
              )}
            </StructuredListBody>
          </StructuredListWrapper>
        </Tile>
      </Column>

      <Modal
        open={eventOpen}
        modalHeading="Add Event"
        primaryButtonText="Add"
        secondaryButtonText="Cancel"
        onRequestSubmit={handleAddEvent}
        onRequestClose={() => { setEventOpen(false); setEventTitle(''); setEventDesc(''); setEventOccurredAt(''); }}
        primaryButtonDisabled={!eventTitle.trim() || addEvent.isPending}
      >
        <TextInput
          id="evt-title"
          labelText="Event title"
          placeholder="e.g. Pod kill triggered"
          value={eventTitle}
          onChange={(e) => setEventTitle(e.target.value)}
          style={{ marginBottom: 'var(--cds-spacing-05)' }}
        />
        <TextArea
          id="evt-desc"
          labelText="Description (optional)"
          value={eventDesc}
          onChange={(e) => setEventDesc(e.target.value)}
          rows={2}
          style={{ marginBottom: 'var(--cds-spacing-05)' }}
        />
        <TextInput
          id="evt-occurred"
          labelText="Occurred at (ISO, optional)"
          placeholder={new Date().toISOString()}
          value={eventOccurredAt}
          onChange={(e) => setEventOccurredAt(e.target.value)}
        />
      </Modal>

      <Modal
        open={postmortemOpen}
        modalHeading={gameDay.postmortem ? 'Edit Postmortem' : 'Write Postmortem'}
        primaryButtonText="Save"
        secondaryButtonText="Cancel"
        onRequestSubmit={handleSavePostmortem}
        onRequestClose={() => setPostmortemOpen(false)}
        primaryButtonDisabled={!pmSummary.trim() || createPostmortem.isPending || updatePostmortem.isPending}
        size="lg"
      >
        <TextArea
          id="pm-summary"
          labelText="Summary"
          placeholder="What happened during the game day?"
          value={pmSummary}
          onChange={(e) => setPmSummary(e.target.value)}
          rows={4}
          style={{ marginBottom: 'var(--cds-spacing-05)' }}
        />
        <TextArea
          id="pm-lessons"
          labelText="Lessons Learned"
          placeholder="What did the team learn?"
          value={pmLessons}
          onChange={(e) => setPmLessons(e.target.value)}
          rows={4}
          style={{ marginBottom: 'var(--cds-spacing-05)' }}
        />
        <TextArea
          id="pm-actions"
          labelText="Action Items"
          placeholder="What follow-up actions are needed?"
          value={pmActions}
          onChange={(e) => setPmActions(e.target.value)}
          rows={4}
        />
      </Modal>
    </Grid>
  );
}
