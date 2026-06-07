"""Real adapters for external systems — opt-in, import-guarded.

Each adapter here calls the actual system's SDK and reports REAL token usage and
latency. They require the system's Python package, a running backend, and (for
LLM-based ones) provider API keys, so they are unavailable by default and the
runner skips them with a clear message. Fill in the TODOs to enable a system.

Honesty over coverage: rather than fake these, we ship them as clearly-marked
stubs whose `available` is False until you wire them up. Use the `sim-*`
estimators (adapters/simulator.py) for a dependency-free comparison today.
"""

from __future__ import annotations

from framework.adapter import IngestResult, MemoryAdapter, MemoryItem, Query, RecallResult


class _UnavailableAdapter(MemoryAdapter):
    """Base for real adapters that are not yet wired up in this environment."""

    _import_name: str = ""

    def __init__(self) -> None:
        self._module = None

    @property
    def available(self) -> bool:
        if not self._import_name:
            return False
        try:  # pragma: no cover - depends on optional deps
            import importlib

            self._module = importlib.import_module(self._import_name)
            return self._configured()
        except Exception:
            return False

    def _configured(self) -> bool:
        """Override: check for API keys / running backend before claiming available."""
        return False

    def ingest(self, item: MemoryItem) -> IngestResult:  # pragma: no cover
        raise NotImplementedError(f"{self.name}: real adapter not wired up — see adapters/external.py")

    def recall(self, query: Query) -> RecallResult:  # pragma: no cover
        raise NotImplementedError(f"{self.name}: real adapter not wired up — see adapters/external.py")


class Mem0Adapter(_UnavailableAdapter):
    name = "mem0"
    footprint = "vector + graph DB"
    requires = ["pip install mem0ai", "LLM key", "vector store", "(optional) Neo4j"]
    _import_name = "mem0"
    # TODO: from mem0 import Memory; m = Memory(); m.add(...) / m.search(...)
    #       report usage from the LLM/embedding responses into OpMetrics.


class ZepAdapter(_UnavailableAdapter):
    name = "zep"
    footprint = "temporal graph DB"
    requires = ["pip install zep-cloud OR graphiti-core", "LLM key", "graph DB"]
    _import_name = "graphiti_core"


class LettaAdapter(_UnavailableAdapter):
    name = "letta"
    footprint = "server + Postgres/pgvector"
    requires = ["pip install letta", "LLM key", "Postgres"]
    _import_name = "letta"


class MemOSAdapter(_UnavailableAdapter):
    name = "memos"
    footprint = "graph + vector + KV-cache"
    requires = ["pip install MemoryOS / memos", "LLM key", "Neo4j", "Qdrant"]
    _import_name = "memos"


class Mem9Adapter(_UnavailableAdapter):
    name = "mem9"
    footprint = "TiDB (vector + full-text)"
    requires = ["mem9 mnemo-server", "TiDB", "LLM key"]
    _import_name = "mem9"  # likely via REST; placeholder import name


class MilvusAdapter(_UnavailableAdapter):
    name = "milvus"
    footprint = "vector DB (cluster or Lite)"
    requires = ["pip install pymilvus", "embeddings", "Milvus/Lite"]
    _import_name = "pymilvus"
    # Note: Milvus is a vector store, not a memory abstraction — pair it with an
    # embedding model; its memory-system cost is embeddings only (no extraction LLM).
