## 1. Provider Interface & Pipeline Skeleton

- [x] 1.1 Create `pkg/memory/provider/provider.go` with the `Usage` struct (LLM input/output tokens, embed tokens, latency ms, context tokens) and an `add`/sum helper, matching the field shapes in `benchmark/framework/metrics.py`
- [x] 1.2 Define stage interfaces in `pkg/memory/provider`: `Extractor`, `Distiller`, `Embedder`, `Index`, `Retriever`, `Reranker`, `ContextBudgeter`, `KVCacheReuse` — each operation returning `Usage` (KVCacheReuse left as a reserved/marker concept)
- [x] 1.3 Add provider metadata (`Meta`: name/cost/latency/`Requires`, `Available()`) to a common `Provider` interface embedded by each stage
- [x] 1.4 Implement `Pipeline` with `Ingest(item)` and `Recall(query, budget)` that run the stages in order and accumulate `Usage`
- [x] 1.5 Unit tests: a pipeline of stubs runs in order and sums usage correctly

## 2. Deterministic Default Providers

- [x] 2.1 Implement no-op `Extractor`, `Distiller`, `Embedder`, `Reranker` and a top-k `ContextBudgeter`, all reporting zero token usage
- [x] 2.2 Implement a key+keyword/tag `Index` and `Retriever` (`KeywordStore`) that mirror the existing store's filter/rank semantics
- [x] 2.3 Build a `lite` default pipeline from these providers (via the registry)
- [x] 2.4 Tests: the default pipeline returns correct hits and reports zero LLM/embed usage

## 3. Route memory.Store Through the Pipeline

- [x] 3.1 Add `StoreBackedStore` + `NewStorePipeline` in `pkg/memory`: the deterministic pipeline persists through the real `memory.Store` (KV→LSM→S3). Done **additively** (wraps the typed Store) rather than rewriting `Put*/List*`, preserving the existing API and key layout
- [x] 3.2 Tests: existing `pkg/memory` suite passes unchanged; store-backed pipeline persists and recalls (`composition_test.go`)

## 4. Profiles & Configuration

- [x] 4.1 Add a `memory:` block to `pkg/config` (`profile`, per-stage overrides for `custom`, and `budget`); absent block resolves to `lite`
- [x] 4.2 Add a provider registry + profile resolver building a `Pipeline` from config (`lite`/`standard`/`pro`/`custom`)
- [x] 4.3 Validate provider requirements at build time; fail fast with a clear error when a selected provider is unavailable
- [x] 4.4 Add a profile-aware `yazictl memory recall --query --profile --tag --top-k --max-context-tokens` command (additive; typed CRUD intact). It builds the pipeline per `--tenant`/`--profile`, runs Recall over the durable store, and prints hits + metered Usage + costUSD. NOTE: metering is per-invocation (client-side); **server-side persistent metering** remains a follow-up
- [x] 4.5 Tests: `custom` overrides only named stages; unset config behaves like today; unavailable custom store fails fast (per-tenant profiles are supported by construction — a pipeline is built per tenant)

## 5. Reference "standard" Providers (one real paid path)

- [x] 5.1 Implement a real `Embedder` (dependency-free local hashing `LocalEmbedder`, reporting embed-token usage). NOTE: shipped as the offline reference; a real model (Ollama) implements the same interface later
- [x] 5.2 Implement a vector `Index`+`Retriever` (in-memory cosine `VectorIndex`) as the reference; **durable pgvector/Milvus Lite shipped as `PgVectorIndex` stub** (open design question: vector persistence vs snapshot)
- [x] 5.3 Define the `standard` profile wiring deterministic core + embedder + vector retriever
- [x] 5.4 Register `LLMExtractor`, `GraphRetriever`, `LLMReranker`, `PgVectorIndex` as interface-only stubs reporting `Available()=false`
- [x] 5.5 Tests: standard pipeline embeds on write and retrieves via vector search; selecting an unavailable stub fails fast

## 6. Cost Metering & Budgets

- [x] 6.1 Implement a `Meter` aggregating `Usage` per tenant and memory class, with a configurable `Pricing` table and $ estimate
- [x] 6.2 Expose accumulated usage/cost programmatically (`Meter.Usage/Totals/CostFor`); re-pricing recomputes from recorded usage (tested). NOTE: a CLI/RPC surface for cost rides on task 4.4
- [x] 6.3 Budget enforcement at the pipeline boundary: `ContextBudgeter` caps recall context tokens; over-budget ingest is rejected or degraded per mode
- [x] 6.4 Tests: recall capped to budget; over-budget ingest rejected/degraded; no budget = no limit; cost attributed to the correct tenant

## 7. Benchmark Integration

- [x] 7.1 Add `benchmark/` `yazi-lite` / `yazi-standard` adapters (selector `yazi-profiles`) that drive the live `yazictl memory recall --profile` surface and report the engine's REAL metered Usage/cost — verified against a running server (lite $0, standard ~$1e-6) alongside the `sim-*` estimators
- [x] 7.2 In-process accuracy-vs-cost frontier across profiles via `provider.RunProfiles` (lite=$0 < standard); shares `Usage`/`Pricing` shapes with the benchmark (`frontier_test.go`)

## 8. Docs & Validation

- [x] 8.1 Document the provider interfaces, profiles, budgets, and the `memory:` config in `COMPOSITION.md` (linked from the design)
- [x] 8.2 Architecture conformance test (`pkg/arch`) asserts the provider layer is self-contained (imports no other Yazi package)
- [x] 8.3 `CGO_ENABLED=0 go build ./...` and `go test ./...` pass; the `lite` default is unchanged
