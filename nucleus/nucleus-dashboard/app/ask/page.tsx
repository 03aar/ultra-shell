'use client';

import { useState, useRef, useEffect } from 'react';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';
const API_KEY = process.env.NEXT_PUBLIC_API_KEY || 'dev-nucleus-key-local';

interface AskResult {
  command: string;
  explanation: string;
  risk_level: string;
  executed?: boolean;
  output?: string;
}

interface HistoryEntry {
  question: string;
  result: AskResult;
  timestamp: string;
}

const examples = [
  'What files did I change today?',
  'Why did my last deployment fail?',
  'How do I set up a new Python project?',
  'What process is using port 8080?',
  'Show me all large files over 100MB',
  'Who committed last in this repo?',
  'What are the listening ports?',
  'How much disk space is left?',
];

export default function AskPage() {
  const [query, setQuery] = useState('');
  const [loading, setLoading] = useState(false);
  const [currentResult, setCurrentResult] = useState<AskResult | null>(null);
  const [history, setHistory] = useState<HistoryEntry[]>([]);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  const handleAsk = async (q?: string) => {
    const question = q || query;
    if (!question.trim()) return;
    setLoading(true);
    setCurrentResult(null);

    try {
      const res = await fetch(`${API_BASE}/natural/translate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-Nucleus-Key': API_KEY },
        body: JSON.stringify({ input: question }),
      });
      const json = await res.json();
      if (json.data) {
        const result: AskResult = {
          command: json.data.command,
          explanation: json.data.explanation,
          risk_level: json.data.risk_level,
        };
        setCurrentResult(result);
        setHistory(prev => [{ question, result, timestamp: new Date().toISOString() }, ...prev.slice(0, 49)]);
      }
    } catch {
      setCurrentResult({ command: 'Error connecting to API', explanation: 'Check that nucleus-api is running', risk_level: 'none' });
    } finally {
      setLoading(false);
      setQuery('');
    }
  };

  const handleExecute = async () => {
    if (!currentResult) return;
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/agent/execute`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-Nucleus-Key': API_KEY },
        body: JSON.stringify({ command: currentResult.command, dry_run: false }),
      });
      const json = await res.json();
      setCurrentResult(prev => prev ? { ...prev, executed: true, output: json.data?.stdout || 'Command executed' } : null);
    } catch {
      setCurrentResult(prev => prev ? { ...prev, executed: true, output: 'Execution failed' } : null);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="h-full flex flex-col">
      <div className="px-6 py-4 border-b border-nucleus-border">
        <h1 className="font-heading text-2xl font-bold">Ask Nucleus</h1>
        <p className="text-sm text-nucleus-muted mt-1">Natural language interface — ask anything about your system</p>
      </div>

      <div className="flex-1 flex overflow-hidden">
        {/* Left: History */}
        <div className="w-72 border-r border-nucleus-border overflow-auto">
          <div className="p-3 border-b border-nucleus-border">
            <h3 className="text-xs text-nucleus-muted uppercase tracking-wider">History</h3>
          </div>
          {history.length === 0 ? (
            <div className="p-4 text-xs text-nucleus-muted text-center">No questions yet</div>
          ) : (
            history.map((entry, i) => (
              <div
                key={i}
                className="p-3 border-b border-nucleus-border/50 cursor-pointer hover:bg-nucleus-panel/50 transition-colors"
                onClick={() => { setCurrentResult(entry.result); }}
              >
                <p className="text-xs text-nucleus-text truncate">{entry.question}</p>
                <p className="text-[10px] text-nucleus-muted mt-1 font-mono">{new Date(entry.timestamp).toLocaleTimeString()}</p>
              </div>
            ))
          )}
        </div>

        {/* Main */}
        <div className="flex-1 flex flex-col">
          {/* Search input */}
          <div className="p-6 border-b border-nucleus-border">
            <div className="max-w-2xl mx-auto">
              <div className="relative">
                <input
                  ref={inputRef}
                  type="text"
                  value={query}
                  onChange={e => setQuery(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && handleAsk()}
                  placeholder="Ask Nucleus anything..."
                  className="w-full px-5 py-4 bg-nucleus-surface border border-nucleus-border rounded-xl text-lg font-body text-nucleus-text placeholder-nucleus-muted focus:border-nucleus-accent focus:outline-none transition-colors"
                />
                <button
                  onClick={() => handleAsk()}
                  disabled={loading || !query.trim()}
                  className="absolute right-3 top-1/2 -translate-y-1/2 px-4 py-2 bg-nucleus-accent text-black font-bold rounded-lg text-sm hover:bg-nucleus-accent/80 transition-colors disabled:opacity-50"
                >
                  {loading ? '...' : 'Ask'}
                </button>
              </div>

              {/* Examples */}
              {!currentResult && (
                <div className="mt-6 grid grid-cols-2 gap-2">
                  {examples.map((ex, i) => (
                    <button
                      key={i}
                      onClick={() => { setQuery(ex); handleAsk(ex); }}
                      className="text-left px-3 py-2 bg-nucleus-panel border border-nucleus-border/50 rounded-lg text-xs text-nucleus-muted hover:text-nucleus-text hover:border-nucleus-accent/30 transition-colors"
                    >
                      {ex}
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Result */}
          {currentResult && (
            <div className="flex-1 overflow-auto p-6">
              <div className="max-w-2xl mx-auto space-y-4">
                {/* Translated command */}
                <div className="bg-nucleus-surface border border-nucleus-border rounded-xl p-5">
                  <div className="flex items-center justify-between mb-3">
                    <h3 className="text-xs text-nucleus-muted uppercase tracking-wider">Command</h3>
                    <span className={`text-xs px-2 py-0.5 rounded-full border risk-${currentResult.risk_level}`} style={{ borderColor: 'currentColor' }}>
                      {currentResult.risk_level}
                    </span>
                  </div>
                  <code className="block font-mono text-nucleus-accent text-sm bg-nucleus-panel p-3 rounded-lg">
                    {currentResult.command}
                  </code>
                  <p className="text-sm text-nucleus-muted mt-3">{currentResult.explanation}</p>

                  {!currentResult.executed && (
                    <button
                      onClick={handleExecute}
                      disabled={loading}
                      className="mt-4 px-4 py-2 bg-nucleus-accent text-black font-bold rounded-lg text-sm hover:bg-nucleus-accent/80 transition-colors disabled:opacity-50"
                    >
                      {loading ? 'Executing...' : 'Execute This Command'}
                    </button>
                  )}
                </div>

                {/* Execution output */}
                {currentResult.executed && currentResult.output && (
                  <div className="bg-nucleus-surface border border-nucleus-border rounded-xl p-5">
                    <h3 className="text-xs text-nucleus-muted uppercase tracking-wider mb-3">Output</h3>
                    <pre className="font-mono text-xs text-nucleus-text bg-nucleus-panel p-3 rounded-lg overflow-auto max-h-64">
                      {currentResult.output}
                    </pre>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
