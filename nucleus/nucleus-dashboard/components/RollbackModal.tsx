'use client';

import { ExecutionNode, rollback } from '@/lib/api';
import { useState } from 'react';

interface RollbackModalProps {
  execution: ExecutionNode;
  onClose: () => void;
  onRollbackComplete: () => void;
}

export default function RollbackModal({ execution, onClose, onRollbackComplete }: RollbackModalProps) {
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<{ success: boolean; message: string } | null>(null);

  const handleRollback = async () => {
    setLoading(true);
    try {
      const res = await rollback(execution.id);
      setResult({ success: res.success, message: res.message });
      if (res.success) {
        setTimeout(onRollbackComplete, 1500);
      }
    } catch (e: any) {
      setResult({ success: false, message: e.message || 'Rollback failed' });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50" onClick={onClose}>
      <div
        className="bg-nucleus-surface border border-nucleus-border rounded-lg p-6 max-w-md w-full mx-4"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="font-heading text-lg font-bold text-nucleus-text mb-4">Confirm Rollback</h2>

        <div className="mb-4">
          <p className="text-sm text-nucleus-muted mb-2">Rolling back execution:</p>
          <code className="text-sm font-mono text-nucleus-accent block bg-nucleus-panel p-2 rounded">
            {execution.command.raw}
          </code>
        </div>

        {execution.files_mutated.length > 0 && (
          <div className="mb-4">
            <p className="text-sm text-nucleus-muted mb-1">Files to restore:</p>
            <ul className="text-xs font-mono text-yellow-400 space-y-0.5">
              {execution.files_mutated.map((f, i) => (
                <li key={i}>{f}</li>
              ))}
            </ul>
          </div>
        )}

        {result && (
          <div className={`mb-4 p-3 rounded text-sm ${result.success ? 'bg-green-900/30 text-green-400' : 'bg-red-900/30 text-red-400'}`}>
            {result.message}
          </div>
        )}

        <div className="flex gap-3 justify-end">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm text-nucleus-muted hover:text-nucleus-text border border-nucleus-border rounded transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleRollback}
            disabled={loading || result?.success === true}
            className="px-4 py-2 text-sm bg-red-600 hover:bg-red-500 text-white rounded transition-colors disabled:opacity-50"
          >
            {loading ? 'Rolling back...' : 'Rollback'}
          </button>
        </div>
      </div>
    </div>
  );
}
