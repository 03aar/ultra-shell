'use client';

import { useState } from 'react';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

interface Endpoint {
  method: string;
  path: string;
  description: string;
  params?: { name: string; type: string; description: string }[];
  body?: string;
}

const endpoints: Endpoint[] = [
  {
    method: 'GET',
    path: '/executions',
    description: 'List recent command executions',
    params: [
      { name: 'limit', type: 'number', description: 'Max results (default 50)' },
      { name: 'offset', type: 'number', description: 'Skip N results' },
      { name: 'session_id', type: 'string', description: 'Filter by session' },
      { name: 'risk_level', type: 'string', description: 'Filter by risk level' },
      { name: 'command_category', type: 'string', description: 'Filter by category' },
    ],
  },
  { method: 'GET', path: '/executions/:id', description: 'Get full execution details' },
  { method: 'GET', path: '/executions/:id/context', description: 'Get execution with dependency context' },
  { method: 'GET', path: '/context', description: 'Get current environment state' },
  { method: 'GET', path: '/context/graph', description: 'Get full session DAG', params: [{ name: 'session_id', type: 'string', description: 'Filter by session' }] },
  { method: 'POST', path: '/rollback/:execution_id', description: 'Rollback an execution' },
  { method: 'GET', path: '/sessions', description: 'List all sessions' },
  { method: 'POST', path: '/sessions', description: 'Create a new session', body: '{ "name": "my-session" }' },
  { method: 'GET', path: '/sessions/:id/replay', description: 'Get session replay data' },
  { method: 'WS', path: '/ws/stream', description: 'WebSocket live event stream' },
];

const methodColors: Record<string, string> = {
  GET: 'text-green-400 bg-green-400/10',
  POST: 'text-blue-400 bg-blue-400/10',
  WS: 'text-purple-400 bg-purple-400/10',
};

