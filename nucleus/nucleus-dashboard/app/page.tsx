'use client';

import { useEffect, useState } from 'react';
import { ExecutionNode, getExecutions, rollback as rollbackExec } from '@/lib/api';
import CommandRow from '@/components/CommandRow';
import ContextPanel from '@/components/ContextPanel';
import LiveStream from '@/components/LiveStream';
import RollbackModal from '@/components/RollbackModal';
import Terminal from '@/components/Terminal';

export default function Dashboard() {
  const [executions, setExecutions] = useState<ExecutionNode[]>([]);
  const [stats, setStats] = useState({ total: 0, riskWarnings: 0, rollbacks: 0, sessionStart: '' });
  const [rollbackTarget, setRollbackTarget] = useState<ExecutionNode | null>(null);
  const [selectedExec, setSelectedExec] = useState<ExecutionNode | null>(null);
  const [showTerminal, setShowTerminal] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const data = await getExecutions({ limit: 50 });
        setExecutions(data.executions || []);
        const execs = data.executions || [];
        const riskCount = execs.filter((e) => e.risk_flags.length > 0).length;
        const rollbackCount = execs.filter((e) => e.rolled_back).length;
        setStats({
          total: data.total,
          riskWarnings: riskCount,
          rollbacks: rollbackCount,
          sessionStart: execs.length > 0 ? execs[execs.length - 1].timestamp : new Date().toISOString(),
        });
      } catch (e) {
        console.error('Failed to fetch executions:', e);
      }
    };
    fetchData();
    const interval = setInterval(fetchData, 10000);
    return () => clearInterval(interval);
  }, []);

  const handleRollback = (id: string) => {
    const exec = executions.find((e) => e.id === id);
    if (exec) setRollbackTarget(exec);
  };

  const sessionDuration = () => {
    if (!stats.sessionStart) return '0m';
    const diff = Date.now() - new Date(stats.sessionStart).getTime();
    const mins = Math.floor(diff / 60000);
    const hrs = Math.floor(mins / 60);
    if (hrs > 0) return `${hrs}h ${mins % 60}m`;
    return `${mins}m`;
  };

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="px-6 py-4 border-b border-nucleus-border">
        <div className="flex items-center justify-between">
          <h1 className="font-heading text-2xl font-bold">Dashboard</h1>
          <button
            onClick={() => setShowTerminal(!showTerminal)}
            className="px-3 py-1.5 text-sm bg-nucleus-panel border border-nucleus-border rounded hover:border-nucleus-accent hover:text-nucleus-accent transition-colors font-mono"
          >
            {showTerminal ? 'Hide Terminal' : 'Show Terminal'}
          </button>
        </div>
      </div>

      {/* Stats Bar */}
      <div className="grid grid-cols-4 gap-4 px-6 py-4 border-b border-nucleus-border">
        <div className="bg-nucleus-surface border border-nucleus-border rounded-lg p-4">
          <p className="text-xs text-nucleus-muted uppercase tracking-wider">Executions</p>
          <p className="text-2xl font-heading font-bold text-nucleus-accent mt-1">{stats.total}</p>
        </div>
        <div className="bg-nucleus-surface border border-nucleus-border rounded-lg p-4">
          <p className="text-xs text-nucleus-muted uppercase tracking-wider">Risk Warnings</p>
          <p className="text-2xl font-heading font-bold text-yellow-400 mt-1">{stats.riskWarnings}</p>
        </div>
        <div className="bg-nucleus-surface border border-nucleus-border rounded-lg p-4">
          <p className="text-xs text-nucleus-muted uppercase tracking-wider">Rollbacks</p>
          <p className="text-2xl font-heading font-bold text-orange-400 mt-1">{stats.rollbacks}</p>
        </div>
        <div className="bg-nucleus-surface border border-nucleus-border rounded-lg p-4">
          <p className="text-xs text-nucleus-muted uppercase tracking-wider">Session Duration</p>
          <p className="text-2xl font-heading font-bold text-nucleus-text mt-1">{sessionDuration()}</p>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Left: Command Feed */}
        <div className="flex-1 flex flex-col border-r border-nucleus-border">
          <div className="px-4 py-2 border-b border-nucleus-border">
            <h2 className="text-sm font-heading font-semibold text-nucleus-muted uppercase tracking-wider">
              Live Command Feed
            </h2>
          </div>
          <div className="flex-1 overflow-auto">
            <LiveStream onRollback={handleRollback} />
          </div>

          {/* Terminal (toggleable) */}
          {showTerminal && (
            <div className="h-64 border-t border-nucleus-border">
              <Terminal />
            </div>
          )}
        </div>

        {/* Right: Context Panel */}
        <div className="w-80 overflow-auto">
          <div className="px-4 py-2 border-b border-nucleus-border">
            <h2 className="text-sm font-heading font-semibold text-nucleus-muted uppercase tracking-wider">
              Context
            </h2>
          </div>
          <div className="p-4">
            <ContextPanel />
          </div>
        </div>
      </div>

      {/* Execution Detail Panel */}
      {selectedExec && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-40" onClick={() => setSelectedExec(null)}>
          <div className="bg-nucleus-surface border border-nucleus-border rounded-lg p-6 max-w-2xl w-full mx-4 max-h-[80vh] overflow-auto" onClick={(e) => e.stopPropagation()}>
            <h2 className="font-heading text-lg font-bold mb-4">Execution Detail</h2>
            <div className="space-y-3">
              <div>
                <span className="text-xs text-nucleus-muted">Command</span>
                <code className="block font-mono text-sm text-nucleus-accent mt-1">{selectedExec.command.raw}</code>
              </div>
              <div className="grid grid-cols-3 gap-4">
                <div>
                  <span className="text-xs text-nucleus-muted">Exit Code</span>
                  <p className={`font-mono ${selectedExec.exit_code === 0 ? 'text-green-400' : 'text-red-400'}`}>{selectedExec.exit_code}</p>
                </div>
                <div>
                  <span className="text-xs text-nucleus-muted">Duration</span>
                  <p className="font-mono">{selectedExec.duration_ms}ms</p>
                </div>
                <div>
                  <span className="text-xs text-nucleus-muted">Category</span>
                  <p className={`font-mono category-${selectedExec.command.category}`}>{selectedExec.command.category}</p>
                </div>
              </div>
              {selectedExec.stdout && (
                <div>
                  <span className="text-xs text-nucleus-muted">Stdout</span>
                  <pre className="bg-nucleus-panel p-3 rounded mt-1 text-xs font-mono overflow-auto max-h-48">{selectedExec.stdout}</pre>
                </div>
              )}
              {selectedExec.stderr && (
                <div>
                  <span className="text-xs text-nucleus-muted">Stderr</span>
                  <pre className="bg-red-950/30 p-3 rounded mt-1 text-xs font-mono text-red-300 overflow-auto max-h-48">{selectedExec.stderr}</pre>
                </div>
              )}
            </div>
            <button onClick={() => setSelectedExec(null)} className="mt-4 px-4 py-2 text-sm border border-nucleus-border rounded hover:border-nucleus-accent transition-colors">
              Close
            </button>
          </div>
        </div>
      )}

      {/* Rollback Modal */}
      {rollbackTarget && (
        <RollbackModal
          execution={rollbackTarget}
          onClose={() => setRollbackTarget(null)}
          onRollbackComplete={() => {
            setRollbackTarget(null);
            // Refresh
          }}
        />
      )}
    </div>
  );
}
