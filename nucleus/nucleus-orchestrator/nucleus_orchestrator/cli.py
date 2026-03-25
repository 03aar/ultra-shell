"""Typer CLI for nucleus-orchestrator."""

from __future__ import annotations

import asyncio
from typing import Annotated, Optional

import typer
from rich.console import Console
from rich.panel import Panel
from rich.table import Table

from nucleus_orchestrator.agent import NucleusAgent
from nucleus_orchestrator.client import NucleusClient
from nucleus_orchestrator.planner import GoalPlanner
from nucleus_orchestrator.types import RiskLevel

app = typer.Typer(
    name="nucleus-agent",
    help="AI agent orchestration layer for the Nucleus shell engine.",
    no_args_is_help=True,
)
console = Console()


def _resolve_provider(provider_name: str, model: str | None = None):
    """Instantiate the requested LLM provider.

    Supported providers:
      - claude / anthropic  (requires ANTHROPIC_API_KEY)
      - openai / gpt        (requires OPENAI_API_KEY)
      - gemini / google     (requires GOOGLE_API_KEY)
      - ollama              (requires local Ollama server, no API key)
    """
    import os
    from nucleus_orchestrator.providers.claude import ClaudeProvider
    from nucleus_orchestrator.providers.gemini import GeminiProvider
    from nucleus_orchestrator.providers.ollama import OllamaProvider
    from nucleus_orchestrator.providers.openai_provider import OpenAIProvider

    name = provider_name.lower()
    if name in ("claude", "anthropic"):
        if not os.environ.get("ANTHROPIC_API_KEY"):
            console.print("[yellow]Warning: ANTHROPIC_API_KEY not set. Set it in .env or environment.[/yellow]")
        return ClaudeProvider(model=model or "claude-sonnet-4-5-20250514")
    elif name in ("openai", "gpt"):
        if not os.environ.get("OPENAI_API_KEY"):
            console.print("[yellow]Warning: OPENAI_API_KEY not set. Set it in .env or environment.[/yellow]")
        return OpenAIProvider(model=model or "gpt-4o")
    elif name in ("gemini", "google"):
        if not os.environ.get("GOOGLE_API_KEY"):
            console.print("[yellow]Warning: GOOGLE_API_KEY not set. Set it in .env or environment.[/yellow]")
        return GeminiProvider(model=model or "gemini-2.0-flash")
    elif name == "ollama":
        return OllamaProvider(model=model or "llama3")
    else:
        console.print(f"[red]Unknown provider: {provider_name}[/red]")
        console.print("[dim]Available: claude, openai, gemini, ollama[/dim]")
        raise typer.Exit(1)


@app.command()
def run(
    goal: Annotated[str, typer.Argument(help="The goal for the agent to accomplish.")],
    provider: Annotated[str, typer.Option("--provider", "-p", help="LLM provider.")] = "claude",
    model: Annotated[Optional[str], typer.Option("--model", "-m", help="Model override.")] = None,
    interactive: Annotated[bool, typer.Option("--interactive/--no-interactive", "-i/-I", help="Interactive rich output.")] = True,
    session: Annotated[Optional[str], typer.Option("--session", "-s", help="Session ID.")] = None,
    auto_approve: Annotated[str, typer.Option("--auto-approve", "-a", help="Max risk to auto-approve (none/low/medium/high/critical).")] = "low",
) -> None:
    """Run the Nucleus agent to accomplish a goal."""
    try:
        risk = RiskLevel(auto_approve.lower())
    except ValueError:
        console.print(f"[red]Invalid risk level: {auto_approve}[/red]")
        raise typer.Exit(1)

    llm = _resolve_provider(provider, model)
    nucleus = NucleusClient()
    agent = NucleusAgent(
        provider=llm,
        nucleus_client=nucleus,
        auto_approve_risk=risk,
    )

    async def _run() -> None:
        try:
            if interactive:
                await agent.run_interactive(goal)
            else:
                result = await agent.run(goal, session_id=session)
                console.print(Panel(result.summary, title="Result", border_style="green"))
                console.print(f"Steps: {result.steps_taken}  Duration: {result.total_duration_s}s")
        finally:
            await nucleus.close()

    asyncio.run(_run())


