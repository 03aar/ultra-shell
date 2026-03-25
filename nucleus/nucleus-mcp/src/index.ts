#!/usr/bin/env node

import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";
import WebSocket from "ws";
import { loadConfig, Logger } from "./config.js";
import { NucleusAPIClient, NucleusAPIError } from "./client.js";

const config = loadConfig();
const logger = new Logger(config.logLevel);
const client = new NucleusAPIClient(config.apiUrl, config.apiKey, logger);

const server = new McpServer({
  name: "nucleus",
  version: "1.0.0",
});

// ---------------------------------------------------------------------------
// Helper: format error responses
// ---------------------------------------------------------------------------
function errorContent(err: unknown): { content: Array<{ type: "text"; text: string }> } {
  const message =
    err instanceof NucleusAPIError
      ? `Error (${err.statusCode ?? "unknown"}): ${err.message}`
      : err instanceof Error
        ? err.message
        : "Unknown error";
  return { content: [{ type: "text" as const, text: message }] };
}

// ---------------------------------------------------------------------------
// Tools
// ---------------------------------------------------------------------------

// @ts-ignore — MCP SDK deep type instantiation with Zod schemas
server.tool(
  "execute_command",
  "Run a shell command through Nucleus with risk evaluation, execution tracking, and rollback support",
  {
    command: z.string().describe("The shell command to execute"),
    dry_run: z.boolean().optional().describe("If true, evaluate the command without executing it"),
    session_id: z.string().optional().describe("Session ID to associate this execution with"),
  },
  async ({ command, dry_run, session_id }) => {
    try {
      const result = await client.execute(command, dry_run ?? false, session_id);
      const lines = [
        `Execution ID: ${result.execution_id}`,
        `Command: ${result.command}`,
        `Exit Code: ${result.exit_code}`,
        `Duration: ${result.duration_ms}ms`,
        `Risk Level: ${result.risk_level} (score: ${result.risk_score})`,
        `Session: ${result.session_id}`,
        `Dry Run: ${result.dry_run}`,
        `Timestamp: ${result.timestamp}`,
      ];
      if (result.stdout) {
        lines.push("", "--- STDOUT ---", result.stdout);
      }
      if (result.stderr) {
        lines.push("", "--- STDERR ---", result.stderr);
      }
      return { content: [{ type: "text" as const, text: lines.join("\n") }] };
    } catch (err) {
      return errorContent(err);
    }
  }
);

server.tool(
  "get_context",
  "Get the current shell environment context including hostname, user, cwd, OS, and optionally running processes, git status, and environment variables",
  {
    include_processes: z.boolean().optional().describe("Include running process list"),
    include_git: z.boolean().optional().describe("Include git repository status"),
    include_env: z.boolean().optional().describe("Include environment variables"),
  },
  async ({ include_processes, include_git, include_env }) => {
    try {
      const ctx = await client.getContext(
        include_processes ?? false,
        include_git ?? false,
        include_env ?? false
      );
      const lines = [
        `Hostname: ${ctx.hostname}`,
        `User: ${ctx.username}`,
        `Shell: ${ctx.shell}`,
        `CWD: ${ctx.cwd}`,
        `OS: ${ctx.os}`,
        `Arch: ${ctx.arch}`,
      ];
      if (ctx.git) {
        lines.push(
          "",
          "--- Git ---",
          `Branch: ${ctx.git.branch}`,
          `Repo: ${ctx.git.repo}`,
          `Status: ${ctx.git.status}`,
          `Remote: ${ctx.git.remote}`
        );
      }
      if (ctx.processes && ctx.processes.length > 0) {
        lines.push("", "--- Top Processes ---");
        for (const p of ctx.processes.slice(0, 20)) {
          lines.push(`  PID ${p.pid}: ${p.name} (CPU: ${p.cpu}%, MEM: ${p.memory}%)`);
        }
      }
      if (ctx.env) {
        lines.push("", "--- Environment Variables ---");
        for (const [k, v] of Object.entries(ctx.env)) {
          lines.push(`  ${k}=${v}`);
        }
      }
      return { content: [{ type: "text" as const, text: lines.join("\n") }] };
    } catch (err) {
      return errorContent(err);
    }
  }
);

