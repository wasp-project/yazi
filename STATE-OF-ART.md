# State of the Art: Agent Memory Systems vs. Yazi

This document surveys popular agent-memory systems and positions **Yazi**
against them. The thesis is deliberate and narrow:

> Most memory systems compete on **capability** — richer recall, temporal
> reasoning, self-editing memory. Yazi competes on **cost**. Our final goal is a
> **cost-aware agent memory system**: one where every expensive operation
> (LLM extraction, embeddings, RAM-resident vector indexes, LLM-in-the-loop
> editing) is *optional, tiered, metered, and budgeted* on top of a cheap
> deterministic core that always works at ~zero marginal token cost.

This is not a claim that Yazi is more capable than the systems below today — it
is not. It is a claim that **cost is an unserved axis**, and that at agent scale
(many agents × high message volume × long horizons) cost is what actually
breaks.

---

## 1. The Cost Model of Agent Memory

Before comparing systems, it helps to name where the money goes. An agent memory
system can incur cost at five distinct points:

| # | Cost point | What drives it | Who pays it |
| --- | --- | --- | --- |
| C1 | **Write / ingest** | LLM calls to extract, summarize, deduplicate, and reconcile facts from each message | $/1M tokens, per message |
| C2 | **Embedding** | Encoding text into vectors at write *and* query time | $/1M tokens or GPU time |
| C3 | **Storage** | Where vectors/graphs/records live; RAM-resident ANN indexes are the expensive case | $/GB-month, RAM vs disk vs object |
| C4 | **Retrieval → context** | Tokens injected into the prompt each turn; LLM calls made *during* retrieval | $/1M input tokens, per turn |
| C5 | **Inference reuse** | Re-encoding the same memory every call vs. reusing a cached KV representation | GPU time / time-to-first-token |
| C6 | **Operational** | Infra footprint: vector DB + graph DB + queue + coordinators, ops headcount | $/month + people |

The key observation: **C1, C2, and C4 are usually charged per message or per
turn.** They are not one-time. A system that calls an LLM to extract facts on
every message has a cost that scales linearly with traffic, forever. That is the
cost Yazi exists to avoid by default — and to *meter and bound* when you opt in.

---

## 2. The Systems

### 2.1 mem0 — universal memory layer
*Apache-2.0 · Python/TS · self-host or cloud · https://github.com/mem0ai/mem0*

A hybrid memory layer: a vector store for semantic memory (20+ backends — Qdrant,
pgvector, Milvus, Pinecone…) plus a graph store for entity/relationship memory
(Neo4j, Memgraph, Neptune…). Default self-host is FastAPI + Postgres/pgvector +
Neo4j.

- **Creation (C1):** LLM-based. Every `add` runs an extraction pass and a
  consolidation/update pass — LLM tokens **per ingest**.
- **Retrieval (C2/C4):** vector + BM25 + entity-link boosting; query must be
  embedded.
- **Cost profile:** dominant cost is LLM tokens on every write; plus embeddings
  and running Postgres (+ optionally Neo4j).

### 2.2 Zep / Graphiti — temporal knowledge graph
*Apache-2.0 · Python (Go/TS SDKs) · cloud (CE deprecated Apr 2025) · https://github.com/getzep/zep*

Zep's engine **Graphiti** builds a **bi-temporal knowledge graph**: facts carry
`valid_at` / `invalid_at` windows, and contradictions **invalidate** old facts
rather than deleting them (full lineage preserved). Backends: Neo4j, FalkorDB,
Kuzu.

- **Creation (C1):** LLM entity/edge extraction on each event, incremental and
  **async** — still LLM tokens per ingest, but off the hot path.
- **Retrieval (C2/C4):** "graph RAG" — semantic + BM25 + graph traversal +
  temporal filters; returns pre-formatted context. No LLM summarization call at
  read time (a deliberate latency/cost choice).
