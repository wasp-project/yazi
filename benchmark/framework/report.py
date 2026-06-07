"""Reporting — turn AdapterResult records into JSON / Markdown / CSV."""

from __future__ import annotations

import csv
import io
import json

from framework.metrics import AdapterResult, Pricing


def to_json(results: list[AdapterResult], pricing: Pricing, meta: dict | None = None) -> str:
    payload = {
        "meta": meta or {},
        "pricing": {
            "llm_input_per_m": pricing.llm_input_per_m,
            "llm_output_per_m": pricing.llm_output_per_m,
            "embed_per_m": pricing.embed_per_m,
        },
        "results": [r.to_dict(pricing) for r in results],
    }
    return json.dumps(payload, indent=2)


_COLUMNS = [
    ("adapter", "Adapter"),
    ("accuracy", "Acc."),
    ("avg_recall_latency_ms", "Recall ms"),
    ("avg_ingest_latency_ms", "Ingest ms"),
    ("system_cost_usd", "Sys cost $"),
    ("cost_per_1k_ops_usd", "$/1k ops"),
    ("cost_per_correct_usd", "$/correct"),
    ("downstream_context_tokens", "Ctx tok"),
    ("footprint", "Footprint"),
]


def _fmt(key: str, val) -> str:
    if val is None:
        return "—"
    if key == "accuracy":
        return f"{val:.0%}"
    if key.endswith("_usd"):
        return f"{val:.6f}".rstrip("0").rstrip(".") if val else "0"
    if key.endswith("_ms"):
        return f"{val:.1f}"
    return str(val)


def to_markdown(results: list[AdapterResult], pricing: Pricing) -> str:
    rows = [r.to_dict(pricing) for r in results]
    head = "| " + " | ".join(label for _, label in _COLUMNS) + " |"
    sep = "| " + " | ".join("---" for _ in _COLUMNS) + " |"
    lines = [head, sep]
    for row in rows:
        cells = []
        for key, _ in _COLUMNS:
            val = _fmt(key, row.get(key))
            if key == "adapter" and row.get("is_estimate"):
                val += " *"
            cells.append(val)
        lines.append("| " + " | ".join(cells) + " |")
    table = "\n".join(lines)
    note = (
        "\n\n`*` = cost **estimate** (sim-\\* adapter; no real service called). "
        f"Pricing: LLM in ${pricing.llm_input_per_m}/M, out ${pricing.llm_output_per_m}/M, "
        f"embed ${pricing.embed_per_m}/M per 1M tokens.\n"
        "`Sys cost $` is the recurring memory-system spend (extraction + embeddings + "
        "recall LLM) for the whole suite; it excludes the downstream prompt cost of "
        "injecting recalled context (`Ctx tok`), which every system pays."
    )
    return table + note


def to_csv(results: list[AdapterResult], pricing: Pricing) -> str:
    buf = io.StringIO()
    rows = [r.to_dict(pricing) for r in results]
    fields = [k for k, _ in _COLUMNS] + ["is_estimate", "n_ingests", "n_queries", "n_correct"]
    w = csv.DictWriter(buf, fieldnames=fields, extrasaction="ignore")
    w.writeheader()
    for row in rows:
        w.writerow(row)
    return buf.getvalue()
