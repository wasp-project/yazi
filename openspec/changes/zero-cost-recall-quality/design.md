## Context

`StoreBackedStore.Retrieve` is the entire `lite` read path:

```go
list, _ = store.ListBasic(BasicFilter{Tag: q.Tag, Limit: 50})   // or all, Limit 50
for _, m := range list {
    overlap := |terms(query) ∩ terms(content+subject+tags)|
    score   := overlap / (len(queryTerms)+1)
}
sort by score desc; take topK
```

Three defects follow directly from that code, all confirmed by running it:

1. **No term discrimination.** Every term counts 1, so *"the"* scores like
   *"postgres"*. Observed: query *"where is the deploy script"* returned an
   unrelated Postgres decision.
2. **Candidate set is a blind prefix of the corpus.** `Limit: 50` is applied
   *before* scoring; beyond 50 records, relevance depends on index insertion
   order.
3. **No signal other than text.** A superseded preference from a year ago and the
   one that replaced it score identically.

The existing index space (`/_memory/index/basic/{kind,scope,subject,tag}/<v>` →
JSON id list) already proves the pattern this change needs: postings lists stored
as ordinary KV values. This change adds a term dimension to it.

Constraints: zero tokens (that is the point), no new dependency, indexes must
live in the same KV so they inherit tiering and tenancy for free, and the write
path must not become slow enough to matter for an interactive agent.

## Goals / Non-Goals

**Goals:**
- Free-tier recall that is good enough that upgrading is a choice, not a
  necessity.
- Ranking that is explainable — every score decomposable into text, recency and
  frequency contributions.
- Candidate selection driven by query terms, not by scan order.
- Deterministic handling of the two things agents do constantly: writing the same
  fact twice, and changing their mind.
- Quality becomes a measured number in the benchmark, per profile.

**Non-Goals:**
- Semantic similarity, synonyms, or embeddings — that is what `standard` is for,
  and confusing the two would blur the cost story.
- LLM-based extraction, summarization, deduplication or conflict resolution —
  explicitly the thing Yazi refuses to do by default.
- A knowledge graph or entity resolution. Temporal validity here is two
  timestamps and a filter, not Graphiti.
- Multilingual analysis beyond Unicode-aware tokenization; a CJK-aware analyzer
  is a follow-up.

## Decisions

### Decision 1: Inverted index as postings lists in the existing KV space

Key layout `/_memory/index/term/<class>/<term>` → JSON `{docID: termFreq}` plus
`/_memory/index/stats/<class>` → `{docCount, avgFieldLen, docLens}`. Written in
the same put transaction as the record, removed on delete.

- **Why**: Reuses the pattern already in `pkg/memory/store.go`, so postings
  inherit tenant prefixing, LSM durability and (after
  `tiered-memory-storage`) tiering with no extra work. No new storage component.
- **Trade-off**: A write now touches one key per distinct term (~10–30 for a
  typical memory) instead of ~4. Mitigated by batching through `MSet` and by the
  fact that agent memory is read-light/write-light in absolute terms. Measured in
  the benchmark's ingest-latency column.
- **Alternatives**: A separate index file/engine (breaks the "one store, one
  durability path" property); scanning at query time with a smarter scorer
  (leaves defect 2 unfixed).

### Decision 2: BM25 with field weights, not TF-IDF and not cosine

Standard BM25 (`k1=1.2`, `b=0.75`, both configurable) over a weighted
concatenation: `subject ×3`, `tags ×2`, `content ×1`.

- **Why BM25**: It is the strongest classical baseline, it needs only counts we
  already have, and its saturation behavior (`k1`) is what prevents a long
  rambling memory from dominating on repeated terms. Field weighting encodes the
  domain fact that a memory's `subject` and `tags` are curated by the agent while
  `content` is prose.
- **Why not TF-IDF**: no length normalization; long transcript memories would
  swamp short preferences.
- **Why not cosine over sparse vectors**: mathematically similar, more machinery,
  and it invites confusion with the `standard` profile's dense vectors.

### Decision 3: Stopwords and tokenization are explicit and inspectable

A built-in English stopword list, Unicode-aware word segmentation, lowercasing,
and no stemming in this change. The list is overridable in config; queries reduced
to only stopwords fall back to the unfiltered terms rather than returning nothing.

- **Why no stemming**: Porter-style stemming helps recall and hurts precision on
  short, high-signal texts like preferences, and it is language-specific. Deferred
  until the quality suite can show it helps.
- **Why the empty-query fallback**: a user searching for *"is it the one"* should
  get imperfect results, not zero results.

### Decision 4: Final score = text × recency × frequency, multiplicatively

```
score = bm25 · exp(-ln2 · age / recencyHalfLife) · (1 + log(1 + accessCount)/frequencyBoost)
```

Superseded records are excluded unless explicitly requested; records outside
their validity interval are excluded when a point-in-time is given.

- **Why multiplicative**: A memory with zero textual match must not be surfaced
  by being recent; recency modulates relevance rather than substituting for it.
- **Why exponential decay with a configured half-life**: it is the standard
  primitive, it is one intuitive parameter, and different memory classes want
  different values — `preference` should decay far slower than `context`, so the
  half-life is per-class with sane defaults (preference 365d, decision 180d,
  history 30d, context 7d).
- **Why access count matters**: it is the same recency-and-frequency principle
  the storage tiering uses, applied to ranking — the memories worth keeping warm
  are usually the memories worth returning. The log dampens runaway feedback where
  a returned memory becomes more returnable.
- **Explainability**: `Hit` carries the three factors so a caller can see why
  something ranked where it did. This is cheap and makes the free tier debuggable.

### Decision 5: Consolidation is exact-match and subject-scoped, never fuzzy

Two rules, both deterministic:

1. **Duplicate collapse** — an incoming record whose normalized text, kind, scope
   and subject hash equals an existing live record updates that record's
   `updatedAt` and access counters instead of creating a second one; the existing
   id is returned.
2. **Supersede** — a new `preference` (and only `preference`) with the same
   `scope` + `subject` as a live record marks the old one `supersededBy: <newID>`.
   Superseded records stay readable and are excluded from recall by default.

- **Why only exact match and only `preference`**: Fuzzy merging without an LLM
  produces wrong merges, and wrong merges in a memory system are worse than
  duplicates. `preference` is the one class whose semantics are inherently
  single-valued ("prefers dark theme" replaces "prefers light theme");
  `history`/`context` are append-only by nature and `decision` needs its lineage.
- **Why supersede instead of delete**: it is Zep's insight (invalidate, don't
  destroy) obtained for free, because the trigger is a structural key match rather
  than an LLM contradiction check.

