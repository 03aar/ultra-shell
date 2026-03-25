'use client';

import { useState, useRef, useEffect } from 'react';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';
const API_KEY = process.env.NEXT_PUBLIC_API_KEY || 'dev-nucleus-key-local';

interface AgentStep {
  type: 'thinking' | 'tool_call' | 'result' | 'error';
  content: string;
  timestamp: string;
  toolName?: string;
  riskLevel?: string;
}

const providers = [
  { id: 'claude', name: 'Claude', models: ['claude-sonnet-4-5', 'claude-opus-4-5', 'claude-haiku-4-5'] },
  { id: 'openai', name: 'OpenAI', models: ['gpt-4o', 'gpt-4o-mini', 'o1', 'o3-mini'] },
  { id: 'gemini', name: 'Gemini', models: ['gemini-2.0-flash', 'gemini-1.5-pro'] },
  { id: 'ollama', name: 'Ollama', models: ['llama3', 'codellama', 'mistral', 'deepseek-coder'] },
];

const approvalLevels = ['none', 'low', 'medium', 'high'];

export default function AgentPage() {
  const [selectedProvider, setSelectedProvider] = useState('claude');
  const [selectedModel, setSelectedModel] = useState('claude-sonnet-4-5');
  const [goal, setGoal] = useState('');
  const [autoApprove, setAutoApprove] = useState('low');
  const [running, setRunning] = useState(false);
  const [steps, setSteps] = useState<AgentStep[]>([]);
  const [plan, setPlan] = useState<{steps: {command: string; rationale: string; risk_level: string}[]} | null>(null);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [steps]);

  const currentModels = providers.find(p => p.id === selectedProvider)?.models || [];

  useEffect(() => {
    if (currentModels.length > 0 && !currentModels.includes(selectedModel)) {
      setSelectedModel(currentModels[0]);
    }
  }, [selectedProvider]);

  const handlePlan = async () => {
    if (!goal.trim()) return;
    setRunning(true);
    setSteps([{ type: 'thinking', content: 'Generating execution plan...', timestamp: new Date().toISOString() }]);

    try {
      const res = await fetch(`${API_BASE}/agent/plan`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-Nucleus-Key': API_KEY },
        body: JSON.stringify({ goal, context: '', agent_id: selectedProvider }),
      });
      const json = await res.json();
      if (json.data) {
        setPlan(json.data);
        setSteps(prev => [...prev, {
          type: 'result',
          content: `Plan generated with ${json.data.steps?.length || 0} steps`,
          timestamp: new Date().toISOString(),
        }]);
      }
    } catch (e: unknown) {
      setSteps(prev => [...prev, {
        type: 'error',
        content: `Plan failed: ${e instanceof Error ? e.message : String(e)}`,
        timestamp: new Date().toISOString(),
      }]);
    } finally {
      setRunning(false);
    }
  };

  const handleExecute = async () => {
    if (!goal.trim()) return;
    setRunning(true);
    setPlan(null);
    setSteps([{ type: 'thinking', content: `Starting agent with ${selectedProvider}/${selectedModel}...`, timestamp: new Date().toISOString() }]);

    // Execute through the agent API
    try {
      const res = await fetch(`${API_BASE}/agent/execute`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-Nucleus-Key': API_KEY },
        body: JSON.stringify({ command: goal, agent_id: `${selectedProvider}:${selectedModel}`, dry_run: false }),
      });
      const json = await res.json();
      if (json.data) {
        setSteps(prev => [...prev,
          { type: 'tool_call', content: `Executed: ${goal}`, toolName: 'execute_command', timestamp: new Date().toISOString() },
          { type: 'result', content: json.data.stdout || 'Command completed', timestamp: new Date().toISOString() },
        ]);
      } else if (json.error) {
        setSteps(prev => [...prev, { type: 'error', content: json.error, timestamp: new Date().toISOString() }]);
      }
    } catch (e: unknown) {
      setSteps(prev => [...prev, { type: 'error', content: `Error: ${e instanceof Error ? e.message : String(e)}`, timestamp: new Date().toISOString() }]);
    } finally {
      setRunning(false);
    }
  };

  return (
    <div className="h-full flex flex-col">
      <div className="px-6 py-4 border-b border-nucleus-border">
        <h1 className="font-heading text-2xl font-bold">Agent Console</h1>
        <p className="text-sm text-nucleus-muted mt-1">AI agent orchestration with full context awareness</p>
      </div>

      <div className="flex-1 flex overflow-hidden">
        {/* Left: Config + Goal */}
        <div className="w-96 border-r border-nucleus-border overflow-auto">
          {/* Provider selector */}
          <div className="p-4 border-b border-nucleus-border">
            <h3 className="text-xs text-nucleus-muted uppercase tracking-wider mb-2">Provider</h3>
            <div className="grid grid-cols-2 gap-2">
              {providers.map(p => (
                <button
                  key={p.id}
                  onClick={() => setSelectedProvider(p.id)}
                  className={`px-3 py-2 text-xs rounded font-mono transition-colors ${
                    selectedProvider === p.id
                      ? 'bg-nucleus-accent/10 text-nucleus-accent border border-nucleus-accent/30'
                      : 'bg-nucleus-panel border border-nucleus-border text-nucleus-muted hover:text-nucleus-text'
                  }`}
                >
                  {p.name}
                </button>
              ))}
            </div>
          </div>

          {/* Model selector */}
          <div className="p-4 border-b border-nucleus-border">
            <h3 className="text-xs text-nucleus-muted uppercase tracking-wider mb-2">Model</h3>
            <select
              value={selectedModel}
              onChange={e => setSelectedModel(e.target.value)}
              className="w-full px-3 py-2 bg-nucleus-panel border border-nucleus-border rounded text-sm font-mono text-nucleus-text focus:border-nucleus-accent focus:outline-none"
            >
              {currentModels.map(m => (
                <option key={m} value={m}>{m}</option>
              ))}
            </select>
          </div>

          {/* Auto-approve */}
          <div className="p-4 border-b border-nucleus-border">
            <h3 className="text-xs text-nucleus-muted uppercase tracking-wider mb-2">Auto-Approve Up To</h3>
            <div className="flex gap-1">
              {approvalLevels.map(level => (
                <button
                  key={level}
                  onClick={() => setAutoApprove(level)}
                  className={`flex-1 px-2 py-1.5 text-xs rounded font-mono transition-colors ${
                    autoApprove === level
                      ? 'bg-nucleus-accent/10 text-nucleus-accent border border-nucleus-accent/30'
                      : 'bg-nucleus-panel border border-nucleus-border text-nucleus-muted'
                  }`}
                >
                  {level}
                </button>
              ))}
            </div>
          </div>

          {/* Goal input */}
          <div className="p-4">
            <h3 className="text-xs text-nucleus-muted uppercase tracking-wider mb-2">Goal</h3>
            <textarea
              value={goal}
              onChange={e => setGoal(e.target.value)}
              placeholder="e.g., Set up a new Node.js project with tests and Docker"
              className="w-full h-32 px-3 py-2 bg-nucleus-panel border border-nucleus-border rounded text-sm font-mono text-nucleus-text placeholder-nucleus-muted focus:border-nucleus-accent focus:outline-none resize-none"
            />
            <div className="flex gap-2 mt-3">
              <button
                onClick={handlePlan}
                disabled={running || !goal.trim()}
                className="flex-1 px-4 py-2 text-sm bg-nucleus-panel border border-nucleus-border rounded hover:border-nucleus-accent hover:text-nucleus-accent transition-colors disabled:opacity-50 font-mono"
              >
                Plan
              </button>
              <button
                onClick={handleExecute}
                disabled={running || !goal.trim()}
                className="flex-1 px-4 py-2 text-sm bg-nucleus-accent text-black font-bold rounded hover:bg-nucleus-accent/80 transition-colors disabled:opacity-50 font-mono"
              >
                {running ? 'Running...' : 'Execute'}
              </button>
            </div>
          </div>

          {/* Plan display */}
          {plan && plan.steps && (
            <div className="p-4 border-t border-nucleus-border">
              <h3 className="text-xs text-nucleus-muted uppercase tracking-wider mb-2">Plan</h3>
              <div className="space-y-2">
                {plan.steps.map((step, i) => (
                  <div key={i} className="bg-nucleus-panel p-3 rounded border border-nucleus-border/50">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-xs text-nucleus-muted font-mono">Step {i + 1}</span>
                      <span className={`text-xs px-1.5 py-0.5 rounded risk-${step.risk_level}`}
                        style={{ borderColor: 'currentColor', border: '1px solid' }}>
                        {step.risk_level}
                      </span>
                    </div>
                    <code className="text-xs font-mono text-nucleus-accent block">{step.command}</code>
                    <p className="text-xs text-nucleus-muted mt-1">{step.rationale}</p>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Right: Execution output */}
        <div className="flex-1 flex flex-col overflow-hidden">
          <div className="px-4 py-3 border-b border-nucleus-border flex items-center justify-between">
            <h2 className="text-sm font-heading font-semibold text-nucleus-muted uppercase tracking-wider">
              Execution Output
            </h2>
            {running && (
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 rounded-full bg-nucleus-accent animate-pulse" />
                <span className="text-xs text-nucleus-accent font-mono">Running</span>
              </div>
            )}
          </div>

          <div className="flex-1 overflow-auto p-4 space-y-3">
            {steps.length === 0 ? (
              <div className="flex items-center justify-center h-full text-nucleus-muted text-sm">
                Enter a goal and click Execute or Plan to start
              </div>
            ) : (
              steps.map((step, i) => (
                <div key={i} className={`rounded-lg border p-3 ${
                  step.type === 'thinking' ? 'border-blue-800/50 bg-blue-950/20' :
                  step.type === 'tool_call' ? 'border-nucleus-accent/30 bg-nucleus-accent/5' :
                  step.type === 'error' ? 'border-red-800/50 bg-red-950/20' :
                  'border-nucleus-border bg-nucleus-panel'
                }`}>
                  <div className="flex items-center gap-2 mb-1">
                    <span className={`text-[10px] uppercase font-mono tracking-wider ${
                      step.type === 'thinking' ? 'text-blue-400' :
                      step.type === 'tool_call' ? 'text-nucleus-accent' :
                      step.type === 'error' ? 'text-red-400' :
                      'text-nucleus-muted'
                    }`}>
                      {step.type === 'tool_call' ? `Tool: ${step.toolName}` : step.type}
                    </span>
                    <span className="text-[10px] text-nucleus-muted font-mono">
                      {new Date(step.timestamp).toLocaleTimeString()}
                    </span>
                  </div>
                  <pre className="text-xs font-mono whitespace-pre-wrap text-nucleus-text">{step.content}</pre>
                </div>
              ))
            )}
            <div ref={bottomRef} />
          </div>
        </div>
      </div>
    </div>
  );
}
