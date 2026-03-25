"""NucleusAgent — agentic loop that connects an LLM provider to the Nucleus engine."""

from __future__ import annotations

import json
import time
import uuid
from typing import Any

from rich.console import Console
from rich.live import Live
from rich.markdown import Markdown
from rich.panel import Panel
from rich.spinner import Spinner
from rich.text import Text

from nucleus_orchestrator.client import NucleusClient
from nucleus_orchestrator.providers.base import BaseProvider
from nucleus_orchestrator.types import (
    AgentResult,
    CompletionResult,
    ExecutionResult,
    RiskLevel,
)

# ---------------------------------------------------------------------------
# Nucleus tool definitions (passed to the LLM so it can call back)
# ---------------------------------------------------------------------------

NUCLEUS_TOOLS: list[dict[str, Any]] = [
    {
        "name": "execute_command",
        "description": "Execute a shell command through the Nucleus engine. Returns stdout, stderr, exit code, and execution metadata.",
        "parameters": [
            {"name": "command", "type": "string", "description": "The shell command to execute.", "required": True},
            {"name": "dry_run", "type": "boolean", "description": "If true, simulate without executing.", "required": False, "default": False},
        ],
    },
    {
        "name": "get_context",
        "description": "Get the current shell context: working directory, user, git info, environment, recent commands.",
        "parameters": [],
    },
    {
        "name": "get_history",
        "description": "Retrieve recent command execution history.",
        "parameters": [
            {"name": "limit", "type": "integer", "description": "Max entries to return.", "required": False, "default": 20},
            {"name": "search", "type": "string", "description": "Filter by substring.", "required": False},
        ],
    },
    {
        "name": "rollback",
        "description": "Rollback (undo) a previous command execution by its execution_id.",
        "parameters": [
            {"name": "execution_id", "type": "string", "description": "The execution ID to rollback.", "required": True},
        ],
    },
    {
        "name": "search_history",
        "description": "Semantic search over command history.",
        "parameters": [
            {"name": "query", "type": "string", "description": "Natural language search query.", "required": True},
            {"name": "limit", "type": "integer", "description": "Max results.", "required": False, "default": 10},
        ],
    },
    {
        "name": "get_risk_assessment",
        "description": "Assess the risk level of a command before executing it. Returns risk level, score, and reasons.",
        "parameters": [
            {"name": "command", "type": "string", "description": "The command to assess.", "required": True},
        ],
    },
    {
        "name": "create_session",
        "description": "Create a new Nucleus session for grouping related commands.",
        "parameters": [
            {"name": "name", "type": "string", "description": "Session name.", "required": False},
            {"name": "tags", "type": "array", "description": "Tags for the session.", "required": False},
        ],
    },
    {
        "name": "run_skill",
        "description": "Run a predefined Nucleus skill (automation recipe) by name.",
        "parameters": [
            {"name": "name", "type": "string", "description": "Skill name.", "required": True},
            {"name": "params", "type": "object", "description": "Skill parameters.", "required": False},
        ],
    },
]

# ---------------------------------------------------------------------------
# System prompt
# ---------------------------------------------------------------------------

SYSTEM_PROMPT_TEMPLATE = """\
You are Nucleus Agent, an AI assistant that operates inside the Nucleus shell engine.
You have access to a powerful set of tools that let you execute shell commands, inspect
context, search history, assess risk, manage sessions, and run skills.

Current context:
{context}

Guidelines:
- Always assess risk before running destructive or elevated commands.
- Prefer dry-run first for high-risk operations, then execute if the user approves.
- Use sessions to group related work.
- Explain what you are doing and why before executing commands.
- If a command fails, inspect the error and try a different approach.
- Be concise but thorough.
- Never execute commands that could cause irreversible damage without explicit confirmation.
- When the goal is achieved, summarize what was done.
"""


