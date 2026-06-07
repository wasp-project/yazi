"""Adapter interface — the seam every memory system plugs into.

A benchmark run drives each adapter through the same lifecycle:

    setup() -> ingest(item) * N -> recall(query) * M -> teardown()

Adapters report per-operation metrics (latency, tokens). The runner times the
calls as a fallback, but adapters should fill in token usage from their own SDK
responses so the cost model is accurate.
"""

from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass, field

from framework.metrics import OpMetrics


@dataclass
class MemoryItem:
    """A piece of memory to store."""

    id: str
    text: str
    kind: str = "context"        # preference | history | decision | context
    scope: str = "user"
    subject: str = ""
    tags: list[str] = field(default_factory=list)


@dataclass
class Query:
    """A recall request plus the ground truth used to score it."""

    text: str
    expected_ids: list[str] = field(default_factory=list)
    expected_substring: str = ""
    tag: str = ""                # optional structured filter hint
    top_k: int = 5


@dataclass
class RetrievedItem:
    id: str
    text: str
    score: float = 0.0


@dataclass
class RecallResult:
    items: list[RetrievedItem]
    metrics: OpMetrics


@dataclass
class IngestResult:
    metrics: OpMetrics


class MemoryAdapter(ABC):
    """Base class for all memory-system adapters."""

    #: display name, e.g. "yazi", "mem0", "sim-zep"
    name: str = "base"
    #: True for cost *estimators* (no real service is called)
    is_estimate: bool = False
    #: short operational-footprint label for reporting (C6)
    footprint: str = "n/a"
    #: external infra/services this adapter needs to run for real
    requires: list[str] = []

    @property
    def available(self) -> bool:
        """Whether this adapter can actually run in the current environment."""
        return True

    def setup(self) -> None:  # noqa: B027 - optional hook
        """Prepare a clean store. Called once before a suite."""

    def teardown(self) -> None:  # noqa: B027 - optional hook
        """Release resources / clear state. Called once after a suite."""

    @abstractmethod
    def ingest(self, item: MemoryItem) -> IngestResult:
        """Store one memory item and report the cost of doing so."""

    @abstractmethod
    def recall(self, query: Query) -> RecallResult:
        """Retrieve memories for a query and report the cost of doing so."""
