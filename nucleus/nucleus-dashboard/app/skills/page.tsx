'use client';

import { useEffect, useState } from 'react';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';
const API_KEY = process.env.NEXT_PUBLIC_API_KEY || 'dev-nucleus-key-local';

interface Skill {
  name: string;
  description: string;
  parameters: { name: string; type: string; required: boolean; default?: string; description?: string }[];
  source: string;
}

interface SkillRunResult {
  success: boolean;
  steps: { command: string; description: string; exit_code: number; output: string }[];
  message: string;
}

export default function SkillsPage() {
  const [skills, setSkills] = useState<Skill[]>([]);
  const [selected, setSelected] = useState<Skill | null>(null);
  const [params, setParams] = useState<Record<string, string>>({});
  const [running, setRunning] = useState(false);
  const [result, setResult] = useState<SkillRunResult | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch(`${API_BASE}/skills`, { headers: { 'X-Nucleus-Key': API_KEY } })
      .then(res => res.json())
      .then(json => setSkills(json.data || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const handleRun = async () => {
    if (!selected) return;
    setRunning(true);
    setResult(null);
    try {
      const res = await fetch(`${API_BASE}/skills/${selected.name}/run`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-Nucleus-Key': API_KEY },
        body: JSON.stringify({ params }),
      });
      const json = await res.json();
      setResult(json.data);
    } catch {
      setResult({ success: false, steps: [], message: 'Failed to run skill' });
    } finally {
      setRunning(false);
    }
  };

  const categoryIcons: Record<string, string> = {
    git_cleanup: '🔀', docker_cleanup: '🐳', project_setup: '📁',
    deploy_check: '🚀', env_audit: '🔒', process_debug: '🔍',
  };

  return (
    <div className="h-full flex flex-col">
      <div className="px-6 py-4 border-b border-nucleus-border">
        <h1 className="font-heading text-2xl font-bold">Skills Library</h1>
        <p className="text-sm text-nucleus-muted mt-1">Reusable command workflows with safety checks</p>
      </div>

      <div className="flex-1 flex overflow-hidden">
        {/* Skills grid */}
        <div className="flex-1 overflow-auto p-6">
          {loading ? (
            <div className="text-nucleus-muted text-center py-12">Loading skills...</div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 max-w-5xl">
              {skills.map(skill => (
                <div
                  key={skill.name}
                  onClick={() => { setSelected(skill); setParams({}); setResult(null); }}
                  className={`bg-nucleus-surface border rounded-xl p-5 cursor-pointer transition-all hover:border-nucleus-accent/50 ${
                    selected?.name === skill.name ? 'border-nucleus-accent' : 'border-nucleus-border'
                  }`}
                >
                  <div className="flex items-center gap-3 mb-3">
                    <span className="text-2xl">{categoryIcons[skill.name] || '⚡'}</span>
                    <div>
                      <h3 className="font-mono text-sm font-semibold text-nucleus-text">{skill.name}</h3>
                      <span className="text-[10px] text-nucleus-muted uppercase">{skill.source}</span>
                    </div>
                  </div>
                  <p className="text-xs text-nucleus-muted leading-relaxed">{skill.description}</p>
                  {skill.parameters && skill.parameters.length > 0 && (
                    <div className="mt-3 flex gap-1 flex-wrap">
                      {skill.parameters.map(p => (
                        <span key={p.name} className="text-[10px] px-1.5 py-0.5 bg-nucleus-panel rounded text-nucleus-muted font-mono">
                          {p.name}{p.required ? '*' : ''}
                        </span>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Detail panel */}
        {selected && (
          <div className="w-96 border-l border-nucleus-border overflow-auto bg-nucleus-surface">
            <div className="p-4 border-b border-nucleus-border">
              <h2 className="font-heading text-lg font-bold">{selected.name}</h2>
              <p className="text-xs text-nucleus-muted mt-1">{selected.description}</p>
            </div>

            {/* Parameters form */}
            {selected.parameters && selected.parameters.length > 0 && (
              <div className="p-4 border-b border-nucleus-border space-y-3">
                <h3 className="text-xs text-nucleus-muted uppercase tracking-wider">Parameters</h3>
                {selected.parameters.map(p => (
                  <div key={p.name}>
                    <label className="text-xs text-nucleus-muted font-mono block mb-1">
                      {p.name} {p.required && <span className="text-red-400">*</span>}
                      {p.description && <span className="ml-1 text-nucleus-muted/50">— {p.description}</span>}
                    </label>
                    <input
                      type="text"
                      value={params[p.name] || p.default || ''}
                      onChange={e => setParams({ ...params, [p.name]: e.target.value })}
                      placeholder={p.default || `Enter ${p.name}`}
                      className="w-full px-3 py-2 bg-nucleus-panel border border-nucleus-border rounded text-sm font-mono text-nucleus-text placeholder-nucleus-muted focus:border-nucleus-accent focus:outline-none"
                    />
                  </div>
                ))}
              </div>
            )}

            <div className="p-4 border-b border-nucleus-border">
              <button
                onClick={handleRun}
                disabled={running}
                className="w-full px-4 py-2.5 bg-nucleus-accent text-black font-bold rounded-lg text-sm hover:bg-nucleus-accent/80 transition-colors disabled:opacity-50"
              >
                {running ? 'Running...' : `Run ${selected.name}`}
              </button>
            </div>

            {/* Result */}
            {result && (
              <div className="p-4 space-y-3">
                <div className={`px-3 py-2 rounded text-sm ${result.success ? 'bg-green-900/20 text-green-400' : 'bg-red-900/20 text-red-400'}`}>
                  {result.message}
                </div>
                {result.steps.map((step, i) => (
                  <div key={i} className="bg-nucleus-panel rounded p-3">
                    <div className="flex items-center gap-2 mb-1">
                      <span className={`text-xs font-mono ${step.exit_code === 0 ? 'text-green-400' : 'text-red-400'}`}>
                        [{step.exit_code}]
                      </span>
                      <span className="text-xs text-nucleus-muted">{step.description}</span>
                    </div>
                    <code className="text-xs font-mono text-nucleus-accent block">{step.command}</code>
                    {step.output && (
                      <pre className="text-[10px] font-mono text-nucleus-muted mt-1 overflow-auto max-h-20">{step.output}</pre>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
