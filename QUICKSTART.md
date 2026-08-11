# Yazi Quick Start — Minimal Usable Agent Memory (MVP)

This is the **shortest path from `git clone` to an agent that remembers things at
$0 token cost.** Every command below was run end-to-end against this repo; the
outputs shown are real.

**What the MVP is:** one Go binary (`yazi`) + one CLI (`yazictl`) giving you a
durable, structured, deterministic memory store for an LLM agent — `remember`
and `recall` with **zero LLM calls and zero embedding calls** on the default
profile, persisted to local disk (LSM/WAL/SSTable), with a metered cost report
on every recall.

**What the MVP is not (yet):** no HTTP API (CLI/gRPC only), no persistent vector
index, no automatic RAM→disk→S3 tiering. See [Known limits](#8-known-limits).

---

## 1. Prerequisites

- **Go ≥ 1.21** — that is all. No database, no vector store, no API key, no Docker.

> macOS + some Go toolchains abort with a `missing LC_UUID` linker error. Prefix
> every `go` command with `CGO_ENABLED=0` if you hit it. This guide already does.

---

## 2. Build (30 seconds)

```bash
git clone https://github.com/wasp-project/yazi.git
cd yazi
CGO_ENABLED=0 go build -o bin/yazi    ./cmd/yazi   # server
CGO_ENABLED=0 go build -o bin/yazictl ./cmd/cli    # client
export PATH="$PWD/bin:$PATH"
```

---

## 3. Configure

> **Important:** `config/**` is in `.gitignore`, so a fresh clone has **no
> config file**. The server reads `./config/default.yml` relative to its working
> directory. Create it:

```bash
mkdir -p config
cat > config/default.yml <<'YAML'
protocol: grpc
storage: local
engine: lsm          # durable: WAL + SSTable on disk
persistent: scheduled
capacity: 1024
loglevel: info
port: 3456

lsm:
  dir: data/lsm
  memtableMaxEntries: 8
  compactionMaxTables: 4
  walMaxSegmentEntries: 0

# No `tenant:` block => single-tenant (bypass, no key namespacing).
# No `memory:` block => `lite` profile: deterministic, zero-token recall.
YAML
```

---

## 4. Run the server

```bash
yazi        # gRPC on :3456, data under ./data/lsm
```

Leave it running. Everything below goes in a second terminal (with `bin/` on
`PATH`).

---

## 5. Remember and recall

### Write a memory (no LLM involved)

The agent hands Yazi an already-structured record. Yazi never runs a second
extraction pass over it — that is the whole cost thesis.

```bash
yazictl memory basic put --json '{
  "kind":"preference","scope":"user","subject":"ui",
  "content":"prefers dark theme","tags":["theme","ui"]}'
# -> 20260809024939.650454000-6b595114   (the memory id)

yazictl memory basic put --json '{
  "kind":"decision","scope":"project","subject":"db",
  "content":"use postgres for the billing service","tags":["arch","db"]}'
```

`kind` is one of `preference | history | decision | context`.

### Deterministic lookup (exact, indexed, $0)

```bash
yazictl memory basic get   <memory-id>
yazictl memory basic list  --filter '{"tag":"ui","limit":10}'
yazictl memory basic list  --filter '{"kind":"decision","scope":"project"}'
```

Filters (`kind`, `scope`, `subject`, `tag`, `limit`) are served from secondary
indexes, not a full scan.

### Recall by query, with the meter running

```bash
yazictl memory recall --query "which database do we use" --top-k 3
```

```json
{
  "costUSD": 0,
  "hits": [
    { "id": "20260809024939.670434000-37fc8e12",
      "text": "use postgres for the billing service",
      "score": 0.1666 }
  ],
  "profile": "lite",
  "usage": { "llmInputTokens": 0, "llmOutputTokens": 0,
             "embedTokens": 0, "contextTokens": 9, "latencyMs": 0 }
}
```

**`costUSD: 0` is the point.** `contextTokens` is what you will pay downstream
when you paste those hits into a prompt — cap it with
`--max-context-tokens 500`.

### Two other memory layers

```bash
# distilled experience
yazictl memory advanced put --json '{"title":"planning-pattern","summary":"prefers step-by-step planning","template":"plan","tags":["workflow"]}'
# role/domain/scenario policy
yazictl memory policy   put --json '{"role":"doctor","domain":"medical","scenario":"triage","policy":"clarify symptoms before advice"}'
```

Both support `get | list | del` with their own filters (`template`/`tag`,
`role`/`domain`/`scenario`/`tag`).

---

## 6. Verify durability

```bash
# stop the server (Ctrl-C), then:
ls data/lsm            # sstable-1.dat  sstable-2.dat  wal-3.log
yazi &                 # restart
yazictl memory basic list --filter '{"tag":"ui","limit":10}'   # still there
```

---

## 7. Wire it into an agent (the actual MVP deliverable)

Two shell functions are enough for any agent harness that can run commands
(OpenClaw, Claude Code, a cron script). Drop these in the agent's environment:

```bash
remember() {  # remember "<content>" "<tag>"
  yazictl memory basic put --json "$(printf '{"kind":"context","scope":"user","content":%s,"tags":[%s]}' \
    "$(printf '%s' "$1" | python3 -c 'import json,sys;print(json.dumps(sys.stdin.read()))')" \
    "\"${2:-general}\"")"
}

recall() {    # recall "<query>"  -> newline-separated memory lines for the prompt
  yazictl memory recall --query "$1" --top-k 5 --max-context-tokens 500 \
    | python3 -c 'import json,sys;[print("-",h["text"]) for h in json.load(sys.stdin)["hits"]]'
}
```

The agent loop becomes: `recall` → inject the lines into the system prompt →
act → `remember` the outcome. Recall costs nothing but the tokens of the lines
you chose to inject.

For the full local stack (OpenClaw + a local Gemma via Ollama), see
[LOCAL-DEPLOYMENT.md](./LOCAL-DEPLOYMENT.md).

---

## 8. Optional: pay for more recall quality

Profiles trade cost for quality. `lite` is the default and the only one that is
free.

| Profile | Composition | Cost per recall | Status |
| --- | --- | --- | --- |
| `lite` | keyword + secondary index over the durable store | **$0** | production path |
| `standard` | local hashing embedder + in-memory cosine index | embed tokens only | works; **re-embeds the whole store on every call** |
| `pro` | + LLM extraction / LLM rerank | LLM tokens | **not implemented** — fails fast |
| `custom` | bind each stage yourself | you decide | — |

```bash
yazictl memory recall --query "which database do we use" --profile standard --top-k 3
#  "costUSD": 1.2e-07,  "usage": { "embedTokens": 6, ... }
```

`--profile pro` intentionally refuses to run (the LLM providers are interface
stubs, not fake implementations). Today it surfaces as a Go panic rather than a
clean error message.

---

## 9. Optional: multi-tenant and cloud (S3)

### Per-tenant isolation

```bash
yazictl --tenant acme   memory basic put  --json '{"kind":"preference","scope":"user","content":"acme secret"}'
yazictl --tenant acme   memory basic list --filter '{"limit":10}'   # sees it
yazictl --tenant globex memory basic list --filter '{"limit":10}'   # [] — isolated
```

Keys become `tenant/acme/_memory/...`. Without `--tenant`, nothing is namespaced.

### S3 / S3-compatible snapshot persistence

```yaml
# config/default.yml
protocol: grpc
storage: s3
persistent: scheduled
scheduledPeriod: 10
port: 3456
s3:
  bucket: my-yazi-memory
  region: us-east-1
  key: yazi.data
  # endpoint: http://127.0.0.1:9000   # MinIO / S3-compatible
tenant:
  mode: static
  tenant: acme
```

```bash
export AWS_ACCESS_KEY_ID=... AWS_SECRET_ACCESS_KEY=...
yazi
```

Startup **fails fast** if bucket/region/key/credentials are missing — it never
silently falls back to local disk.

> Note: `engine: lsm` and `storage: s3` are currently **mutually exclusive**.
> Selecting the LSM engine bypasses the snapshot-persistence path entirely, so
> "local LSM with an S3 cold tier" is not yet possible. S3 mode stores the
> **whole dataset as one snapshot object**, and the working set lives in RAM.

---

## 10. Measure the cost claim

```bash
cd benchmark
python3 run.py --adapters all --plot     # stdlib only, no deps
```

Compares Yazi against cost *estimators* for mem0 / Zep / Letta / MemOS / mem9 on
accuracy, latency, recurring system cost, and context tokens. `sim-*` rows are
documented estimates (flagged `*`); `yazi` and `mock` are real measurements. The
`yazi` adapter needs a running server and `yazictl` on `PATH`, otherwise it is
reported as `SKIPPED` rather than faked.

---

## 11. Known limits

Read these before building on the MVP:

1. **No config file in a fresh clone** — `config/**` is gitignored and the path
   `./config/default.yml` is hardcoded (no `--config` flag). Step 3 is mandatory.
2. **No HTTP/REST API.** Memory keys and indexes are built **client-side** in
   `cmd/cli`; the server is a generic KV server. Non-Go agents must shell out to
   `yazictl`.
3. **No real storage tiering.** The `mem` engine holds everything in RAM and
   snapshots the whole dataset; the `lsm` engine loads every SSTable fully into
   RAM on startup. Both are durable, neither is memory-bounded.
4. **`standard` recall is O(corpus) per query** — the vector index is ephemeral
   and reseeded from the store on every call.
5. **CLI errors panic** with a Go stack trace instead of a message and exit code.
6. Tenant keys contain a cosmetic double slash (`tenant/acme//_memory/...`).

---

## 12. Where to go next

| Doc | What it covers |
| --- | --- |
| [README.md](./README.md) | project overview, deployment modes |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | the four layers, key namespacing, data flow |
| [MEMORY.md](./MEMORY.md) | memory schemas and filters in full |
| [COMPOSITION.md](./COMPOSITION.md) | the provider pipeline and profiles |
| [STATE-OF-ART.md](./STATE-OF-ART.md) | cost model and comparison vs. mem0/Zep/Letta/MemOS/mem9 |
| [LOCAL-DEPLOYMENT.md](./LOCAL-DEPLOYMENT.md) | OpenClaw + local Gemma end-to-end |
| [benchmark/README.md](./benchmark/README.md) | the cost benchmark methodology |
