"""Retrieval-quality evaluation.

A query is scored as correct when the retrieved set (top_k) contains the ground
truth — either by memory id (`expected_ids`) or by a case-insensitive substring
match on the retrieved text (`expected_substring`). Substring matching lets the
same suite score systems that mint their own ids (so id matching is impossible).
"""

from __future__ import annotations

from dataclasses import dataclass

from framework.adapter import Query, RetrievedItem


@dataclass
class QueryEval:
    correct: bool
    hit_rank: int  # 1-based rank of the first correct item, or 0 if none
    retrieved: int


def evaluate(query: Query, items: list[RetrievedItem]) -> QueryEval:
    topk = items[: query.top_k] if query.top_k else items
    expected_ids = set(query.expected_ids)
    needle = query.expected_substring.lower().strip()

    for rank, it in enumerate(topk, start=1):
        id_hit = bool(expected_ids) and it.id in expected_ids
        sub_hit = bool(needle) and needle in (it.text or "").lower()
        if id_hit or sub_hit:
            return QueryEval(correct=True, hit_rank=rank, retrieved=len(topk))
    return QueryEval(correct=False, hit_rank=0, retrieved=len(topk))
