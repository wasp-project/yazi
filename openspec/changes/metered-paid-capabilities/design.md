## Context

The composition layer is well-built and half-connected. What exists:

- `provider.Usage` / `Meter` / `Pricing` / `Budget` with `reject`/`degrade`
  modes, and `Pipeline.Ingest` already performs a pre-embedding budget check
  (`pipeline.go:86-96`).
- `LocalEmbedder` (hashing) and an in-process cosine `VectorIndex`
  (`provider/vector.go`) — real code, but the index lives only in the process.
- Honest stubs: `LLMExtractor`, `LLMReranker`, `GraphRetriever`, `PgVectorIndex`
  all report `Available() == false`, and `requireAvailable` fails the build of a
  pipeline that selects them.

What is broken or missing:

- `RecallWithProfile` for any non-`lite` profile constructs a **fresh** pipeline
  and re-ingests up to 1000 basic memories before serving the query
  (`composition.go:141-152`). Verified live: a two-record store reports
  `embedTokens: 6` per recall; at 100k records the same query would embed the
  corpus. Embedding is a write-time cost being paid at read time, every time.
- The meter is per-process and per-request in practice — the CLI constructs a new
  `Meter` for each invocation (`yazi_ctl.go:484`), so totals never accumulate.
- No LLM client exists, so the budget machinery has never bounded a real charge.

Constraints: `lite` must remain bit-for-bit free; no mandatory dependency; a
local model (Ollama) must be a first-class target, since `LOCAL-DEPLOYMENT.md`
positions the offline stack as the sweet spot; and every spend must be attributed
before it is incurred, not after.

## Goals / Non-Goals

**Goals:**
- Embedding is paid once per embedded memory, at write time.
- Only memories worth embedding are embedded.
- A real LLM path exists, works against a local model, and is bounded by budgets
  that refuse before spending.
- Cost is durable, queryable, and attributable per tenant, class and provider.
- `pro` becomes selectable rather than a stub that panics.

**Non-Goals:**
- Building an ANN index. The durable index in this change is exact search over
  stored vectors with the tiering from `tiered-memory-storage` keeping it cheap;
  approximate indexing is a later change driven by measured latency.
- Reimplementing pgvector, Milvus or a graph store. Those remain wrapper targets;
  `PgVectorIndex` and `GraphRetriever` stay stubs.
- Billing, invoicing, or quota enforcement across processes. This is metering and
  local enforcement.
- Prompt engineering research for extraction quality. One documented default
  prompt, overridable.
- Making any paid capability the default. `lite` remains the default forever.

## Decisions

### Decision 1: Kill re-ingest-on-recall; the index is durable and incremental

`Pipeline.Ingest` writes vectors to a persistent index keyed alongside the
record (`/_memory/vector/<class>/<id>` → quantized vector + dimension + model
id). `Recall` embeds only the query and scans/probes the stored vectors. The
`RecallWithProfile` seeding loop is deleted.

- **Why**: It converts embedding from an O(corpus) per-query cost into an O(1)
  per-write cost. This is the single largest cost defect in the codebase and it
  contradicts the project's own thesis.
- **Model identity is part of the record**: a vector is only comparable to
  vectors from the same embedder and dimension. Changing the configured embedder
  invalidates the index; the system detects the mismatch and refuses to mix,
  offering an explicit re-embed pass whose cost is projected before it runs.
- **Alternatives**: Cache the in-process index and rebuild on cold start
  (rebuild cost is unbounded and pays embedding twice); require an external
  vector DB (violates the single-binary floor).

### Decision 2: Exact search first, approximate later, chosen by measurement

Recall scores the query vector against stored vectors, with candidate
pre-filtering by the BM25/index path from `zero-cost-recall-quality` when a text
query is present (a hybrid retrieval), so the vector scan is bounded.

- **Why**: An exact scan over 100k float32 vectors is milliseconds; the honest
  answer is that most Yazi deployments will never need ANN, and shipping an ANN
  index we cannot maintain would import exactly the RAM-heavy cost profile
  `STATE-OF-ART.md` criticizes in Milvus. Hybrid pre-filtering keeps the scan
  small in the common case.
- **Exit criterion, stated now**: when the benchmark shows p95 vector recall
  latency exceeding 50 ms at a target corpus size, ANN becomes its own change.

### Decision 3: Embed by value — promotion rules, not blanket coverage

Embedding applies to advanced-class memories plus any basic memory matching a
configured promotion rule (kind allow-list, tag allow-list, minimum length, or an
access-count threshold). Everything else remains text-indexed only and is still
recallable via the free path.

- **Why**: `STATE-OF-ART.md` §5.2 commits to "only promoted memories are
  embedded — so embedding spend tracks value, not volume." This is that
  sentence, implemented. The access-count threshold means a memory that keeps
  getting recalled *earns* its embedding, which is the same recency-and-frequency
  principle used by storage tiering and ranking.
- **Consequence to state plainly**: semantic recall covers a subset of the
  corpus. Recall results mark which hits came from the vector path, so this is
  visible rather than a silent coverage gap.

### Decision 4: One OpenAI-compatible HTTP client, no SDK

`pkg/llm` speaks chat-completions and embeddings over HTTP/JSON against a
configured `baseURL`, `model` and optional API key. Ollama, vLLM,
llama.cpp servers and most hosted vendors accept this shape.

