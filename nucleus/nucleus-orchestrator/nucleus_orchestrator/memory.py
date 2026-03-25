"""Persistent memory layer for agents across sessions."""

from __future__ import annotations
from dataclasses import dataclass, field
from typing import Any, Optional
import json

from .client import NucleusClient


@dataclass
class ProjectContext:
    """Inferred project context from recent sessions."""
    detected_stack: str = "unknown"
    common_commands: list[str] = field(default_factory=list)
    common_failures: list[dict[str, str]] = field(default_factory=list)
    project_structure: list[str] = field(default_factory=list)
    team_members: list[str] = field(default_factory=list)


@dataclass
class AgentPreferences:
    """Learned preferences from agent history."""
    preferred_test_command: str = ""
    preferred_build_command: str = ""
    preferred_deploy_flow: list[str] = field(default_factory=list)
    packages_used: list[str] = field(default_factory=list)


class AgentMemory:
    """Persistent memory for AI agents across Nucleus sessions.

    Stores key-value pairs via tagged execution records in Nucleus,
    enabling agents to remember context across sessions.
    """

    def __init__(self, agent_id: str, nucleus: NucleusClient) -> None:
        self.agent_id = agent_id
        self.nucleus = nucleus
        self._cache: dict[str, Any] = {}

    async def remember(self, key: str, value: Any) -> None:
        """Store a value in persistent memory."""
        self._cache[key] = value
        encoded = json.dumps({"_memory": True, "agent": self.agent_id, "key": key, "value": value})
        await self.nucleus.execute(
            f"echo '{encoded}' > /dev/null  # nucleus-memory:{self.agent_id}:{key}",
            dry_run=False,
        )

    async def recall(self, key: str) -> Any:
        """Retrieve a stored value from memory."""
        if key in self._cache:
            return self._cache[key]

        results = await self.nucleus.search_history(
            f"nucleus-memory:{self.agent_id}:{key}", limit=1
        )
        if results:
            try:
                for r in results:
                    cmd = r.get("command", {}).get("raw", "")
                    if f"nucleus-memory:{self.agent_id}:{key}" in cmd:
                        start = cmd.find("echo '") + 6
                        end = cmd.find("' > /dev/null")
                        if start > 5 and end > start:
                            data = json.loads(cmd[start:end])
                            self._cache[key] = data.get("value")
                            return self._cache[key]
            except (json.JSONDecodeError, KeyError, IndexError):
                pass
        return None

    async def get_project_context(self) -> ProjectContext:
        """Infer project context from recent session history."""
        ctx = ProjectContext()

        try:
            history = await self.nucleus.get_history(limit=200)
        except Exception:
            return ctx

        command_counts: dict[str, int] = {}
        failure_commands: list[dict[str, str]] = []

        for entry in history:
            cmd_data = entry if isinstance(entry, dict) else {}
            raw = cmd_data.get("command", {}).get("raw", "")
            binary = cmd_data.get("command", {}).get("binary", "")
            exit_code = cmd_data.get("exit_code", 0)

            if binary:
                command_counts[binary] = command_counts.get(binary, 0) + 1

            if exit_code != 0 and raw:
                failure_commands.append({
                    "command": raw,
                    "stderr": cmd_data.get("stderr", "")[:200],
                })

        # Detect stack
        if "npm" in command_counts or "node" in command_counts or "npx" in command_counts:
            ctx.detected_stack = "node"
        elif "python" in command_counts or "pip" in command_counts or "pytest" in command_counts:
            ctx.detected_stack = "python"
        elif "cargo" in command_counts or "rustc" in command_counts:
            ctx.detected_stack = "rust"
        elif "go" in command_counts:
            ctx.detected_stack = "go"

        # Most common commands
        sorted_cmds = sorted(command_counts.items(), key=lambda x: x[1], reverse=True)
        ctx.common_commands = [cmd for cmd, _ in sorted_cmds[:15]]

        # Failures
        ctx.common_failures = failure_commands[:10]

        return ctx

    async def get_preferences(self) -> AgentPreferences:
        """Learn agent preferences from history."""
        prefs = AgentPreferences()

        try:
            history = await self.nucleus.get_history(limit=200)
        except Exception:
            return prefs

        for entry in history:
            cmd_data = entry if isinstance(entry, dict) else {}
            raw = cmd_data.get("command", {}).get("raw", "")
            binary = cmd_data.get("command", {}).get("binary", "")
            exit_code = cmd_data.get("exit_code", 0)

            if exit_code == 0:
                if binary in ("pytest", "jest", "mocha", "vitest") or "test" in raw:
                    prefs.preferred_test_command = raw
                if binary in ("make", "npm", "cargo", "go") and "build" in raw:
                    prefs.preferred_build_command = raw
                if binary in ("pip", "npm", "cargo") and "install" in raw:
                    pkg = raw.split("install")[-1].strip().split()[0] if "install" in raw else ""
                    if pkg and pkg not in prefs.packages_used:
                        prefs.packages_used.append(pkg)

        return prefs
