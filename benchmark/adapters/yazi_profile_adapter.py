"""`yazi-<profile>` — real adapters that drive Yazi's cost-aware recall profiles.

Unlike the `sim-*` estimators, these call the actual engine: ingest writes typed
memories, and recall invokes `yazictl memory recall --profile <profile>`, parsing
the REAL metered Usage (embed tokens, cost) the server reports. This lets the
benchmark place Yazi's own profiles (student vs standard) on the accuracy-vs-cost
frontier with ground-truth numbers, not estimates.

Requires a running yazi server + yazictl (same as the `yazi` adapter).
"""

from __future__ import annotations

import json
import time

from adapters.yazi_adapter import YaziAdapter
from framework.adapter import Query, RecallResult, RetrievedItem
from framework.metrics import OpMetrics


class YaziProfileAdapter(YaziAdapter):
    is_estimate = False

    def __init__(self, profile: str) -> None:
        super().__init__()
        self.profile = profile
        self.name = f"yazi-{profile}"
        self.footprint = f"single Go binary (profile={profile})"

    def recall(self, query: Query) -> RecallResult:
        args = [
            "memory", "recall",
            "--query", query.text,
            "--profile", self.profile,
            "--top-k", str(query.top_k or 5),
        ]
        if query.tag:
            args += ["--tag", query.tag]

        t0 = time.perf_counter()
        proc = self._run(args)
        dt = (time.perf_counter() - t0) * 1000

        items: list[RetrievedItem] = []
        m = OpMetrics(latency_ms=dt)
        out = (proc.stdout or "").strip()
        if out:
            try:
                d = json.loads(out)
                for h in d.get("hits", []):
                    items.append(RetrievedItem(id=h.get("id", ""), text=h.get("text", ""), score=h.get("score", 0.0)))
                u = d.get("usage", {})
                m.llm_input_tokens = u.get("llmInputTokens", 0)
                m.llm_output_tokens = u.get("llmOutputTokens", 0)
                m.embed_tokens = u.get("embedTokens", 0)
                m.context_tokens = u.get("contextTokens", 0)
            except json.JSONDecodeError:
                pass
        return RecallResult(items=items, metrics=m)
