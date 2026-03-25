'use client';

import { useState } from 'react';
import ExecutionGraph from '@/components/ExecutionGraph';
import { ExecutionNode } from '@/lib/api';

export default function GraphPage() {
  const [selectedNode, setSelectedNode] = useState<ExecutionNode | null>(null);

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="px-6 py-4 border-b border-nucleus-border flex items-center justify-between">
        <h1 className="font-heading text-2xl font-bold">Execution Graph</h1>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            {[
              { label: 'Filesystem', color: '#fbbf24' },
              { label: 'Network', color: '#60a5fa' },
              { label: 'Git', color: '#4ade80' },
              { label: 'Build', color: '#a78bfa' },
              { label: 'Error', color: '#ef4444' },
            ].map((item) => (
              <div key={item.label} className="flex items-center gap-1">
                <div className="w-3 h-3 rounded-full" style={{ backgroundColor: item.color }} />
                <span className="text-xs text-nucleus-muted">{item.label}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Graph */}
      <div className="flex-1 flex">
        <div className={`flex-1 ${selectedNode ? 'w-2/3' : 'w-full'}`}>
          <ExecutionGraph onNodeClick={setSelectedNode} />
        </div>

        {/* Detail Panel */}
        {selectedNode && (
          <div className="w-1/3 border-l border-nucleus-border overflow-auto bg-nucleus-surface">
            <div className="p-4 border-b border-nucleus-border flex items-center justify-between">
              <h2 className="font-heading text-sm font-semibold uppercase tracking-wider text-nucleus-muted">
                Node Detail
              </h2>
              <button
                onClick={() => setSelectedNode(null)}
                className="text-nucleus-muted hover:text-nucleus-text text-lg"
              >
                x
              </button>
            </div>
            <div className="p-4 space-y-4">
              <div>
                <span className="text-xs text-nucleus-muted">ID</span>
                <p className="font-mono text-xs text-nucleus-accent">{selectedNode.id}</p>
              </div>
              <div>
                <span className="text-xs text-nucleus-muted">Command</span>
                <code className="block font-mono text-sm bg-nucleus-panel p-2 rounded mt-1">
                  {selectedNode.command.raw}
                </code>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <span className="text-xs text-nucleus-muted">Exit Code</span>
                  <p className={`font-mono text-lg ${selectedNode.exit_code === 0 ? 'text-green-400' : 'text-red-400'}`}>
                    {selectedNode.exit_code}
                  </p>
                </div>
                <div>
                  <span className="text-xs text-nucleus-muted">Duration</span>
                  <p className="font-mono text-lg">{selectedNode.duration_ms}ms</p>
                </div>
              </div>
              <div>
                <span className="text-xs text-nucleus-muted">Category</span>
                <p className={`font-mono text-sm category-${selectedNode.command.category}`}>
                  {selectedNode.command.category}
                </p>
              </div>
              <div>
                <span className="text-xs text-nucleus-muted">Timestamp</span>
                <p className="font-mono text-xs">{new Date(selectedNode.timestamp).toLocaleString()}</p>
              </div>
              {selectedNode.risk_flags.length > 0 && (
                <div>
                  <span className="text-xs text-nucleus-muted">Risk Flags</span>
                  <div className="space-y-1 mt-1">
                    {selectedNode.risk_flags.map((flag, i) => (
                      <div key={i} className={`text-xs font-mono risk-${flag.level} bg-nucleus-panel p-2 rounded`}>
                        [{flag.level.toUpperCase()}] {flag.message}
                      </div>
                    ))}
                  </div>
                </div>
              )}
              {selectedNode.files_mutated.length > 0 && (
                <div>
                  <span className="text-xs text-nucleus-muted">Files Mutated</span>
                  <ul className="mt-1 space-y-0.5">
                    {selectedNode.files_mutated.map((f, i) => (
                      <li key={i} className="text-xs font-mono text-yellow-400">{f}</li>
                    ))}
                  </ul>
                </div>
              )}
              {selectedNode.stdout && (
                <div>
                  <span className="text-xs text-nucleus-muted">Output</span>
                  <pre className="bg-nucleus-panel p-2 rounded mt-1 text-xs font-mono overflow-auto max-h-48">
                    {selectedNode.stdout}
                  </pre>
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