### Decision 6: Temporal validity is two optional timestamps and a filter

`validFrom` / `validUntil` on every memory class; `asOf` on recall and list
filters selects records live at that instant. Absent bounds mean always valid.

- **Why**: It answers the majority of temporal questions agents actually ask
  ("what was true then") at the cost of two nullable fields and a comparison —
  versus a bi-temporal knowledge graph with LLM-extracted edges.
- **Limitation, stated plainly**: this is validity time only, not transaction
  time, and it records no reason for invalidation. `supersededBy` supplies
  lineage for the one case that generates it automatically.

### Decision 7: Quality is a benchmark column, per profile

A labelled dataset (query → relevant memory ids) drives precision@k, recall@k and
MRR. The report prints quality **beside** cost for `lite`, `standard` and `pro`.

- **Why**: `benchmark/README.md` currently states that accuracy is deliberately
  held constant so the cost axis is isolated. That was the right call when `lite`
  had no ranking to speak of; it is the wrong call once ranking is the feature.
  The comparison Yazi wants to publish is "free tier reaches X% of paid-tier
  quality at 0% of paid-tier recurring cost", and that sentence needs both axes
  measured.

## Risks / Trade-offs

- **[Index writes slow the write path]** → Batch postings updates via `MSet`;
  benchmark the ingest-latency column before and after; if a class of very long
  memories dominates, cap indexed terms per record and log the truncation rather
  than silently dropping terms.
- **[Index rebuild on upgrade is expensive for large corpora]** → Rebuild is
  incremental and resumable, records progress in a marker key, and the store
  serves reads (via the old scan path) while it runs.
- **[Postings lists for common terms grow large]** → Cap postings length per term
  with a documented policy (keep the highest term-frequency documents) and expose
  the truncation in the index statistics; no silent caps.
- **[Auto-supersede hides a memory the user wanted]** → Only `preference` is
  affected, superseded records remain readable via an explicit flag, and every
  supersede is logged with both ids.
- **[Per-class half-lives are guesses]** → They are configuration with defaults,
  and the quality suite is what turns the guess into a tuned value.
- **[Tuning to a small labelled set overfits]** → Keep the labelled set separate
  from the existing functional suites, report per-suite numbers rather than one
  aggregate, and treat it as a regression guard rather than a leaderboard.

## Migration Plan

1. Add the model fields (`validFrom`, `validUntil`, `supersededBy`, access
   counters) — additive and backward-compatible with stored JSON.
2. Land tokenizer, stopwords, postings write path and the incremental rebuild;
   recall still uses the old scorer. Verify indexes match a scan-derived oracle.
3. Switch candidate selection and scoring to index + BM25 behind a config flag
   defaulting to on; the previous scorer remains selectable for one release.
4. Add recency/frequency weighting and per-class half-lives.
5. Add consolidation and temporal filtering.
6. Add the labelled quality suite and publish before/after numbers.

Rollback: steps 3–4 are a config flag; the scan-based scorer stays in the tree
until the quality suite shows a regression-free release.

## Open Questions

- Should postings be per memory class or global across classes? Per class matches
  the existing index layout and keeps recall scoped; global would allow
  cross-class recall in one pass. Leaning per class, with recall iterating the
  classes it was asked for.
- Is CJK segmentation in scope for the first release? Yazi's users are likely to
  need it. Leaning: ship Unicode word segmentation now, add a bigram fallback for
  scripts without spaces as a fast follow, and say so in the docs.
- Should `access_count` be persisted on every read (a write per recall) or
  sampled/batched? A write per read would double the write volume; leaning
  toward in-memory counters flushed periodically, accepting loss on crash.
