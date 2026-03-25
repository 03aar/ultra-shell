"""Pydantic models for nucleus-orchestrator."""

from __future__ import annotations

import enum
from datetime import datetime
from typing import Any

from pydantic import BaseModel, Field


class RiskLevel(str, enum.Enum):
    """Risk severity classification for shell commands."""

    NONE = "none"
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"

    def __ge__(self, other: "RiskLevel") -> bool:
        order = list(RiskLevel)
        return order.index(self) >= order.index(other)

    def __gt__(self, other: "RiskLevel") -> bool:
        order = list(RiskLevel)
        return order.index(self) > order.index(other)

    def __le__(self, other: "RiskLevel") -> bool:
        order = list(RiskLevel)
        return order.index(self) <= order.index(other)

    def __lt__(self, other: "RiskLevel") -> bool:
        order = list(RiskLevel)
        return order.index(self) < order.index(other)


class RiskAssessment(BaseModel):
    """Risk analysis result for a command."""

    command: str
    risk_level: RiskLevel = RiskLevel.NONE
    score: float = Field(default=0.0, ge=0.0, le=1.0)
    reasons: list[str] = Field(default_factory=list)
    suggestions: list[str] = Field(default_factory=list)
    requires_approval: bool = False


class ExecutionResult(BaseModel):
    """Result of executing a command through Nucleus."""

    execution_id: str = ""
    command: str = ""
    exit_code: int = 0
    stdout: str = ""
    stderr: str = ""
    duration_ms: float = 0.0
    risk_level: RiskLevel = RiskLevel.NONE
    session_id: str = ""
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    dry_run: bool = False
    metadata: dict[str, Any] = Field(default_factory=dict)


class ContextSnapshot(BaseModel):
    """Current shell context from Nucleus."""

    working_directory: str = ""
    user: str = ""
    hostname: str = ""
    shell: str = ""
    environment: dict[str, str] = Field(default_factory=dict)
    git_branch: str | None = None
    git_repo: str | None = None
    git_status: str | None = None
    recent_commands: list[str] = Field(default_factory=list)
    active_session: str | None = None


class ExecutionNode(BaseModel):
    """A node in the execution dependency graph."""

    execution_id: str = ""
    command: str = ""
    exit_code: int = 0
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    parent_id: str | None = None
    children: list[str] = Field(default_factory=list)
    tags: list[str] = Field(default_factory=list)
    session_id: str = ""


class RollbackResult(BaseModel):
    """Result of a rollback operation."""

    execution_id: str = ""
    rollback_command: str = ""
    success: bool = False
    message: str = ""
    steps_undone: int = 0


class GraphData(BaseModel):
    """Execution dependency graph."""

    nodes: list[ExecutionNode] = Field(default_factory=list)
    edges: list[tuple[str, str]] = Field(default_factory=list)
    session_id: str = ""


class Session(BaseModel):
    """A Nucleus session."""

    session_id: str = ""
    name: str = ""
    created_at: datetime = Field(default_factory=datetime.utcnow)
    tags: list[str] = Field(default_factory=list)
    command_count: int = 0
    active: bool = True


class SkillResult(BaseModel):
    """Result of running a Nucleus skill."""

    skill_name: str = ""
    success: bool = False
    output: str = ""
    steps_completed: int = 0
    steps_total: int = 0
    artifacts: dict[str, Any] = Field(default_factory=dict)


class PlanStep(BaseModel):
    """A single step within an execution plan."""

    step_id: int = 0
    description: str = ""
    command: str | None = None
    tool: str | None = None
    tool_params: dict[str, Any] = Field(default_factory=dict)
    depends_on: list[int] = Field(default_factory=list)
    risk_level: RiskLevel = RiskLevel.NONE
    estimated_duration_s: float = 0.0
    rollback_command: str | None = None
    completed: bool = False
    result: ExecutionResult | None = None


class ExecutionPlan(BaseModel):
    """A plan decomposed from a high-level goal."""

    goal: str = ""
    steps: list[PlanStep] = Field(default_factory=list)
    total_risk: RiskLevel = RiskLevel.NONE
    estimated_duration_s: float = 0.0
    context_summary: str = ""
    validated: bool = False
    validation_errors: list[str] = Field(default_factory=list)


class CompletionResult(BaseModel):
    """Result from an LLM completion call."""

    content: str = ""
    tool_calls: list[dict[str, Any]] = Field(default_factory=list)
    finish_reason: str = ""
    model: str = ""
    usage: dict[str, int] = Field(default_factory=dict)
    raw: dict[str, Any] = Field(default_factory=dict)


class AgentResult(BaseModel):
    """Final result from an agent run."""

    goal: str = ""
    success: bool = False
    summary: str = ""
    steps_taken: int = 0
    executions: list[ExecutionResult] = Field(default_factory=list)
    plan: ExecutionPlan | None = None
    session_id: str = ""
    total_duration_s: float = 0.0
    error: str | None = None
