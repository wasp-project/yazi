---
title: From KV Store to Agent Memory
date: 2026-08-10
change: (baseline — the state before the roadmap)
author: Yazi maintainers
---

# From KV Store to Agent Memory

Yazi started as a small Go key-value server and is now positioned as a
**cost-aware memory engine for LLM agents**: structured memory, tenant isolation,
a provider pipeline with metered cost, and a documented RAM → disk → S3 storage
hierarchy. On a 10,000-record store it answers a recall in **7.4 ms for exactly
$0.00** — no LLM call, no embedding call, no vector database.

That number is real and it is the whole point of the project. This post explains
how the code got there. It also reports the five things the documentation
currently claims that the code does not yet do — including a `standard`-profile
recall that reports a cost of `$0.00000018` while actually spending
**$0.00042**, a factor of **2,334**. Post 001 onward is about closing that list.

---

## 1. Design — how a KV server became a memory system

### The starting point

The original Yazi was a conventional KV server with three pluggable axes: a
protocol (`naive` or gRPC), a cache policy (LRU or a plain map), and a
persistence target (a local snapshot file). About 4,400 lines of Go at commit
`7a3265c`, against 8,570 today. `pkg/storage` defined two interfaces that still
anchor the whole system:

```go
type KVStore interface {
    Get/Set/Expire/Del/MSet/MGet/Keys(...)
    Encode() []byte        // snapshot serialization
    Decode([]byte) error
}

type PersistentStorage interface {
    Write(data []byte) (int, error)   // persist a snapshot
    Read(data []byte) (int, error)    // load a snapshot
}
```

Note the shape of `PersistentStorage`: it moves **the entire dataset as one
blob**. That decision, reasonable for a cache with a snapshot file, propagates
through everything below and is the root of the largest gap described in §4.

### The problem that changed the project

Agents need to remember things. The obvious move — "use a KV store, the agent can
serialize whatever it wants" — fails immediately, because the agent then has to
invent its own key scheme, its own indexes, and its own filtering. Every agent
reinvents a slightly broken one.

But the popular answer at the other extreme is expensive. Surveying mem0, Zep,
Milvus, Letta, MemOS and mem9 (written up in
[STATE-OF-ART.md](../STATE-OF-ART.md)) produced one observation that set the
project's direction:

> Every capability-first memory system has at least one cost that is charged
> **per message or per turn** — LLM extraction on write, embeddings on write and
> query, or an LLM call inside retrieval. Those costs do not amortize. They scale
> linearly with traffic, forever.

So Yazi took a narrow position: **compete on cost, not capability.** Keep a
deterministic core that always works at ~zero marginal token cost, and make every
expensive feature opt-in, metered and budgeted.

### Constraints adopted, and kept

Four rules have survived every change so far, and they explain most of the code:

1. **Single static binary.** No mandatory database, queue, or coordinator.
   `go.mod` still has six direct dependencies.
2. **No new dependency for a solvable problem.** S3 support required AWS SigV4;
   `aws-sdk-go-v2` needs Go ≥ 1.24 and pulls a large tree, so
   [`pkg/storage/s3`](../pkg/storage/s3/s3.go) hand-rolls the signer over the
   standard library instead.
3. **The default costs zero tokens.** Any feature that spends must be selected
   explicitly.
4. **Layers depend downward only**, enforced by a test rather than a convention —
   [`pkg/arch`](../pkg/arch/arch_test.go) fails the build if the storage engine
   imports the tenant or UI layer.

---

## 2. Principle — why cost is the axis

The cost of an agent memory system is not one number. It has six distinct points,
and they behave differently:

| | Cost point | What drives it | Amortizes? |
| --- | --- | --- | --- |
| C1 | Write / ingest | LLM calls to extract and reconcile facts | **No** — per message |
| C2 | Embedding | Encoding text at write *and* query time | **No** — per message and per query |
| C3 | Storage | Where records live; RAM-resident indexes are the expensive case | Yes — $/GB-month |
| C4 | Retrieval → context | Tokens injected into the prompt each turn | **No** — per turn |
| C5 | Inference reuse | Re-encoding the same memory every call | — |
| C6 | Operational | Vector DB + graph DB + queue + coordinators + people | Yes — $/month |

The systems that compete on capability accept recurring C1/C2/C4 as the price of
semantic recall. That is a defensible trade at demo scale. At fleet scale — many
agents × high message volume × long horizons — it is the thing that breaks.