server.tool(
  "get_execution_history",
  "Retrieve recent command execution history with optional filtering by session, category, or search term",
  {
    limit: z.number().optional().describe("Maximum number of entries to return (default 50)"),
    session_id: z.string().optional().describe("Filter by session ID"),
    filter_category: z.string().optional().describe("Filter by command category"),
    search: z.string().optional().describe("Search term to filter commands"),
  },
  async ({ limit, session_id, filter_category, search }) => {
    try {
      let searchTerm = search;
      if (filter_category && !search) {
        searchTerm = `category:${filter_category}`;
      } else if (filter_category && search) {
        searchTerm = `category:${filter_category} ${search}`;
      }
      const entries = await client.getHistory(limit ?? 50, session_id, searchTerm);
      if (entries.length === 0) {
        return { content: [{ type: "text" as const, text: "No execution history found." }] };
      }
      const lines = [`Execution History (${entries.length} entries):`, ""];
      for (const entry of entries) {
        lines.push(
          `[${entry.timestamp}] ${entry.execution_id}`,
          `  Command: ${entry.command}`,
          `  Exit: ${entry.exit_code} | Duration: ${entry.duration_ms}ms | Risk: ${entry.risk_level}`,
          `  Session: ${entry.session_id} | Category: ${entry.category}`,
          ""
        );
      }
      return { content: [{ type: "text" as const, text: lines.join("\n") }] };
    } catch (err) {
      return errorContent(err);
    }
  }
);

server.tool(
  "rollback_execution",
  "Rollback a previously executed command by its execution ID. Can preview rollback steps without executing them.",
  {
    execution_id: z.string().describe("The execution ID to rollback"),
    preview_only: z.boolean().optional().describe("If true, show rollback plan without executing"),
  },
  async ({ execution_id, preview_only }) => {
    try {
      const result = await client.rollback(execution_id, preview_only ?? false);
      const lines = [
        `Rollback for Execution: ${result.execution_id}`,
        `Status: ${result.status}`,
        `Preview Only: ${result.preview_only}`,
        "",
        "Rollback Commands:",
      ];
      for (const cmd of result.rollback_commands) {
        lines.push(`  - ${cmd}`);
      }
      return { content: [{ type: "text" as const, text: lines.join("\n") }] };
    } catch (err) {
      return errorContent(err);
    }
  }
);

// @ts-ignore — MCP SDK deep type instantiation with Zod schemas
server.tool(
  "get_execution_graph",
  "Get the execution dependency graph (DAG) for a session, showing how commands relate to each other",
  {
    session_id: z.string().optional().describe("Session ID to get graph for"),
    format: z
      .enum(["json", "mermaid"])
      .optional()
      .describe("Output format: json or mermaid diagram"),
  },
  async ({ session_id, format }) => {
    try {
      const graph = await client.getGraph(session_id, format ?? "json");
      if (format === "mermaid" && graph.mermaid) {
        return {
          content: [
            {
              type: "text" as const,
              text: `Execution Graph (Mermaid) for session ${graph.session_id}:\n\n\`\`\`mermaid\n${graph.mermaid}\n\`\`\``,
            },
          ],
        };
      }
      const lines = [
        `Execution Graph for session: ${graph.session_id}`,
        `Format: ${graph.format}`,
        "",
        `Nodes (${graph.nodes.length}):`,
      ];
      for (const node of graph.nodes) {
        lines.push(`  [${node.id}] ${node.command} (${node.status}) @ ${node.timestamp}`);
      }
      lines.push("", `Edges (${graph.edges.length}):`);
      for (const edge of graph.edges) {
        lines.push(`  ${edge.source} --${edge.relation}--> ${edge.target}`);
      }
      return { content: [{ type: "text" as const, text: lines.join("\n") }] };
    } catch (err) {
      return errorContent(err);
    }
  }
);

