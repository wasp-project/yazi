"""`sim-*` — cost ESTIMATORS for LLM-based memory systems.

These adapters do NOT call mem0/zep/letta/memos/mem9. They reuse the shared
keyword store for retrieval (so accuracy is comparable) and attach a documented
per-system COST PROFILE that models the recurring token/embedding spend each
system incurs per write and per recall (see ../../STATE-OF-ART.md).

They exist so the benchmark runs end-to-end with zero external setup and shows
the COST CONTRAST immediately. Results from sim-* are marked `is_estimate=true`.
Swap in the real adapters (adapters/external.py) for ground-truth numbers.

The multipliers below are illustrative and tunable; they are not measurements.
"""

from __future__ import annotations

import time
from dataclasses import dataclass, field

from adapters.keyword_store import KeywordStore
from framework.adapter import IngestResult, MemoryAdapter, MemoryItem, Query, RecallResult
from framework.metrics import OpMetrics, estimate_tokens


@dataclass
class CostProfile:
    label: str
    footprint: str
    requires: list[str] = field(default_factory=list)
    # ingest (C1 extraction + C2 embedding)
    extract_input_mult: float = 0.0    # LLM input tokens per ingested token
    extract_output_ratio: float = 0.0  # output tokens as fraction of extraction input
    embed_on_write: bool = False
    # recall (C2 query embed + C4 retrieval LLM)
    embed_on_query: bool = False
    recall_llm: bool = False
    recall_input_mult: float = 0.0     # LLM input per retrieved context token
    recall_output_tokens: int = 0
    # downstream prompt context (C4); KV-cache reuse (C5) lowers it
    context_reuse: float = 0.0         # 0..1 reduction in injected context tokens
    # simulated hot-path latency (ms)
    ingest_latency_ms: float = 0.0
    recall_latency_ms: float = 0.0


# Profiles grounded in the qualitative cost descriptions in STATE-OF-ART.md.
PROFILES: dict[str, CostProfile] = {
    "mem0": CostProfile(
        label="sim-mem0", footprint="vector + graph DB (PG + Neo4j)", requires=["LLM", "embeddings", "Postgres", "Neo4j"],
        extract_input_mult=2.0, extract_output_ratio=0.30, embed_on_write=True,
        embed_on_query=True, ingest_latency_ms=350, recall_latency_ms=60,
    ),
    "zep": CostProfile(
        label="sim-zep", footprint="temporal graph DB", requires=["LLM", "embeddings", "graph DB"],
        extract_input_mult=2.5, extract_output_ratio=0.35, embed_on_write=True,
        embed_on_query=True, ingest_latency_ms=20, recall_latency_ms=70,  # ingest is async (low hot-path latency, cost still paid)
    ),
    "letta": CostProfile(
        label="sim-letta", footprint="server + Postgres/pgvector", requires=["LLM (in loop)", "embeddings", "Postgres"],
        extract_input_mult=1.5, extract_output_ratio=0.40, embed_on_write=True,
        embed_on_query=True, recall_llm=True, recall_input_mult=1.2, recall_output_tokens=80,
        ingest_latency_ms=500, recall_latency_ms=900,  # memory edits/searches are LLM tool calls
    ),
    "memos": CostProfile(
        label="sim-memos", footprint="graph + vector + KV-cache", requires=["LLM", "embeddings", "Neo4j", "Qdrant"],
        extract_input_mult=2.0, extract_output_ratio=0.30, embed_on_write=True,
        embed_on_query=True, context_reuse=0.50,  # KV-cache reuse cuts injected context
        ingest_latency_ms=300, recall_latency_ms=40,
    ),
    "mem9": CostProfile(
        label="sim-mem9", footprint="TiDB (vector + full-text)", requires=["LLM", "embeddings", "TiDB"],
        extract_input_mult=2.0, extract_output_ratio=0.30, embed_on_write=True,
        embed_on_query=True, ingest_latency_ms=250, recall_latency_ms=55,
    ),
}


class SimulatedAdapter(MemoryAdapter):
    is_estimate = True

    def __init__(self, system: str) -> None:
        if system not in PROFILES:
            raise KeyError(f"unknown sim profile: {system} (have {sorted(PROFILES)})")
        self.profile = PROFILES[system]
        self.name = self.profile.label
        self.footprint = self.profile.footprint
        self.requires = self.profile.requires
        self._store = KeywordStore()

    def setup(self) -> None:
        self._store.clear()

    def ingest(self, item: MemoryItem) -> IngestResult:
        t0 = time.perf_counter()
        self._store.add(item)
        wall = (time.perf_counter() - t0) * 1000
        p = self.profile
        base = estimate_tokens(item.text)
        llm_in = round(base * p.extract_input_mult)
        llm_out = round(llm_in * p.extract_output_ratio)
        embed = base if p.embed_on_write else 0
        m = OpMetrics(
            latency_ms=wall + p.ingest_latency_ms,
            llm_input_tokens=llm_in,
            llm_output_tokens=llm_out,
            embed_tokens=embed,
        )
        return IngestResult(metrics=m)

    def recall(self, query: Query) -> RecallResult:
        t0 = time.perf_counter()
        items = self._store.search(query.text, query.top_k, query.tag)
        wall = (time.perf_counter() - t0) * 1000
        p = self.profile
        ctx_raw = sum(estimate_tokens(it.text) for it in items)
        ctx = round(ctx_raw * (1 - p.context_reuse))
        embed = estimate_tokens(query.text) if p.embed_on_query else 0
        llm_in = 0
        llm_out = 0
        if p.recall_llm:
            llm_in = round(ctx_raw * p.recall_input_mult) + estimate_tokens(query.text)
            llm_out = p.recall_output_tokens
        m = OpMetrics(
            latency_ms=wall + p.recall_latency_ms,
            llm_input_tokens=llm_in,
            llm_output_tokens=llm_out,
            embed_tokens=embed,
            context_tokens=ctx,
        )
        return RecallResult(items=items, metrics=m)
