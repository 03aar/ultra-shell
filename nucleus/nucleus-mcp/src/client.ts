import axios, { AxiosInstance, AxiosError } from "axios";
import { Logger } from "./config.js";

export interface ExecutionResult {
  execution_id: string;
  command: string;
  exit_code: number;
  stdout: string;
  stderr: string;
  duration_ms: number;
  risk_level: string;
  risk_score: number;
  session_id: string;
  timestamp: string;
  dry_run: boolean;
}

export interface ContextResult {
  hostname: string;
  username: string;
  shell: string;
  cwd: string;
  os: string;
  arch: string;
  processes?: Array<{ pid: number; name: string; cpu: number; memory: number }>;
  git?: {
    branch: string;
    repo: string;
    status: string;
    remote: string;
  };
  env?: Record<string, string>;
}

export interface HistoryEntry {
  execution_id: string;
  command: string;
  exit_code: number;
  timestamp: string;
  duration_ms: number;
  risk_level: string;
  session_id: string;
  category: string;
}

export interface RollbackResult {
  execution_id: string;
  rollback_commands: string[];
  status: string;
  preview_only: boolean;
}

export interface GraphResult {
  session_id: string;
  format: string;
  nodes: Array<{ id: string; command: string; status: string; timestamp: string }>;
  edges: Array<{ source: string; target: string; relation: string }>;
  mermaid?: string;
}

export interface SessionResult {
  session_id: string;
  name: string;
  tags: string[];
  created_at: string;
  command_count: number;
  status: string;
}

export interface RiskAssessment {
  command: string;
  risk_level: string;
  risk_score: number;
  risk_factors: Array<{ factor: string; severity: string; description: string }>;
  recommendation: string;
  requires_confirmation: boolean;
}

export interface SkillResult {
  skill_name: string;
  status: string;
  output: unknown;
  duration_ms: number;
}

export interface SearchResult {
  results: HistoryEntry[];
  total: number;
  query: string;
}

export class NucleusAPIError extends Error {
  public statusCode: number | undefined;
  public responseBody: unknown;

  constructor(message: string, statusCode?: number, responseBody?: unknown) {
    super(message);
    this.name = "NucleusAPIError";
    this.statusCode = statusCode;
    this.responseBody = responseBody;
  }
}

export class NucleusAPIClient {
  private client: AxiosInstance;
  private logger: Logger;
  private maxRetries = 3;
  private baseDelay = 1000;

  constructor(apiUrl: string, apiKey: string, logger: Logger) {
    this.logger = logger;
    this.client = axios.create({
      baseURL: apiUrl,
      headers: {
        Authorization: `Bearer ${apiKey}`,
        "Content-Type": "application/json",
      },
      timeout: 30000,
    });
  }

  private async withRetry<T>(operation: () => Promise<T>, context: string): Promise<T> {
    let lastError: Error | undefined;
    for (let attempt = 0; attempt <= this.maxRetries; attempt++) {
      try {
        if (attempt > 0) {
          const delay = this.baseDelay * Math.pow(2, attempt - 1);
          this.logger.debug(`Retry attempt ${attempt}/${this.maxRetries} for ${context}, waiting ${delay}ms`);
          await new Promise((resolve) => setTimeout(resolve, delay));
        }
        return await operation();
      } catch (err) {
        lastError = this.normalizeError(err, context);
        const axErr = err as AxiosError;
        const status = axErr.response?.status;
        // Do not retry client errors (4xx) except 429
        if (status && status >= 400 && status < 500 && status !== 429) {
          throw lastError;
        }
        this.logger.warn(`Attempt ${attempt + 1} failed for ${context}: ${lastError.message}`);
      }
    }
    throw lastError;
  }

  private normalizeError(err: unknown, context: string): NucleusAPIError {
    if (err instanceof NucleusAPIError) return err;
    if (axios.isAxiosError(err)) {
      const axErr = err as AxiosError;
      const status = axErr.response?.status;
      const body = axErr.response?.data;
      const message =
        (body as { error?: string })?.error ||
        (body as { message?: string })?.message ||
        axErr.message;
      return new NucleusAPIError(`${context}: ${message}`, status, body);
    }
    if (err instanceof Error) {
      return new NucleusAPIError(`${context}: ${err.message}`);
    }
    return new NucleusAPIError(`${context}: Unknown error`);
  }

  async getContext(
    includeProcesses = false,
    includeGit = false,
    includeEnv = false
  ): Promise<ContextResult> {
    return this.withRetry(async () => {
      const params: Record<string, string> = {};
      if (includeProcesses) params.include_processes = "true";
      if (includeGit) params.include_git = "true";
      if (includeEnv) params.include_env = "true";
      const resp = await this.client.get("/api/v1/context", { params });
      return resp.data;
    }, "getContext");
  }

  async getHistory(
    limit = 50,
    sessionId?: string,
    search?: string
  ): Promise<HistoryEntry[]> {
    return this.withRetry(async () => {
      const params: Record<string, string | number> = { limit };
      if (sessionId) params.session_id = sessionId;
      if (search) params.search = search;
      const resp = await this.client.get("/api/v1/history", { params });
      return resp.data;
    }, "getHistory");
  }

  async execute(
    command: string,
    dryRun = false,
    sessionId?: string
  ): Promise<ExecutionResult> {
    return this.withRetry(async () => {
      const body: Record<string, unknown> = { command, dry_run: dryRun };
      if (sessionId) body.session_id = sessionId;
      const resp = await this.client.post("/api/v1/execute", body);
      return resp.data;
    }, "execute");
  }

  async rollback(
    executionId: string,
    previewOnly = false
  ): Promise<RollbackResult> {
    return this.withRetry(async () => {
      const resp = await this.client.post("/api/v1/rollback", {
        execution_id: executionId,
        preview_only: previewOnly,
      });
      return resp.data;
    }, "rollback");
  }

  async getGraph(
    sessionId?: string,
    format: "json" | "mermaid" = "json"
  ): Promise<GraphResult> {
    return this.withRetry(async () => {
      const params: Record<string, string> = { format };
      if (sessionId) params.session_id = sessionId;
      const resp = await this.client.get("/api/v1/graph", { params });
      return resp.data;
    }, "getGraph");
  }

  async createSession(
    name: string,
    tags: string[] = []
  ): Promise<SessionResult> {
    return this.withRetry(async () => {
      const resp = await this.client.post("/api/v1/sessions", { name, tags });
      return resp.data;
    }, "createSession");
  }

  async getRiskAssessment(command: string): Promise<RiskAssessment> {
    return this.withRetry(async () => {
      const resp = await this.client.post("/api/v1/risk", { command });
      return resp.data;
    }, "getRiskAssessment");
  }

  async getSessions(): Promise<SessionResult[]> {
    return this.withRetry(async () => {
      const resp = await this.client.get("/api/v1/sessions");
      return resp.data;
    }, "getSessions");
  }

  async runSkill(
    name: string,
    params: Record<string, unknown> = {}
  ): Promise<SkillResult> {
    return this.withRetry(async () => {
      const resp = await this.client.post("/api/v1/skills/run", {
        skill_name: name,
        params,
      });
      return resp.data;
    }, "runSkill");
  }

  async searchHistory(
    query: string,
    limit = 20,
    sessionId?: string
  ): Promise<SearchResult> {
    return this.withRetry(async () => {
      const params: Record<string, string | number> = { query, limit };
      if (sessionId) params.session_id = sessionId;
      const resp = await this.client.get("/api/v1/history/search", { params });
      return resp.data;
    }, "searchHistory");
  }
}
