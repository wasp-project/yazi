## 1. Model & Storage Fields

- [ ] 1.1 Add optional `validFrom`, `validUntil`, `supersededBy` and access-counter fields to the basic/advanced/policy models in `pkg/memory/model.go`, keeping stored JSON backward-compatible
- [ ] 1.2 Round-trip test: records written before this change decode unchanged; new fields survive put→get

## 2. Tokenizer & Index Write Path

- [ ] 2.1 Create `pkg/memory/index` with Unicode-aware, case-insensitive tokenization and a configurable English stopword list
- [ ] 2.2 Define the postings key layout `/_memory/index/term/<class>/<term>` → `{docID: termFreq}` and the corpus-statistics key (`docCount`, `avgFieldLen`, per-doc lengths)
- [ ] 2.3 Maintain postings and statistics on put/update/delete, batching writes through `MSet`
- [ ] 2.4 Cap postings length per term with a documented policy and expose the truncation in index statistics — never truncate silently
- [ ] 2.5 Implement an incremental, resumable index rebuild that records progress in a marker key and keeps the store readable while it runs
- [ ] 2.6 Test: postings match a scan-derived oracle after randomized put/update/delete sequences; delete removes all postings; two tenants' postings and statistics are isolated; an interrupted rebuild resumes

## 3. BM25 Ranking

- [ ] 3.1 Create `pkg/memory/rank` implementing BM25 with configurable `k1`/`b` and field weights (subject ×3, tags ×2, content ×1)
- [ ] 3.2 Replace `StoreBackedStore.Retrieve` candidate selection in `pkg/memory/composition.go` with index lookups over query terms, removing the hard-coded `Limit: 50` scan
- [ ] 3.3 Implement the all-stopword query fallback to unfiltered terms
- [ ] 3.4 Keep the previous scan-based scorer selectable by config for one release; default to the new ranker
- [ ] 3.5 Test: rare term beats common term; stopword-only overlap returns nothing; subject match outranks body match; long repetitive memory is saturated and length-normalized; a relevant memory beyond the old 50-record window is found
- [ ] 3.6 Regression test for the observed defect: query "where is the deploy script" returns the deploy memory and does not return the unrelated Postgres memory

## 4. Recency & Frequency

- [ ] 4.1 Implement multiplicative scoring `bm25 × recencyDecay × frequencyBoost` with exponential decay
- [ ] 4.2 Add per-kind half-life defaults (preference 365d, decision 180d, history 30d, context 7d) and make them configurable
- [ ] 4.3 Track access counts in memory and flush them periodically rather than writing on every read; document the crash-loss trade-off
- [ ] 4.4 Extend `provider.Hit` to carry the relevance, recency and frequency components of the score
- [ ] 4.5 Test: newer wins at equal relevance; a zero-relevance memory is never surfaced by recency; preference outlives context at equal age; frequency boost is sub-linear; hits expose their components

## 5. Deterministic Consolidation

- [ ] 5.1 Implement exact-duplicate collapse keyed on normalized content + kind + scope + subject, returning the existing id and refreshing its timestamp
- [ ] 5.2 Implement preference supersession: a new `preference` for the same scope+subject marks the previous record `supersededBy` and keeps it stored
- [ ] 5.3 Exclude superseded records from recall by default; add an explicit option to include them
- [ ] 5.4 Log every collapse and supersession with both record ids
- [ ] 5.5 Test: duplicate writes yield one record and one id; near-duplicates stay separate; superseded preference is excluded by default and retrievable on request; history/context/decision are never auto-superseded

## 6. Temporal Validity

- [ ] 6.1 Apply validity filtering to recall and list, defaulting to "valid now"
- [ ] 6.2 Add an `asOf` parameter to the memory service, the CLI, and the HTTP API from `agent-integration-surface`
- [ ] 6.3 Apply validity filtering before ranking and compose it with supersession without double-exclusion
- [ ] 6.4 Test: historical query excludes later facts; expired memory absent from current recall; point-in-time recall reports zero token cost

## 7. Quality Benchmark

- [ ] 7.1 Add a labelled dataset (query → relevant memory ids) under `benchmark/testcases/datasets/`, separate from the functional suites
- [ ] 7.2 Implement precision@k, recall@k and MRR in `benchmark/evaluation/evaluators.py`
- [ ] 7.3 Report quality columns beside the cost columns per profile; keep `sim-*` rows flagged as estimates and exclude them from measured-quality claims
- [ ] 7.4 Add a regression mode comparing against recorded baselines and reporting declines with the affected queries
- [ ] 7.5 Update `benchmark/README.md`: accuracy is no longer held constant for Yazi's own profiles; explain why and how the two axes are now read together
- [ ] 7.6 Run the suite before and after this change and record the delta

## 8. Documentation & Verification

- [ ] 8.1 Update `MEMORY.md` with the new fields, filters (`asOf`, include-superseded) and consolidation rules
- [ ] 8.2 Update `COMPOSITION.md` and `README.md` to describe `lite` as BM25 + recency/frequency at zero token cost
- [ ] 8.3 Update `QUICKSTART.md`, removing the ranking-quality entry from Known limits
- [ ] 8.4 Publish the measured free-tier-versus-paid-tier quality/cost comparison in `STATE-OF-ART.md`
- [ ] 8.5 Run `CGO_ENABLED=0 go build ./...` and `CGO_ENABLED=0 go test ./...`; confirm the full suite passes

## 9. Engineering Blog

- [ ] 9.1 Write the milestone blog post under `blog/` per the `engineering-blog` capability: why "the" broke recall, the BM25 + decay + frequency scoring model, exact-match consolidation as an anti-LLM design choice, and the measured precision/MRR before and after at unchanged zero token cost