class NucleusAgent:
    """Orchestrates an LLM provider with the Nucleus shell engine in an agentic loop."""

    def __init__(
        self,
        provider: BaseProvider,
        nucleus_client: NucleusClient | None = None,
        auto_approve_risk: RiskLevel = RiskLevel.LOW,
        max_steps: int = 50,
    ) -> None:
        self.provider = provider
        self.nucleus = nucleus_client or NucleusClient()
        self.auto_approve_risk = auto_approve_risk
        self.max_steps = max_steps
        self._console = Console()

    # ------------------------------------------------------------------
    # Tool dispatch
    # ------------------------------------------------------------------

    async def _dispatch_tool(self, name: str, arguments: dict[str, Any]) -> str:
        """Execute a tool call and return the JSON-serialised result."""
        try:
            if name == "execute_command":
                command = arguments["command"]
                dry_run = arguments.get("dry_run", False)

                # Risk gate
                if not dry_run:
                    risk = await self.nucleus.get_risk_assessment(command)
                    if risk.risk_level > self.auto_approve_risk:
                        return json.dumps(
                            {
                                "blocked": True,
                                "reason": f"Command risk ({risk.risk_level.value}) exceeds auto-approve threshold ({self.auto_approve_risk.value}). "
                                f"Reasons: {', '.join(risk.reasons)}. "
                                "Re-run with dry_run=true or request explicit approval.",
                            }
                        )

                result = await self.nucleus.execute(
                    command=command,
                    dry_run=dry_run,
                    session_id=arguments.get("session_id"),
                )
                return result.model_dump_json()

            elif name == "get_context":
                ctx = await self.nucleus.get_context()
                return ctx.model_dump_json()

            elif name == "get_history":
                entries = await self.nucleus.get_history(
                    limit=arguments.get("limit", 20),
                    search=arguments.get("search"),
                )
                return json.dumps([e.model_dump(mode="json") for e in entries])

            elif name == "rollback":
                result = await self.nucleus.rollback(arguments["execution_id"])
                return result.model_dump_json()

            elif name == "search_history":
                entries = await self.nucleus.search_history(
                    query=arguments["query"],
                    limit=arguments.get("limit", 10),
                )
                return json.dumps([e.model_dump(mode="json") for e in entries])

            elif name == "get_risk_assessment":
                risk = await self.nucleus.get_risk_assessment(arguments["command"])
                return risk.model_dump_json()

            elif name == "create_session":
                session = await self.nucleus.create_session(
                    name=arguments.get("name", ""),
                    tags=arguments.get("tags"),
                )
                return session.model_dump_json()

            elif name == "run_skill":
                result = await self.nucleus.run_skill(
                    name=arguments["name"],
                    params=arguments.get("params"),
                )
                return result.model_dump_json()

            else:
                return json.dumps({"error": f"Unknown tool: {name}"})

        except Exception as exc:
            return json.dumps({"error": str(exc)})

    # ------------------------------------------------------------------
    # Context builder
    # ------------------------------------------------------------------

    async def _build_system_prompt(self) -> str:
        try:
            ctx = await self.nucleus.get_context()
            context_str = (
                f"Directory: {ctx.working_directory}\n"
                f"User: {ctx.user}@{ctx.hostname}\n"
                f"Shell: {ctx.shell}\n"
            )
            if ctx.git_branch:
                context_str += f"Git: {ctx.git_repo or 'repo'} on {ctx.git_branch}\n"
            if ctx.recent_commands:
                context_str += f"Recent: {', '.join(ctx.recent_commands[:5])}\n"
        except Exception:
            context_str = "(Nucleus context unavailable — the engine may not be running.)"
        return SYSTEM_PROMPT_TEMPLATE.format(context=context_str)

    # ------------------------------------------------------------------
    # Main agentic loop
    # ------------------------------------------------------------------

    async def run(
        self,
        goal: str,
        session_id: str | None = None,
    ) -> AgentResult:
        """Run the agent to completion for the given goal."""
        start = time.monotonic()
        session_id = session_id or uuid.uuid4().hex[:12]
        system_prompt = await self._build_system_prompt()
        executions: list[ExecutionResult] = []

        messages: list[dict[str, Any]] = [
            {"role": "user", "content": goal},
        ]

        steps = 0
        last_content = ""

        while steps < self.max_steps:
            steps += 1

            completion: CompletionResult = await self.provider.complete(
                messages=messages,
                system=system_prompt,
                tools=NUCLEUS_TOOLS,
                max_tokens=4096,
            )

            last_content = completion.content

            # No tool calls — the model is done
            if not completion.tool_calls:
                break

            # Build assistant message with content + tool use indication
            assistant_msg: dict[str, Any] = {"role": "assistant", "content": completion.content}
            messages.append(assistant_msg)

            # Process each tool call
            for tc in completion.tool_calls:
                tool_name = tc["name"]
                tool_args = tc.get("arguments", {})
                tool_id = tc.get("id", f"tool_{steps}")

                result_str = await self._dispatch_tool(tool_name, tool_args)

                # Track executions
                if tool_name == "execute_command":
                    try:
                        exec_data = json.loads(result_str)
                        if "execution_id" in exec_data:
                            executions.append(ExecutionResult(**exec_data))
                    except (json.JSONDecodeError, Exception):
                        pass

                # Append tool result as the next message
                messages.append(
                    {
                        "role": "tool_result",
                        "tool_use_id": tool_id,
                        "content": result_str,
                    }
                )

        elapsed = time.monotonic() - start

        return AgentResult(
            goal=goal,
            success=True,
            summary=last_content,
            steps_taken=steps,
            executions=executions,
            session_id=session_id,
            total_duration_s=round(elapsed, 2),
        )

    # ------------------------------------------------------------------
    # Interactive (rich terminal) mode
    # ------------------------------------------------------------------

    async def run_interactive(self, goal: str) -> None:
        """Run the agent with rich terminal streaming output."""
        console = self._console

        console.print(
            Panel(
                f"[bold cyan]Goal:[/bold cyan] {goal}",
                title="Nucleus Agent",
                border_style="bright_blue",
            )
        )

        session_id = uuid.uuid4().hex[:12]
        system_prompt = await self._build_system_prompt()

        messages: list[dict[str, Any]] = [
            {"role": "user", "content": goal},
        ]

        steps = 0

        while steps < self.max_steps:
            steps += 1

            with Live(Spinner("dots", text="Thinking..."), console=console, transient=True):
                completion = await self.provider.complete(
                    messages=messages,
                    system=system_prompt,
                    tools=NUCLEUS_TOOLS,
                    max_tokens=4096,
                )

            if completion.content:
                console.print(
                    Panel(
                        Markdown(completion.content),
                        title=f"[bold]Agent (step {steps})[/bold]",
                        border_style="green",
                    )
                )

            if not completion.tool_calls:
                break

            assistant_msg: dict[str, Any] = {"role": "assistant", "content": completion.content}
            messages.append(assistant_msg)

            for tc in completion.tool_calls:
                tool_name = tc["name"]
                tool_args = tc.get("arguments", {})
                tool_id = tc.get("id", f"tool_{steps}")

                console.print(
                    Text.assemble(
                        ("  Tool: ", "bold yellow"),
                        (tool_name, "cyan"),
                        ("  ", ""),
                        (json.dumps(tool_args, indent=None), "dim"),
                    )
                )

                with Live(Spinner("dots", text=f"Running {tool_name}..."), console=console, transient=True):
                    result_str = await self._dispatch_tool(tool_name, tool_args)

                # Show truncated result
                display = result_str[:500] + ("..." if len(result_str) > 500 else "")
                console.print(
                    Panel(display, title="Result", border_style="dim", expand=False)
                )

                messages.append(
                    {
                        "role": "tool_result",
                        "tool_use_id": tool_id,
                        "content": result_str,
                    }
                )

        console.print(
            Panel(
                f"[bold green]Completed in {steps} step(s)[/bold green]",
                border_style="bright_green",
            )
        )
