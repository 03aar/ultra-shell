'use client';

import { useEffect, useState } from 'react';
import { Session, ExecutionNode, getSessions, getSessionReplay } from '@/lib/api';
import CommandRow from '@/components/CommandRow';

export default function SessionsPage() {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [selectedSession, setSelectedSession] = useState<string | null>(null);
  const [replayData, setReplayData] = useState<ExecutionNode[]>([]);
  const [replaying, setReplaying] = useState(false);
  const [replayIndex, setReplayIndex] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchSessions = async () => {
      try {
        const data = await getSessions();
        setSessions(data || []);
      } catch (e) {
        console.error('Failed to fetch sessions:', e);
      } finally {
        setLoading(false);
      }
    };
    fetchSessions();
  }, []);

  const loadReplay = async (sessionId: string) => {
    setSelectedSession(sessionId);
    try {
      const data = await getSessionReplay(sessionId);
      setReplayData(data.executions || []);
      setReplayIndex(0);
      setReplaying(false);
    } catch (e) {
      console.error('Failed to load replay:', e);
    }
  };

  const startReplay = () => {
    setReplaying(true);
    setReplayIndex(0);
  };

  useEffect(() => {
    if (!replaying || replayIndex >= replayData.length) {
      setReplaying(false);
      return;
    }
    const timer = setTimeout(() => {
      setReplayIndex((prev) => prev + 1);
    }, 500);
    return () => clearTimeout(timer);
  }, [replaying, replayIndex, replayData.length]);

  const exportSession = () => {
    if (!replayData.length) return;
    const blob = new Blob([JSON.stringify(replayData, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `nucleus-session-${selectedSession}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const formatDuration = (session: Session) => {
    const start = new Date(session.start_time).getTime();
    const end = session.end_time ? new Date(session.end_time).getTime() : Date.now();
    const mins = Math.floor((end - start) / 60000);
    if (mins < 60) return `${mins}m`;
    return `${Math.floor(mins / 60)}h ${mins % 60}m`;
  };

  return (
    <div className="h-full flex flex-col">
      <div className="px-6 py-4 border-b border-nucleus-border">
        <h1 className="font-heading text-2xl font-bold">Sessions</h1>
      </div>

      <div className="flex-1 flex overflow-hidden">
        {/* Session List */}
        <div className="w-96 border-r border-nucleus-border overflow-auto">
          {loading ? (
            <div className="p-8 text-center text-nucleus-muted text-sm">Loading sessions...</div>
          ) : sessions.length === 0 ? (
            <div className="p-8 text-center text-nucleus-muted text-sm">No sessions yet. Start Nucleus to create one.</div>
          ) : (
            sessions.map((session) => (
              <div
                key={session.id}
                className={`p-4 border-b border-nucleus-border cursor-pointer transition-colors ${
                  selectedSession === session.id ? 'bg-nucleus-panel' : 'hover:bg-nucleus-panel/50'
                }`}
                onClick={() => loadReplay(session.id)}
              >
                <div className="flex items-center justify-between">
                  <span className="font-mono text-sm font-semibold text-nucleus-text">{session.name}</span>
                  {!session.end_time && (
                    <span className="text-xs px-2 py-0.5 bg-nucleus-accent/10 text-nucleus-accent rounded-full">Active</span>
                  )}
                </div>
                <div className="mt-2 grid grid-cols-3 gap-2 text-xs text-nucleus-muted">
                  <div>
                    <span className="block text-[10px] uppercase tracking-wider">Commands</span>
                    <span className="text-nucleus-text font-mono">{session.execution_count}</span>
                  </div>
                  <div>
                    <span className="block text-[10px] uppercase tracking-wider">Risks</span>
                    <span className="text-yellow-400 font-mono">{session.risk_warning_count}</span>
                  </div>
                  <div>
                    <span className="block text-[10px] uppercase tracking-wider">Duration</span>
                    <span className="text-nucleus-text font-mono">{formatDuration(session)}</span>
                  </div>
                </div>
                <p className="text-xs text-nucleus-muted mt-2 font-mono">
                  {new Date(session.start_time).toLocaleString()}
                </p>
              </div>
            ))
          )}
        </div>

        {/* Replay Area */}
        <div className="flex-1 flex flex-col overflow-hidden">
          {selectedSession ? (
            <>
              <div className="px-4 py-3 border-b border-nucleus-border flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <h2 className="font-heading text-sm font-semibold text-nucleus-muted uppercase">
                    Session Replay
                  </h2>
                  {replaying && (
                    <span className="text-xs font-mono text-nucleus-accent">
                      {replayIndex}/{replayData.length}
                    </span>
                  )}
                </div>
                <div className="flex items-center gap-2">
                  <button
                    onClick={startReplay}
                    disabled={replayData.length === 0}
                    className="px-3 py-1 text-xs bg-nucleus-panel border border-nucleus-border rounded hover:border-nucleus-accent hover:text-nucleus-accent transition-colors disabled:opacity-50 font-mono"
                  >
                    {replaying ? 'Replaying...' : 'Replay'}
                  </button>
                  <button
                    onClick={exportSession}
                    disabled={replayData.length === 0}
                    className="px-3 py-1 text-xs bg-nucleus-panel border border-nucleus-border rounded hover:border-nucleus-accent hover:text-nucleus-accent transition-colors disabled:opacity-50 font-mono"
                  >
                    Export JSON
                  </button>
                </div>
              </div>

              {/* Replay progress bar */}
              {replaying && (
                <div className="h-1 bg-nucleus-panel">
                  <div
                    className="h-full bg-nucleus-accent transition-all duration-500"
                    style={{ width: `${(replayIndex / Math.max(replayData.length, 1)) * 100}%` }}
                  />
                </div>
              )}

              <div className="flex-1 overflow-auto">
                {replayData.length === 0 ? (
                  <div className="p-8 text-center text-nucleus-muted text-sm">
                    No executions in this session
                  </div>
                ) : (
                  (replaying ? replayData.slice(0, replayIndex) : replayData).map((exec) => (
                    <CommandRow key={exec.id} execution={exec} />
                  ))
                )}
              </div>
            </>
          ) : (
            <div className="flex-1 flex items-center justify-center text-nucleus-muted text-sm">
              Select a session to view its replay
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
