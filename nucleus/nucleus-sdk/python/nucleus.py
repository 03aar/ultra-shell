"""
NUCLEUS Python SDK

A Python client for the NUCLEUS AI-native shell runtime API.
Provides typed access to executions, context, sessions, rollback, and real-time streaming.

Usage:
    from nucleus import NucleusClient

    client = NucleusClient("http://localhost:8080/api/v1", "dev-nucleus-key-local")
    executions = client.get_executions(limit=10)
    context = client.get_context()
"""

from __future__ import annotations

import json
import threading
from dataclasses import dataclass, field
from typing import Any, Callable, Dict, List, Optional
from urllib.request import Request, urlopen
from urllib.error import URLError, HTTPError
from urllib.parse import urlencode


@dataclass
class RiskFlag:
    """A risk flag associated with a command execution."""
    code: str
    message: str
    level: str  # none, low, medium, high, critical


@dataclass
class ParsedCommand:
    """A parsed shell command."""
    raw: str
    binary: str
    args: List[str]
    flags: List[str]
    pipes: List[Dict[str, Any]]
    redirections: List[Dict[str, str]]
    env_assignments: Dict[str, str]
    category: str
    is_background: bool
    is_chained: bool


@dataclass
class ExecutionNode:
    """Represents a single command execution in the NUCLEUS runtime."""
    id: str
    timestamp: str
    command: ParsedCommand
    exit_code: int
    duration_ms: int
    stdout: str
    stderr: str
    files_mutated: List[str]
    env_changes: Dict[str, str]
    processes_spawned: List[int]
    risk_flags: List[RiskFlag]
    rollback_available: bool
    rolled_back: bool
    session_id: str

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> ExecutionNode:
        cmd = data.get("command", {})
        return cls(
            id=data["id"],
            timestamp=data["timestamp"],
            command=ParsedCommand(
                raw=cmd.get("raw", ""),
                binary=cmd.get("binary", ""),
                args=cmd.get("args", []),
                flags=cmd.get("flags", []),
                pipes=cmd.get("pipes", []),
                redirections=cmd.get("redirections", []),
                env_assignments=cmd.get("env_assignments", {}),
                category=cmd.get("category", "unknown"),
                is_background=cmd.get("is_background", False),
                is_chained=cmd.get("is_chained", False),
            ),
            exit_code=data.get("exit_code", -1),
            duration_ms=data.get("duration_ms", 0),
            stdout=data.get("stdout", ""),
            stderr=data.get("stderr", ""),
            files_mutated=data.get("files_mutated", []),
            env_changes=data.get("env_changes", {}),
            processes_spawned=data.get("processes_spawned", []),
            risk_flags=[
                RiskFlag(code=f["code"], message=f["message"], level=f["level"])
                for f in data.get("risk_flags", [])
            ],
            rollback_available=data.get("rollback_available", False),
            rolled_back=data.get("rolled_back", False),
            session_id=data.get("session_id", ""),
        )


@dataclass
class ContextState:
    """Current environment state from the NUCLEUS runtime."""
    cwd: str
    git_branch: str
    git_status: str
    running_processes: List[Dict[str, Any]]
    env_vars: Dict[str, str]
    disk_usage: Dict[str, Any]
    open_files_count: int

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> ContextState:
        return cls(
            cwd=data.get("cwd", ""),
            git_branch=data.get("git_branch", ""),
            git_status=data.get("git_status", ""),
            running_processes=data.get("running_processes", []),
            env_vars=data.get("env_vars", {}),
            disk_usage=data.get("disk_usage", {}),
            open_files_count=data.get("open_files_count", 0),
        )


@dataclass
class Edge:
    """A dependency edge between execution nodes."""
    from_id: str
    to_id: str
    dependency_type: str

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> Edge:
        return cls(
            from_id=data["from"],
            to_id=data["to"],
            dependency_type=data["dependency_type"],
        )


