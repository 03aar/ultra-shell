"""Ollama local-model provider implementation."""

from __future__ import annotations

import json
from typing import Any

import httpx

from nucleus_orchestrator.providers.base import BaseProvider
from nucleus_orchestrator.types import CompletionResult


def _nucleus_tools_to_ollama(tools: list[dict[str, Any]]) -> list[dict[str, Any]]:
    """Convert Nucleus tool definitions to Ollama's OpenAI-compatible format."""
    converted: list[dict[str, Any]] = []
    for tool in tools:
        properties: dict[str, Any] = {}
        required: list[str] = []
        for param in tool.get("parameters", []):
            prop: dict[str, Any] = {
                "type": param.get("type", "string"),
                "description": param.get("description", ""),
            }
            if "enum" in param:
                prop["enum"] = param["enum"]
            properties[param["name"]] = prop
            if param.get("required", False):
                required.append(param["name"])
        schema: dict[str, Any] = {"type": "object", "properties": properties}
        if required:
            schema["required"] = required
        converted.append(
            {
                "type": "function",
                "function": {
                    "name": tool["name"],
                    "description": tool.get("description", ""),
                    "parameters": schema,
                },
            }
        )
    return converted


class OllamaProvider(BaseProvider):
    """Provider for locally-running Ollama models."""

    def __init__(
        self,
        model: str = "llama3",
        base_url: str = "http://localhost:11434",
    ) -> None:
        self._model = model
        self._base_url = base_url.rstrip("/")
        self._http = httpx.AsyncClient(
            base_url=self._base_url,
            timeout=httpx.Timeout(120.0, connect=10.0),
        )

    def get_name(self) -> str:
        return f"ollama ({self._model})"

    def supports_tool_use(self) -> bool:
        # Ollama supports tool calling via the chat API for compatible models
        return True

    async def complete(
        self,
        messages: list[dict[str, Any]],
        system: str | None = None,
        tools: list[dict[str, Any]] | None = None,
        max_tokens: int = 4096,
    ) -> CompletionResult:
        formatted = self._format_messages(messages, system)

        payload: dict[str, Any] = {
            "model": self._model,
            "messages": formatted,
            "stream": False,
            "options": {"num_predict": max_tokens},
        }
        if tools:
            payload["tools"] = _nucleus_tools_to_ollama(tools)

        resp = await self._http.post("/api/chat", json=payload)
        resp.raise_for_status()
        data = resp.json()

        message = data.get("message", {})
        content = message.get("content", "")

        tool_calls: list[dict[str, Any]] = []
        for tc in message.get("tool_calls", []):
            func = tc.get("function", {})
            arguments = func.get("arguments", {})
            if isinstance(arguments, str):
                try:
                    arguments = json.loads(arguments)
                except json.JSONDecodeError:
                    arguments = {}
            tool_calls.append(
                {
                    "id": f"ollama_{func.get('name', 'call')}",
                    "name": func.get("name", ""),
                    "arguments": arguments,
                }
            )

        prompt_tokens = data.get("prompt_eval_count", 0)
        completion_tokens = data.get("eval_count", 0)

        return CompletionResult(
            content=content,
            tool_calls=tool_calls,
            finish_reason="tool_calls" if tool_calls else "stop",
            model=self._model,
            usage={
                "input_tokens": prompt_tokens,
                "output_tokens": completion_tokens,
            },
            raw=data,
        )

    @staticmethod
    def _format_messages(
        messages: list[dict[str, Any]],
        system: str | None = None,
    ) -> list[dict[str, Any]]:
        formatted: list[dict[str, Any]] = []
        if system:
            formatted.append({"role": "system", "content": system})
        for msg in messages:
            role = msg.get("role", "user")
            content = msg.get("content", "")
            if role == "tool_result":
                formatted.append(
                    {
                        "role": "tool",
                        "content": content if isinstance(content, str) else json.dumps(content),
                    }
                )
            else:
                formatted.append({
                    "role": role,
                    "content": content if isinstance(content, str) else str(content),
                })
        return formatted

    async def close(self) -> None:
        await self._http.aclose()
