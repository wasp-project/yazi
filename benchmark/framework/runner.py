"""Benchmark runner — drives adapters over suites and aggregates results."""

from __future__ import annotations

from dataclasses import dataclass, field

from evaluation.evaluators import evaluate
from framework.adapter import MemoryAdapter
from framework.metrics import AdapterResult, Pricing
from testcases.cases import Suite


@dataclass
class RunConfig:
    pricing: Pricing = field(default_factory=Pricing)
    verbose: bool = True


def _log(cfg: RunConfig, msg: str) -> None:
    if cfg.verbose:
        print(msg, flush=True)


def run_adapter(adapter: MemoryAdapter, suites: list[Suite], cfg: RunConfig) -> AdapterResult:
    res = AdapterResult(
        adapter=adapter.name,
        is_estimate=adapter.is_estimate,
        footprint=adapter.footprint,
        requires=list(adapter.requires),
    )
    for suite in suites:
        adapter.setup()
        try:
            for item in suite.items:
                r = adapter.ingest(item)
                res.ingest.add(r.metrics)
                res.n_ingests += 1
            for query in suite.queries:
                rr = adapter.recall(query)
                res.recall.add(rr.metrics)
                res.n_queries += 1
                if evaluate(query, rr.items).correct:
                    res.n_correct += 1
        finally:
            adapter.teardown()
    return res


def run(adapters: list[MemoryAdapter], suites: list[Suite], cfg: RunConfig) -> list[AdapterResult]:
    results: list[AdapterResult] = []
    suite_names = ", ".join(s.name for s in suites)
    total_items = sum(len(s.items) for s in suites)
    total_q = sum(len(s.queries) for s in suites)
    _log(cfg, f"Suites: {suite_names}  ({total_items} items, {total_q} queries)\n")

    for adapter in adapters:
        if not adapter.available:
            _log(cfg, f"  - {adapter.name:12s} SKIPPED (unavailable; requires: {', '.join(adapter.requires) or 'n/a'})")
            continue
        tag = " [estimate]" if adapter.is_estimate else ""
        _log(cfg, f"  - {adapter.name:12s} running{tag} ...")
        res = run_adapter(adapter, suites, cfg)
        cost = res.system_cost(cfg.pricing)
        _log(
            cfg,
            f"      accuracy={res.accuracy:.0%}  "
            f"recall_latency={res.avg_recall_latency_ms:.1f}ms  "
            f"system_cost=${cost:.6f}",
        )
        results.append(res)
    _log(cfg, "")
    return results
