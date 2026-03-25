"""Multi-agent orchestration — multiple AI agents working together via Nucleus."""

from __future__ import annotations
import asyncio
from dataclasses import dataclass, field
from typing import Any

from .agent import NucleusAgent
from .client import NucleusClient
from .types import AgentResult, RiskLevel


@dataclass
class SubTask:
    """A sub-task assigned to an agent."""
    id: str
    description: str
    agent_index: int
    depends_on: list[str] = field(default_factory=list)
    result: AgentResult | None = None
    status: str = "pending"  # pending, running, completed, failed


@dataclass
class SwarmResult:
    """Result of a multi-agent swarm execution."""
    goal: str
    subtasks: list[SubTask]
    total_commands: int
    total_duration_ms: int
    success: bool
    summary: str


class NucleusSwarm:
    """Multi-agent orchestration — multiple AI agents working on a shared goal.

    Decomposes goals into parallel subtasks, assigns to agents based on
    capability, coordinates via shared Nucleus session, and aggregates results.

    Example:
        agents = [NucleusAgent(claude_provider, nucleus), NucleusAgent(gemini_provider, nucleus)]
        swarm = NucleusSwarm(agents, nucleus)
        result = await swarm.run("refactor this codebase for performance")
    """

    def __init__(
        self,
        agents: list[NucleusAgent],
        nucleus: NucleusClient,
        max_parallel: int = 3,
    ) -> None:
        self.agents = agents
        self.nucleus = nucleus
        self.max_parallel = max_parallel

    async def run(self, goal: str, session_id: str | None = None) -> SwarmResult:
        """Execute a goal using multiple agents in parallel.

        1. Decompose goal into subtasks using the first agent as coordinator
        2. Assign subtasks to agents based on dependency graph
        3. Execute independent subtasks in parallel
        4. Aggregate results
        """
        if not self.agents:
            return SwarmResult(
                goal=goal,
                subtasks=[],
                total_commands=0,
                total_duration_ms=0,
                success=False,
                summary="No agents available",
            )

        # Create shared session
        if session_id is None:
            try:
                session = await self.nucleus.create_session(
                    f"swarm-{goal[:30].replace(' ', '-')}"
                )
                session_id = session.get("id", "")
            except Exception:
                session_id = "swarm-default"

        # Step 1: Decompose goal into subtasks using coordinator agent
        coordinator = self.agents[0]
        subtasks = await self._decompose_goal(coordinator, goal)

        if not subtasks:
            # Single agent fallback
            result = await coordinator.run(goal, session_id=session_id)
            return SwarmResult(
                goal=goal,
                subtasks=[SubTask(id="0", description=goal, agent_index=0, result=result)],
                total_commands=result.commands_executed,
                total_duration_ms=0,
                success=result.success if hasattr(result, 'success') else True,
                summary=result.final_response if hasattr(result, 'final_response') else str(result),
            )

        # Step 2: Execute subtasks respecting dependencies
        completed: set[str] = set()
        total_commands = 0

        while any(t.status == "pending" for t in subtasks):
            # Find ready tasks (all dependencies completed)
            ready = [
                t for t in subtasks
                if t.status == "pending"
                and all(dep in completed for dep in t.depends_on)
            ]

            if not ready:
                # Deadlock — mark remaining as failed
                for t in subtasks:
                    if t.status == "pending":
                        t.status = "failed"
                break

            # Execute ready tasks in parallel (up to max_parallel)
            batch = ready[:self.max_parallel]
            tasks = []
            for subtask in batch:
                subtask.status = "running"
                agent = self.agents[subtask.agent_index % len(self.agents)]
                tasks.append(self._run_subtask(agent, subtask, session_id))

            results = await asyncio.gather(*tasks, return_exceptions=True)

            for subtask, result in zip(batch, results):
                if isinstance(result, Exception):
                    subtask.status = "failed"
                    subtask.result = None
                else:
                    subtask.status = "completed"
                    subtask.result = result
                    completed.add(subtask.id)
                    if hasattr(result, 'commands_executed'):
                        total_commands += result.commands_executed

        # Step 3: Aggregate
        all_success = all(t.status == "completed" for t in subtasks)
        summaries = []
        for t in subtasks:
            status_icon = "done" if t.status == "completed" else "FAILED"
            summaries.append(f"[{status_icon}] {t.description}")

        return SwarmResult(
            goal=goal,
            subtasks=subtasks,
            total_commands=total_commands,
            total_duration_ms=0,
            success=all_success,
            summary="\n".join(summaries),
        )

    async def _decompose_goal(
        self, coordinator: NucleusAgent, goal: str
    ) -> list[SubTask]:
        """Use the coordinator agent to decompose a goal into subtasks."""
        decompose_prompt = (
            f"Break this goal into 2-5 independent subtasks that can be "
            f"executed by different AI agents. Return ONLY a JSON array of "
            f"objects with fields: id (string), description (string), "
            f"depends_on (array of id strings). Goal: {goal}"
        )

        try:
            result = await coordinator.run(decompose_prompt)
            response = result.final_response if hasattr(result, 'final_response') else str(result)

            # Try to parse JSON from response
            import json
            # Find JSON array in response
            start = response.find('[')
            end = response.rfind(']')
            if start >= 0 and end > start:
                tasks_data = json.loads(response[start:end + 1])
                subtasks = []
                for i, task in enumerate(tasks_data):
                    subtasks.append(SubTask(
                        id=str(task.get("id", str(i))),
                        description=task.get("description", ""),
                        agent_index=i % len(self.agents),
                        depends_on=task.get("depends_on", []),
                    ))
                return subtasks
        except Exception:
            pass

        # Fallback: single task
        return [SubTask(id="0", description=goal, agent_index=0)]

    async def _run_subtask(
        self, agent: NucleusAgent, subtask: SubTask, session_id: str
    ) -> AgentResult:
        """Execute a single subtask with an agent."""
        return await agent.run(subtask.description, session_id=session_id)