@dataclass
class Graph:
    """The execution dependency graph."""
    nodes: List[ExecutionNode]
    edges: List[Edge]

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> Graph:
        return cls(
            nodes=[ExecutionNode.from_dict(n) for n in data.get("nodes", [])],
            edges=[Edge.from_dict(e) for e in data.get("edges", [])],
        )


@dataclass
class Session:
    """A NUCLEUS shell session."""
    id: str
    name: str
    shell_pid: int
    start_time: str
    end_time: Optional[str]
    git_repo: Optional[str]
    cwd: str
    execution_count: int = 0
    risk_warning_count: int = 0

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> Session:
        return cls(
            id=data["id"],
            name=data.get("name", ""),
            shell_pid=data.get("shell_pid", 0),
            start_time=data.get("start_time", ""),
            end_time=data.get("end_time"),
            git_repo=data.get("git_repo"),
            cwd=data.get("cwd", ""),
            execution_count=data.get("execution_count", 0),
            risk_warning_count=data.get("risk_warning_count", 0),
        )


@dataclass
class RollbackResult:
    """Result of a rollback operation."""
    success: bool
    files_restored: List[str]
    env_restored: Dict[str, str]
    message: str

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> RollbackResult:
        return cls(
            success=data.get("success", False),
            files_restored=data.get("files_restored", []),
            env_restored=data.get("env_restored", {}),
            message=data.get("message", ""),
        )


class NucleusError(Exception):
    """Exception raised for NUCLEUS API errors."""
    def __init__(self, message: str, status_code: int = 0):
        self.status_code = status_code
        super().__init__(message)


