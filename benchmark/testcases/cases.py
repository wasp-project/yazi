"""Test-case model and loader.

A suite is a JSON file under testcases/datasets/ shaped like:

    {
      "name": "preferences",
      "description": "...",
      "items":   [ {"id","text","kind","scope","subject","tags"}, ... ],
      "queries": [ {"text","expected_ids","expected_substring","tag","top_k"}, ... ]
    }

Each query's ground truth (`expected_ids` / `expected_substring`) is what the
evaluator scores retrieval against.
"""

from __future__ import annotations

import glob
import json
import os
from dataclasses import dataclass, field

from framework.adapter import MemoryItem, Query

DATASETS_DIR = os.path.join(os.path.dirname(__file__), "datasets")


@dataclass
class Suite:
    name: str
    description: str
    items: list[MemoryItem]
    queries: list[Query]
    path: str = ""

    @property
    def stats(self) -> dict:
        return {"items": len(self.items), "queries": len(self.queries)}


def _suite_from_dict(d: dict, path: str = "") -> Suite:
    items = [
        MemoryItem(
            id=i["id"],
            text=i["text"],
            kind=i.get("kind", "context"),
            scope=i.get("scope", "user"),
            subject=i.get("subject", ""),
            tags=i.get("tags", []),
        )
        for i in d.get("items", [])
    ]
    queries = [
        Query(
            text=q["text"],
            expected_ids=q.get("expected_ids", []),
            expected_substring=q.get("expected_substring", ""),
            tag=q.get("tag", ""),
            top_k=q.get("top_k", 5),
        )
        for q in d.get("queries", [])
    ]
    return Suite(
        name=d.get("name", os.path.splitext(os.path.basename(path))[0]),
        description=d.get("description", ""),
        items=items,
        queries=queries,
        path=path,
    )


def list_suite_files() -> list[str]:
    return sorted(glob.glob(os.path.join(DATASETS_DIR, "*.json")))


def load_suite(path: str) -> Suite:
    with open(path) as f:
        return _suite_from_dict(json.load(f), path)


def load_suites(names: list[str] | None = None) -> list[Suite]:
    """Load suites by name (filename without .json) or all of them when None."""
    files = list_suite_files()
    suites = [load_suite(p) for p in files]
    if names and "all" not in names:
        wanted = set(names)
        suites = [s for s in suites if s.name in wanted]
    return suites
