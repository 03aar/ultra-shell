"""Claude / Anthropic provider implementation."""

from __future__ import annotations

import os
from typing import Any

import anthropic

from nucleus_orchestrator.providers.base import BaseProvider
from nucleus_orchestrator.types import CompletionResult


def _nucleus_tools_to_anthropic(tools: list[dict[str, Any]]) -> list[dict[str, Any]]:
    """Convert Nucleus tool definitions to Anthropic tool_use format."""
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
                "name": tool["name"],
                "description": tool.get("description", ""),
                "input_schema": schema,
            }
        )
    return converted


class ClaudeProvider(BaseProvider):
    """Anthropic Claude provider with full tool-use support."""

    def __init__(
        self,
        model: str = "claude-sonnet-4-5-20250514",
        api_key: str | None = None,
    ) -> None:
        self._model = model
        self._client = anthropic.AsyncAnthropic(
            api_key=api_key or os.environ.get("ANTHROPIC_API_KEY", ""),
        )

    def get_name(self) -> str:
        return f"claude ({self._model})"

    def supports_tool_use(self) -> bool:
        return True

    async def complete(
        self,
        messages: list[dict[str, Any]],
        system: str | None = None,
        tools: list[dict[str, Any]] | None = None,
        max_tokens: int = 4096,
    ) -> CompletionResult:
        kwargs: dict[str, Any] = {
            "model": self._model,
            "max_tokens": max_tokens,
            "messages": self._format_messages(messages),
        }
        if system:
            kwargs["system"] = system
        if tools:
            kwargs["tools"] = _nucleus_tools_to_anthropic(tools)

        response = await self._client.messages.create(**kwargs)

        text_parts: list[str] = []
        tool_calls: list[dict[str, Any]] = []
        for block in response.content:
            if block.type == "text":
                text_parts.append(block.text)
            elif block.type == "tool_use":
                tool_calls.append(
                    {
                        "id": block.id,
                        "name": block.name,
                        "arguments": block.input,
                    }
                )

        return CompletionResult(
            content="\n".join(text_parts),
            tool_calls=tool_calls,
            finish_reason=response.stop_reason or "",
            model=response.model,
            usage={
                "input_tokens": response.usage.input_tokens,
                "output_tokens": response.usage.output_tokens,
            },
            raw=response.model_dump(),
        )

    @staticmethod
    def _format_messages(messages: list[dict[str, Any]]) -> list[dict[str, Any]]:
        """Ensure messages conform to Anthropic's expected structure."""
        formatted: list[dict[str, Any]] = []
        for msg in messages:
            role = msg.get("role", "user")
            if role == "system":
                # Anthropic system messages are handled via the system kwarg;
                # if one slips in here, convert to user message.
                role = "user"

            content = msg.get("content")
            if isinstance(content, str):
                formatted.append({"role": role, "content": content})
            elif isinstance(content, list):
                # Already block-formatted (tool_result, images, etc.)
                formatted.append({"role": role, "content": content})
            else:
                formatted.append({"role": role, "content": str(content)})
        return formatted