// @ts-ignore — MCP SDK deep type instantiation with Zod schemas
server.tool(
  "run_skill",
  "Execute a Nucleus skill (a predefined automation workflow) with the given parameters",
  {
    skill_name: z.string().describe("Name of the skill to run"),
    params: z
      .record(z.unknown())
      .optional()
      .describe("Parameters to pass to the skill"),
  },
  async ({ skill_name, params }) => {
    try {
      const result = await client.runSkill(skill_name, (params ?? {}) as Record<string, unknown>);
      const lines = [
        `Skill: ${result.skill_name}`,
        `Status: ${result.status}`,
        `Duration: ${result.duration_ms}ms`,
        "",
        "Output:",
        typeof result.output === "string"
          ? result.output
          : JSON.stringify(result.output, null, 2),
      ];
      return { content: [{ type: "text" as const, text: lines.join("\n") }] };
    } catch (err) {
      return errorContent(err);
    }
  }
);

server.tool(
  "search_history",
  "Full-text search across execution history with relevance ranking",
  {
    query: z.string().describe("Search query string"),
    session_id: z.string().optional().describe("Limit search to a specific session"),
    limit: z.number().optional().describe("Maximum results to return (default 20)"),
  },
  async ({ query, session_id, limit }) => {
    try {
      const result = await client.searchHistory(query, limit ?? 20, session_id);
      if (result.results.length === 0) {
        return {
          content: [{ type: "text" as const, text: `No results found for query: "${query}"` }],
        };
      }
      const lines = [
        `Search Results for "${result.query}" (${result.total} total, showing ${result.results.length}):`,
        "",
      ];
      for (const entry of result.results) {
        lines.push(
          `[${entry.timestamp}] ${entry.execution_id}`,
          `  Command: ${entry.command}`,
          `  Exit: ${entry.exit_code} | Risk: ${entry.risk_level} | Session: ${entry.session_id}`,
          ""
        );
      }
      return { content: [{ type: "text" as const, text: lines.join("\n") }] };
    } catch (err) {
      return errorContent(err);
    }
  }
);

server.tool(
  "get_risk_assessment",
  "Perform a dry-run risk assessment for a command without executing it. Returns risk level, score, factors, and recommendation.",
  {
    command: z.string().describe("The command to assess"),
  },
  async ({ command }) => {
    try {
      const assessment = await client.getRiskAssessment(command);
      const lines = [
        `Risk Assessment for: ${assessment.command}`,
        "",
        `Risk Level: ${assessment.risk_level}`,
        `Risk Score: ${assessment.risk_score}`,
        `Requires Confirmation: ${assessment.requires_confirmation}`,
        `Recommendation: ${assessment.recommendation}`,
        "",
        "Risk Factors:",
      ];
      if (assessment.risk_factors.length === 0) {
        lines.push("  No risk factors identified.");
      } else {
        for (const f of assessment.risk_factors) {
          lines.push(`  - [${f.severity}] ${f.factor}: ${f.description}`);
        }
      }
      return { content: [{ type: "text" as const, text: lines.join("\n") }] };
    } catch (err) {
      return errorContent(err);
    }
  }
);

// @ts-ignore — MCP SDK deep type instantiation with Zod schemas
server.tool(
  "create_session",
  "Create a new Nucleus session for grouping related command executions together",
  {
    name: z.string().describe("Human-readable session name"),
    tags: z.array(z.string()).optional().describe("Optional tags for categorization"),
  },
  async ({ name, tags }) => {
    try {
      const session = await client.createSession(name, tags ?? []);
      const lines = [
        `Session Created:`,
        `  ID: ${session.session_id}`,
        `  Name: ${session.name}`,
        `  Tags: ${session.tags.join(", ") || "(none)"}`,
        `  Created: ${session.created_at}`,
        `  Status: ${session.status}`,
      ];
      return { content: [{ type: "text" as const, text: lines.join("\n") }] };
    } catch (err) {
      return errorContent(err);
    }
  }
);

