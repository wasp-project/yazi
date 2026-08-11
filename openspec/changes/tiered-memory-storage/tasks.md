## 1. SSTable v2 — Block Index

- [ ] 1.1 Define the v2 file layout in `pkg/storage/lsm`: sorted blocks (target size configurable, default 4 KiB), sparse index (first key → offset/length per block), footer (magic, version, index offset, entry count, min/max key, seq range)
- [ ] 1.2 Implement the v2 writer: buffer entries into blocks, emit the sparse index and footer, fsync, return the table descriptor
- [ ] 1.3 Implement the v2 reader: open reads footer + sparse index only; `Get` binary-searches the index, reads one block, and skips the table entirely when the key falls outside the recorded range
- [ ] 1.4 Implement range/iteration reads over blocks for `Keys()` and compaction merge without materializing whole tables
- [ ] 1.5 Detect v1 (full-map) tables by absent footer magic and rewrite them to v2 on startup; log the migration; add a `--no-migrate` mode that exits with an error instead
- [ ] 1.6 Ship a `v2→v1` rewrite tool so step 1 is reversible
- [ ] 1.7 Test: table round-trip; key outside range skips block reads; migration preserves every key/value; property test comparing engine reads against an in-memory oracle over randomized write/delete/compact sequences

## 2. Block Cache & Resident Budget

- [ ] 2.1 Add a process-wide block cache keyed by `(tableID, blockOffset)` with LRU eviction and a byte budget
- [ ] 2.2 Add `storage.residentBudgetMB` config; subtract memtable size from the budget and size the block cache with the remainder
- [ ] 2.3 Export cache counters: hits, misses, evictions, resident bytes
- [ ] 2.4 Test: with a dataset many times the budget, resident memory stays within budget and all keys remain readable; evicted blocks are re-read correctly

## 3. SegmentStore Abstraction

- [ ] 3.1 Define `SegmentStore` (`Put`, `Get`, `GetRange`, `List`, `Delete`, returning a distinguishable not-found) in `pkg/storage`
- [ ] 3.2 Implement the local-disk `SegmentStore` and route all sealed-table IO in the LSM engine through it (no behavior change yet)
- [ ] 3.3 Test: local segment store round-trip including ranged reads; LSM tests pass unchanged through the new indirection

## 4. Object-Storage Segments & Manifest

- [ ] 4.1 Extend `pkg/storage/s3` with multi-object support: `Put`/`Get`/`Delete` by name, `List` by prefix, and ranged `GET` via the `Range` header, all through the existing SigV4 signer
- [ ] 4.2 Implement the S3 `SegmentStore` writing segments as immutable objects under `segments/<seq>-<id>.sst`
- [ ] 4.3 Implement the manifest: live segment list with key ranges, sizes and tiers; publish-manifest is the commit point; reconstruct state from the manifest at startup
- [ ] 4.4 Implement orphan GC: list objects unreferenced by the manifest and older than the grace period, delete them, log the reclaim
- [ ] 4.5 Add a directory lock file so a second process cannot open the same data directory
- [ ] 4.6 Test against the in-process S3-compatible `httptest` endpoint: segment put/get/ranged-get; crash between upload and manifest publish leaves committed state intact and the orphan reclaimable; upload volume is proportional to new data, not corpus size

## 5. Orthogonal Engine × Persistence

- [ ] 5.1 Replace the `if engine == lsm / else if storage != ""` branch in `pkg/server/server.go` with independent engine and backend selection that compose
- [ ] 5.2 Add the compatibility mapping so existing `config/local.yml` and `config/cloud.yml` produce equivalent behavior unmodified
- [ ] 5.3 Keep whole-blob `PersistentStorage` snapshotting for the `mem` engine; route the `lsm` engine to `SegmentStore`
- [ ] 5.4 Preserve fail-fast on incomplete backend configuration (missing bucket/region/credentials) with no silent fallback
- [ ] 5.5 Test: all four engine × backend combinations start and serve correctly; both legacy config files start unchanged; missing S3 settings abort startup

## 6. Tiering Policy

- [ ] 6.1 Track per-segment access metadata: last-access time, decaying access count, size
- [ ] 6.2 Implement the hot/warm/cold classifier scoring segments by recency **and** frequency
- [ ] 6.3 Implement demotion (warm→cold): upload if absent, confirm durability, only then unlink the local file; retain the file and log on upload failure
- [ ] 6.4 Implement promotion (cold→warm) on first read touching the segment's key range, with a bounded local segment cache evicting by the same score
- [ ] 6.5 Add `storage.localCacheMB`, the demotion window, and the classifier interval to config
- [ ] 6.6 Export per-tier counters: segment count, bytes, read hits/misses, promotions, demotions; surface them in the storage status output
- [ ] 6.7 Test: an old-but-hot segment survives demotion while a new-but-cold one is demoted; a full write→demote→read cycle returns every original value; local cache respects its limit; failed upload aborts demotion without data loss

## 7. Storage Cost Metering

- [ ] 7.1 Add `StorageFootprint` (bytes and segment counts per tier, per tenant) as a point-in-time snapshot distinct from `provider.Usage`
- [ ] 7.2 Extend `provider.Pricing` with per-tier `$/GB-month` and per-1k object-storage request prices
- [ ] 7.3 Report per-tenant cost as two separate line items — recurring token cost and at-rest storage cost — never blended into one total
- [ ] 7.4 Expose the footprint and cost report through the memory service and the HTTP API added in `agent-integration-surface`
- [ ] 7.5 Test: footprint per tier and tenant; monthly storage cost recomputes when prices change without re-running operations; deterministic profile reports zero token cost with non-zero storage cost

## 8. Benchmark: The Cost Number

- [ ] 8.1 Add storage-cost fields to `benchmark/framework/metrics.py` mirroring the Go `Pricing`/`StorageFootprint` shapes
- [ ] 8.2 Add a corpus-scale parameter to the runner so a suite can be replayed at 10k / 100k / 1M memories
- [ ] 8.3 Extend the `yazi` adapter to read the real footprint from the running server; extend the `sim-*` estimators with each system's documented storage profile, still flagged as estimates
- [ ] 8.4 Add `$/month @ 1M memories` and a separate cold-recall latency column to the report; do not average cold and warm latency together
- [ ] 8.5 Run the comparison and record the resulting figures in `STATE-OF-ART.md`, replacing the qualitative "cheap (RAM→disk→S3)" cell with a measured number

## 9. Documentation

- [ ] 9.1 Update `ARCHITECTURE.md`: the tier model, the `SegmentStore` seam beside `PersistentStorage`, the manifest commit protocol, and the v2 file layout
- [ ] 9.2 Update `README.md`, `QUICKSTART.md` and `STATE-OF-ART.md` to describe tiering as implemented, and remove the Known-limits entries this change closes
- [ ] 9.3 Document operational guidance: sizing `residentBudgetMB` and `localCacheMB`, expected cold-read latency, and object-storage request-cost behavior
- [ ] 9.4 Run `CGO_ENABLED=0 go build ./...` and `CGO_ENABLED=0 go test ./...`; confirm the full suite passes

## 10. Engineering Blog

- [ ] 10.1 Write the milestone blog post under `blog/` per the `engineering-blog` capability: why a "tiered" system was not tiered, the SSTable v2 and manifest design, the recency-and-frequency classifier, and measured resident memory / cold-read latency / $-per-month before and after