@app.command()
def plan(
    goal: Annotated[str, typer.Argument(help="The goal to plan for.")],
    provider: Annotated[str, typer.Option("--provider", "-p")] = "claude",
    model: Annotated[Optional[str], typer.Option("--model", "-m")] = None,
) -> None:
    """Generate an execution plan without running it."""
    llm = _resolve_provider(provider, model)
    nucleus = NucleusClient()
    planner = GoalPlanner()

    async def _plan() -> None:
        try:
            try:
                ctx = await nucleus.get_context()
            except Exception:
                ctx = None

            execution_plan = await planner.decompose(goal, context=ctx, provider=llm)
            execution_plan = await planner.validate(execution_plan, nucleus)

            table = Table(title=f"Plan: {execution_plan.goal}", show_lines=True)
            table.add_column("#", style="bold", width=4)
            table.add_column("Description", min_width=30)
            table.add_column("Command/Tool", min_width=20)
            table.add_column("Risk", width=10)
            table.add_column("Deps", width=8)

            for step in execution_plan.steps:
                cmd = step.command or f"tool:{step.tool}" if step.tool else "-"
                risk_style = {
                    RiskLevel.NONE: "dim",
                    RiskLevel.LOW: "green",
                    RiskLevel.MEDIUM: "yellow",
                    RiskLevel.HIGH: "red",
                    RiskLevel.CRITICAL: "bold red",
                }.get(step.risk_level, "")
                table.add_row(
                    str(step.step_id),
                    step.description,
                    cmd,
                    f"[{risk_style}]{step.risk_level.value}[/{risk_style}]",
                    ",".join(str(d) for d in step.depends_on) or "-",
                )

            console.print(table)

            if execution_plan.validation_errors:
                console.print("[bold red]Validation errors:[/bold red]")
                for err in execution_plan.validation_errors:
                    console.print(f"  - {err}")
            else:
                console.print("[bold green]Plan validated successfully.[/bold green]")

            console.print(
                f"Total risk: {execution_plan.total_risk.value}  "
                f"Estimated: {execution_plan.estimated_duration_s}s"
            )
        finally:
            await nucleus.close()

    asyncio.run(_plan())


@app.command("providers")
def list_providers() -> None:
    """List available LLM providers."""
    table = Table(title="Available Providers")
    table.add_column("Name", style="bold cyan")
    table.add_column("Models")
    table.add_column("Tool Use")
    table.add_column("Env Var")

    table.add_row(
        "claude",
        "claude-sonnet-4-5, claude-opus-4, claude-3-haiku, ...",
        "[green]Yes[/green]",
        "ANTHROPIC_API_KEY",
    )
    table.add_row(
        "openai",
        "gpt-4o, gpt-4-turbo, gpt-3.5-turbo, ...",
        "[green]Yes[/green]",
        "OPENAI_API_KEY",
    )
    table.add_row(
        "ollama",
        "llama3, mistral, codellama, (any local model)",
        "[green]Yes[/green]",
        "(none — local)",
    )
    console.print(table)


@app.command("test-provider")
def test_provider(
    provider: Annotated[str, typer.Argument(help="Provider to test.")],
    model: Annotated[Optional[str], typer.Option("--model", "-m")] = None,
) -> None:
    """Test connectivity to an LLM provider."""
    llm = _resolve_provider(provider, model)

    async def _test() -> None:
        console.print(f"Testing [bold]{llm.get_name()}[/bold] ...")
        try:
            result = await llm.complete(
                messages=[{"role": "user", "content": "Say 'hello' in one word."}],
                max_tokens=32,
            )
            console.print(f"[green]Success![/green] Response: {result.content.strip()}")
            console.print(f"Model: {result.model}  Tokens: {result.usage}")
        except Exception as exc:
            console.print(f"[red]Failed:[/red] {exc}")
            raise typer.Exit(1)

    asyncio.run(_test())


if __name__ == "__main__":
    app()
