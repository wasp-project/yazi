## 1. Persistent Vector Index

- [ ] 1.1 Define the stored vector record (`/_memory/vector/<class>/<id>` → vector bytes + dimension + embedder model id + format tag) in a new `pkg/memory/provider/vectorstore`
- [ ] 1.2 Implement the durable `Index`/`Retriever` pair over the KV layer so vectors inherit tenant namespacing, durability and tiering
- [ ] 1.3 Add a `Persistent() bool` declaration to the provider interface and have the pipeline skip population for persistent indexes
- [ ] 1.4 **Delete the recall-time re-ingest loop** in `memory.RecallWithProfile` (`composition.go:141-152`); recall embeds the query only
- [ ] 1.5 Detect embedder/dimension mismatch against stored vectors and refuse to compare across them; add an explicit re-embed pass that projects its cost and requires confirmation
- [ ] 1.6 Implement hybrid retrieval: pre-filter candidates with the deterministic index when a text query is present, and label each hit with the path that produced it
- [ ] 1.7 Remove the vector on memory delete
- [ ] 1.8 Test: recall embedding usage is independent of corpus size; vectors survive restart; delete removes the vector; tenants are isolated; embedder change is detected and refused; hits carry their retrieval path
- [ ] 1.9 Benchmark before/after: embedding tokens per recall at 100 / 10k / 100k memories; publish the delta

## 2. Value-Based Embedding

- [ ] 2.1 Add an eligibility rule engine (class, kind, tag, minimum length, minimum access count) with the default "advanced class only"
- [ ] 2.2 Apply eligibility in the ingest path so ineligible memories are stored with deterministic indexing and no embedding usage
- [ ] 2.3 Promote a memory to embedded once its access count crosses the configured threshold, embedding exactly once and reporting the cost
- [ ] 2.4 Report embedded count and fraction in memory statistics
- [ ] 2.5 Test: ineligible memory reports no embedding usage; default embeds advanced only; tag rule works; a memory recalled past the threshold gets embedded once; unembedded memories remain recallable deterministically

## 3. LLM Client

- [ ] 3.1 Create `pkg/llm`: OpenAI-compatible chat-completions and embeddings over stdlib `net/http`, configured by base URL, model, optional API key, timeout and retry policy — no new module dependency
- [ ] 3.2 Parse token usage from responses; when a provider omits usage, fall back to the conservative estimator and mark the figure as estimated
- [ ] 3.3 Add an `llm:` config block and log the configured endpoint at startup
- [ ] 3.4 Implement a startup reachability probe backing `Available()`
- [ ] 3.5 Test against an in-process `httptest` OpenAI-compatible endpoint: completion, embedding, usage parsing, timeout, retry, and unreachable-endpoint behavior

## 4. Real Extractor & Reranker

- [ ] 4.1 Implement `LLMExtractor` against `pkg/llm` with a documented default prompt, returning structured memory items and reporting usage
- [ ] 4.2 Implement `LLMReranker` against `pkg/llm` with a documented default prompt, returning reordered hits and reporting usage
- [ ] 4.3 Make both prompts overridable in config and printable via a CLI command
- [ ] 4.4 Implement failure fallback: rerank failure returns deterministic order, extraction failure stores the original item — both reported, never partial or fabricated
- [ ] 4.5 Wire real providers into `provider/registry.go` so `pro` resolves to them; keep `GraphRetriever` and `PgVectorIndex` as honest stubs
- [ ] 4.6 Make `pro` fail at startup (not at request time) with a message naming missing settings
- [ ] 4.7 Test: extraction and rerank happy paths with metered usage; prompt override; both fallback paths; unconfigured `pro` aborts startup; configured `pro` serves

## 5. Pre-Call Projection & Ceilings

- [ ] 5.1 Add `Project(input) Usage` to the costly provider interfaces, conservative and rounding up; deterministic providers project zero
- [ ] 5.2 Generalize `Pipeline`'s pre-embedding budget check into a pre-call projection for every costly stage, evaluated before dispatch
- [ ] 5.3 Add a per-tenant rolling `maxSpendUSD` ceiling that refuses over-ceiling operations regardless of reject/degrade mode, while letting zero-cost operations through
- [ ] 5.4 Record projected and actual usage per operation and expose the aggregate projection error
- [ ] 5.5 Test: rejected operation issues no LLM/embedding request; ceiling refuses even in degrade mode; free operations pass under a exhausted ceiling; allowance resets with the period; projection error is reported

## 6. Durable Metering & Cost Report

- [ ] 6.1 Persist meter aggregates under `/_meta/cost/<tenant>/<period>`, updated after each metered operation
- [ ] 6.2 Build the cost report: breakdown by tenant, memory class and provider, with token cost and storage cost as separate line items
- [ ] 6.3 Expose the report through the memory service and the HTTP API from `agent-integration-surface`
- [ ] 6.4 Add a `yazictl cost` command; stop constructing a throwaway `Meter` per CLI invocation (`yazi_ctl.go:484`)
- [ ] 6.5 Test: totals accumulate across restarts; tenant scoping; provider attribution; deterministic-only deployment reports zero token cost with non-zero storage cost; CLI and API reports agree

## 7. Documentation, Benchmark & Verification

- [ ] 7.1 Update `COMPOSITION.md`: profiles as implemented, write-time embedding economics, eligibility rules, ceilings, and what remains a stub
- [ ] 7.2 Update `README.md`, `QUICKSTART.md` and `MEMORY.md`: the `pro` setup path, the local-model configuration, and removal of the closed Known-limits entries
- [ ] 7.3 Add an end-to-end local recipe to `LOCAL-DEPLOYMENT.md`: `pro` against a local model with a spend ceiling configured
- [ ] 7.4 Extend the benchmark to report quality and cost for `lite` / `standard` / `pro` side by side using the quality suite from `zero-cost-recall-quality`
- [ ] 7.5 Publish the headline comparison in `STATE-OF-ART.md`: what fraction of paid-tier quality the free tier reaches, and what the paid tiers cost per 1k operations
- [ ] 7.6 Run `CGO_ENABLED=0 go build ./...` and `CGO_ENABLED=0 go test ./...`; confirm the full suite passes

## 8. Engineering Blog

- [ ] 8.1 Write the milestone blog post under `blog/` per the `engineering-blog` capability: the re-embed-per-recall bug and what it cost, why embedding follows value rather than volume, projecting spend before incurring it, and the measured cost curve across the three profiles
