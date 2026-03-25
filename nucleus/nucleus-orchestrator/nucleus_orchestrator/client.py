"""Nucleus API client — HTTP + WebSocket interface to the Nucleus shell engine."""

from __future__ import annotations

import asyncio
import json
import uuid
from typing import Any, Callable, Coroutine, Sequence

import httpx
import websockets
import websockets.asyncio.client

from nucleus_orchestrator.types import (
    ContextSnapshot,
    ExecutionNode,
    ExecutionResult,
    GraphData,
    RiskAssessment,
    RiskLevel,
    RollbackResult,
    Session,
    SkillResult,
)


class NucleusClient:
    """Async client for the Nucleus shell engine API."""

    def __init__(
        self,
        api_url: str = "http://localhost:8080/api/v1",
        api_key: str = "dev-nucleus-key-local",
    ) -> None:
        self._api_url = api_url.rstrip("/")
        self._api_key = api_key
        self._http: httpx.AsyncClient | None = None

    # -- lifecycle --------------------------------------------------------

    async def _client(self) -> httpx.AsyncClient:
        if self._http is None or self._http.is_closed:
            self._http = httpx.AsyncClient(
                base_url=self._api_url,
                headers={
                    "Authorization": f"Bearer {self._api_key}",
                    "Content-Type": "application/json",
                },
                timeout=httpx.Timeout(30.0, connect=10.0),
            )
        return self._http

    async def close(self) -> None:
        if self._http and not self._http.is_closed:
            await self._http.aclose()

    # -- helpers ----------------------------------------------------------

    async def _get(self, path: str, params: dict[str, Any] | None = None) -> dict[str, Any]:
        client = await self._client()
        resp = await client.get(path, params=params)
        resp.raise_for_status()
        return resp.json()

    async def _post(self, path: str, body: dict[str, Any] | None = None) -> dict[str, Any]:
        client = await self._client()
        resp = await client.post(path, json=body or {})
        resp.raise_for_status()
        return resp.json()

    # -- commands ---------------------------------------------------------

    async def execute(
        self,
        command: str,
        dry_run: bool = False,
        session_id: str | None = None,
    ) -> ExecutionResult:
        """Execute a shell command through Nucleus."""
        payload: dict[str, Any] = {"command": command, "dry_run": dry_run}
        if session_id:
            payload["session_id"] = session_id
        data = await self._post("/execute", payload)
        return ExecutionResult(**data)

    async def get_context(self) -> ContextSnapshot:
        """Retrieve the current shell context."""
        data = await self._get("/context")
        return ContextSnapshot(**data)

    async def get_history(
        self,
        limit: int = 50,
        session_id: str | None = None,
        search: str | None = None,
    ) -> list[ExecutionResult]:
        """Fetch command history."""
        params: dict[str, Any] = {"limit": limit}
        if session_id:
            params["session_id"] = session_id
        if search:
            params["search"] = search
        data = await self._get("/history", params=params)
        items = data if isinstance(data, list) else data.get("items", [])
        return [ExecutionResult(**item) for item in items]

    async def rollback(self, execution_id: str) -> RollbackResult:
        """Rollback a previous execution."""
        data = await self._post(f"/rollback/{execution_id}")
        return RollbackResult(**data)

    async def get_graph(self, session_id: str | None = None) -> GraphData:
        """Retrieve the execution dependency graph."""
        params: dict[str, Any] = {}
        if session_id:
            params["session_id"] = session_id
        data = await self._get("/graph", params=params)
        nodes = [ExecutionNode(**n) for n in data.get("nodes", [])]
        edges = [(e[0], e[1]) for e in data.get("edges", [])]
        return GraphData(
            nodes=nodes,
            edges=edges,
            session_id=session_id or "",
        )

    async def create_session(
        self,
        name: str = "",
        tags: Sequence[str] | None = None,
    ) -> Session:
        """Create a new Nucleus session."""
        payload: dict[str, Any] = {
            "name": name or f"session-{uuid.uuid4().hex[:8]}",
        }
        if tags:
            payload["tags"] = list(tags)
        data = await self._post("/sessions", payload)
        return Session(**data)

    async def search_history(
        self,
        query: str,
        limit: int = 20,
    ) -> list[ExecutionResult]:
        """Semantic search over command history."""
        data = await self._post("/history/search", {"query": query, "limit": limit})
        items = data if isinstance(data, list) else data.get("items", [])
        return [ExecutionResult(**item) for item in items]

    async def run_skill(
        self,
        name: str,
        params: dict[str, Any] | None = None,
    ) -> SkillResult:
        """Run a named Nucleus skill."""
        data = await self._post(f"/skills/{name}/run", params or {})
        return SkillResult(**data)

    async def get_risk_assessment(self, command: str) -> RiskAssessment:
        """Get a risk assessment for a command before executing."""
        data = await self._post("/risk", {"command": command})
        return RiskAssessment(**data)

    # -- websocket --------------------------------------------------------

    async def stream(
        self,
        on_event: Callable[[dict[str, Any]], Coroutine[Any, Any, None]],
        event_types: Sequence[str] | None = None,
        duration: float | None = None,
    ) -> None:
        """Subscribe to Nucleus events over WebSocket.

        Parameters
        ----------
        on_event:
            Async callback invoked for each event dict.
        event_types:
            Optional filter — only forward events whose ``type`` field matches.
        duration:
            If set, disconnect after *duration* seconds.
        """
        ws_url = self._api_url.replace("http://", "ws://").replace("https://", "wss://")
        ws_url = f"{ws_url}/ws"

        async def _run() -> None:
            async with websockets.asyncio.client.connect(
                ws_url,
                additional_headers={"Authorization": f"Bearer {self._api_key}"},
            ) as ws:
                # Send subscription filter if requested
                if event_types:
                    await ws.send(json.dumps({"subscribe": list(event_types)}))

                async for raw in ws:
                    try:
                        event = json.loads(raw)
                    except json.JSONDecodeError:
                        continue
                    if event_types and event.get("type") not in event_types:
                        continue
                    await on_event(event)

        if duration is not None:
            try:
                await asyncio.wait_for(_run(), timeout=duration)
            except asyncio.TimeoutError:
                pass
        else:
            await _run()