Yazi's bet is that a deterministic core makes C1, C2 and C4 approximately zero,
and that the resulting quality gap is smaller than it looks for the memories
agents actually store: preferences, decisions, project facts. Those are short,
structured, and retrieved by attributes far more often than by fuzzy similarity.

**The alternative we rejected:** build a good vector-search memory layer and
optimize its cost afterwards. Rejected because cost is not a constant factor you
tune at the end — it is set by whether an LLM sits on the write path. Once
extraction is mandatory, no amount of optimization removes the per-message
charge. The cheap path has to be the default path, or it is not a cost story.

The corollary is the part that is easy to miss: **cost-aware does not mean
free.** It means the expensive path exists, and the meter is visible. That is why
`Usage`, `Meter`, `Pricing` and `Budget` were built before any provider that
could actually spend money.

---

## 3. Implementation — what exists today

### Four layers

```
Layer 1 · User Interface   cmd/cli (yazictl), pkg/server (gRPC/naive), pkg/client
Layer 2 · Tenant Service   pkg/tenant — Resolve(Request) → TenantContext.KeyPrefix()
Layer 3 · Storage Engines  pkg/memory (model + indexes), pkg/storage (cache, LSM, replication)
Layer 4 · Cloud Vendors    pkg/storage/local (disk), pkg/storage/s3 (SigV4, no SDK)
```

Each boundary is a Go interface; the direction is enforced by
[`pkg/arch`](../pkg/arch/arch_test.go). Full diagrams are in
[ARCHITECTURE.md](../ARCHITECTURE.md).

### The memory model

Three record types on top of plain KV, each stored as JSON with secondary indexes
as id-lists:

| Type | Key | Indexes |
| --- | --- | --- |
| `BasicMemory` — preference / history / decision / context | `/_memory/basic/{id}` | kind, scope, subject, tag |
| `AdvancedMemory` — distilled experience | `/_memory/advanced/{id}` | template, tag |
| `CognitivePolicy` — role/domain/scenario policy | `/_memory/policy/{id}` | role, domain, scenario, tag |

A filtered `list` reads one index key and one `MGet`, so it never scans the
corpus. Measured at 10,000 records: **2 KV calls, 0.67 ms** (§4).

### Tenancy as a key prefix

`tenant.Service.Resolve()` returns a prefix; a `prefixKV` decorator prepends it on
writes and point reads, and strips it in `Keys()` so the memory store still sees
plain `/_memory/...` shapes. An empty prefix means no wrapping at all — the
single-tenant path is byte-identical to the pre-tenancy layout.

```
Bypass:          /_memory/basic/<id>
Tenant "acme":   tenant/acme/_memory/basic/<id>
```

**Why a prefix rather than a storage wrapper:** isolation is a namespacing
concern, and Yazi already namespaces (`/_data`, `/_meta`, `/_raft`). The engine
never learns what a tenant is, so an external multi-tenant implementation only
has to satisfy one two-method interface.

### Cost as a first-class type

The composition layer ([`pkg/memory/provider`](../pkg/memory/provider)) models the
memory path as a pipeline of swappable stages:

```
WRITE:  Item ─▶ [Extractor] ─▶ [Distiller] ─▶ [Embedder] ─▶ [Index]
READ:   Query ─▶ [Retriever] ─▶ [Reranker]  ─▶ [ContextBudgeter] ─▶ Hits
```

Every stage returns a `Usage` (LLM in/out tokens, embed tokens, context tokens,
latency). A `Meter` aggregates it per tenant and memory class; a `Pricing` table
converts it to dollars and can re-price recorded usage without re-running
anything. `Budget` caps recall context tokens and ingest cost, with `reject` or
`degrade` enforcement.

Profiles bind the stages: `lite` (deterministic, $0), `standard` (local embedder +
vector index), `pro` (+ LLM extraction and rerank), `custom`.

The part worth defending: `LLMExtractor`, `LLMReranker`, `GraphRetriever` and
`PgVectorIndex` exist as interfaces that report `Available() == false`. Selecting
one fails fast rather than silently degrading. **Stubs that admit they are stubs
are better than implementations that pretend.** The cost of that honesty is that
`pro` currently does not run at all.

---

## 4. Practice — the measurements, and the gap list

