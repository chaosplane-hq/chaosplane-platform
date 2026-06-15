'use client';

import { useEffect, useRef, useState } from 'react';
import { getWsUrl } from '@/lib/api';
import { getAccessToken } from '@/lib/auth';
import type { ExperimentStatusMessage } from '@/lib/types';

export function useExperimentWs(name: string) {
  const [lastMessage, setLastMessage] = useState<ExperimentStatusMessage | null>(null);
  const [connected, setConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const unmounted = useRef(false);

  useEffect(() => {
    if (!name) return;
    unmounted.current = false;

    function connect() {
      if (unmounted.current) return;
      const token = getAccessToken();
      if (!token) return;
      const ws = new WebSocket(getWsUrl(`/ws/experiments/${name}?token=${encodeURIComponent(token)}`));
      wsRef.current = ws;

      ws.onopen = () => setConnected(true);

      ws.onmessage = (evt) => {
        try {
          const msg = JSON.parse(evt.data as string) as ExperimentStatusMessage;
          setLastMessage(msg);
        } catch {
          /* empty */
        }
      };

      ws.onclose = () => {
        setConnected(false);
        if (!unmounted.current) {
          reconnectTimer.current = setTimeout(connect, 3000);
        }
      };

      ws.onerror = () => ws.close();
    }

    connect();

    return () => {
      unmounted.current = true;
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current);
      wsRef.current?.close();
    };
  }, [name]);

  return { lastMessage, connected };
}
