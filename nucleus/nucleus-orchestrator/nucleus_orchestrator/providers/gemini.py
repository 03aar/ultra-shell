"""Google Gemini provider implementation."""

from __future__ import annotations

import os
from typing import Any

import httpx

from nucleus_orchestrator.providers.base import BaseProvider
from nucleus_orchestrator.types import CompletionResult


def _nucleus_tools_to_gemini(tools: list[dict[str, Any]]) -> list[dict[str, Any]]:
    """Convert Nucleus tool definitions to Gemini function_declarations format."""
    declarations: list[dict[str, Any]] = []
    for tool in tools:
        properties: dict[str, Any] = {}
        required: list[str] = []
        for param in tool.get("parameters", []):
            prop: dict[str, Any] = {
                "type": param.get("type", "STRING").upper(),
                "description": param.get("description", ""),
            }
            if "enum" in param:
                prop["enum"] = param["enum"]
            properties[param["name"]] = prop
            if param.get("required", False):
                required.append(param["name"])

        schema: dict[str, Any] = {
            "type": "OBJECT",
            "properties": properties,
        }
        if required:
            schema["required"] = required

        declarations.append({
            "name": tool["name"],
            "description": tool.get("description", ""),
            "parameters": schema,
        })
    return declarations


class GeminiProvider(BaseProvider):
    """Google Gemini provider with function calling support."""

    def __init__(
        self,
        model: str = "gemini-2.0-flash",
        api_key: str | None = None,
    ) -> None:
        self._model = model
        self._api_key = api_key or os.environ.get("GOOGLE_API_KEY", "")
        self._base_url = "https://generativelanguage.googleapis.com/v1beta"

    def get_name(self) -> str:
        return f"gemini ({self._model})"

    def supports_tool_use(self) -> bool:
        return True

    async def complete(
        self,
        messages: list[dict[str, Any]],
        system: str | None = None,
        tools: list[dict[str, Any]] | None = None,
        max_tokens: int = 4096,
    ) -> CompletionResult:
        url = f"{self._base_url}/models/{self._model}:generateContent?key={self._api_key}"

        contents = self._format_messages(messages)

        body: dict[str, Any] = {
            "contents": contents,
            "generationConfig": {
                "maxOutputTokens": max_tokens,
                "temperature": 0.1,
            },
        }

        if system:
            body["systemInstruction"] = {
                "parts": [{"text": system}]
            }

        if tools:
            body["tools"] = [{
                "functionDeclarations": _nucleus_tools_to_gemini(tools)
            }]

        async with httpx.AsyncClient(timeout=30.0) as client:
            resp = await client.post(url, json=body)
            resp.raise_for_status()
            data = resp.json()

        text_parts: list[str] = []
        tool_calls: list[dict[str, Any]] = []

        candidates = data.get("candidates", [])
        if candidates:
            content = candidates[0].get("content", {})
            for part in content.get("parts", []):
                if "text" in part:
                    text_parts.append(part["text"])
                elif "functionCall" in part:
                    fc = part["functionCall"]
                    tool_calls.append({
                        "id": fc["name"],
                        "name": fc["name"],
                        "arguments": fc.get("args", {}),
                    })

        usage_meta = data.get("usageMetadata", {})

        return CompletionResult(
            content="\n".join(text_parts),
            tool_calls=tool_calls,
            finish_reason=candidates[0].get("finishReason", "") if candidates else "",
            model=self._model,
            usage={
                "input_tokens": usage_meta.get("promptTokenCount", 0),
                "output_tokens": usage_meta.get("candidatesTokenCount", 0),
            },
            raw=data,
        )

    @staticmethod
    def _format_messages(messages: list[dict[str, Any]]) -> list[dict[str, Any]]:
        """Convert messages to Gemini contents format."""
        contents: list[dict[str, Any]] = []
        for msg in messages:
            role = msg.get("role", "user")
            if role == "system":
                role = "user"
            elif role == "assistant":
                role = "model"

            content_val = msg.get("content", "")
            if isinstance(content_val, str):
                parts = [{"text": content_val}]
            elif isinstance(content_val, list):
                parts = []
                for item in content_val:
                    if isinstance(item, dict) and "text" in item:
                        parts.append({"text": item["text"]})
                    elif isinstance(item, str):
                        parts.append({"text": item})
                    else:
                        parts.append({"text": str(item)})
            else:
                parts = [{"text": str(content_val)}]

            contents.append({"role": role, "parts": parts})
        return contents
