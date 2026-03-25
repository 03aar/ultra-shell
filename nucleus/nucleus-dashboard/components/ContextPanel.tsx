'use client';

import { useEffect, useState } from 'react';
import { ContextState, getContext } from '@/lib/api';

export default function ContextPanel() {
  const [ctx, setCtx] = useState<ContextState | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchContext = async () => {
      try {
        const data = await getContext();
        setCtx(data);
      } catch (e) {
        console.error('Failed to fetch context:', e);
      } finally {
        setLoading(false);
      }
    };
    fetchContext();
    const interval = setInterval(fetchContext, 5000);
    return () => clearInterval(interval);
  }, []);

  if (loading) return <div className="text-nucleus-muted text-sm p-4">Loading context...</div>;
  if (!ctx) return <div className="text-red-400 text-sm p-4">Failed to load context</div>;

  const diskPercent = ctx.disk_usage.percent.toFixed(1);

  return (
    <div className="space-y-4">
      {/* Working Directory */}
      <div>
        <h3 className="text-xs text-nucleus-muted uppercase tracking-wider mb-1">CWD</h3>
        <p className="font-mono text-sm text-nucleus-accent">{ctx.cwd}</p>
      </div>

      {/* Git Info */}
      {ctx.git_branch && (
        <div>
          <h3 className="text-xs text-nucleus-muted uppercase tracking-wider mb-1">Git</h3>
          <p className="font-mono text-sm">
            <span className="text-green-400">{ctx.git_branch}</span>
          </p>
          {ctx.git_status && (
            <pre className="text-xs text-nucleus-muted mt-1 font-mono">{ctx.git_status}</pre>
          )}
        </div>
      )}

      {/* Disk Usage */}
      <div>
        <h3 className="text-xs text-nucleus-muted uppercase tracking-wider mb-1">Disk</h3>
        <div className="w-full h-2 bg-nucleus-panel rounded-full overflow-hidden">
          <div
            className="h-full rounded-full transition-all"
            style={{
              width: `${diskPercent}%`,
              backgroundColor: ctx.disk_usage.percent > 90 ? '#ef4444' : ctx.disk_usage.percent > 70 ? '#fbbf24' : '#00ff88',
            }}
          />
        </div>
        <p className="text-xs text-nucleus-muted mt-1 font-mono">{diskPercent}% used</p>
      </div>

      {/* Top Processes */}
      <div>
        <h3 className="text-xs text-nucleus-muted uppercase tracking-wider mb-1">
          Processes ({ctx.running_processes.length})
        </h3>
        <div className="space-y-1 max-h-40 overflow-auto">
          {ctx.running_processes.slice(0, 8).map((proc) => (
            <div key={proc.pid} className="flex items-center gap-2 text-xs font-mono">
              <span className="text-nucleus-muted w-12">{proc.pid}</span>
              <span className="text-nucleus-text flex-1 truncate">{proc.name}</span>
              <span className="text-blue-400 w-12 text-right">{proc.cpu.toFixed(1)}%</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