> Measured on an Intel Core i7-9750H @ 2.60 GHz, 16 GB RAM, macOS 26.4.1,
> go 1.22.6 darwin/amd64. LSM engine (`memtableMaxEntries: 1024`), gRPC, single
> tenant. Memory operations issued over one long-lived connection by a probe
> program that wraps the client KV in a counter — so "KV calls" below are actual
> wire round-trips.

### Reproducing the basics

```bash
CGO_ENABLED=0 go build -o bin/yazi ./cmd/yazi && CGO_ENABLED=0 go build -o bin/yazictl ./cmd/cli
# config/ is gitignored, so a fresh clone has no config file — write one first.
# QUICKSTART.md §3 has a working config/default.yml to paste. See gap ⑤ below.
bin/yazi &
bin/yazictl memory basic put --json '{"kind":"decision","scope":"project","subject":"db","content":"use postgres for the billing service","tags":["arch","db"]}'
bin/yazictl memory recall --query "which database do we use" --top-k 3
```

```json
{
  "costUSD": 0,
  "hits": [{ "id": "…", "text": "use postgres for the billing service", "score": 0.166 }],
  "profile": "lite",
  "usage": { "llmInputTokens": 0, "llmOutputTokens": 0, "embedTokens": 0, "contextTokens": 9, "latencyMs": 0 }
}
```

A warm `CGO_ENABLED=0 go build ./...` takes **1.8 s**; `go test ./...` passes on
all nine test packages.

### 4.1 What the deterministic core actually costs

| Operation (10,000 records) | Latency | KV round-trips | Tokens | $ |
| --- | --- | --- | --- | --- |
| `PutBasic` | 10.8 ms | **9** | 0 | 0 |
| `GetBasic` | 0.17 ms | 1 | 0 | 0 |
| `ListBasic{tag, limit:10}` | 0.67 ms | 2 (`Get` + `MGet` 10) | 0 | 0 |
| `recall --profile lite` | 7.35 ms | 2 (`Keys` + `MGet` 50) | 0 | **0** |
| `recall --profile standard` | 18.5 ms | 2 (`Keys` + `MGet` 1000) | 9 reported | see 4.4 |

Storage and residency as the corpus grows:

| Records | Resident (RSS) | On disk | `lite` recall |
| --- | --- | --- | --- |
| 1,000 | 16.3 MB | 892 KB | 1.63 ms |
| 5,000 | 24.8 MB | 8.1 MB | 5.30 ms |
| 10,000 | 40.4 MB | 22.8 MB | 7.35 ms |

Restart with 10,000 records: **61 ms** from process start to first served query,
all records intact, RSS 28.2 MB.

The headline holds: **recall costs zero tokens and single-digit milliseconds, on
a laptop, with one binary and no database.** No other system in the comparison
does that.

### 4.2 Five things the docs claim and the code does not do

This is the honest part, and it is why the roadmap looks the way it does.

**① The storage tiering does not exist.** README, ARCHITECTURE and STATE-OF-ART
all describe "RAM → LSM on disk → S3, tiered". In the code,
`pkg/storage/lsm/sstable.go:76-125` loads **every SSTable fully into a
`map[string]entry`** at startup. The LSM engine buys durability, not a memory
bound. The table above shows RSS tracking the corpus at roughly 2.5 KB per record
— *extrapolating* that slope (an extrapolation, not a measurement) puts 10 million
memories at ~25 GB resident. It would OOM long before that.

Worse, `pkg/server/server.go:80` selects the engine with
`if engine == lsm { … } else if storage != "" { … }`. LSM and S3 are **mutually
exclusive**: choosing the durable local engine bypasses cloud persistence
entirely. "Local disk with an S3 cold tier" is not reachable by configuration.
And when S3 *is* selected, `PersistentStorage.Write([]byte)` uploads the **whole
dataset as one object** — cold storage means re-uploading everything.

→ Post 001, change `tiered-memory-storage`.

