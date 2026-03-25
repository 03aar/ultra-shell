"""PlanExecutor — runs an ExecutionPlan step-by-step through Nucleus."""

from __future__ import annotations

import json
from typing import Any, Callable, Coroutine

from nucleus_orchestrator.client import NucleusClient
from nucleus_orchestrator.types import (
    ExecutionPlan,
    ExecutionResult,
    PlanStep,
    RiskLevel,
)

# Callback type for progress/approval hooks
StepCallback = Callable[[PlanStep, str], Coroutine[Any, Any, bool]]


class PlanExecutor:
    """Execute an :class:`ExecutionPlan` through the Nucleus client."""

    async def execute(
        self,
        plan: ExecutionPlan,
        nucleus: NucleusClient,
        auto_approve: RiskLevel = RiskLevel.LOW,
        callbacks: StepCallback | None = None,
    ) -> list[ExecutionResult]:
        """Walk through each step in the plan and execute it.

        Parameters
        ----------
        plan:
            The validated execution plan.
        nucleus:
            A connected :class:`NucleusClient`.
        auto_approve:
            Risk threshold below which commands execute without asking.
        callbacks:
            Optional async callback ``(step, phase) -> bool``.
            Called with phase ``"before"`` prior to execution and ``"after"``
            once complete. Returning ``False`` from a ``"before"`` callback
            skips the step.

        Returns
        -------
        list[ExecutionResult]
            Results for each executed step (skipped steps are omitted).
        """
        results: list[ExecutionResult] = []
        completed_ids: set[int] = set()

        # Topological order: steps are expected to be ordered already,
        # but we respect depends_on by waiting for dependencies.
        pending = list(plan.steps)

        while pending:
            ready = [
                s for s in pending
                if all(dep in completed_ids for dep in s.depends_on)
            ]
            if not ready:
                # Remaining steps have unsatisfiable deps (predecessor failed/skipped)
                break

            for step in ready:
                pending.remove(step)

                # --- before callback / approval gate ---
                if callbacks:
                    proceed = await callbacks(step, "before")
                    if not proceed:
                        completed_ids.add(step.step_id)
                        step.completed = True
                        continue

                # --- risk gate ---
                if step.command and step.risk_level > auto_approve:
                    # Without an interactive callback we skip dangerous steps
                    if callbacks is None:
                        completed_ids.add(step.step_id)
                        step.completed = True
                        continue
                    proceed = await callbacks(step, "approve")
                    if not proceed:
                        completed_ids.add(step.step_id)
                        step.completed = True
                        continue

                # --- execute ---
                result: ExecutionResult | None = None

                if step.command:
                    result = await nucleus.execute(command=step.command)
                elif step.tool:
                    result = await self._dispatch_tool(step.tool, step.tool_params, nucleus)

                if result is not None:
                    step.result = result
                    results.append(result)

                    # If a step fails, attempt rollback of previous steps
                    if result.exit_code != 0:
                        await self._rollback_completed(results[:-1], nucleus)
                        step.completed = True
                        completed_ids.add(step.step_id)
                        break

                step.completed = True
                completed_ids.add(step.step_id)

                if callbacks:
                    await callbacks(step, "after")

        return results

    # ------------------------------------------------------------------

    @staticmethod
    async def _dispatch_tool(
        tool: str,
        params: dict[str, Any],
        nucleus: NucleusClient,
    ) -> ExecutionResult:
        """Map a plan-step tool invocation to a NucleusClient method."""
        if tool == "execute_command":
            return await nucleus.execute(
                command=params.get("command", "echo 'no command'"),
                dry_run=params.get("dry_run", False),
            )

        if tool == "run_skill":
            skill_result = await nucleus.run_skill(
                name=params.get("name", ""),
                params=params.get("params"),
            )
            return ExecutionResult(
                command=f"skill:{params.get('name', '')}",
                exit_code=0 if skill_result.success else 1,
                stdout=skill_result.output,
            )

        if tool == "get_context":
            ctx = await nucleus.get_context()
            return ExecutionResult(
                command="get_context",
                exit_code=0,
                stdout=ctx.model_dump_json(),
            )

        if tool == "search_history":
            entries = await nucleus.search_history(
                query=params.get("query", ""),
                limit=params.get("limit", 10),
            )
            return ExecutionResult(
                command="search_history",
                exit_code=0,
                stdout=json.dumps([e.model_dump(mode="json") for e in entries]),
            )

        if tool == "create_session":
            session = await nucleus.create_session(
                name=params.get("name", ""),
                tags=params.get("tags"),
            )
            return ExecutionResult(
                command="create_session",
                exit_code=0,
                stdout=session.model_dump_json(),
            )

        # Fallback
        return ExecutionResult(
            command=f"unknown_tool:{tool}",
            exit_code=1,
            stderr=f"Unknown tool: {tool}",
        )

    @staticmethod
    async def _rollback_completed(
        results: list[ExecutionResult],
        nucleus: NucleusClient,
    ) -> None:
        """Best-effort rollback of previously completed steps (reverse order)."""
        for result in reversed(results):
            if result.execution_id:
                try:
                    await nucleus.rollback(result.execution_id)
                except Exception:
                    pass  # best effort
