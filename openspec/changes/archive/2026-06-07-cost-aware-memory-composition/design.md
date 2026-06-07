## Context

Yazi today has a fixed memory path: `cmd/cli` builds a `memory.Store` over a
`KV`, writes structured `/_memory/...` records with secondary indexes, and reads
them back by key/filter — deterministic, $0 in tokens, single binary. The newly
added tenant layer namespaces keys; the S3 backend gives cloud persistence; the
`benchmark/` harness already models a per-system cost profile through a common
adapter interface.

The competitive survey (STATE-OF-ART.md) shows the alternatives buy semantic
recall, temporal reasoning, and self-editing memory at a recurring per-message
cost (LLM extraction, embeddings, RAM-resident vector indexes). Users want to
pick where they sit on that quality/cost curve — some want the free
deterministic core; others with budget want embeddings + vector search; a
well-funded team wants the full stack. This change makes the memory path a
**composable pipeline of providers** with **declared, metered, budgetable cost**,
so Yazi orchestrates capabilities instead of reimplementing competitors.

Constraints: keep the default ($0, single-binary, offline) behavior identical;
never make a heavy dependency mandatory; reuse the existing tenant seam and the
benchmark harness; wrap external systems rather than reimplement them.

## Goals / Non-Goals

**Goals:**
- A provider interface for each pipeline stage, each reporting `Usage` and
  declaring cost/latency/requirements.
- Named profiles (`lite`/`standard`/`pro`/`custom`) selecting providers via
  config, with `lite` as the unchanged default.
- Per-tenant, per-class cost metering and profile-declared budget enforcement.
- Ship deterministic defaults + one real "standard" path (embedder + vector
  index); leave extraction/graph/rerank as interface-only stubs.
- Make each profile a runnable benchmark composition.

**Non-Goals:**
- Implementing every competitor capability (graph/temporal, self-editing, LLM
  rerank) now — only their interfaces.
- Building our own vector DB or embedding model (we wrap pgvector / Milvus Lite /
  a local embedder).
- Changing the memory data model, key layout, or the deterministic read/write
  semantics of the `lite` profile.
- A pricing/billing system (we estimate cost from usage + a price table; billing
  is out of scope).

## Decisions

### Decision 1: A staged pipeline of small provider interfaces
Define `pkg/memory/provider` with one interface per stage — `Extractor`,
`Distiller`, `Embedder`, `Index`, `Retriever`, `Reranker`, `ContextBudgeter`,
`KVCacheReuse`. A `Pipeline` holds one provider per stage and runs them in order
for `Ingest` and `Recall`, threading a `Usage` accumulator.
- **Why**: Mirrors the benchmark adapter seam and the existing layered design.
  Small interfaces keep each capability independently swappable and testable, and
  let external systems be wrapped one stage at a time.
