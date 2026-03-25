const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';
const API_KEY = process.env.NEXT_PUBLIC_API_KEY || 'dev-nucleus-key-local';
const WS_BASE = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/api/v1';

export interface ParsedCommand {
  raw: string;
  binary: string;
  args: string[];
  flags: string[];
  pipes: { binary: string; args: string[]; flags: string[] }[];
  redirections: { direction: string; target: string }[];
  env_assignments: Record<string, string>;
  category: string;
  is_background: boolean;
  is_chained: boolean;
}

export interface RiskFlag {
  code: string;
  message: string;
  level: 'none' | 'low' | 'medium' | 'high' | 'critical';
}

export interface ExecutionNode {
  id: string;
  timestamp: string;
  command: ParsedCommand;
  exit_code: number;
  duration_ms: number;
  stdout: string;
  stderr: string;
  files_mutated: string[];
  env_changes: Record<string, string>;
  processes_spawned: number[];
  risk_flags: RiskFlag[];
  rollback_available: boolean;
  rolled_back: boolean;
  session_id: string;
}

export interface Edge {
  from: string;
  to: string;
  dependency_type: string;
}

export interface ContextState {
  cwd: string;
  git_branch: string;
  git_status: string;
  running_processes: { pid: number; name: string; cpu: number; memory: number; command: string }[];
  env_vars: Record<string, string>;
  disk_usage: { total: number; used: number; available: number; percent: number };
  open_files_count: number;
}

export interface Session {
  id: string;
  name: string;
  shell_pid: number;
  start_time: string;
  end_time: string | null;
  git_repo: string | null;
  cwd: string;
  execution_count: number;
  risk_warning_count: number;
}

export interface APIResponse<T> {
  data: T;
  meta: { timestamp: string; session_id: string; version: string };
  error: string | null;
}

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      'X-Nucleus-Key': API_KEY,
      ...options?.headers,
    },
  });
  const json: APIResponse<T> = await res.json();
  if (json.error) throw new Error(json.error);
  return json.data;
}

export async function getExecutions(params?: {
  limit?: number;
  offset?: number;
  session_id?: string;
  risk_level?: string;
  command_category?: string;
}) {
  const query = new URLSearchParams();
  if (params?.limit) query.set('limit', String(params.limit));
  if (params?.offset) query.set('offset', String(params.offset));
  if (params?.session_id) query.set('session_id', params.session_id);
  if (params?.risk_level) query.set('risk_level', params.risk_level);
  if (params?.command_category) query.set('command_category', params.command_category);
  const qs = query.toString();
  return apiFetch<{ executions: ExecutionNode[]; total: number; session_id: string }>(
    `/executions${qs ? '?' + qs : ''}`
  );
}

export async function getExecution(id: string) {
  return apiFetch<ExecutionNode>(`/executions/${id}`);
}

export async function getExecutionContext(id: string) {
  return apiFetch<{
    execution: ExecutionNode;
    edges: Edge[];
    recent_prior: ExecutionNode[];
    env_diff_since_start: Record<string, string>;
  }>(`/executions/${id}/context`);
}

export async function getContext() {
  return apiFetch<ContextState>('/context');
}

export async function getContextGraph(sessionId?: string) {
  const qs = sessionId ? `?session_id=${sessionId}` : '';
  return apiFetch<{ nodes: ExecutionNode[]; edges: Edge[] }>(`/context/graph${qs}`);
}

export async function rollback(executionId: string) {
  return apiFetch<{
    success: boolean;
    files_restored: string[];
    env_restored: Record<string, string>;
    message: string;
  }>(`/rollback/${executionId}`, { method: 'POST' });
}

export async function getSessions() {
  return apiFetch<Session[]>('/sessions');
}

export async function createSession(name: string) {
  return apiFetch<Session>('/sessions', {
    method: 'POST',
    body: JSON.stringify({ name }),
  });
}

export async function getSessionReplay(id: string) {
  return apiFetch<{ session: Session; executions: ExecutionNode[] }>(`/sessions/${id}/replay`);
}

export interface WSMessage {
  type: 'execution_complete' | 'risk_warning' | 'rollback_complete' | 'session_started';
  data: any;
}

export function connectWebSocket(onMessage: (msg: WSMessage) => void): WebSocket {
  const ws = new WebSocket(`${WS_BASE}/ws/stream?api_key=${API_KEY}`);
  ws.onmessage = (event) => {
    try {
      const msg: WSMessage = JSON.parse(event.data);
      onMessage(msg);
    } catch (e) {
      console.error('Failed to parse WS message:', e);
    }
  };
  ws.onclose = () => {
    // Reconnect after 3 seconds
    setTimeout(() => connectWebSocket(onMessage), 3000);
  };
  return ws;
}