**② The memory model runs on the client.** `cmd/cli/yazi_ctl.go:620` builds the
`memory.Store` client-side; only raw KV crosses the wire. The server holds a
`tenant.Service` it never uses (`pkg/server/server.go:47`, comment: "the choke
point for a future server-side memory RPC"). Consequences: tenant isolation is
enforced by a *cooperative client*, a memory write costs 9 round-trips, and any
non-Go agent must shell out to `yazictl` — there is no HTTP API and no MCP
server. Measured CLI overhead per invocation: **19.8 ms** of process spawn and
dial before any work happens.

→ Post 001's sibling, change `agent-integration-surface`.

**③ Free-tier ranking is a term-overlap count.**
`pkg/memory/composition.go:82-100` scores by counting shared words, with no
stopwords, no term weighting and no recency. Query *"where is the deploy script"*
returns the correct memory **and** an unrelated Postgres decision, because both
contain *"the"*. Candidates are also capped at `Limit: 50` **before** scoring, so
past 50 records relevance partly depends on insertion order.

Related, and visible in the write numbers above: `indexAdd`
(`pkg/memory/store.go:483`) does a read-modify-write of the entire posting list on
every write. Per-put latency rose from **5.27 ms** (growing 1k→5k) to **10.81 ms**
(growing 5k→10k) for exactly that reason — ingest is quadratic in the number of
records sharing a tag.

→ Change `zero-cost-recall-quality`.

**④ The `standard` profile's cost report is off by 2,334×.**
`memory.RecallWithProfile` (`composition.go:141-152`) rebuilds the vector index on
**every recall** by re-ingesting up to 1,000 records. Embedding — a write-time
cost — is being paid at read time, per query. And the returned `Usage` covers only
the query embedding, because the per-item ingest usages are discarded by the
caller.

Measured on the 10,000-record store, one `standard` recall:

| | Embed tokens | Cost |
| --- | --- | --- |
| **Reported** to the caller | 9 | $0.00000018 |
| **Actually metered** for the same call | 21,009 | $0.00042018 |

For a project whose thesis is that cost should be visible, a cost report that
under-states spend by three orders of magnitude is the most serious defect in the
tree.

→ Change `metered-paid-capabilities`.

**⑤ Small things that block the first five minutes.** `config/**` is
gitignored and the config path is hardcoded to `./config/default.yml` with no
`--config` flag, so a fresh clone has no config file and no way to point at one.
`--profile pro` terminates with a Go panic and a stack trace. `-P/-H/-p` are
registered as non-persistent root flags, so `yazictl -P 3466 memory basic list`
fails with `unknown shorthand flag: 'P'` — **and exits 0**. Tenant keys contain a
doubled separator (`tenant/acme//_memory/...`).

→ Change `agent-integration-surface`.

### 4.3 What to do differently, today

- **Create `config/default.yml` yourself.** It is not in the clone. The
  [QUICKSTART](../QUICKSTART.md) inlines a working one.
- **Use `--profile lite`** (the default). It is the only profile whose cost report
  you can trust today.
- **Assume the working set fits in RAM.** Budget roughly 2.5 KB resident per
  memory record until post 001 lands.
- **Talk to the server on the default port**, or use the Go client directly —
  the port flag does not reach the memory subcommands.

### 4.4 What did not improve, and what is not measured

Nothing regressed here — this is a baseline, so there is nothing to compare
against. Two honest gaps in the measurements themselves:

- **No retrieval-quality number exists.** The benchmark deliberately holds
  accuracy constant across dependency-free adapters
  ([`benchmark/README.md`](../benchmark/README.md)) so the cost axis is isolated.
  That was right when `lite` had no ranking worth measuring; it means this post
  cannot tell you how *good* the free recall is, only what it costs. A labelled
  quality suite is part of `zero-cost-recall-quality`.
- **No storage cost number exists.** The benchmark prices tokens, where Yazi
  already scores zero. It does not yet price $/GB-month, which is the axis where
  the tiering work will actually show up. Added in `tiered-memory-storage`.

---

## Where this goes

The pattern in the gap list is consistent: the **architecture** is sound and the
**seams** are in the right places — `PersistentStorage`, `tenant.Service`,
`provider.Index`, the four-layer test — but several of the loudest claims are
still descriptions of intent. The next four milestones each convert one claim into
running code with a number attached:

| Post | Change | Converts |
| --- | --- | --- |
| 001 | `agent-integration-surface` | "usable by an agent" → HTTP API + MCP server + a server-side memory service |
| 002 | `tiered-memory-storage` | "RAM → disk → S3, tiered" → block-indexed SSTables, segment objects, a real tier policy, and $/GB-month |
| 003 | `zero-cost-recall-quality` | "free recall" → free recall that is actually good: BM25, recency, consolidation |
| 004 | `metered-paid-capabilities` | "metered cost" → a meter that reports what you actually spent |

If you only take one thing from this post: the recall really does cost $0.00, and
the `standard` profile really does under-report by 2,334×. Both are true, both are
measured, and a project that only told you the first one would not be worth
reading.
