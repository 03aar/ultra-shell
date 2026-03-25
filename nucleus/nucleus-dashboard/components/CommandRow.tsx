'use client';

import { ExecutionNode } from '@/lib/api';

interface CommandRowProps {
  execution: ExecutionNode;
  onRollback?: (id: string) => void;
  onClick?: (exec: ExecutionNode) => void;
}

export default function CommandRow({ execution, onRollback, onClick }: CommandRowProps) {
  const time = new Date(execution.timestamp).toLocaleTimeString();
  const exitColor = execution.exit_code === 0 ? 'text-green-400' : 'text-red-400';
  const maxRisk = execution.risk_flags.length > 0
    ? execution.risk_flags.reduce((max, f) => {
        const levels = ['none', 'low', 'medium', 'high', 'critical'];
        return levels.indexOf(f.level) > levels.indexOf(max) ? f.level : max;
      }, 'none')
    : 'none';

  return (
    <div
      className="flex items-center gap-3 px-4 py-2 border-b border-nucleus-border hover:bg-nucleus-panel/50 cursor-pointer transition-colors group"
      onClick={() => onClick?.(execution)}
    >
      {/* Timestamp */}
      <span className="text-xs text-nucleus-muted font-mono w-20 shrink-0">{time}</span>

      {/* Command */}
      <div className="flex-1 min-w-0">
        <span className="font-mono text-sm text-nucleus-text truncate block">
          {execution.rolled_back && (
            <span className="text-yellow-500 mr-2">[ROLLED BACK]</span>
          )}
          <span className={`category-${execution.command.category}`}>
            {execution.command.binary}
          </span>{' '}
          <span className="text-nucleus-muted">
            {execution.command.args.join(' ')} {execution.command.flags.join(' ')}
          </span>
        </span>
      </div>

      {/* Exit Code */}
      <span className={`font-mono text-xs ${exitColor} w-8 text-center shrink-0`}>
        {execution.exit_code}
      </span>

      {/* Risk Badge */}
      {maxRisk !== 'none' && (
        <span className={`text-xs px-2 py-0.5 rounded-full border risk-${maxRisk} shrink-0`}
          style={{ borderColor: 'currentColor', opacity: 0.8 }}>
          {maxRisk}
        </span>
      )}

      {/* Duration */}
      <span className="text-xs text-nucleus-muted font-mono w-16 text-right shrink-0">
        {execution.duration_ms}ms
      </span>

      {/* Rollback Button */}
      {execution.rollback_available && !execution.rolled_back && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onRollback?.(execution.id);
          }}
          className="text-xs px-2 py-1 bg-nucleus-panel border border-nucleus-border rounded hover:border-nucleus-accent hover:text-nucleus-accent transition-colors opacity-0 group-hover:opacity-100 shrink-0"
        >
          Rollback
        </button>
      )}
    </div>
  );
}
