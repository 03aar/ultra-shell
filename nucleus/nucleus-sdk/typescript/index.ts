/**
 * NUCLEUS TypeScript SDK
 *
 * A TypeScript client for the NUCLEUS AI-native shell runtime API.
 * Works in both Node.js and browser environments.
 *
 * @example
 * ```typescript
 * import { NucleusClient } from 'nucleus-sdk';
 *
 * const client = new NucleusClient('http://localhost:8080/api/v1', 'dev-nucleus-key-local');
 * const executions = await client.getExecutions({ limit: 10 });
 * const context = await client.getContext();
 * ```
 */

// ── Types ──────────────────────────────────────────

export interface RiskFlag {
  code: string;
  message: string;
  level: 'none' | 'low' | 'medium' | 'high' | 'critical';
}

export interface PipeSegment {
  binary: string;
  args: string[];
  flags: string[];
}

export interface Redirection {
  direction: string;
  target: string;
}

export interface ParsedCommand {
  raw: string;
  binary: string;
  args: string[];
  flags: string[];
  pipes: PipeSegment[];
  redirections: Redirection[];
  env_assignments: Record<string, string>;
  category: string;
  is_background: boolean;
  is_chained: boolean;
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

export interface ProcessInfo {
  pid: number;
  name: string;
  cpu: number;
  memory: number;
  command: string;
}

export interface DiskUsage {
  total: number;
  used: number;
  available: number;
  percent: number;
}

export interface ContextState {
  cwd: string;
  git_branch: string;
  git_status: string;
  running_processes: ProcessInfo[];
  env_vars: Record<string, string>;
  disk_usage: DiskUsage;
  open_files_count: number;
}

export interface Graph {
  nodes: ExecutionNode[];
  edges: Edge[];
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

export interface RollbackResult {
  success: boolean;
  files_restored: string[];
  env_restored: Record<string, string>;
  message: string;
}

export interface ExecutionListResponse {
  executions: ExecutionNode[];
  total: number;
  session_id: string;
}

export interface ExecutionContextResponse {
  execution: ExecutionNode;
  edges: Edge[];
  recent_prior: ExecutionNode[];
  env_diff_since_start: Record<string, string>;
}

export interface SessionReplayResponse {
  session: Session;
  executions: ExecutionNode[];
}

export interface APIResponse<T> {
  data: T;
  meta: { timestamp: string; session_id: string; version: string };
  error: string | null;
}

export interface WSMessage {
  type: 'execution_complete' | 'risk_warning' | 'rollback_complete' | 'session_started';
  data: unknown;
}

export interface GetExecutionsOptions {
  limit?: number;
  offset?: number;
  session_id?: string;
  risk_level?: string;
  command_category?: string;
}

// ── Errors ─────────────────────────────────────────

export class NucleusError extends Error {
  statusCode: number;

  constructor(message: string, statusCode = 0) {
    super(message);
    this.name = 'NucleusError';
    this.statusCode = statusCode;
  }
}

// ── Client ─────────────────────────────────────────

export class NucleusClient {
  private apiUrl: string;
  private apiKey: string;

  /**
   * Create a NUCLEUS API client.
   *
   * @param apiUrl - Base URL for the NUCLEUS API (e.g., "http://localhost:8080/api/v1")
   * @param apiKey - API key for authentication (e.g., "dev-nucleus-key-local")
   */
  constructor(apiUrl: string, apiKey: string) {
    this.apiUrl = apiUrl.replace(/\/+$/, '');
    this.apiKey = apiKey;
  }

  private async request<T>(method: string, path: string, body?: unknown): Promise<T> {
    const url = `${this.apiUrl}${path}`;
    const options: RequestInit = {
      method,
      headers: {
        'Content-Type': 'application/json',
        'X-Nucleus-Key': this.apiKey,
      },
    };

    if (body) {
      options.body = JSON.stringify(body);
    }

    const response = await fetch(url, options);
    const json: APIResponse<T> = await response.json();

    if (json.error) {
      throw new NucleusError(json.error, response.status);
    }

    return json.data;
  }

