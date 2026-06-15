'use client';

import { useState, useRef, useLayoutEffect } from 'react';
import {
  Button,
  TextInput,
  SkeletonText,
  InlineNotification,
} from '@carbon/react';
import { Add, TrashCan, SendAlt } from '@carbon/icons-react';
import {
  useChatSessions,
  useCreateChatSession,
  useDeleteChatSession,
  useChatMessages,
  useSendMessage,
} from '@/lib/hooks/use-ai-chat';
import type { ChatSession, ChatMessage } from '@/lib/types';

function SessionList({
  sessions,
  selectedId,
  onSelect,
  onDelete,
  onCreate,
  isCreating,
}: {
  sessions: ChatSession[];
  selectedId: string | null;
  onSelect: (id: string) => void;
  onDelete: (id: string) => void;
  onCreate: () => void;
  isCreating: boolean;
}) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%', borderRight: '1px solid var(--cds-border-subtle)' }}>
      <div style={{ padding: '1rem', borderBottom: '1px solid var(--cds-border-subtle)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <span style={{ fontWeight: 600 }}>Sessions</span>
        <Button kind="ghost" size="sm" renderIcon={Add} iconDescription="New session" hasIconOnly disabled={isCreating} onClick={onCreate} />
      </div>
      <div style={{ overflowY: 'auto', flexGrow: 1 }}>
        {sessions.length === 0 ? (
          <p style={{ padding: '1rem', color: 'var(--cds-text-secondary)', fontSize: '0.875rem' }}>No sessions yet.</p>
        ) : sessions.map((s) => (
          <button
            key={s.id}
            type="button"
            onClick={() => onSelect(s.id)}
            style={{
              padding: '0.75rem 1rem',
              cursor: 'pointer',
              background: selectedId === s.id ? 'var(--cds-layer-selected)' : 'transparent',
              borderBottom: '1px solid var(--cds-border-subtle-00)',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              gap: '0.5rem',
              width: '100%',
              border: 'none',
              textAlign: 'left',
              color: 'inherit',
            }}
          >
            <div style={{ overflow: 'hidden' }}>
              <p style={{ margin: 0, fontWeight: 500, fontSize: '0.875rem', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                {s.title || 'Untitled'}
              </p>
              <p style={{ margin: 0, fontSize: '0.75rem', color: 'var(--cds-text-secondary)' }}>
                {new Date(s.updatedAt).toLocaleDateString()}
              </p>
            </div>
            <Button
              kind="ghost"
              size="sm"
              renderIcon={TrashCan}
              iconDescription="Delete session"
              hasIconOnly
              onClick={(e) => { e.stopPropagation(); onDelete(s.id); }}
            />
          </button>
        ))}
      </div>
    </div>
  );
}

function MessageBubble({ message }: { message: ChatMessage }) {
  const isUser = message.role === 'user';
  return (
    <div style={{ display: 'flex', justifyContent: isUser ? 'flex-end' : 'flex-start', marginBottom: '0.75rem' }}>
      <div style={{
        maxWidth: '70%',
        padding: '0.75rem 1rem',
        borderRadius: isUser ? '1rem 1rem 0.25rem 1rem' : '1rem 1rem 1rem 0.25rem',
        background: isUser ? 'var(--cds-interactive)' : 'var(--cds-layer-02)',
        color: isUser ? 'var(--cds-text-on-color)' : 'var(--cds-text-primary)',
        fontSize: '0.875rem',
        lineHeight: 1.5,
        whiteSpace: 'pre-wrap',
        wordBreak: 'break-word',
      }}>
        {message.content}
      </div>
    </div>
  );
}

function ChatArea({ sessionId }: { sessionId: string }) {
  const { data, isLoading, isError } = useChatMessages(sessionId);
  const send = useSendMessage(sessionId);
  const [input, setInput] = useState('');
  const bottomRef = useRef<HTMLDivElement>(null);

  const messages = data?.items ?? [];

  useLayoutEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  });

  function handleSend() {
    const content = input.trim();
    if (!content || send.isPending) return;
    setInput('');
    send.mutate({ content });
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  }

  if (isLoading) return <div style={{ padding: '2rem' }}><SkeletonText paragraph lineCount={4} /></div>;
  if (isError) return <div style={{ padding: '2rem' }}><InlineNotification kind="error" title="Failed to load messages" subtitle="" /></div>;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ flexGrow: 1, overflowY: 'auto', padding: '1.5rem' }}>
        {messages.length === 0 ? (
          <p style={{ color: 'var(--cds-text-secondary)', textAlign: 'center', marginTop: '4rem' }}>
            Start the conversation by typing a message below.
          </p>
        ) : messages.map((m) => (
          <MessageBubble key={m.id} message={m} />
        ))}
        {send.isPending && (
          <div style={{ display: 'flex', justifyContent: 'flex-start', marginBottom: '0.75rem' }}>
            <div style={{ padding: '0.75rem 1rem', borderRadius: '1rem 1rem 1rem 0.25rem', background: 'var(--cds-layer-02)' }}>
              <SkeletonText width="120px" />
            </div>
          </div>
        )}
        <div ref={bottomRef} />
      </div>
      <div style={{ padding: '1rem', borderTop: '1px solid var(--cds-border-subtle)', display: 'flex', gap: '0.5rem' }}>
        <TextInput
          id="chat-input"
          labelText=""
          hideLabel
          placeholder="Ask anything about your chaos experiments…"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          disabled={send.isPending}
        />
        <Button
          renderIcon={SendAlt}
          iconDescription="Send"
          hasIconOnly
          kind="primary"
          disabled={!input.trim() || send.isPending}
          onClick={handleSend}
        />
      </div>
    </div>
  );
}

export default function AIChatPage() {
  const { data, isLoading } = useChatSessions();
  const createSession = useCreateChatSession();
  const deleteSession = useDeleteChatSession();
  const [selectedId, setSelectedId] = useState<string | null>(null);

  const sessions = data?.items ?? [];

  function handleCreate() {
    createSession.mutate(undefined, {
      onSuccess: (s) => setSelectedId(s.id),
    });
  }

  function handleDelete(id: string) {
    deleteSession.mutate(id, {
      onSuccess: () => {
        if (selectedId === id) setSelectedId(null);
      },
    });
  }

  return (
    <div style={{ display: 'flex', height: 'calc(100vh - 3rem)', overflow: 'hidden' }}>
      <div style={{ width: '280px', flexShrink: 0 }}>
        {isLoading ? (
          <div style={{ padding: '1rem' }}><SkeletonText paragraph lineCount={5} /></div>
        ) : (
          <SessionList
            sessions={sessions}
            selectedId={selectedId}
            onSelect={setSelectedId}
            onDelete={handleDelete}
            onCreate={handleCreate}
            isCreating={createSession.isPending}
          />
        )}
      </div>
      <div style={{ flexGrow: 1, overflow: 'hidden' }}>
        {selectedId ? (
          <ChatArea sessionId={selectedId} />
        ) : (
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%', flexDirection: 'column', gap: '1rem' }}>
            <p style={{ color: 'var(--cds-text-secondary)' }}>Select a session or create a new one to start chatting.</p>
            <Button renderIcon={Add} kind="primary" disabled={createSession.isPending} onClick={handleCreate}>
              New Session
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}
