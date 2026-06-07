"""A tiny in-process keyword store used by the dependency-free adapters.

It gives every offline adapter the SAME retrieval mechanic, so accuracy
differences in the report come from the COST MODEL we attach per system, not
from a different search implementation. Real adapters (mem0/zep/...) override
retrieval with their actual engines.
"""

from __future__ import annotations

import re

from framework.adapter import MemoryItem, RetrievedItem

_WORD = re.compile(r"[a-z0-9]+")


def _tokens(text: str) -> set[str]:
    return set(_WORD.findall(text.lower()))


class KeywordStore:
    def __init__(self) -> None:
        self._items: list[MemoryItem] = []

    def clear(self) -> None:
        self._items = []

    def add(self, item: MemoryItem) -> None:
        self._items.append(item)

    def search(self, query_text: str, top_k: int, tag: str = "") -> list[RetrievedItem]:
        q = _tokens(query_text)
        scored: list[tuple[float, MemoryItem]] = []
        for it in self._items:
            doc = _tokens(it.text + " " + it.subject + " " + " ".join(it.tags))
            overlap = len(q & doc)
            if not overlap:
                continue
            score = overlap / (len(q) or 1)
            if tag and tag in it.tags:
                score += 0.25  # light structured-filter boost
            scored.append((score, it))
        scored.sort(key=lambda x: x[0], reverse=True)
        return [RetrievedItem(id=it.id, text=it.text, score=s) for s, it in scored[:top_k]]

    def all_texts(self) -> list[str]:
        return [it.text for it in self._items]
