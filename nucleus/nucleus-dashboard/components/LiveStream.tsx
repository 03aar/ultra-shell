'use client';

import { useEffect, useRef, useState } from 'react';
import { connectWebSocket, WSMessage, ExecutionNode } from '@/lib/api';
import CommandRow from './CommandRow';

interface LiveStreamProps {
  onExecution?: (exec: ExecutionNode) => void;
  onRollback?: (id: string) => void;
}

export default function LiveStream({ onExecution, onRollback }: LiveStreamProps) {
  const [executions, setExecutions] = useState<ExecutionNode[]>([]);
  const [connected, setConnected] = useState(false);
  const [toasts, setToasts] = useState<{ id: string; message: string; level: string }[]>([]);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const ws = connectWebSocket((msg: WSMessage) => {
      setConnected(true);

      switch (msg.type) {
        case 'execution_complete': {
          const exec = msg.data as ExecutionNode;
          setExecutions((prev) => [...prev.slice(-200), exec]);
          onExecution?.(exec);

          // Show risk toasts
          if (exec.risk_flags?.length > 0) {
            const maxRisk = exec.risk_flags.reduce((max, f) => {
              const levels = ['none', 'low', 'medium', 'high', 'critical'];
              return levels.indexOf(f.level) > levels.indexOf(max) ? f.level : max;
            }, 'none');

            if (['medium', 'high', 'critical'].includes(maxRisk)) {
              const toast = {
                id: exec.id,
                message: `${maxRisk.toUpperCase()}: ${exec.risk_flags[0].message}`,
                level: maxRisk,
              };
              setToasts((prev) => [...prev, toast]);
              setTimeout(() => {
                setToasts((prev) => prev.filter((t) => t.id !== toast.id));
              }, 5000);
            }
          }
          break;
        }
        case 'rollback_complete':
          // Refresh could be done here
          break;
      }
    });

    ws.onopen = () => setConnected(true);
    ws.onclose = () => setConnected(false);

    return () => ws.close();
  }, []);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [executions]);

  return (
    <div className="relative">
      {/* Connection Status */}
      <div className="flex items-center gap-2 px-4 py-2 border-b border-nucleus-border">
        <div className={`w-2 h-2 rounded-full ${connected ? 'bg-nucleus-accent' : 'bg-red-500'} ${connected ? 'animate-pulse' : ''}`} />
        <span className="text-xs text-nucleus-muted font-mono">
          {connected ? 'Live' : 'Disconnected'}
        </span>
        <span className="text-xs text-nucleus-muted ml-auto font-mono">
          {executions.length} events
        </span>
      </div>

      {/* Command Feed */}
      <div className="max-h-[600px] overflow-auto">
        {executions.length === 0 ? (
          <div className="p-8 text-center text-nucleus-muted text-sm">
            Waiting for commands...
          </div>
        ) : (
          executions.map((exec) => (
            <CommandRow key={exec.id} execution={exec} onRollback={onRollback} />
          ))
        )}
        <div ref={bottomRef} />
      </div>

      {/* Toast Notifications */}
      <div className="fixed top-4 right-4 z-50 space-y-2">
        {toasts.map((toast) => (
          <div
            key={toast.id}
            className={`toast-enter px-4 py-3 rounded-lg border shadow-lg max-w-sm font-mono text-sm ${
              toast.level === 'critical'
                ? 'bg-red-950 border-red-500 text-red-300'
                : toast.level === 'high'
                ? 'bg-orange-950 border-orange-500 text-orange-300'
                : 'bg-yellow-950 border-yellow-500 text-yellow-300'
            }`}
          >
            {toast.message}
          </div>
        ))}
      </div>
    </div>
  );
}