- **Cost profile:** LLM extraction per ingest + embeddings + the operational
  weight of running a graph database. Practical self-hosting is now the raw
  Graphiti library (Zep Community Edition is unmaintained).

### 2.3 Milvus — vector database (infrastructure)
*Apache-2.0 · Go/C++ · self-host or Zilliz Cloud · https://github.com/milvus-io/milvus*

Milvus is **not** an agent-memory abstraction; it is the **storage substrate**
others build on. A distributed vector DB scaling to billions of vectors, with
HNSW / IVF / DiskANN / GPU indexes and hybrid (dense+sparse+BM25) search.

- **No LLM in the loop (C1=0).** It stores and searches vectors; embeddings are
  produced upstream.
- **Cost profile:** C3 and C6 dominate. HNSW/IVF are **RAM-hungry**; DiskANN
  trades RAM for disk/latency. Distributed mode is **heavy**: etcd + object
  storage/MinIO + Pulsar/Kafka + query/data/index nodes, typically on
  Kubernetes. (Milvus Lite/Standalone are far lighter.)
- **Relevance to Yazi:** Milvus is the kind of component Yazi wants to make
  *optional* — you should not need a billion-vector cluster to remember a user's
  preferences.

### 2.4 Letta (formerly MemGPT) — self-editing memory
*Apache-2.0 · Python · self-host or cloud · https://github.com/letta-ai/letta*

Letta gives agents an OS-inspired memory hierarchy they **edit themselves**:
- **Core memory** — small always-in-context blocks (RAM-like), edited via tools
  `core_memory_append` / `core_memory_replace`.
- **Recall memory** — searchable conversation history.
- **Archival memory** — long-term store, embedded and **vector-searched**
  (pgvector).

- **Creation/editing (C1/C4):** **LLM is required in the loop.** Every
  remember/promote/evict is an LLM **tool call**, and large in-context core
  blocks inflate prompt tokens every turn.
- **Cost profile:** token-dominated — this is the most "LLM-hungry" model here,
  by design (the agent reasons about its own memory). Infra is light (a server +
  Postgres).

### 2.5 MemOS — memory operating system
*Apache-2.0 · Python · arXiv 2507.03724 · ~9.6k★ · https://github.com/MemTensor/MemOS*

MemOS treats memory as a schedulable OS resource. Its **MemCube** wraps content +
metadata and moves between three substrates:
- **Plaintext** memory (graph-structured facts; Neo4j + Qdrant),
- **Activation** memory (**KV-cache** representations reused at inference),
- **Parameter** memory (knowledge in weights/adapters).

- **The cost-relevant idea (C5):** "internalize" frequently used plaintext into
  **activation (KV-cache) memory** so it isn't re-encoded each call. MemOS
  reports (on LoCoMo, vendor numbers) ~**61% token-overhead reduction** and up to
  ~**91% time-to-first-token reduction** vs. naive memory.
- **Cost profile:** still pays C1 (extraction) + C2 (embeddings) + C6 (graph +
  vector infra), but is the clearest precedent that **memory systems can and
  should optimize cost** — specifically inference cost via KV-cache reuse.

### 2.6 mem9 — TiDB-backed memory for coding agents
*Apache-2.0 · Go server + TS plugins · self-host or cloud · https://github.com/mem9-ai/mem9*

From the **PingCAP/TiDB** team; tagline "**Unlimited memory for OpenClaw**." A Go
**mnemo-server** (REST) with stateless TS plugins for OpenClaw / Claude Code /
OpenCode. Backend is **TiDB Cloud** (distributed MySQL with native `VECTOR` +
full-text).

- **Creation (C1):** "smart ingest" uses **LLM extraction** + dedup/taxonomy;
  a raw mode exists.
- **Retrieval (C2/C4):** hybrid vector + full-text, ranked by relevance.
- **Cost profile:** LLM tokens on smart ingest + embeddings + TiDB
  storage/compute. **This is Yazi's closest neighbor** — same Go client/server
  shape, same OpenClaw target — but it mandates a distributed SQL+vector backend
  (TiDB) and LLM-based ingest as the primary path.