- **Why**: One protocol covers the local-first story and the hosted story with
  ~200 lines and zero dependencies — consistent with the hand-rolled SigV4
  precedent. An SDK per vendor would multiply `go.mod` and pin release cadences.
- **`Available()` becomes real**: a provider is available when its `llm:` block
  is configured and a startup probe succeeds. `pro` then fails at startup with
  "requires: llm.baseURL" rather than at request time.
- **Trade-off**: vendor-specific features (structured outputs, prompt caching,
  native tool calls) are not exposed. Acceptable — extraction and reranking need
  plain completions.

### Decision 5: Budgets enforce *before* the call, with a hard tenant ceiling

Each costly provider gains `Project(input) Usage` — an estimate computed without
calling out. `Pipeline` sums projections, prices them, and applies the budget
before dispatch. Beyond per-operation budgets, a per-tenant `maxSpendUSD` over a
rolling window refuses operations that would exceed it, regardless of mode.

- **Why**: `degrade` and `reject` are only meaningful if the decision precedes
  the spend. The existing pre-embedding check already does this for embeddings
  (`pipeline.go:87`); this generalizes the pattern to every costly stage.
- **Why a hard ceiling in addition to per-operation budgets**: per-operation
  limits bound the blast radius of one bad request; only a rolling ceiling bounds
  the bill. The ceiling refuses — it never silently degrades — because a user who
  set a spend cap wants an error, not a quiet quality drop.
- **Projection accuracy**: projections use the same ~4-chars/token heuristic as
  `estimateTokens`; after each call the actual usage is recorded and the
  projection error is tracked so systematic bias is visible.

### Decision 6: The meter becomes durable and queryable

Counters persist under `/_meta/cost/<tenant>/<period>` in the same store, updated
after each operation and readable via the service, the HTTP API and
`yazictl cost`. Reports break down by tenant, memory class, and provider, and
show token cost and storage cost as separate figures.

- **Why**: A per-process meter that the CLI recreates per invocation
  (`yazi_ctl.go:484`) cannot answer "what did memory cost me this month" — which
  is the question the whole project exists to answer. Persisting in the same
  store means the cost data inherits tenancy, durability and tiering.
- **Trade-off**: one small extra write per metered operation. Bounded by writing
  aggregates, not per-operation rows.

### Decision 7: `pro` ships as a real profile with a documented default prompt

`pro` = local or hosted embedder + persistent vector index + LLM extractor + LLM
reranker. The extraction prompt and rerank prompt ship as defaults, are
overridable in config, and are printed by `yazictl` on request so users can see
exactly what is being sent on their behalf.

- **Why print the prompts**: sending a user's memories to an LLM is the most
  consequential thing this system can do. Making the payload inspectable is
  consistent with making the cost inspectable.

## Risks / Trade-offs

- **[Re-embedding after a model change is expensive and surprising]** → Vectors
  record their model id and dimension; a mismatch refuses to serve semantic
  recall and reports the cost of a re-embed pass before running it. Never
  silently re-embed.
- **[Sending memories to a hosted LLM is a privacy event]** → `pro` is opt-in,
  the target endpoint is logged at startup, the default documented configuration
  points at a local endpoint, and the prompts are inspectable.
- **[Projection under-estimates and the ceiling is overshot]** → Projections are
  conservative (round up), actual usage is reconciled after each call, and the
  projection error is exported so bias is measurable.
- **[Exact vector scan degrades at scale]** → Hybrid pre-filtering bounds the
  candidate set; a measured p95 latency exit criterion is written into Decision 2
  rather than left implicit.
- **[Selective embedding creates invisible coverage gaps]** → Hits are labelled
  with the retrieval path that produced them, and the report states what fraction
  of the corpus is embedded.
- **[Paid features dilute the cost identity]** → Mitigated by construction:
  `lite` stays the default, every paid stage reports its usage, and the benchmark
  publishes free-versus-paid quality and cost side by side.

## Migration Plan

1. Land the persistent vector index and delete the recall-time re-ingest;
   `standard` cost moves from per-recall to per-write. Publish the before/after
   numbers — this is a user-visible cost improvement.
2. Add promotion rules; default to advanced-class-only so existing `standard`
   users see reduced embedding volume, announced in the release notes.
3. Land `pkg/llm` and the real extractor/reranker; `pro` becomes selectable and
   fails clearly when unconfigured.
4. Generalize pre-call projection and add the per-tenant hard ceiling.
5. Persist the meter and ship the cost report surfaces.

Rollback: each step is behind profile or config selection; `lite` is untouched
throughout, so a rollback never affects the default deployment.

## Open Questions

- Vector storage format: raw float32 (4 bytes/dim, exact) or int8 quantization
  (4× smaller, small recall loss)? Leaning float32 first with the format tagged
  in the record so quantization can be added without a breaking change.
- Should the rolling spend ceiling be enforced per tenant only, or also globally
  for the process? Leaning both, with the global ceiling defaulting to unset.
- Should `standard` embed synchronously on write (raising write latency) or
  asynchronously with a queue (raising complexity and creating a window where a
  memory is not semantically recallable)? Leaning synchronous for correctness and
  simplicity, revisited if the benchmark shows write latency becoming a problem.
