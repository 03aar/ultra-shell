"""OpenAI / GPT provider implementation."""

from __future__ import annotations

import json
import os
from typing import Any

import openai

from nucleus_orchestrator.providers.base import BaseProvider
from nucleus_orchestrator.types import CompletionResult


def _nucleus_tools_to_openai(tools: list[dict[str, Any]]) -> list[dict[str, Any]]:
    """Convert Nucleus tool definitions to OpenAI function-calling format."""
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
            if "default" in param:
                prop["default"] = param["default"]
            properties[param["name"]] = prop
            if param.get("required", False):
                required.append(param["name"])
        schema: dict[str, Any] = {
            "type": "object",
            "properties": properties,
        }
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


class OpenAIProvider(BaseProvider):
    """OpenAI GPT provider with function-calling support."""

    def __init__(
        self,
        model: str = "gpt-4o",
        api_key: str | None = None,
    ) -> None:
        self._model = model
        self._client = openai.AsyncOpenAI(
            api_key=api_key or os.environ.get("OPENAI_API_KEY", ""),
        )

    def get_name(self) -> str:
        return f"openai ({self._model})"

    def supports_tool_use(self) -> bool:
        return True

    async def complete(
        self,
        messages: list[dict[str, Any]],
        system: str | None = None,
        tools: list[dict[str, Any]] | None = None,
        max_tokens: int = 4096,
    ) -> CompletionResult:
        formatted = self._format_messages(messages, system)

        kwargs: dict[str, Any] = {
            "model": self._model,
            "messages": formatted,
            "max_tokens": max_tokens,
        }
        if tools:
            kwargs["tools"] = _nucleus_tools_to_openai(tools)
            kwargs["tool_choice"] = "auto"

        response = await self._client.chat.completions.create(**kwargs)
        choice = response.choices[0]
        message = choice.message

        tool_calls: list[dict[str, Any]] = []
        if message.tool_calls:
            for tc in message.tool_calls:
                try:
                    arguments = json.loads(tc.function.arguments)
                except (json.JSONDecodeError, TypeError):
                    arguments = {}
                tool_calls.append(
                    {
                        "id": tc.id,
                        "name": tc.function.name,
                        "arguments": arguments,
                    }
                )

        usage: dict[str, int] = {}
        if response.usage:
            usage = {
                "input_tokens": response.usage.prompt_tokens,
                "output_tokens": response.usage.completion_tokens,
            }

        return CompletionResult(
            content=message.content or "",
            tool_calls=tool_calls,
            finish_reason=choice.finish_reason or "",
            model=response.model or self._model,
            usage=usage,
            raw=response.model_dump(),
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
                # Map Anthropic-style tool_result back to OpenAI "tool" role
                formatted.append(
                    {
                        "role": "tool",
                        "tool_call_id": msg.get("tool_use_id", ""),
                        "content": content if isinstance(content, str) else json.dumps(content),
                    }
                )
            elif role == "assistant" and "tool_calls_raw" in msg:
                # Re-attach native tool_calls on the assistant message
                formatted.append(
                    {
                        "role": "assistant",
                        "content": content or None,
                        "tool_calls": msg["tool_calls_raw"],
                    }
                )
            else:
                formatted.append({"role": role, "content": content if isinstance(content, str) else str(content)})
        return formatted