---

## 3. Comparison

### 3.1 Architecture & approach

| System | Category | Primary store | Memory creation | Retrieval | LLM in loop? |
| --- | --- | --- | --- | --- | --- |
| **mem0** | memory layer | vector + graph | LLM extract per add | vector + BM25 + entity | **Yes** (write) |
| **Zep/Graphiti** | temporal KG | graph DB | LLM extract per event (async) | graph RAG + temporal | **Yes** (write) |
| **Milvus** | vector DB (infra) | vector index | — (none) | ANN / hybrid | No |
| **Letta** | stateful agent | pgvector + Postgres | LLM self-edit via tools | core in-context + vector | **Yes** (heavy) |
| **MemOS** | memory OS | graph + vector + KV-cache | LLM extract + tiering | vector + graph; KV reuse | **Yes** (write) |
| **mem9** | memory server | TiDB (vector+FTS) | LLM "smart ingest" | vector + full-text | **Yes** (write) |
| **Yazi (today)** | memory engine | KV cache → LSM → S3 | **client supplies structured JSON** | **deterministic key + secondary index** | **No** |

### 3.2 Cost drivers (the point of this doc)

| System | C1 write tokens | C2 embeddings | C3 storage | C4 retrieval tokens | C5 KV-cache reuse | C6 infra footprint |
| --- | --- | --- | --- | --- | --- | --- |
| mem0 | per add | yes | vector+graph | query embed | no | medium (PG+Neo4j) |
| Zep | per event | yes | graph DB | query embed | no | medium-heavy (graph DB) |
| Milvus | none | upstream | **RAM-heavy** | none | no | **heavy** (cluster) |
| Letta | **per edit** | yes (archival) | Postgres | **large core blocks** | no | light |
| MemOS | per add | yes | graph+vector | reduced | **yes** | medium-heavy |
| mem9 | per smart-ingest | yes | TiDB | query embed | no | medium (TiDB) |
| **Yazi (today)** | **~0** | **none** | **cheap (RAM→disk→S3)** | **~0 (no query-side LLM/embed)** | roadmap | **light (single Go binary)** |

> Reading the table: every capability-first system has at least one cost that is
> **charged per message or per turn** (bold cells). Yazi's deterministic core has
> none — its marginal cost per write and per read is dominated by bytes moved,
> not tokens spent. The trade-off is equally real: Yazi's core does **not** do
> semantic recall today.

---

## 4. Where Yazi Stands Today (Honest Assessment)

**What Yazi is now:** a lightweight Go KV/storage engine with a structured memory
model (basic / advanced / policy), secondary indexes (kind, scope, subject, tag,
template, role, domain, scenario), tiered persistence (in-memory cache → LSM with
WAL/SSTable → S3 snapshot), per-tenant key namespacing, and a layered
architecture (UI → tenant → storage → cloud). See
[ARCHITECTURE.md](./ARCHITECTURE.md).

**What that buys (the cost story):**
- **C1 ≈ 0:** Yazi does not require an LLM to store memory. The agent (which is
  already running an LLM) can hand Yazi already-structured records; Yazi never
  imposes a second extraction pass.
- **C2 = 0 today:** retrieval is by key and secondary index — exact, ordered,
  filterable — with **no embedding** at write or query time.
- **C3 cheap & tiered:** hot data in RAM, warm data in LSM on local disk, cold
  snapshots in S3. No mandatory RAM-resident vector index.
- **C4 ≈ 0 retrieval overhead:** a `list`/`get` returns records deterministically;
  no LLM call happens *inside* retrieval.
- **C6 light:** a single static binary; no graph DB, no vector cluster, no queue,
  no coordinator. Runs on a laptop or on S3-backed cloud with the same image.

**What Yazi is *not* (yet):** it has no semantic/vector search, no LLM-based fact
extraction or consolidation, no temporal knowledge graph, no self-editing agent
memory, and no KV-cache reuse. Against mem0/Zep/MemOS/Letta on *capability*, Yazi
is a substrate, not a peer.

