'use client';

import { useEffect, useState } from 'react';
import { ContextState, ExecutionNode, getContext, getExecutions } from '@/lib/api';

export default function ContextPage() {
  const [ctx, setCtx] = useState<ContextState | null>(null);
  const [executions, setExecutions] = useState<ExecutionNode[]>([]);
  const [loading, setLoading] = useState(true);
  const [envFilter, setEnvFilter] = useState('');

  useEffect(() => {
    const fetchAll = async () => {
      try {
        const [ctxData, execData] = await Promise.all([
          getContext(),
          getExecutions({ limit: 100, command_category: 'filesystem_mutation' }),
        ]);
        setCtx(ctxData);
        setExecutions(execData.executions || []);
      } catch (e) {
        console.error('Failed to fetch context:', e);
      } finally {
        setLoading(false);
      }
    };
    fetchAll();
    const interval = setInterval(fetchAll, 5000);
    return () => clearInterval(interval);
  }, []);

  if (loading) {
    return (
      <div className="h-full flex items-center justify-center text-nucleus-muted">
        Loading environment state...
      </div>
    );
  }

  if (!ctx) {
    return (
      <div className="h-full flex items-center justify-center text-red-400">
        Failed to load environment state
      </div>
    );
  }

  const filteredEnvVars = Object.entries(ctx.env_vars).filter(
    ([key]) => !envFilter || key.toLowerCase().includes(envFilter.toLowerCase())
  );

  const mutatedFiles = new Map<string, ExecutionNode[]>();
  for (const exec of executions) {
    for (const file of exec.files_mutated) {
      const existing = mutatedFiles.get(file) || [];
      existing.push(exec);
      mutatedFiles.set(file, existing);
    }
  }

  return (
    <div className="h-full flex flex-col">
      <div className="px-6 py-4 border-b border-nucleus-border">
        <h1 className="font-heading text-2xl font-bold">Environment State</h1>
      </div>

      <div className="flex-1 overflow-auto p-6 space-y-6">
        {/* System Info */}
        <div className="grid grid-cols-3 gap-4">
          <div className="bg-nucleus-surface border border-nucleus-border rounded-lg p-4">
            <h3 className="text-xs text-nucleus-muted uppercase tracking-wider mb-2">Working Directory</h3>
            <p className="font-mono text-sm text-nucleus-accent">{ctx.cwd}</p>
          </div>
          <div className="bg-nucleus-surface border border-nucleus-border rounded-lg p-4">
            <h3 className="text-xs text-nucleus-muted uppercase tracking-wider mb-2">Git Branch</h3>
            <p className="font-mono text-sm text-green-400">{ctx.git_branch || 'N/A'}</p>
            {ctx.git_status && (
              <pre className="text-xs text-nucleus-muted font-mono mt-2 max-h-20 overflow-auto">{ctx.git_status}</pre>
            )}
          </div>
          <div className="bg-nucleus-surface border border-nucleus-border rounded-lg p-4">
            <h3 className="text-xs text-nucleus-muted uppercase tracking-wider mb-2">Disk Usage</h3>
            <div className="w-full h-3 bg-nucleus-panel rounded-full overflow-hidden mb-2">
              <div
                className="h-full rounded-full"
                style={{
                  width: `${ctx.disk_usage.percent}%`,
                  backgroundColor: ctx.disk_usage.percent > 90 ? '#ef4444' : ctx.disk_usage.percent > 70 ? '#fbbf24' : '#00ff88',
                }}
              />
            </div>
            <p className="font-mono text-xs text-nucleus-muted">
              {(ctx.disk_usage.used / 1e9).toFixed(1)}GB / {(ctx.disk_usage.total / 1e9).toFixed(1)}GB ({ctx.disk_usage.percent.toFixed(1)}%)
            </p>
          </div>
        </div>

        {/* Running Processes */}
        <div className="bg-nucleus-surface border border-nucleus-border rounded-lg">
          <div className="px-4 py-3 border-b border-nucleus-border">
            <h3 className="text-xs text-nucleus-muted uppercase tracking-wider">
              Running Processes ({ctx.running_processes.length})
            </h3>
          </div>
          <div className="overflow-auto max-h-64">
            <table className="w-full text-xs font-mono">
              <thead>
                <tr className="border-b border-nucleus-border text-nucleus-muted">
                  <th className="text-left p-2 w-20">PID</th>
                  <th className="text-left p-2">Name</th>
                  <th className="text-right p-2 w-20">CPU %</th>
                  <th className="text-right p-2 w-20">MEM %</th>
                  <th className="text-left p-2">Command</th>
                </tr>
              </thead>
              <tbody>
                {ctx.running_processes.map((proc) => (
                  <tr key={proc.pid} className="border-b border-nucleus-border/50 hover:bg-nucleus-panel/50">
                    <td className="p-2 text-nucleus-muted">{proc.pid}</td>
                    <td className="p-2 text-nucleus-text">{proc.name}</td>
                    <td className="p-2 text-right text-blue-400">{proc.cpu.toFixed(1)}</td>
                    <td className="p-2 text-right text-purple-400">{proc.memory.toFixed(1)}</td>
                    <td className="p-2 text-nucleus-muted truncate max-w-xs">{proc.command}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Environment Variables */}
        <div className="bg-nucleus-surface border border-nucleus-border rounded-lg">
          <div className="px-4 py-3 border-b border-nucleus-border flex items-center justify-between">
            <h3 className="text-xs text-nucleus-muted uppercase tracking-wider">
              Environment Variables ({filteredEnvVars.length})
            </h3>
            <input
              type="text"
              placeholder="Filter..."
              value={envFilter}
              onChange={(e) => setEnvFilter(e.target.value)}
              className="px-2 py-1 text-xs bg-nucleus-panel border border-nucleus-border rounded font-mono text-nucleus-text placeholder-nucleus-muted focus:border-nucleus-accent focus:outline-none"
            />
          </div>
          <div className="overflow-auto max-h-64">
            {filteredEnvVars.map(([key, value]) => (
              <div key={key} className="flex px-4 py-1.5 border-b border-nucleus-border/30 text-xs font-mono hover:bg-nucleus-panel/50">
                <span className="text-nucleus-accent w-48 shrink-0 truncate">{key}</span>
                <span className={`text-nucleus-muted truncate ${value === '***MASKED***' ? 'text-red-400' : ''}`}>
                  {value}
                </span>
              </div>
            ))}
          </div>
        </div>

        {/* File Mutation Tracker */}
        <div className="bg-nucleus-surface border border-nucleus-border rounded-lg">
          <div className="px-4 py-3 border-b border-nucleus-border">
            <h3 className="text-xs text-nucleus-muted uppercase tracking-wider">
              File Mutations This Session ({mutatedFiles.size} files)
            </h3>
          </div>
          <div className="overflow-auto max-h-64">
            {mutatedFiles.size === 0 ? (
              <div className="p-4 text-center text-nucleus-muted text-xs">No file mutations recorded</div>
            ) : (
              Array.from(mutatedFiles.entries()).map(([file, execs]) => (
                <div key={file} className="px-4 py-2 border-b border-nucleus-border/30 hover:bg-nucleus-panel/50">
                  <p className="font-mono text-xs text-yellow-400">{file}</p>
                  <p className="text-[10px] text-nucleus-muted mt-0.5">
                    Modified by {execs.length} command(s): {execs.map((e) => e.command.binary).join(', ')}
                  </p>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
