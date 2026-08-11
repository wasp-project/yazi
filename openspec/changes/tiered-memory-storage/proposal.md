## Why

Yazi's entire positioning rests on "cheap, tiered storage: RAM → LSM on disk →
S3" (`STATE-OF-ART.md` §5.2, `README.md`, `ARCHITECTURE.md`). **That tiering does
not exist in the code.** Three findings, all verified:

1. `pkg/storage/lsm/sstable.go` loads every SSTable fully into a
   `map[string]entry` at startup. The LSM engine gives durability, not a memory
   bound — resident memory still scales with total data.
2. `pkg/server/server.go` selects the engine with `if engine == lsm { … } else if
   storage != "" { … }`. LSM and S3 are **mutually exclusive**: choosing the
   durable local engine bypasses cloud persistence entirely, so "local disk with
   an S3 cold tier" is unreachable by configuration.
3. When S3 *is* selected, persistence is a **whole-dataset snapshot** written to
   a single object. Cold storage means re-uploading and re-downloading
   everything, and the working set is still entirely resident in RAM.

Consequently the honest answer to "what does it cost to store 10 million
memories?" is "it OOMs". Token cost is already zero and cannot be improved;
**storage cost (C3) and operational footprint (C6) are the only remaining cost
axes**, and they are architectural — the longer they wait, the more expensive
they are to fix. This change turns the tiering story into running code and makes
it measurable.

## What Changes

- Rewrite the SSTable format with a **sparse block index and on-demand block
  reads**, so a table contributes bounded memory (index + block cache) instead of
  one map entry per key. Resident memory becomes a configured budget, not a
  function of dataset size.
- Add a **block cache** with an eviction policy, so hot keys stay in RAM and cold
  keys cost one disk read.
- Make **engine and cloud persistence orthogonal**: `engine: lsm` composes with
  `storage: s3` instead of short-circuiting it. **BREAKING** for the current
  `storage`/`engine` config semantics; a compatibility mapping keeps existing
  files working.
- Replace snapshot persistence with **segment-level cloud persistence**: sealed
  SSTables and manifest are uploaded as individual immutable objects, and cold
  segments are fetched on demand and cached locally. Uploading a new segment no
  longer rewrites the whole dataset.
- Add a **tiering policy** that classifies segments hot/warm/cold by access
  recency and frequency (not only age), demoting cold segments to object storage
  and promoting them back on access, with the local disk acting as a bounded
  cache.
- Extend cost metering with **storage cost**: bytes and object counts per tier,
  priced at $/GB-month, reported per tenant alongside the existing token
  accounting.
- Extend the benchmark harness with a **storage-cost dimension** so the
  comparison table can state monthly cost at a given corpus size — the number
  that actually differentiates Yazi from mem0/Zep/mem9.

## Capabilities

### New Capabilities
- `bounded-resident-memory`: SSTable block index, on-demand block reads, block
  cache, and a configurable resident-memory budget that data volume cannot
  exceed.
- `orthogonal-engine-and-persistence`: engine selection and cloud persistence
  become independent axes that compose.
- `segment-level-cloud-persistence`: immutable per-segment objects in S3 with a
  manifest, replacing whole-dataset snapshots; on-demand fetch with a local
  cache.
- `storage-tiering-policy`: hot/warm/cold classification by access recency and
  frequency, with demotion and promotion between RAM, local disk and object
  storage.

### Modified Capabilities
- `cost-metering-and-budgets`: the meter gains storage-cost accounting (bytes and
  objects per tier, $/GB-month) in addition to token usage, and the pricing table
  gains storage prices.

## Impact

- **New code**: block-index SSTable reader/writer and block cache in
  `pkg/storage/lsm`; `pkg/storage/tiering` (classification + promote/demote);
  segment-object store and manifest handling in `pkg/storage/s3`.
- **Modified code**: `pkg/storage/lsm/sstable.go` and `store.go` (format and read
  path), `pkg/server/server.go` (compose engine × persistence instead of
  branching), `pkg/config/config.go` (block/cache/tiering/`residentBudget`
  settings), `pkg/memory/provider/meter.go` and `pricing` (storage cost),
  `benchmark/framework/metrics.py` and `report.py` (storage-cost column).
- **On-disk format**: SSTable v2 is a new format. A one-time migration reads v1
  tables and rewrites them; the version is recorded in the file header.
- **Dependencies**: none new; the existing hand-rolled S3 client gains
  multi-object and range-read support.
- **Compatibility**: existing `config/local.yml` and `config/cloud.yml` continue
  to work through the compatibility mapping; existing data files are migrated on
  first start.
