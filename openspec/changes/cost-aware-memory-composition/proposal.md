## Why

Yazi's deterministic core gives memory at ~zero token cost, but it has no
semantic recall, LLM extraction, or temporal reasoning — capabilities the
alternatives (mem0, Zep, Letta, MemOS, mem9) provide at a recurring per-message
cost. Rather than reimplement those (too crowded, and it would betray the
single-binary, $0-floor value), Yazi should let users **explicitly compose** their
memory pipeline from pluggable capabilities and **see and bound the cost** of each
composition. This turns Yazi into a cost-aware memory *orchestrator*: cheap by
default, expensive only by deliberate, metered choice.

## What Changes

- Introduce a **provider/capability interface layer** inside the memory engine.
  Each pipeline stage becomes a swappable provider that returns a `Usage` record
  and declares its cost/latency/infra requirements:
  - write path: `Extractor` → `Distiller` → `Embedder` → `Index`
  - read path: `Retriever` → `Reranker` → `ContextBudgeter` (+ `KVCacheReuse`)
- Add **named memory profiles** (`student` | `standard` | `pro` | `custom`)
  selectable via config. The default (`student`) wires the existing deterministic
  KV+index/LSM core — behavior unchanged, cost $0. Higher profiles add providers.
- Add **cost metering and budgets**: roll per-operation `Usage` up per tenant and
  per memory class, and let a profile declare budgets (e.g. max recall-context
  tokens, max ingest $/1k) that the orchestrator enforces (reject or degrade).
- Ship **reference providers**: the deterministic defaults (already exist, wrapped
  as providers) plus one real "standard" path — a local embedder + a vector index.
  Leave `Extractor` (LLM), graph/temporal, and `Reranker` as **interface-only
  stubs** so Yazi wraps external systems later without reimplementing them.
- Extend the **benchmark/** harness so each profile is a runnable composition,
  letting users/CI compare profiles on the accuracy-vs-cost frontier.

## Capabilities

### New Capabilities
- `memory-provider-interface`: The pluggable provider contracts (Extractor,
  Embedder, Index, Retriever, Reranker, KVCacheReuse), the `Usage` and
  provider-metadata (cost/latency/requires) they report, and the composed
  write/read pipeline that runs them in order with deterministic defaults.
- `memory-profiles`: Named, config-selected compositions (`student` | `standard`
  | `pro` | `custom`) that bind providers to pipeline stages; `student` is the
  default and reproduces today's $0 deterministic behavior.
- `cost-metering-and-budgets`: Per-operation `Usage` accounting aggregated per
  tenant and memory class, plus profile-declared budgets the orchestrator
  enforces by rejecting or degrading over-budget operations.

### Modified Capabilities
<!-- No existing specs under openspec/specs/; all behavior is introduced as new capabilities. -->

## Impact

- **New code**: `pkg/memory/provider/` (interfaces, `Usage`, pipeline, metering,
  budgets), default deterministic providers wrapping the current store path, one
  reference embedder + vector index provider, interface-only stubs for the rest.
- **Modified code**: `pkg/memory` (route Put/List through the pipeline),
  `pkg/config` (a `memory:` block: profile + per-stage overrides + budgets),
  `cmd/cli`/`pkg/server` wiring to select a profile, `benchmark/` adapters to run
  per-profile compositions.
- **Config**: new `memory.profile` and `memory.budget` settings; default `student`
  keeps current behavior. Higher profiles may require external services (embedder,
  vector store) — opt-in only.
- **Dependencies**: none by default; reference providers add optional deps (e.g. a
  local embedding runtime, a vector index) compiled in but inactive unless selected.
- **Compatibility**: Backward compatible — absent `memory` config resolves to the
  `student` profile (deterministic core, $0), identical to current behavior.
