"""GoalPlanner — decomposes high-level goals into executable step plans."""

from __future__ import annotations

import json
from typing import Any

from nucleus_orchestrator.client import NucleusClient
from nucleus_orchestrator.providers.base import BaseProvider
from nucleus_orchestrator.types import (
    ContextSnapshot,
    ExecutionPlan,
    PlanStep,
    RiskLevel,
)

_PLANNER_SYSTEM = """\
You are a planning engine for the Nucleus shell agent.
Given a high-level goal and the current shell context, decompose the goal into
a sequence of concrete steps that can be executed through Nucleus.

Return a JSON object with this exact structure:
{
  "goal": "<the original goal>",
  "context_summary": "<brief summary of relevant context>",
  "steps": [
    {
      "step_id": 1,
      "description": "<what this step does>",
      "command": "<shell command to run, or null if using a tool>",
      "tool": "<nucleus tool name, or null if using a command>",
      "tool_params": {},
      "depends_on": [],
      "risk_level": "none|low|medium|high|critical",
      "estimated_duration_s": 1.0,
      "rollback_command": "<command to undo, or null>"
    }
  ],
  "total_risk": "none|low|medium|high|critical",
  "estimated_duration_s": 10.0
}

Guidelines:
- Order steps logically; use depends_on to mark dependencies.
- Assess risk realistically based on what each command does.
- Include rollback commands where feasible.
- Keep steps atomic — one logical action per step.
- Always include a verification/validation step at the end.
Return ONLY the JSON, no extra text.
"""


class GoalPlanner:
    """Decomposes a natural-language goal into an ExecutionPlan."""

    async def decompose(
        self,
        goal: str,
        context: ContextSnapshot | None = None,
        provider: BaseProvider | None = None,
    ) -> ExecutionPlan:
        """Use the LLM provider to break *goal* into concrete steps."""
        if provider is None:
            raise ValueError("A provider is required for planning.")

        context_str = "No context available."
        if context:
            parts = [f"Directory: {context.working_directory}"]
            if context.git_branch:
                parts.append(f"Git branch: {context.git_branch}")
            if context.user:
                parts.append(f"User: {context.user}")
            if context.recent_commands:
                parts.append(f"Recent commands: {', '.join(context.recent_commands[:5])}")
            context_str = "\n".join(parts)

        messages: list[dict[str, Any]] = [
            {
                "role": "user",
                "content": f"Goal: {goal}\n\nCurrent context:\n{context_str}",
            },
        ]

        completion = await provider.complete(
            messages=messages,
            system=_PLANNER_SYSTEM,
            max_tokens=4096,
        )

        plan = self._parse_plan(completion.content, goal)
        return plan

    async def validate(
        self,
        plan: ExecutionPlan,
        nucleus: NucleusClient,
    ) -> ExecutionPlan:
        """Validate a plan by risk-checking each command step against Nucleus."""
        errors: list[str] = []
        max_risk = RiskLevel.NONE

        for step in plan.steps:
            if step.command:
                try:
                    risk = await nucleus.get_risk_assessment(step.command)
                    step.risk_level = risk.risk_level
                    if risk.risk_level > max_risk:
                        max_risk = risk.risk_level
                    if risk.risk_level >= RiskLevel.CRITICAL:
                        errors.append(
                            f"Step {step.step_id} ({step.command!r}) has CRITICAL risk: "
                            f"{', '.join(risk.reasons)}"
                        )
                except Exception as exc:
                    errors.append(f"Step {step.step_id}: risk check failed — {exc}")

            # Validate dependency references
            all_ids = {s.step_id for s in plan.steps}
            for dep in step.depends_on:
                if dep not in all_ids:
                    errors.append(
                        f"Step {step.step_id} depends on non-existent step {dep}"
                    )

        plan.total_risk = max_risk
        plan.validation_errors = errors
        plan.validated = len(errors) == 0
        return plan

    @staticmethod
    def _parse_plan(raw: str, goal: str) -> ExecutionPlan:
        """Parse LLM output JSON into an ExecutionPlan."""
        # Strip markdown fences if present
        cleaned = raw.strip()
        if cleaned.startswith("```"):
            lines = cleaned.split("\n")
            lines = lines[1:]  # drop opening fence
            if lines and lines[-1].strip() == "```":
                lines = lines[:-1]
            cleaned = "\n".join(lines)

        try:
            data = json.loads(cleaned)
        except json.JSONDecodeError:
            # Fallback: create a single-step plan with the raw content
            return ExecutionPlan(
                goal=goal,
                steps=[
                    PlanStep(
                        step_id=1,
                        description=f"Execute goal directly: {goal}",
                        command=None,
                        tool="execute_command",
                        tool_params={"command": goal},
                    )
                ],
                context_summary="Failed to parse plan from LLM output.",
                validated=False,
                validation_errors=["Could not parse LLM plan output as JSON."],
            )

        steps: list[PlanStep] = []
        for s in data.get("steps", []):
            risk_str = s.get("risk_level", "none")
            try:
                risk = RiskLevel(risk_str)
            except ValueError:
                risk = RiskLevel.NONE
            steps.append(
                PlanStep(
                    step_id=s.get("step_id", len(steps) + 1),
                    description=s.get("description", ""),
                    command=s.get("command"),
                    tool=s.get("tool"),
                    tool_params=s.get("tool_params", {}),
                    depends_on=s.get("depends_on", []),
                    risk_level=risk,
                    estimated_duration_s=s.get("estimated_duration_s", 0.0),
                    rollback_command=s.get("rollback_command"),
                )
            )

        total_risk_str = data.get("total_risk", "none")
        try:
            total_risk = RiskLevel(total_risk_str)
        except ValueError:
            total_risk = RiskLevel.NONE

        return ExecutionPlan(
            goal=data.get("goal", goal),
            steps=steps,
            total_risk=total_risk,
            estimated_duration_s=data.get("estimated_duration_s", 0.0),
            context_summary=data.get("context_summary", ""),
            validated=False,
        )