That gap is the opportunity. The systems above prove the capabilities; **none of
them treats cost as the primary design constraint.** That is the lane.

---

## 5. The Goal: A Cost-Aware Agent Memory System

"Cost-aware" is a concrete design stance, not a slogan. It means four
commitments:

### 5.1 Cheap by default, expensive by choice
The deterministic KV/index core is always available at ~zero token cost. Semantic
search, LLM extraction, and graph/temporal reasoning are **opt-in layers**, never
mandatory. A user who only needs "remember this preference and give it back"
should never pay for an embedding model or a vector cluster.

### 5.2 Tiered everything (by access frequency, not just age)
Extend the existing RAM → LSM → S3 tiering into a **cost-tiered memory policy**:
- *Hot* memories stay in cache and are returned verbatim (cheapest).
- *Warm* memories live in LSM; only *promoted* (distilled "advanced") memories
  are embedded — so embedding spend tracks value, not volume.
- *Cold* memories are snapshotted to S3 and rehydrated on demand.
- Borrowing MemOS's insight (C5): frequently injected memories become candidates
  for **KV-cache reuse** to cut time-to-first-token.

### 5.3 Metered and budgeted (cost as a first-class signal)
Make cost *observable and enforceable*, which no system here does natively:
- Account token / embedding / storage cost **per tenant and per memory class**
  (the tenant layer already namespaces; cost accounting rides on the same seam).
- Support **retrieval budgets**: "answer within N context tokens / $X" — return
  fewer, cheaper, higher-value memories under a cap instead of dumping the
  top-k. Cost-bounded retrieval is the read-side analogue of rate limiting.

### 5.4 No mandatory heavy dependencies
Stay a single Go binary. Where a vector index is genuinely needed, prefer an
**embeddable / pluggable** index (or an S3/disk-resident ANN) over standing up a
separate Milvus/TiDB cluster — so the *floor* cost of running Yazi stays near
zero and scales up only when the workload demands it.

### 5.5 Positioning summary

| | Capability-first systems | Yazi (cost-aware) |
| --- | --- | --- |
| Default write path | LLM extraction (per message) | structured put (no LLM) |
| Default read path | embed + vector/graph search | key + index lookup |
| Semantic / temporal / self-edit | built-in, always-on | opt-in layer, metered |
| Storage | vector/graph DB (often RAM-heavy) | RAM → LSM → S3, tiered |
| Footprint | DB(s) + queue + coordinators | single binary |
| Cost visibility | none native | per-tenant / per-class metering, budgets |
| North star | richer memory | **lowest total cost of memory at agent scale** |

The bet: as agents move from demos to fleets, the question stops being "can it
remember?" and becomes "what does remembering *cost* per agent per day?" Yazi
aims to be the answer to the second question — and to make the first question's
expensive features available on demand, with the meter running where you can see
it.

---

## Sources & Caveats

System details verified from project repos, docs, and papers as of mid-2026.
Performance figures (e.g., MemOS LoCoMo token/latency reductions; mem9 user
counts) are **vendor-reported** and not independently benchmarked here. Backend
support lists and default models evolve quickly between releases.

- mem0 — https://github.com/mem0ai/mem0
- Zep / Graphiti — https://github.com/getzep/zep · https://github.com/getzep/graphiti
- Milvus — https://github.com/milvus-io/milvus · https://milvus.io/docs
- Letta (MemGPT) — https://github.com/letta-ai/letta · https://docs.letta.com
- MemOS — https://github.com/MemTensor/MemOS · https://arxiv.org/abs/2507.03724
- mem9 — https://github.com/mem9-ai/mem9 · https://www.pingcap.com/blog/how-we-built-mem9-agent-memory-product/
- Yazi — [ARCHITECTURE.md](./ARCHITECTURE.md) · [MEMORY.md](./MEMORY.md)