server.tool(
  "watch_stream",
  "Subscribe to real-time Nucleus event stream via WebSocket. Returns events collected during the specified duration.",
  {
    event_types: z
      .array(z.string())
      .optional()
      .describe("Event types to filter (e.g. execution, risk, session). Empty = all events."),
    duration_seconds: z
      .number()
      .optional()
      .describe("How long to listen for events in seconds (default 10, max 60)"),
  },
  async ({ event_types, duration_seconds }) => {
    const duration = Math.min(duration_seconds ?? 10, 60);
    const filters = event_types ?? [];

    const wsUrl = config.apiUrl.replace(/^http/, "ws") + "/api/v1/stream";

    return new Promise((resolve) => {
      const events: Array<{ type: string; data: unknown; timestamp: string }> = [];
      let ws: WebSocket | null = null;
      let settled = false;

      const finish = () => {
        if (settled) return;
        settled = true;
        try {
          ws?.close();
        } catch {
          // ignore close errors
        }
        if (events.length === 0) {
          resolve({
            content: [
              {
                type: "text" as const,
                text: `No events received during ${duration}s watch window (filters: ${filters.length > 0 ? filters.join(", ") : "all"}).`,
              },
            ],
          });
          return;
        }
        const lines = [
          `Collected ${events.length} events over ${duration}s:`,
          "",
        ];
        for (const evt of events) {
          lines.push(
            `[${evt.timestamp}] ${evt.type}`,
            `  ${typeof evt.data === "string" ? evt.data : JSON.stringify(evt.data)}`,
            ""
          );
        }
        resolve({ content: [{ type: "text" as const, text: lines.join("\n") }] });
      };

      const timer = setTimeout(finish, duration * 1000);

      try {
        ws = new WebSocket(wsUrl, {
          headers: { Authorization: `Bearer ${config.apiKey}` },
        });

        ws.on("open", () => {
          if (filters.length > 0) {
            ws!.send(JSON.stringify({ subscribe: filters }));
          }
          logger.debug(`WebSocket connected, watching for ${duration}s`);
        });

        ws.on("message", (data: WebSocket.Data) => {
          try {
            const parsed = JSON.parse(data.toString());
            const eventType = parsed.type || parsed.event_type || "unknown";
            if (filters.length === 0 || filters.includes(eventType)) {
              events.push({
                type: eventType,
                data: parsed.data || parsed.payload || parsed,
                timestamp: parsed.timestamp || new Date().toISOString(),
              });
            }
          } catch {
            events.push({
              type: "raw",
              data: data.toString(),
              timestamp: new Date().toISOString(),
            });
          }
        });

        ws.on("error", (err: Error) => {
          logger.warn(`WebSocket error: ${err.message}`);
          clearTimeout(timer);
          finish();
        });

        ws.on("close", () => {
          clearTimeout(timer);
          finish();
        });
      } catch (err) {
        clearTimeout(timer);
        const message = err instanceof Error ? err.message : "Unknown WebSocket error";
        if (!settled) {
          settled = true;
          resolve({
            content: [
              {
                type: "text" as const,
                text: `Failed to connect to event stream: ${message}`,
              },
            ],
          });
        }
      }
    });
  }
);

// ---------------------------------------------------------------------------
// Resources
// ---------------------------------------------------------------------------

server.resource(
  "context",
  "nucleus://context",
  { description: "Current shell environment context", mimeType: "application/json" },
  async () => {
    const ctx = await client.getContext(true, true, false);
    return {
      contents: [
        {
          uri: "nucleus://context",
          mimeType: "application/json" as const,
          text: JSON.stringify(ctx, null, 2),
        },
      ],
    };
  }
);

server.resource(
  "graph",
  "nucleus://graph",
  { description: "Current session execution dependency graph", mimeType: "application/json" },
  async () => {
    const graph = await client.getGraph(undefined, "json");
    return {
      contents: [
        {
          uri: "nucleus://graph",
          mimeType: "application/json" as const,
          text: JSON.stringify(graph, null, 2),
        },
      ],
    };
  }
);

server.resource(
  "history",
  "nucleus://history",
  { description: "Recent execution history", mimeType: "application/json" },
  async () => {
    const history = await client.getHistory(100);
    return {
      contents: [
        {
          uri: "nucleus://history",
          mimeType: "application/json" as const,
          text: JSON.stringify(history, null, 2),
        },
      ],
    };
  }
);

server.resource(
  "sessions",
  "nucleus://sessions",
  { description: "All Nucleus sessions", mimeType: "application/json" },
  async () => {
    const sessions = await client.getSessions();
    return {
      contents: [
        {
          uri: "nucleus://sessions",
          mimeType: "application/json" as const,
          text: JSON.stringify(sessions, null, 2),
        },
      ],
    };
  }
);

