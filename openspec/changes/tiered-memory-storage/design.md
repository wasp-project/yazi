## Context

The current storage stack, as implemented:

- `pkg/storage/memory.go` — an in-RAM cache (`memcache` or LRU) holding the whole
  dataset, with `Encode()`/`Decode()` serializing all of it.
- `pkg/storage/lsm` — WAL + SSTables. `loadSSTables` calls `readSSTable`, which
  builds `map[string]entry` for **every** key in **every** table
  (`sstable.go:76-125`). `Get` probes the memtable, then walks tables newest to
  oldest, hitting those in-RAM maps. Compaction merges tables.
- `pkg/storage/local` and `pkg/storage/s3` — `PersistentStorage.Write([]byte)` /
  `Read([]byte)`, i.e. one blob for the entire store.
- `pkg/server/server.go:80` — `if Engine == lsm { store = lsm.NewStore(...) }
  else if Storage != "" { … persistent … }`. The LSM branch never assigns
  `persistent`, so the manager's snapshot task is never scheduled.

So there are two independent persistence mechanisms that cannot be combined, and
neither bounds memory. Every claim about tiering in the docs describes the
*intended* architecture.

Constraints inherited from the project: single static binary, no new module
dependencies, no mandatory external service, `lite` recall stays $0 in tokens,
and the `/_memory/...` key layout is fixed by the memory layer above.

The important asymmetry for design: **agent memory is write-once, read-rarely,
and overwhelmingly cold.** A year of an agent's memories is mostly never read
again; a small recent/frequent subset carries nearly all reads. That is precisely
the workload object storage is priced for, and precisely the workload a
fully-resident index wastes money on.

## Goals / Non-Goals

**Goals:**
- Resident memory is a **configured budget**, independent of corpus size.
- `engine` and cloud persistence compose freely; local LSM + S3 cold tier works.
- A cold read is a bounded number of range GETs, not a full-dataset download.
- Writes to cloud storage are incremental (new sealed segments only).
- Tier placement is driven by access recency **and** frequency, and is
  observable.
- Storage cost is metered per tenant and reported by the benchmark.

**Non-Goals:**
- A distributed storage layer, sharding, or a replication redesign
  (`pkg/replication` stays as-is).
- Compression or bloom filters in this change — both are natural follow-ups and
  the v2 format reserves header space for them.
- Changing the memory model, index scheme, or recall semantics.
- Multi-writer or cross-process concurrency on the same data directory.
- Sub-key/columnar storage; the unit of storage remains a KV entry.

## Decisions

### Decision 1: SSTable v2 — sorted blocks with a sparse index

Format: a sequence of ~4 KiB blocks of sorted key/value entries, followed by a
sparse index (one entry per block: first key + offset + length), followed by a
footer (magic, version, index offset, entry count, min/max key, sequence range).
Loading a table reads **only the footer and the sparse index**; a `Get` binary-
searches the index and reads exactly one block.

- **Why**: This is the standard LSM design and it is the only way to decouple
  resident memory from dataset size. Memory per table becomes O(blocks), roughly
  1/100th of O(keys) at typical memory-record sizes.
- **Alternatives**: (a) Keep full maps and cap total entries — rejected, it caps
  the product, not the memory. (b) mmap the tables and rely on the page cache —
  tempting and simpler, but it makes resident memory unmeasurable and
  unbudgetable, which defeats the "cost is observable" commitment; it also
  behaves poorly for the S3-backed tier where there is no local file to map.
- **Migration**: v1 tables are detected by their absent footer magic and
  rewritten to v2 on first start, logged as a one-time migration.

### Decision 2: A shared block cache with an explicit byte budget

One process-wide block cache (LRU over `(tableID, blockOffset)`) sized by
`storage.residentBudgetMB`. Memtable size is subtracted from that budget.

- **Why**: Makes "how much RAM does Yazi use" a config value the operator sets,
  which is the prerequisite for the cost claim we want to publish. A single
  shared cache also lets a hot segment keep its blocks resident while cold
  segments hold only their sparse index.
- **Trade-off**: A cache miss on a cold read costs one disk seek (or one S3 range
  GET). Accepted — that is the definition of a cold tier.

### Decision 3: Engine × persistence become orthogonal, via a segment store

Replace the `if lsm / else if storage` branch with two independent selections:
the **engine** (`mem` | `lsm`) and the **segment store** (`local` | `s3`), the
latter being where sealed segments live. `PersistentStorage` (whole-blob
Write/Read) is retained for the `mem` engine only; the LSM engine talks to a new
`SegmentStore` interface:

```
type SegmentStore interface {
    Put(name string, data []byte) error
    GetRange(name string, off, length int64) ([]byte, error)
    Get(name string) ([]byte, error)
    List(prefix string) ([]SegmentInfo, error)
    Delete(name string) error
}
```