class NucleusClient:
    """Client for the NUCLEUS AI-native shell runtime API.

    Args:
        api_url: Base URL for the NUCLEUS API (e.g., "http://localhost:8080/api/v1")
        api_key: API key for authentication (e.g., "dev-nucleus-key-local")
    """

    def __init__(self, api_url: str, api_key: str) -> None:
        self.api_url = api_url.rstrip("/")
        self.api_key = api_key

    def _request(self, method: str, path: str, body: Optional[Dict] = None) -> Any:
        """Make an HTTP request to the NUCLEUS API."""
        url = f"{self.api_url}{path}"
        data = json.dumps(body).encode("utf-8") if body else None

        req = Request(url, data=data, method=method)
        req.add_header("Content-Type", "application/json")
        req.add_header("X-Nucleus-Key", self.api_key)

        try:
            with urlopen(req) as resp:
                response_data = json.loads(resp.read().decode("utf-8"))
                if response_data.get("error"):
                    raise NucleusError(response_data["error"])
                return response_data["data"]
        except HTTPError as e:
            body_text = e.read().decode("utf-8", errors="replace")
            try:
                error_data = json.loads(body_text)
                msg = error_data.get("error", str(e))
            except (json.JSONDecodeError, KeyError):
                msg = str(e)
            raise NucleusError(msg, status_code=e.code)
        except URLError as e:
            raise NucleusError(f"Connection failed: {e.reason}")

    def get_executions(
        self,
        limit: int = 50,
        offset: int = 0,
        session_id: Optional[str] = None,
        risk_level: Optional[str] = None,
        command_category: Optional[str] = None,
    ) -> List[ExecutionNode]:
        """Get recent command executions.

        Args:
            limit: Maximum number of executions to return (default 50)
            offset: Number of executions to skip
            session_id: Filter by session ID
            risk_level: Filter by risk level (none, low, medium, high, critical)
            command_category: Filter by command category

        Returns:
            List of ExecutionNode objects
        """
        params: Dict[str, str] = {"limit": str(limit), "offset": str(offset)}
        if session_id:
            params["session_id"] = session_id
        if risk_level:
            params["risk_level"] = risk_level
        if command_category:
            params["command_category"] = command_category

        qs = urlencode(params)
        data = self._request("GET", f"/executions?{qs}")
        return [ExecutionNode.from_dict(e) for e in data.get("executions", [])]

    def get_execution(self, execution_id: str) -> ExecutionNode:
        """Get full details for a specific execution.

        Args:
            execution_id: UUID of the execution

        Returns:
            ExecutionNode with full details including stdout, stderr
        """
        data = self._request("GET", f"/executions/{execution_id}")
        return ExecutionNode.from_dict(data)

    def get_context(self) -> ContextState:
        """Get current environment state.

        Returns:
            ContextState with cwd, git info, processes, env vars, disk usage
        """
        data = self._request("GET", "/context")
        return ContextState.from_dict(data)

    def get_graph(self, session_id: Optional[str] = None) -> Graph:
        """Get the execution dependency graph.

        Args:
            session_id: Optional session ID to filter the graph

        Returns:
            Graph containing nodes and edges
        """
        path = "/context/graph"
        if session_id:
            path += f"?session_id={session_id}"
        data = self._request("GET", path)
        return Graph.from_dict(data)

    def rollback(self, execution_id: str) -> RollbackResult:
        """Rollback a specific execution, restoring files and env vars.

        Args:
            execution_id: UUID of the execution to rollback

        Returns:
            RollbackResult with success status and restored items
        """
        data = self._request("POST", f"/rollback/{execution_id}")
        return RollbackResult.from_dict(data)

    def get_sessions(self) -> List[Session]:
        """Get all sessions.

        Returns:
            List of Session objects with metadata
        """
        data = self._request("GET", "/sessions")
        return [Session.from_dict(s) for s in (data or [])]

    def create_session(self, name: str) -> Session:
        """Create a new named session.

        Args:
            name: Name for the session

        Returns:
            Created Session object
        """
        data = self._request("POST", "/sessions", {"name": name})
        return Session.from_dict(data)

    def get_session_replay(self, session_id: str) -> List[ExecutionNode]:
        """Get all executions for a session in chronological order.

        Args:
            session_id: UUID of the session

        Returns:
            List of ExecutionNode objects in execution order
        """
        data = self._request("GET", f"/sessions/{session_id}/replay")
        return [ExecutionNode.from_dict(e) for e in data.get("executions", [])]

    def stream(self, callback: Callable[[str, Any], None]) -> threading.Thread:
        """Connect to the WebSocket stream for real-time events.

        Args:
            callback: Function called with (event_type, data) for each event.
                      Event types: execution_complete, risk_warning,
                      rollback_complete, session_started

        Returns:
            Thread running the WebSocket connection
        """
        try:
            import websocket
        except ImportError:
            raise ImportError(
                "websocket-client is required for streaming. "
                "Install it with: pip install websocket-client"
            )

        ws_url = self.api_url.replace("http://", "ws://").replace("https://", "wss://")
        ws_url = f"{ws_url}/ws/stream?api_key={self.api_key}"

        def on_message(ws: Any, message: str) -> None:
            try:
                msg = json.loads(message)
                callback(msg.get("type", "unknown"), msg.get("data"))
            except json.JSONDecodeError:
                callback("raw", message)

        def on_error(ws: Any, error: Any) -> None:
            callback("error", str(error))

        def run() -> None:
            ws = websocket.WebSocketApp(
                ws_url,
                on_message=on_message,
                on_error=on_error,
            )
            ws.run_forever()

        thread = threading.Thread(target=run, daemon=True)
        thread.start()
        return thread


if __name__ == "__main__":
    # Quick test
    client = NucleusClient("http://localhost:8080/api/v1", "dev-nucleus-key-local")
    try:
        ctx = client.get_context()
        print(f"CWD: {ctx.cwd}")
        print(f"Git: {ctx.git_branch}")
        execs = client.get_executions(limit=5)
        print(f"Recent executions: {len(execs)}")
        for e in execs:
            print(f"  [{e.exit_code}] {e.command.raw}")
    except NucleusError as e:
        print(f"API Error: {e}")