// ---------------------------------------------------------------------------
// Prompts
// ---------------------------------------------------------------------------

server.prompt(
  "debug_failure",
  "Analyze a failed command execution and suggest fixes",
  {
    execution_id: z.string().describe("The execution ID of the failed command"),
  },
  async ({ execution_id }) => {
    let contextText: string;
    try {
      const history = await client.getHistory(100);
      const entry = history.find((e) => e.execution_id === execution_id);
      const ctx = await client.getContext(false, true, false);
      contextText = JSON.stringify({ failed_execution: entry, context: ctx }, null, 2);
    } catch (err) {
      contextText = `Failed to fetch context: ${err instanceof Error ? err.message : "unknown error"}`;
    }

    return {
      messages: [
        {
          role: "user" as const,
          content: {
            type: "text" as const,
            text: `A command execution has failed. Please analyze the failure and suggest fixes.

Execution ID: ${execution_id}

Here is the execution data and current environment context:

\`\`\`json
${contextText}
\`\`\`

Please:
1. Identify the root cause of the failure
2. Suggest specific fixes or alternative commands
3. Note any environmental issues that may have contributed
4. Recommend whether a rollback is needed`,
          },
        },
      ],
    };
  }
);

server.prompt(
  "plan_workflow",
  "Plan a multi-step shell workflow with risk assessment",
  {
    goal: z.string().describe("Description of what you want to accomplish"),
    constraints: z.string().optional().describe("Any constraints or requirements"),
  },
  async ({ goal, constraints }) => {
    let contextText: string;
    try {
      const ctx = await client.getContext(false, true, false);
      contextText = JSON.stringify(ctx, null, 2);
    } catch (err) {
      contextText = `Failed to fetch context: ${err instanceof Error ? err.message : "unknown error"}`;
    }

    const constraintBlock = constraints
      ? `\nConstraints:\n${constraints}\n`
      : "";

    return {
      messages: [
        {
          role: "user" as const,
          content: {
            type: "text" as const,
            text: `Plan a multi-step shell workflow to accomplish the following goal.

Goal: ${goal}
${constraintBlock}
Current environment context:

\`\`\`json
${contextText}
\`\`\`

Please:
1. Break the goal into discrete shell commands
2. Identify dependencies between steps
3. Assess the risk level of each step
4. Suggest a session name and tags for tracking
5. Note any commands that should use dry_run first
6. Provide rollback strategies for high-risk steps`,
          },
        },
      ],
    };
  }
);

server.prompt(
  "explain_session",
  "Explain what happened in a Nucleus session in plain language",
  {
    session_id: z.string().describe("The session ID to explain"),
  },
  async ({ session_id }) => {
    let contextText: string;
    try {
      const history = await client.getHistory(200, session_id);
      const graph = await client.getGraph(session_id, "json");
      contextText = JSON.stringify({ history, graph }, null, 2);
    } catch (err) {
      contextText = `Failed to fetch session data: ${err instanceof Error ? err.message : "unknown error"}`;
    }

    return {
      messages: [
        {
          role: "user" as const,
          content: {
            type: "text" as const,
            text: `Please explain what happened in this Nucleus session in plain language.

Session ID: ${session_id}

Here is the full session history and execution graph:

\`\`\`json
${contextText}
\`\`\`

Please:
1. Summarize the overall objective of this session
2. Walk through what each command did and why
3. Highlight any failures or risky operations
4. Note the relationships between commands (from the graph)
5. Provide an overall assessment of the session outcome`,
          },
        },
      ],
    };
  }
);

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

async function main(): Promise<void> {
  logger.info("Starting Nucleus MCP server v1.0.0");
  logger.info(`API URL: ${config.apiUrl}`);
  logger.info(`Log Level: ${config.logLevel}`);

  const transport = new StdioServerTransport();
  await server.connect(transport);

  logger.info("Nucleus MCP server connected and ready");
}

main().catch((err) => {
  logger.error("Fatal error starting Nucleus MCP server", err);
  process.exit(1);
});