- **Why a second interface rather than extending `PersistentStorage`**:
  `ARCHITECTURE.md` §4 already anticipates this ("If it needs per-record rather
  than snapshot semantics, define a sibling interface beside
  `PersistentStorage`"). Snapshot and segment semantics are genuinely different
  contracts; merging them would force every backend to implement both.
- **Config compatibility**: `storage: local` + `engine: lsm` maps to a local
  segment store (today's behavior plus tiering); `storage: s3` + `engine: mem`
  maps to today's snapshot behavior; the newly-reachable `storage: s3` +
  `engine: lsm` is the tiered mode. No config file becomes invalid.

### Decision 4: Immutable segment objects + a manifest, never a rewritten blob

Each sealed SSTable is uploaded once, under a content-addressed name
(`segments/<seq>-<id>.sst`), and never mutated. A small `manifest.json` lists
live segments, their key ranges, sizes, and tier. Compaction writes new segments
and publishes a new manifest; superseded objects are deleted after a grace
period.

- **Why**: Immutability makes S3 semantics trivially correct (no read-modify-
  write races, no partial-object updates) and makes incremental upload cost
  proportional to *new* data, which is the entire point. The manifest is the only
  mutable object, and it is small.
- **Consistency**: manifest publish is the commit point; a crash between segment
  upload and manifest publish leaves an orphan object, cleaned by a GC pass that
  lists objects not referenced by the manifest and older than the grace period.
- **Alternatives**: Append to a single object (S3 has no append), or keep the
  snapshot and diff it (requires reading the whole object to write a delta).

### Decision 5: Tier classification by access recency **and** frequency

Each segment carries counters: last-access time, access count in a decaying
window, and byte size. A background pass classifies:

- **hot** — blocks resident in the cache; recently or frequently read.
- **warm** — segment file present on local disk, sparse index resident.
- **cold** — segment lives only in object storage; only its manifest entry and
  key range are resident.

Demotion (warm→cold) uploads if needed, then deletes the local file. Promotion
(cold→warm) downloads on the first read that touches the segment's key range;
the local disk cache is bounded by `storage.localCacheMB` and evicts by the same
recency/frequency score.

- **Why frequency and not just age**: `STATE-OF-ART.md` §5.2 commits to exactly
  this ("by access frequency, not just age"). A three-month-old preference that
  is read every session is hot; yesterday's transcript chunk is not.
- **Why segment granularity, not record granularity**: records are small
  (hundreds of bytes) and object-per-record would make S3 request cost dominate
  storage cost. Segment granularity keeps request counts proportional to reads,
  and the LSM already produces segments as a natural unit.
- **Risk acknowledged**: segment granularity means one hot key can pin a whole
  segment warm. Compaction naturally re-groups by key range over time; we accept
  the imprecision and measure it (hit rate per tier is exported).

### Decision 6: Storage cost joins the meter as a first-class number

`provider.Pricing` gains `StorageGBMonth` per tier and `RequestsPer1k`;
`provider.Usage` is unchanged (it stays a per-operation token record). Storage
cost is a **stock**, not a flow, so it is reported by a separate
`StorageFootprint` snapshot (bytes and object count per tier, per tenant) that
the meter can price into $/month.

- **Why not fold storage bytes into `Usage`**: `Usage` is accumulated per
  operation and summed; bytes-at-rest are not additive over operations. Mixing
  them would produce nonsense totals.
- **Benchmark**: `benchmark/framework/metrics.py` gains the same two fields so
  the report can print "$/month at 1M memories" beside "$/1k ops". This is the
  headline number the project currently cannot state.

## Risks / Trade-offs

- **[On-disk format change corrupts existing data]** → v1 detection is by footer
  magic; migration writes v2 tables to new files and only unlinks v1 files after
  a successful manifest publish. A `--no-migrate` flag refuses to start rather
  than touching data the operator wants to inspect.
- **[Cold-read latency becomes user-visible]** → An S3 range GET is ~30–100 ms
  versus ~0.1 ms warm. Mitigation: recall reads the index tier first and touches
  cold segments only for ids it has already decided to return; the tier of every
  hit is reported so latency is attributable. Measured in the benchmark as a
  separate cold-recall column rather than averaged away.
- **[S3 request cost replaces storage cost]** → GETs are billed per request.
  Mitigation: block cache + local disk cache + segment (not record) granularity;
  the meter reports request counts so this is visible rather than surprising.
- **[Complexity in the one component that must never lose data]** → The WAL path
  is untouched by this change; all new machinery sits below the seal point.
  Property tests: for random write/read/delete/compact/demote/promote sequences,
  a v2 store must return exactly what an in-memory oracle map returns.
- **[Manifest becomes a single point of contention]** → Single-writer only; the
  design explicitly excludes multi-process access to one data directory, and
  startup takes a directory lock file to enforce it.
- **[Scope]** → This is the largest change in the roadmap. It is sequenced after
  `agent-integration-surface` so that adoption and feedback precede it, and its
  tasks are ordered so each group leaves the tree green.

## Migration Plan

1. Land SSTable v2 (writer, reader, block cache) behind the existing `local`
   path, with v1→v2 migration on start. No config change; behavior identical
   except resident memory.
2. Introduce `SegmentStore` with the local implementation; move sealed-table IO
   behind it. Still no config change.
3. Add the S3 `SegmentStore` and the manifest, and make engine × persistence
   orthogonal in the server factory with the compatibility mapping.
4. Add tiering classification, demotion/promotion, and the local disk cache.
5. Add storage-cost metering and the benchmark column; publish the
   $/month-at-1M-memories figure.

Rollback: steps 1–2 are format-only and reversible by a v2→v1 rewrite tool
(shipped with step 1). Steps 3–5 are additive and disabled by leaving
`storage: local`.

## Open Questions

- Block size default: 4 KiB matches typical memory-record sizes, but agent
  transcripts skew larger. Decide after measuring the real distribution in the
  benchmark corpus.
- Should the manifest be per-tenant or global? Global is simpler; per-tenant
  makes tenant-scoped GC and per-tenant storage billing exact. Leaning per-tenant
  prefixes within one global manifest.
- Does the `mem` engine keep whole-blob snapshot persistence indefinitely, or is
  it eventually demoted to a test-only engine? Leaning: keep it — it is the
  zero-config default and it is honest about being RAM-bound.
