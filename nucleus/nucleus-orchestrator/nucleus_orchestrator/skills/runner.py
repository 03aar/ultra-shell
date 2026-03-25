"""SkillRunner — executes the steps defined in a SkillDefinition."""

from __future__ import annotations

import re
from typing import Any

from nucleus_orchestrator.client import NucleusClient
from nucleus_orchestrator.skills.loader import SkillDefinition, SkillLoader
from nucleus_orchestrator.types import ExecutionResult, SkillResult


def _interpolate(template: str, variables: dict[str, Any]) -> str:
    """Replace ``{{var}}`` placeholders with values from *variables*."""
    def _replace(match: re.Match[str]) -> str:
        key = match.group(1).strip()
        return str(variables.get(key, match.group(0)))

    return re.sub(r"\{\{(.+?)\}\}", _replace, template)


class SkillRunner:
    """Execute skills loaded by :class:`SkillLoader` against a Nucleus client."""

    def __init__(
        self,
        nucleus: NucleusClient,
        loader: SkillLoader | None = None,
    ) -> None:
        self.nucleus = nucleus
        self.loader = loader or SkillLoader()

    async def run(
        self,
        name: str,
        params: dict[str, Any] | None = None,
    ) -> SkillResult:
        """Run a skill by name with the given parameters."""
        skill = self.loader.get(name)
        if skill is None:
            return SkillResult(
                skill_name=name,
                success=False,
                output=f"Skill {name!r} not found.",
            )
        return await self.execute_skill(skill, params or {})

    async def execute_skill(
        self,
        skill: SkillDefinition,
        params: dict[str, Any],
    ) -> SkillResult:
        """Execute all steps of a skill definition."""
        # Validate required parameters
        variables: dict[str, Any] = dict(params)
        for param_def in skill.parameters:
            pname = param_def.get("name", "")
            if param_def.get("required", False) and pname not in variables:
                if "default" in param_def:
                    variables[pname] = param_def["default"]
                else:
                    return SkillResult(
                        skill_name=skill.name,
                        success=False,
                        output=f"Missing required parameter: {pname}",
                    )
            elif pname not in variables and "default" in param_def:
                variables[pname] = param_def["default"]

        outputs: list[str] = []
        artifacts: dict[str, Any] = {}
        completed = 0
        total = len(skill.steps)
        executed_results: list[ExecutionResult] = []

        for step in skill.steps:
            step_type = step.get("type", "command")
            step_name = step.get("name", f"step-{completed + 1}")

            try:
                if step_type == "command":
                    command = _interpolate(step.get("command", ""), variables)
                    result = await self.nucleus.execute(command=command)
                    executed_results.append(result)

                    if result.exit_code != 0:
                        # Step failed — attempt rollback
                        await self._rollback(skill, executed_results, variables)
                        outputs.append(f"[FAIL] {step_name}: {result.stderr}")
                        return SkillResult(
                            skill_name=skill.name,
                            success=False,
                            output="\n".join(outputs),
                            steps_completed=completed,
                            steps_total=total,
                            artifacts=artifacts,
                        )

                    outputs.append(f"[OK] {step_name}: {result.stdout[:200]}")

                    # Capture output as variable for later steps
                    if "capture" in step:
                        variables[step["capture"]] = result.stdout.strip()

                elif step_type == "check":
                    command = _interpolate(step.get("command", ""), variables)
                    result = await self.nucleus.execute(command=command)
                    if result.exit_code != 0:
                        outputs.append(f"[CHECK FAIL] {step_name}: {result.stderr}")
                        if not step.get("continue_on_fail", False):
                            return SkillResult(
                                skill_name=skill.name,
                                success=False,
                                output="\n".join(outputs),
                                steps_completed=completed,
                                steps_total=total,
                                artifacts=artifacts,
                            )
                    else:
                        outputs.append(f"[CHECK OK] {step_name}")

                elif step_type == "set":
                    for key, val in step.get("variables", {}).items():
                        variables[key] = _interpolate(str(val), variables)
                    outputs.append(f"[SET] {step_name}")

                elif step_type == "skill":
                    # Nested skill invocation
                    nested_name = step.get("skill", "")
                    nested_params = {
                        k: _interpolate(str(v), variables)
                        for k, v in step.get("params", {}).items()
                    }
                    nested_result = await self.run(nested_name, nested_params)
                    if not nested_result.success:
                        outputs.append(f"[NESTED FAIL] {step_name}: {nested_result.output}")
                        return SkillResult(
                            skill_name=skill.name,
                            success=False,
                            output="\n".join(outputs),
                            steps_completed=completed,
                            steps_total=total,
                            artifacts=artifacts,
                        )
                    outputs.append(f"[NESTED OK] {step_name}")
                    artifacts.update(nested_result.artifacts)

                elif step_type == "artifact":
                    key = step.get("key", step_name)
                    value = _interpolate(step.get("value", ""), variables)
                    artifacts[key] = value
                    outputs.append(f"[ARTIFACT] {key}")

                else:
                    outputs.append(f"[SKIP] {step_name}: unknown type {step_type!r}")

                completed += 1

            except Exception as exc:
                outputs.append(f"[ERROR] {step_name}: {exc}")
                await self._rollback(skill, executed_results, variables)
                return SkillResult(
                    skill_name=skill.name,
                    success=False,
                    output="\n".join(outputs),
                    steps_completed=completed,
                    steps_total=total,
                    artifacts=artifacts,
                )

        return SkillResult(
            skill_name=skill.name,
            success=True,
            output="\n".join(outputs),
            steps_completed=completed,
            steps_total=total,
            artifacts=artifacts,
        )

    async def _rollback(
        self,
        skill: SkillDefinition,
        executed: list[ExecutionResult],
        variables: dict[str, Any],
    ) -> None:
        """Best-effort rollback using the skill's rollback steps or execution IDs."""
        if skill.rollback:
            for rb_step in skill.rollback:
                command = _interpolate(rb_step.get("command", ""), variables)
                if command:
                    try:
                        await self.nucleus.execute(command=command)
                    except Exception:
                        pass
        else:
            for result in reversed(executed):
                if result.execution_id:
                    try:
                        await self.nucleus.rollback(result.execution_id)
                    except Exception:
                        pass
