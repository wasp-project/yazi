## Why

Yazi's cost thesis is not "never spend money" — it is "spend it deliberately,
where you can see the meter." Today the paid half of that promise is unfinished
in two specific ways:

1. **The `standard` profile's economics are backwards.** `RecallWithProfile`
   (`pkg/memory/composition.go:141-152`) rebuilds the vector index on *every*
   recall by re-ingesting up to 1000 records, so embedding cost scales with
   corpus size **per query** instead of being paid once per write. A system that
   bills itself on cost efficiency cannot ship that.
2. **The `pro` profile does not exist.** `LLMExtractor`, `LLMReranker`,
   `GraphRetriever` and `PgVectorIndex` are honest interface stubs reporting
   `Available() == false`, so `--profile pro` fails — currently as a raw Go
   panic. The metering and budget machinery built to govern LLM spend has never
   governed an actual LLM call.

The infrastructure for doing this right already exists: `Usage`, `Meter`,
`Pricing`, `Budget` with reject/degrade modes, and profile-driven composition.
This change connects real, costly providers to that machinery, so Yazi becomes
the memory system where the expensive path is available, bounded, and visible —
the differentiator none of mem0, Zep, Letta or mem9 offers natively.

## What Changes

- Make embeddings a **write-time, once-per-memory** cost: a durable vector index
  persisted in the KV/segment layer, updated incrementally on ingest, so recall
  embeds only the query. **BREAKING** for `standard`-profile cost expectations —
  cost moves from per-recall to per-write and drops by orders of magnitude at
  steady state.
- Embed **selectively by value, not by volume**: only memories promoted to the
  advanced class (or matching a configured promotion rule) are embedded, so
  embedding spend tracks value as `STATE-OF-ART.md` §5.2 commits.
- Add a **provider-agnostic LLM client** (HTTP/JSON, OpenAI-compatible, which
  covers Ollama and most hosted vendors) with no new module dependency, and
  implement the `LLMExtractor` and `LLMReranker` stubs against it.
- Enforce budgets **live** on those calls: pre-call cost projection, reject or
  degrade per the configured mode, and a hard per-tenant spend ceiling that
  refuses rather than overruns.
- Add a **cost report** surface — per tenant, per class, per provider, token and
  storage cost — over the memory service and HTTP API, plus a `yazictl cost`
  command. The meter becomes durable rather than per-process.
- Make `pro` a real, selectable profile that fails with a clear message when its
  providers are unconfigured, rather than being unreachable.

## Capabilities

### New Capabilities
- `persistent-vector-index`: a durable, incrementally maintained vector index so
  embedding is paid once per memory at write time, not per query.
- `value-based-embedding`: rules that decide which memories are worth embedding,
  so semantic capability is bought for the memories that earn it.
- `llm-provider-integration`: a dependency-free, OpenAI-compatible LLM client
  backing real extractor and reranker providers, usable against a local model.
- `cost-reporting`: durable, queryable per-tenant/per-class/per-provider cost
  accounting exposed through the service, API and CLI.

### Modified Capabilities
- `memory-provider-interface`: providers gain incremental/persistent index
  semantics and a pre-call cost projection so budgets can be enforced before
  spending rather than after.
- `cost-metering-and-budgets`: budget enforcement becomes live against real
  spend, gains a per-tenant hard spend ceiling, and the meter becomes durable
  across restarts.
- `memory-profiles`: `standard` is redefined as write-time embedding with a
  persistent index; `pro` becomes a genuinely selectable profile with real
  providers.

## Impact

- **New code**: `pkg/memory/provider/vectorstore` (durable index over the KV /
  segment layer), `pkg/llm` (OpenAI-compatible HTTP client), real
  `LLMExtractor` / `LLMReranker` implementations, `pkg/memory/provider/report`
  (durable meter + cost report).
- **Modified code**: `pkg/memory/composition.go` (delete the re-ingest-on-recall
  path), `pkg/memory/provider/registry.go` (wire real providers, availability
  from config), `pkg/memory/provider/pipeline.go` (pre-call projection, hard
  ceiling), `pkg/memory/provider/meter.go` (durability), `pkg/config/config.go`
  (`llm:` and `vector:` blocks), CLI and HTTP API from
  `agent-integration-surface`.
- **Stored data**: a vector index in the same store, subject to the tiering from
  `tiered-memory-storage`; durable meter counters per tenant.
- **Dependencies**: none — the LLM client is stdlib `net/http` + JSON, matching
  the precedent set by the hand-rolled SigV4 signer.
- **Cost**: `lite` remains exactly zero and unaffected. `standard` and `pro`
  become genuinely usable and genuinely metered.
- **Sequencing**: depends on `agent-integration-surface` (service seam),
  `tiered-memory-storage` (durable index storage and storage cost), and
  `zero-cost-recall-quality` (the free baseline the paid tiers are measured
  against).
