## Why

The `lite` profile is Yazi's whole value proposition — free recall — and its
ranking is a naive term-overlap count with no stopword handling, no term
weighting, no recency signal and no deduplication
(`pkg/memory/composition.go:82-100`). Measured on a running server: the query
*"where is the deploy script"* returned the correct memory **and** an unrelated
record about Postgres, because both contain the word *"the"*. Candidate selection
is also a hard-coded `Limit: 50` scan before ranking, so on a corpus larger than
50 records the right answer may never be scored at all.

This matters more than it looks. Every bad hit is context tokens the agent pays
for on the next turn (cost point C4), and every missed hit pushes users toward the
`standard`/`pro` profiles — which is to say, pushes them toward paying. **The
cheapest possible improvement to Yazi's cost story is making the free tier's
recall good enough that nobody needs to upgrade.** Everything in this change is
classical information retrieval: deterministic, no model, no embeddings, no
tokens.

## What Changes

- Replace term-overlap scoring with **BM25** over an inverted index, with per-
  field weighting (subject and tags weigh more than body) and stopword removal.
- Build a real **inverted index** in the existing `/_memory/index/...` key space
  so candidate selection is driven by query terms rather than a blanket 50-record
  scan. **BREAKING** for stored index layout: a rebuild pass regenerates indexes
  from the records on first start.
- Add **recency and access-frequency weighting** so a fresh or frequently
  recalled memory outranks a stale one at equal textual relevance, with
  configurable half-lives.
- Add **deterministic consolidation**: an exact-duplicate write is collapsed, and
  a new `preference` for a subject already covered supersedes the old one, which
  is marked superseded rather than deleted (lineage is preserved without any LLM
  in the loop).
- Add **temporal validity**: optional `validFrom` / `validUntil` on memories and a
  point-in-time filter, so "what did we decide in June" is answerable without a
  temporal knowledge graph.
- Add a **retrieval-quality suite** to the benchmark — a labelled set with
  precision@k, recall@k and MRR — so ranking changes are measured rather than
  asserted.

## Capabilities

### New Capabilities
- `deterministic-ranking`: inverted index, BM25 scoring, field weights and
  stopword handling — all at zero token cost.
- `recency-and-frequency-signals`: time decay and access-count weighting folded
  into the final score, with configurable parameters.
- `deterministic-consolidation`: duplicate collapse and supersede-with-lineage
  for subject-scoped memories, without an LLM.
- `temporal-validity`: validity intervals on memory records and point-in-time
  filtering of recall and list operations.
- `retrieval-quality-benchmark`: a labelled evaluation suite reporting
  precision@k, recall@k and MRR per profile.

### Modified Capabilities
- `memory-profiles`: the `lite` profile's guarantee is strengthened — it is
  defined as BM25 + recency/frequency ranking over the durable store, still at
  zero LLM and embedding cost.

## Impact

- **New code**: `pkg/memory/index` (inverted index build/query, tokenizer,
  stopwords), `pkg/memory/rank` (BM25 + decay scoring), consolidation logic in
  `pkg/memory/store.go`, `benchmark/evaluation` quality metrics and a labelled
  dataset.
- **Modified code**: `pkg/memory/composition.go` (`StoreBackedStore.Retrieve`
  replaced by index-driven candidate selection + BM25), `pkg/memory/model.go`
  (`validFrom`, `validUntil`, `supersededBy`, access counters),
  `pkg/memory/store.go` (index maintenance on put/delete), the memory service and
  HTTP API from `agent-integration-surface` (point-in-time filter parameter).
- **Stored data**: new index key space under `/_memory/index/term/...` plus
  per-record statistics; a one-time rebuild regenerates them from existing
  records. Memory records themselves gain optional fields and stay
  backward-compatible.
- **Dependencies**: none — BM25, stopwords and decay are ~300 lines of
  arithmetic.
- **Cost**: unchanged at zero tokens; index maintenance adds bounded KV writes per
  memory write, which the storage meter from `tiered-memory-storage` accounts
  for.
