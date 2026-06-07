# Cost-Aware Memory Composition

Yazi's memory engine is a **composable pipeline of provider stages**. Each stage
is swappable, reports its resource `Usage`, and declares what it requires to run.
Users pick a **profile** to trade memory quality for cost; Yazi *orchestrates*
the chosen capabilities — wrapping external systems (embedders, vector stores,
LLMs) rather than reimplementing them. See [STATE-OF-ART.md](./STATE-OF-ART.md)
for why cost is the axis, and [ARCHITECTURE.md](./ARCHITECTURE.md) for the layers.

Package: `pkg/memory/provider` (self-contained; imports no other Yazi package).

## The pipeline

```
WRITE:  Item ─▶ [Extractor] ─▶ [Distiller] ─▶ [Embedder] ─▶ [Index]
READ:   Query ─▶ [Retriever] ─▶ [Reranker] ─▶ [ContextBudgeter] ─▶ Hits
```

Every stage is a small Go interface whose operation returns a `Usage`
(LLM/embedding tokens, latency, downstream context tokens). Stages default to
deterministic, zero-token providers, so an unconfigured pipeline behaves exactly
like Yazi's original memory store.

## Profiles (the user choice)

Configured via the `memory:` block (absent ⇒ `student`):

```yaml
memory:
  profile: standard        # student | standard | pro | custom
  budget:                  # optional; zero fields = unbounded
    recallContextTokens: 1500
    ingestMaxUSD: 0.50
    mode: degrade          # reject | degrade
  # custom overrides (only when profile: custom):
  embedder: local          # none | local
  store: vector            # keyword | vector | pgvector
  extractor: none          # none | llm
  reranker: none           # none | llm
```

| Profile | Composition | Cost |
| --- | --- | --- |
| **student** | keyword/index store, no embeddings | **$0 tokens**, durable via KV→LSM→S3 |
| **standard** | local embedder + vector retrieval | embedding cost, mostly local |
| **pro** | + LLM extraction + LLM rerank (interface stubs) | full capability, metered |
| **custom** | bind each stage; unset stages use the cheap default | you decide |

`student` keeps Yazi's deterministic, persistent, zero-token behavior. Higher
profiles add providers; selecting a provider whose requirements are unmet **fails
fast at startup** rather than breaking mid-request — so `pro` errors until the LLM
providers are wired up.

## Cost metering & budgets

- A `Meter` aggregates `Usage` per **tenant** and **memory class** (basic /
  advanced / policy). A configurable `Pricing` table converts usage to a $ estimate
  — change prices and totals recompute from recorded usage without re-running.
- A profile may declare **budgets**. `ContextBudgeter` trims recall to a token
  cap; an ingest over its cost cap is **rejected** or **degraded** (skips
  embedding) per the mode. No budget ⇒ unbounded.

## What ships now vs. wrapped later

- **Real, dependency-free, tested:** the deterministic `student` providers
  (persistence-backed via `pkg/memory` `StoreBackedStore`), a local hashing
  `Embedder`, and an in-memory cosine `VectorIndex` for `standard`.
- **Interface-only stubs** (report `Available() == false`): `LLMExtractor`,
  `GraphRetriever`, `LLMReranker`, `PgVectorIndex`. These mark where Yazi will
  **wrap** mem0/Zep/Milvus/pgvector/an LLM — no fabricated capabilities.

## Comparing profiles

`provider.RunProfiles(profiles, items, queries, pricing)` runs the same workload
through several profiles and reports each one's cost — an in-process
accuracy-vs-cost frontier. The `benchmark/` harness uses the same `Usage`/`Pricing`
shapes, so engine and benchmark numbers line up.

## Status

The composition engine (interfaces, pipeline, deterministic + reference providers,
stubs, metering, budgets, profiles, persistence adapter) is implemented and tested
in `pkg/memory/provider` and `pkg/memory`. Live server/CLI wiring of profiles
(per-tenant profile selection, a recall command, server-side persistent metering)
and durable vector backends (pgvector / Milvus Lite) are the next integration step
— see the change `openspec/changes/cost-aware-memory-composition/`.
