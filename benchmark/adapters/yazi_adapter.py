"""`yazi` — the real adapter, driving a running yazi server via `yazictl`.

It writes structured `basic` memories and recalls them with yazi's deterministic
filters (by tag/scope), then ranks the candidates client-side (zero tokens, the
way an agent would). Memory-system cost is $0: no LLM, no embeddings.

Requirements to run:
  - a yazi server reachable on the configured host/port (default 127.0.0.1:3456)
  - `yazictl` on PATH (or set the YAZICTL env var to its path)

Each adapter instance uses a unique --tenant namespace so repeated runs don't
collide. Latency includes CLI process-spawn overhead (a measurement caveat, not
a property of the server); cost is the headline metric for yazi, not latency.
"""

from __future__ import annotations

import json
import os
import shutil
import subprocess
import time

from adapters.keyword_store import _tokens  # reuse the same tokenizer for client-side ranking
from framework.adapter import IngestResult, MemoryAdapter, MemoryItem, Query, RecallResult, RetrievedItem
from framework.metrics import OpMetrics, estimate_tokens

_BIN = os.environ.get("YAZICTL", "yazictl")
_HOST = os.environ.get("YAZI_HOST", "127.0.0.1")
_PORT = os.environ.get("YAZI_PORT", "3456")
_counter = 0


def _next_tenant() -> str:
    global _counter
    _counter += 1
    return f"bench-{os.getpid()}-{_counter}"


class YaziAdapter(MemoryAdapter):
    name = "yazi"
    is_estimate = False
    footprint = "single Go binary (RAM -> LSM -> S3)"
    requires = ["yazi server", "yazictl"]

    def __init__(self) -> None:
        self.tenant = _next_tenant()

    # --- availability ---
    @property
    def available(self) -> bool:
        if shutil.which(_BIN) is None and not os.path.exists(_BIN):
            return False
        try:
            rc = subprocess.run(
                [_BIN, "keys"], capture_output=True, timeout=5,
            ).returncode
            return rc == 0
        except Exception:
            return False

    def setup(self) -> None:
        self.tenant = _next_tenant()

    # --- helpers ---
    def _run(self, args: list[str], timeout: float = 15) -> subprocess.CompletedProcess:
        return subprocess.run(
            [_BIN, "--tenant", self.tenant, *args],
            capture_output=True, text=True, timeout=timeout,
        )

    # --- operations ---
    def ingest(self, item: MemoryItem) -> IngestResult:
        payload = json.dumps(
            {
                "id": item.id,
                "kind": item.kind or "context",
                "scope": item.scope or "user",
                "subject": item.subject,
                "content": item.text,
                "tags": item.tags,
            }
        )
        t0 = time.perf_counter()
        self._run(["memory", "basic", "put", "--json", payload])
        dt = (time.perf_counter() - t0) * 1000
        return IngestResult(metrics=OpMetrics(latency_ms=dt))  # zero token cost

    def recall(self, query: Query) -> RecallResult:
        flt = {"tag": query.tag, "limit": 50} if query.tag else {"scope": "user", "limit": 50}
        t0 = time.perf_counter()
        proc = self._run(["memory", "basic", "list", "--filter", json.dumps(flt)])
        dt = (time.perf_counter() - t0) * 1000

        records = []
        out = (proc.stdout or "").strip()
        if out:
            try:
                records = json.loads(out)
            except json.JSONDecodeError:
                records = []

        # client-side keyword ranking over the candidate set (zero tokens)
        q = _tokens(query.text)
        scored = []
        for r in records:
            text = r.get("content", "")
            doc = _tokens(text + " " + r.get("subject", "") + " " + " ".join(r.get("tags", [])))
            overlap = len(q & doc)
            if overlap:
                scored.append((overlap / (len(q) or 1), r))
        scored.sort(key=lambda x: x[0], reverse=True)
        items = [
            RetrievedItem(id=r.get("id", ""), text=r.get("content", ""), score=s)
            for s, r in scored[: query.top_k]
        ]
        ctx = sum(estimate_tokens(it.text) for it in items)
        return RecallResult(items=items, metrics=OpMetrics(latency_ms=dt, context_tokens=ctx))
