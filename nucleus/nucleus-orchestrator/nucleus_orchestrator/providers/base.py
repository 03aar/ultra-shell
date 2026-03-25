"""Abstract base for LLM providers."""

from __future__ import annotations

import abc
from typing import Any

from nucleus_orchestrator.types import CompletionResult


class BaseProvider(abc.ABC):
    """Every LLM provider must implement these methods."""

    @abc.abstractmethod
    async def complete(
        self,
        messages: list[dict[str, Any]],
        system: str | None = None,
        tools: list[dict[str, Any]] | None = None,
        max_tokens: int = 4096,
    ) -> CompletionResult:
        """Send a chat-completion request and return a structured result."""

    @abc.abstractmethod
    def get_name(self) -> str:
        """Return the human-readable provider name."""

    @abc.abstractmethod
    def supports_tool_use(self) -> bool:
        """Whether this provider natively supports tool/function calling."""