- **Alternatives**: One monolithic `MemoryBackend` interface per system (à la the
  benchmark's `MemoryAdapter`). Rejected for the engine itself — it forces
  all-or-nothing swaps and can't express "deterministic index + hosted reranker".
  (That coarse interface still lives in the benchmark, where whole-system
  comparison is the point.)

### Decision 2: Usage as the universal currency; cost computed from a price table
Every provider op returns `Usage{LLMInput, LLMOutput, EmbedTokens, LatencyMs,
ContextTokens}`. A `Meter` aggregates usage per (tenant, memory-class) and a
configurable `Pricing` table converts it to a $ estimate.
- **Why**: Decouples measurement from pricing (re-price without re-running),
  matches the benchmark's `metrics.py` model exactly so numbers are comparable,
  and gives budgets something concrete to enforce.
- **Alternative**: Each provider returns a $ figure directly. Rejected — bakes in
  prices, breaks re-pricing, and hides the token breakdown.

### Decision 3: Profiles are config that binds providers; default = lite
A `memory:` config block selects `profile` and, for `custom`, per-stage provider
names + options, plus `budget`. A registry resolves names → provider constructors.
Unset config → `lite` (all deterministic defaults).
- **Why**: "Same binary, choose your cost" with zero code change, consistent with
  how `storage`/`engine`/`tenant` are already configured. Per-tenant profiles ride
  the existing tenant seam.
- **Alternative**: Profile chosen per request/API call. Deferred — config-level is
  simpler and sufficient; per-request override can layer on later.

### Decision 4: Wrap, don't reimplement; reference providers prove the seam
Ship deterministic defaults (wrapping today's store) + a real `standard` path: a
local `Embedder` (e.g. an Ollama/embedding HTTP call) and a vector `Index`/`Retriever`
(pgvector or Milvus Lite). `Extractor` (LLM), graph/temporal `Retriever`, and
`Reranker` ship as registered stubs reporting `available=false`.
- **Why**: Proves the pipeline end-to-end with one paid path while keeping scope
  bounded and the $0 single-binary floor intact. Heavy providers are opt-in and may
  require external services — acceptable because the user chose that profile.
- **Alternative**: Build several real providers now. Rejected — provider-matrix
  explosion; honesty (no fabricated capabilities) over coverage.

### Decision 5: Budgets enforced at the pipeline boundary
The `Pipeline` checks a profile's budgets against accumulated/estimated usage:
recall context is truncated to a token budget by the `ContextBudgeter`; an ingest
over its cost budget is rejected or routed to a cheaper provider per an enforcement
mode (`reject` | `degrade`).
- **Why**: Cost-bounded retrieval is the read-side analogue of rate limiting and a
  concrete expression of "cost-aware". Central enforcement keeps it consistent.
- **Alternative**: Advisory budgets (warn only). Offered as a mode, but hard
  enforcement is the differentiator.

### Decision 6: Each profile is a benchmark composition
Extend `benchmark/` so a profile maps to an adapter, letting users/CI plot each
composition on the accuracy-vs-cost frontier — the tool for *choosing* a profile.
- **Why**: We already built the harness; this makes profile selection evidence-based
  and the comparison continuous.

## Risks / Trade-offs

- **Provider-matrix explosion / scope creep** → Ship few first-class providers
  (deterministic + one standard path) and interface-only stubs; gate new providers
  on demand. Document what is stub vs real.
- **Heavy providers pull dependencies / ops weight** → Keep them opt-in and
  compiled-but-inactive; the `lite` default stays single-binary and offline.
  Validate provider requirements at startup (fail fast, like S3).
- **Usage under-reporting skews cost** → Deterministic/local providers report
  tokens=0 honestly; hosted providers report from SDK responses; local-compute cost
  is surfaced as latency, not $ (documented).
- **Pipeline overhead on the hot path** → The `lite` pipeline is a thin pass-through
  of no-op stages over the existing store; benchmark the default to confirm parity.
- **Two interfaces (engine providers vs benchmark adapter) drift** → Keep the
  `Usage`/`Pricing` shapes identical across both; share field names with
  `benchmark/framework/metrics.py`.

## Migration Plan

1. Add `pkg/memory/provider` (interfaces, `Usage`, `Pipeline`, deterministic
   defaults, `Meter`) with no wiring change — defaults reproduce current behavior.
2. Route `memory.Store` Ingest/Recall through the default pipeline; verify parity
   with existing memory tests.
3. Add the `memory:` config block + profile registry; `lite` default.
4. Add the reference `standard` providers (embedder + vector index) behind opt-in
   config and startup requirement checks.
5. Add metering/budgets + enforcement modes.
6. Extend `benchmark/` with per-profile adapters; publish the frontier.
7. **Rollback**: unset `memory` config → `lite` → identical to today; no data
   migration (the deterministic layout is unchanged).

## Open Questions

- First reference vector backend: **pgvector** (reuses a SQL dep, simplest) vs
  **Milvus Lite** (embeddable, no server). Lean pgvector for the standard path.
- Default embedder: local (Ollama) only, or also a hosted option? Lean local-first
  to preserve the offline/cheap ethos.
- Budget enforcement default mode: `reject` vs `degrade`. Lean `degrade` for recall
  (return less context), `reject` for ingest over a hard cost cap.
- Where vector data lives relative to the snapshot model: alongside the KV snapshot,
  or in the vector store only (rehydrated)? Likely vector store owns its index;
  Yazi keeps the source records.
- Should profiles be switchable per request (header) in addition to per tenant?
  Deferred unless needed.