export default function APIExplorerPage() {
  const [selectedEndpoint, setSelectedEndpoint] = useState<Endpoint | null>(null);
  const [response, setResponse] = useState<string>('');
  const [loading, setLoading] = useState(false);
  const [pathParams, setPathParams] = useState<Record<string, string>>({});
  const [wsConnected, setWsConnected] = useState(false);
  const [wsMessages, setWsMessages] = useState<string[]>([]);
  const [wsRef, setWsRef] = useState<WebSocket | null>(null);

  const executeRequest = async (endpoint: Endpoint) => {
    setLoading(true);
    setResponse('');

    try {
      let path = endpoint.path;
      for (const [key, value] of Object.entries(pathParams)) {
        path = path.replace(`:${key}`, value);
      }

      if (endpoint.method === 'WS') {
        const wsUrl = `ws://localhost:8080/api/v1${path}?api_key=dev-nucleus-key-local`;
        const ws = new WebSocket(wsUrl);
        ws.onopen = () => {
          setWsConnected(true);
          setWsMessages((prev) => [...prev, '--- Connected ---']);
        };
        ws.onmessage = (event) => {
          try {
            const formatted = JSON.stringify(JSON.parse(event.data), null, 2);
            setWsMessages((prev) => [...prev.slice(-50), formatted]);
          } catch {
            setWsMessages((prev) => [...prev.slice(-50), event.data]);
          }
        };
        ws.onclose = () => {
          setWsConnected(false);
          setWsMessages((prev) => [...prev, '--- Disconnected ---']);
        };
        setWsRef(ws);
        setLoading(false);
        return;
      }

      const url = `${API_BASE}${path}`;
      const res = await fetch(url, {
        method: endpoint.method,
        headers: {
          'Content-Type': 'application/json',
          'X-Nucleus-Key': 'dev-nucleus-key-local',
        },
        body: endpoint.body ? endpoint.body : undefined,
      });
      const json = await res.json();
      setResponse(JSON.stringify(json, null, 2));
    } catch (e: any) {
      setResponse(`Error: ${e.message}`);
    } finally {
      setLoading(false);
    }
  };

  const generateSnippet = (endpoint: Endpoint, lang: 'python' | 'typescript' | 'curl') => {
    let path = endpoint.path;
    for (const [key, value] of Object.entries(pathParams)) {
      path = path.replace(`:${key}`, value || `{${key}}`);
    }
    const url = `http://localhost:8080/api/v1${path}`;

    switch (lang) {
      case 'curl':
        if (endpoint.method === 'POST') {
          return `curl -X POST "${url}" \\\n  -H "X-Nucleus-Key: dev-nucleus-key-local" \\\n  -H "Content-Type: application/json" \\\n  -d '${endpoint.body || '{}'}'`;
        }
        return `curl "${url}" \\\n  -H "X-Nucleus-Key: dev-nucleus-key-local"`;
      case 'python':
        return `from nucleus import NucleusClient\n\nclient = NucleusClient("${API_BASE}", "dev-nucleus-key-local")\nresult = client.${endpoint.method === 'GET' ? 'get' : 'post'}("${path}")\nprint(result)`;
      case 'typescript':
        return `import { NucleusClient } from 'nucleus-sdk';\n\nconst client = new NucleusClient('${API_BASE}', 'dev-nucleus-key-local');\nconst result = await client.${endpoint.method.toLowerCase()}('${path}');\nconsole.log(result);`;
    }
  };

  const [snippetLang, setSnippetLang] = useState<'curl' | 'python' | 'typescript'>('curl');

  return (
    <div className="h-full flex flex-col">
      <div className="px-6 py-4 border-b border-nucleus-border">
        <h1 className="font-heading text-2xl font-bold">API Explorer</h1>
        <p className="text-sm text-nucleus-muted mt-1 font-body">Interactive NUCLEUS API documentation and tester</p>
      </div>

      <div className="flex-1 flex overflow-hidden">
        {/* Endpoint List */}
        <div className="w-80 border-r border-nucleus-border overflow-auto">
          {endpoints.map((ep, i) => (
            <div
              key={i}
              className={`p-3 border-b border-nucleus-border cursor-pointer transition-colors ${
                selectedEndpoint === ep ? 'bg-nucleus-panel' : 'hover:bg-nucleus-panel/50'
              }`}
              onClick={() => {
                setSelectedEndpoint(ep);
                setResponse('');
                setPathParams({});
              }}
            >
              <div className="flex items-center gap-2">
                <span className={`text-xs font-mono font-bold px-1.5 py-0.5 rounded ${methodColors[ep.method]}`}>
                  {ep.method}
                </span>
                <span className="font-mono text-xs text-nucleus-text">{ep.path}</span>
              </div>
              <p className="text-xs text-nucleus-muted mt-1">{ep.description}</p>
            </div>
          ))}
        </div>

        {/* Request/Response Area */}
        <div className="flex-1 flex flex-col overflow-hidden">
          {selectedEndpoint ? (
            <>
              <div className="p-4 border-b border-nucleus-border space-y-3">
                <div className="flex items-center gap-3">
                  <span className={`text-sm font-mono font-bold px-2 py-1 rounded ${methodColors[selectedEndpoint.method]}`}>
                    {selectedEndpoint.method}
                  </span>
                  <span className="font-mono text-sm text-nucleus-accent">/api/v1{selectedEndpoint.path}</span>
                  <button
                    onClick={() => executeRequest(selectedEndpoint)}
                    disabled={loading}
                    className="ml-auto px-4 py-1.5 text-sm bg-nucleus-accent text-black font-bold rounded hover:bg-nucleus-accent/80 transition-colors disabled:opacity-50 font-mono"
                  >
                    {loading ? 'Loading...' : selectedEndpoint.method === 'WS' ? (wsConnected ? 'Connected' : 'Connect') : 'Send'}
                  </button>
                </div>

                {/* Path params */}
                {selectedEndpoint.path.includes(':') && (
                  <div className="space-y-2">
                    {selectedEndpoint.path.match(/:(\w+)/g)?.map((param) => {
                      const name = param.slice(1);
                      return (
                        <div key={name} className="flex items-center gap-2">
                          <span className="text-xs text-nucleus-muted font-mono w-32">{name}:</span>
                          <input
                            type="text"
                            value={pathParams[name] || ''}
                            onChange={(e) => setPathParams({ ...pathParams, [name]: e.target.value })}
                            placeholder={`Enter ${name}`}
                            className="flex-1 px-2 py-1 text-xs bg-nucleus-panel border border-nucleus-border rounded font-mono text-nucleus-text placeholder-nucleus-muted focus:border-nucleus-accent focus:outline-none"
                          />
                        </div>
                      );
                    })}
                  </div>
                )}

                {/* Query params */}
                {selectedEndpoint.params && (
                  <div className="text-xs">
                    <span className="text-nucleus-muted">Query params: </span>
                    {selectedEndpoint.params.map((p) => (
                      <span key={p.name} className="font-mono text-nucleus-accent mr-2">
                        {p.name}
                      </span>
                    ))}
                  </div>
                )}
              </div>

              {/* Code Snippets */}
              <div className="px-4 py-3 border-b border-nucleus-border">
                <div className="flex items-center gap-2 mb-2">
                  {(['curl', 'python', 'typescript'] as const).map((lang) => (
                    <button
                      key={lang}
                      onClick={() => setSnippetLang(lang)}
                      className={`px-2 py-1 text-xs rounded font-mono ${
                        snippetLang === lang
                          ? 'bg-nucleus-accent/10 text-nucleus-accent border border-nucleus-accent/30'
                          : 'text-nucleus-muted hover:text-nucleus-text'
                      }`}
                    >
                      {lang}
                    </button>
                  ))}
                </div>
                <pre className="bg-nucleus-panel p-3 rounded text-xs font-mono overflow-auto max-h-32 text-nucleus-text">
                  {generateSnippet(selectedEndpoint, snippetLang)}
                </pre>
              </div>

              {/* Response */}
              <div className="flex-1 overflow-auto p-4">
                <h3 className="text-xs text-nucleus-muted uppercase tracking-wider mb-2">Response</h3>
                {selectedEndpoint.method === 'WS' ? (
                  <div className="bg-nucleus-panel rounded p-3 max-h-full overflow-auto">
                    {wsMessages.length === 0 ? (
                      <p className="text-xs text-nucleus-muted">Connect to see live messages...</p>
                    ) : (
                      wsMessages.map((msg, i) => (
                        <pre key={i} className="text-xs font-mono text-nucleus-text mb-2 pb-2 border-b border-nucleus-border/30">
                          {msg}
                        </pre>
                      ))
                    )}
                  </div>
                ) : response ? (
                  <pre className="bg-nucleus-panel p-3 rounded text-xs font-mono overflow-auto text-nucleus-text">
                    {response}
                  </pre>
                ) : (
                  <p className="text-xs text-nucleus-muted">Send a request to see the response</p>
                )}
              </div>
            </>
          ) : (
            <div className="flex-1 flex items-center justify-center text-nucleus-muted text-sm">
              Select an endpoint to explore
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
