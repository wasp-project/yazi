"""Cost and performance metrics for the memory benchmark.

This module defines the cost model used to compare memory systems. It mirrors
the cost taxonomy in ../../STATE-OF-ART.md:

    C1 write/ingest tokens   C2 embedding tokens   C3 storage
    C4 retrieval context     C5 inference reuse     C6 operational footprint

The headline number is **memory-system cost**: the recurring $ a system spends
*producing and retrieving* memory (C1 + C2 + recall-side LLM/embeds). It does NOT
include the downstream prompt cost of injecting recalled context (C4-prompt),
which every system pays and which we report separately as `context_tokens`.
"""

from __future__ import annotations

from dataclasses import dataclass, field, asdict


def estimate_tokens(text: str) -> int:
    """Rough token estimate (~4 chars/token). Good enough for relative cost."""
    if not text:
        return 0
    return max(1, round(len(text) / 4))


@dataclass
class Pricing:
    """USD per 1M tokens. Defaults approximate a small hosted model + embeddings.

    Override via a JSON file (see Pricing.load) to re-price the whole comparison
    without touching adapters — useful for continuous runs as prices change.
    """

    llm_input_per_m: float = 0.15
    llm_output_per_m: float = 0.60
    embed_per_m: float = 0.02

    def llm_cost(self, input_tokens: int, output_tokens: int) -> float:
        return (input_tokens * self.llm_input_per_m + output_tokens * self.llm_output_per_m) / 1_000_000

    def embed_cost(self, tokens: int) -> float:
        return tokens * self.embed_per_m / 1_000_000

    @staticmethod
    def load(path: str | None) -> "Pricing":
        if not path:
            return Pricing()
        import json

        with open(path) as f:
            return Pricing(**json.load(f))


@dataclass
class OpMetrics:
    """Metrics for a single operation (one ingest or one recall)."""

    latency_ms: float = 0.0
    # tokens the memory system itself spent (extraction / consolidation / recall LLM)
    llm_input_tokens: int = 0
    llm_output_tokens: int = 0
    embed_tokens: int = 0
    # tokens of context the recall returned for downstream prompt injection (C4)
    context_tokens: int = 0

    def system_cost(self, pricing: Pricing) -> float:
        """Recurring $ the memory system spent on this op (excludes downstream prompt)."""
        return pricing.llm_cost(self.llm_input_tokens, self.llm_output_tokens) + pricing.embed_cost(self.embed_tokens)

    def add(self, other: "OpMetrics") -> None:
        self.latency_ms += other.latency_ms
        self.llm_input_tokens += other.llm_input_tokens
        self.llm_output_tokens += other.llm_output_tokens
        self.embed_tokens += other.embed_tokens
        self.context_tokens += other.context_tokens


@dataclass
class AdapterResult:
    """Aggregated result for one adapter over one or more suites."""

    adapter: str
    is_estimate: bool
    footprint: str  # e.g. "single binary", "vector DB", "graph DB + queue"
    requires: list[str] = field(default_factory=list)

    n_ingests: int = 0
    n_queries: int = 0
    n_correct: int = 0

    ingest: OpMetrics = field(default_factory=OpMetrics)
    recall: OpMetrics = field(default_factory=OpMetrics)

    # --- derived metrics ---
    @property
    def accuracy(self) -> float:
        return self.n_correct / self.n_queries if self.n_queries else 0.0

    @property
    def avg_ingest_latency_ms(self) -> float:
        return self.ingest.latency_ms / self.n_ingests if self.n_ingests else 0.0

    @property
    def avg_recall_latency_ms(self) -> float:
        return self.recall.latency_ms / self.n_queries if self.n_queries else 0.0

    def system_cost(self, pricing: Pricing) -> float:
        return self.ingest.system_cost(pricing) + self.recall.system_cost(pricing)

    def cost_per_1k_ops(self, pricing: Pricing) -> float:
        ops = self.n_ingests + self.n_queries
        return (self.system_cost(pricing) / ops * 1000) if ops else 0.0

    def cost_per_correct(self, pricing: Pricing) -> float:
        return (self.system_cost(pricing) / self.n_correct) if self.n_correct else float("inf")

    def to_dict(self, pricing: Pricing) -> dict:
        return {
            "adapter": self.adapter,
            "is_estimate": self.is_estimate,
            "footprint": self.footprint,
            "requires": self.requires,
            "n_ingests": self.n_ingests,
            "n_queries": self.n_queries,
            "n_correct": self.n_correct,
            "accuracy": round(self.accuracy, 4),
            "avg_ingest_latency_ms": round(self.avg_ingest_latency_ms, 3),
            "avg_recall_latency_ms": round(self.avg_recall_latency_ms, 3),
            "ingest_metrics": asdict(self.ingest),
            "recall_metrics": asdict(self.recall),
            "system_cost_usd": round(self.system_cost(pricing), 8),
            "cost_per_1k_ops_usd": round(self.cost_per_1k_ops(pricing), 8),
            "cost_per_correct_usd": (
                None if self.cost_per_correct(pricing) == float("inf") else round(self.cost_per_correct(pricing), 8)
            ),
            "downstream_context_tokens": self.recall.context_tokens,
        }
