'use client';

import { useEffect, useRef, useState, useCallback } from 'react';
import { ExecutionNode, ContextState, WSMessage } from '@/lib/api';

const WS_BASE = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/api/v1';
const API_KEY = process.env.NEXT_PUBLIC_API_KEY || 'dev-nucleus-key-local';

interface NucleusStreamState {
  executions: ExecutionNode[];
  warnings: WSMessage[];
  connected: boolean;
  error: string | null;
  lastEvent: WSMessage | null;
}

export function useNucleusStream() {
  const [state, setState] = useState<NucleusStreamState>({
    executions: [],
    warnings: [],
    connected: false,
    error: null,
    lastEvent: null,
  });
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeout = useRef<NodeJS.Timeout>();

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    const ws = new WebSocket(`${WS_BASE}/ws/stream?api_key=${API_KEY}`);
    wsRef.current = ws;

    ws.onopen = () => {
      setState(prev => ({ ...prev, connected: true, error: null }));
    };

    ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data);
        setState(prev => {
          const next = { ...prev, lastEvent: msg };

          switch (msg.type) {
            case 'execution_complete':
              next.executions = [...prev.executions.slice(-200), msg.data as ExecutionNode];
              break;
            case 'risk_warning':
              next.warnings = [...prev.warnings.slice(-50), msg];
              break;
          }

          return next;
        });
      } catch {
        // ignore parse errors
      }
    };

    ws.onclose = () => {
      setState(prev => ({ ...prev, connected: false }));
      reconnectTimeout.current = setTimeout(connect, 3000);
    };

    ws.onerror = () => {
      setState(prev => ({ ...prev, error: 'WebSocket connection failed' }));
    };
  }, []);

  useEffect(() => {
    connect();
    return () => {
      clearTimeout(reconnectTimeout.current);
      wsRef.current?.close();
    };
  }, [connect]);

  return state;
}