  /**
   * Get recent command executions.
   *
   * @param options - Query parameters for filtering
   * @returns List of execution nodes
   */
  async getExecutions(options: GetExecutionsOptions = {}): Promise<ExecutionNode[]> {
    const params = new URLSearchParams();
    if (options.limit) params.set('limit', String(options.limit));
    if (options.offset) params.set('offset', String(options.offset));
    if (options.session_id) params.set('session_id', options.session_id);
    if (options.risk_level) params.set('risk_level', options.risk_level);
    if (options.command_category) params.set('command_category', options.command_category);

    const qs = params.toString();
    const data = await this.request<ExecutionListResponse>(
      'GET',
      `/executions${qs ? '?' + qs : ''}`
    );
    return data.executions;
  }

  /**
   * Get full details for a specific execution.
   *
   * @param executionId - UUID of the execution
   * @returns Full execution node with stdout, stderr, etc.
   */
  async getExecution(executionId: string): Promise<ExecutionNode> {
    return this.request<ExecutionNode>('GET', `/executions/${executionId}`);
  }

  /**
   * Get execution with dependency context (edges, prior executions).
   *
   * @param executionId - UUID of the execution
   * @returns Execution with edges, prior context, and env diff
   */
  async getExecutionContext(executionId: string): Promise<ExecutionContextResponse> {
    return this.request<ExecutionContextResponse>('GET', `/executions/${executionId}/context`);
  }

  /**
   * Get current environment state.
   *
   * @returns Current context including cwd, git, processes, env vars, disk
   */
  async getContext(): Promise<ContextState> {
    return this.request<ContextState>('GET', '/context');
  }

  /**
   * Get the execution dependency graph.
   *
   * @param sessionId - Optional session ID to filter the graph
   * @returns Graph with nodes and edges, compatible with reactflow
   */
  async getGraph(sessionId?: string): Promise<Graph> {
    const qs = sessionId ? `?session_id=${sessionId}` : '';
    return this.request<Graph>('GET', `/context/graph${qs}`);
  }

  /**
   * Rollback a specific execution, restoring files and env vars.
   *
   * @param executionId - UUID of the execution to rollback
   * @returns Result with success status and list of restored items
   */
  async rollback(executionId: string): Promise<RollbackResult> {
    return this.request<RollbackResult>('POST', `/rollback/${executionId}`);
  }

  /**
   * Get all sessions.
   *
   * @returns List of sessions with metadata
   */
  async getSessions(): Promise<Session[]> {
    return this.request<Session[]>('GET', '/sessions');
  }

  /**
   * Create a new named session.
   *
   * @param name - Name for the session
   * @returns Created session
   */
  async createSession(name: string): Promise<Session> {
    return this.request<Session>('POST', '/sessions', { name });
  }

  /**
   * Get all executions for a session in chronological order.
   *
   * @param sessionId - UUID of the session
   * @returns Session replay data with session metadata and executions
   */
  async getSessionReplay(sessionId: string): Promise<SessionReplayResponse> {
    return this.request<SessionReplayResponse>('GET', `/sessions/${sessionId}/replay`);
  }

  /**
   * Connect to the WebSocket stream for real-time events.
   *
   * @param callback - Function called with WSMessage for each event
   * @returns WebSocket instance (call .close() to disconnect)
   */
  stream(callback: (msg: WSMessage) => void): WebSocket {
    const wsUrl = this.apiUrl
      .replace('http://', 'ws://')
      .replace('https://', 'wss://');

    const ws = new WebSocket(`${wsUrl}/ws/stream?api_key=${this.apiKey}`);

    ws.onmessage = (event: MessageEvent) => {
      try {
        const msg: WSMessage = JSON.parse(event.data);
        callback(msg);
      } catch {
        // Ignore parse errors
      }
    };

    return ws;
  }
}

// ── Default export ─────────────────────────────────

export default NucleusClient;
