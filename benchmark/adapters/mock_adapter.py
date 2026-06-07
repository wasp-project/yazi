"""`mock` — a zero-cost keyword memory store.

Represents the *class* yazi belongs to: deterministic, indexed retrieval with no
LLM and no embeddings, so its memory-system cost is $0. Runs anywhere with no
dependencies. Use it as the cheap baseline and as a smoke test for the harness.
"""

from __future__ import annotations

import time

from adapters.keyword_store import KeywordStore
from framework.adapter import IngestResult, MemoryAdapter, MemoryItem, Query, RecallResult
from framework.metrics import OpMetrics, estimate_tokens


class MockAdapter(MemoryAdapter):
    name = "mock"
    is_estimate = False
    footprint = "in-process (single binary class)"
    requires: list[str] = []

    def __init__(self) -> None:
        self._store = KeywordStore()

    def setup(self) -> None:
        self._store.clear()

    def ingest(self, item: MemoryItem) -> IngestResult:
        t0 = time.perf_counter()
        self._store.add(item)
        dt = (time.perf_counter() - t0) * 1000
        # No LLM, no embeddings: zero token cost.
        return IngestResult(metrics=OpMetrics(latency_ms=dt))

    def recall(self, query: Query) -> RecallResult:
        t0 = time.perf_counter()
        items = self._store.search(query.text, query.top_k, query.tag)
        dt = (time.perf_counter() - t0) * 1000
        ctx = sum(estimate_tokens(it.text) for it in items)
        return RecallResult(items=items, metrics=OpMetrics(latency_ms=dt, context_tokens=ctx))
