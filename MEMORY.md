# MEMORY

## Overview

The memory module turns yazi into a structured memory store on top of the existing KV and storage engine layers.

It is designed for scenarios such as:

- OpenClaw local memory store
- user preference persistence
- dialogue history persistence
- decision and context tracking
- structured experience and cognitive policy storage

The current implementation reuses the existing yazi storage path instead of introducing a separate storage backend.

## Storage Model

Memory data is stored as JSON in the selected KV engine.

Logical categories:

- basic memory
- advanced memory
- cognitive policy

Key prefixes:

- `/_memory/basic/{id}`
- `/_memory/advanced/{id}`
- `/_memory/policy/{id}`

Index prefixes:

- `/_memory/index/basic/...`
- `/_memory/index/advanced/...`
- `/_memory/index/policy/...`

This means:

- when yazi uses local snapshot persistence, memory is persisted there
- when yazi uses the LSM engine, memory is stored in WAL + SSTable files
- when replication is enabled, memory follows the same replication path

## Memory Layers

### Basic Memory

Used for raw memory records:

- preference
- history
- decision
- context

Schema:

```json
{
  "id": "optional",
  "kind": "preference",
  "scope": "user",
  "subject": "ui",
  "content": "prefers dark theme",
  "tags": ["theme", "ui"]
}
```

Supported filters:

- `kind`
- `scope`
- `subject`
- `tag`
- `limit`

### Advanced Memory

Used for structured experience distilled from raw memory.

Schema:

```json
{
  "id": "optional",
  "title": "planning-pattern",
  "summary": "prefers step-by-step planning",
  "template": "plan",
  "evidence": ["basic-memory-id-1", "basic-memory-id-2"],
  "tags": ["workflow", "reasoning"]
}
```

Supported filters:

- `template`
- `tag`
- `limit`

### Cognitive Policy

Used for role/domain/scenario specific templates.

Schema:

```json
{
  "id": "optional",
  "role": "doctor",
  "domain": "medical",
  "scenario": "triage",
  "policy": "clarify symptoms before advice",
  "tags": ["safety"]
}
```

Supported filters:

- `role`
- `domain`
- `scenario`
- `tag`
- `limit`

## Local Testing

### Step 1: choose a persistent backend

Recommended config for local testing:

```yaml
protocol: grpc
storage: local
engine: lsm
port: 3456
lsm:
  dir: data/lsm
  memtableMaxEntries: 8
  compactionMaxTables: 4
  walMaxSegmentEntries: 0
```

Then start yazi:

```bash
go run ./cmd/yazi
```

### Step 2: write basic memory

```bash
go run ./cmd/cli memory basic put --json '{"kind":"preference","scope":"user","subject":"assistant","content":"answer briefly first","tags":["style","preference"]}'
```

### Step 3: query memory

```bash
go run ./cmd/cli memory basic list --filter '{"scope":"user","limit":10}'
go run ./cmd/cli memory basic list --filter '{"tag":"preference","limit":10}'
go run ./cmd/cli memory basic get <id>
```

### Step 4: write advanced memory

```bash
go run ./cmd/cli memory advanced put --json '{"title":"response-style","summary":"user prefers concise-first responses","template":"style-template","evidence":["<basic-id>"],"tags":["style"]}'
```

### Step 5: write policy memory

```bash
go run ./cmd/cli memory policy put --json '{"role":"coder","domain":"software","scenario":"bugfix","policy":"verify with tests before final answer","tags":["engineering"]}'
```

### Step 6: verify persistence

- stop yazi
- restart yazi
- run `get` or `list` again

If the records remain, persistence is working through the configured storage engine.

## OpenClaw Integration Notes

At the moment the easiest integration path is CLI-based:

- OpenClaw calls `yazictl memory ...`
- yazi CLI uses the existing client path
- server writes to the selected storage engine

This is suitable when:

- OpenClaw already shells out to local tools
- local functional validation is the primary goal
- simple JSON I/O is acceptable

If OpenClaw later needs lower latency or richer APIs, the next step is:

- expose a dedicated memory RPC service
- keep the same memory data model and storage layout
- let OpenClaw switch from CLI calls to RPC without changing the persistence layer

## Layered Architecture

The memory system is organized into four layers. The storage engine is the
reusable core; the tenant service and cloud backends are thin seams around it.

1. **User Interface** — `cmd/cli` (CLI) and `pkg/server` / `pkg/client` (RPC).
2. **Tenant Service** — `pkg/tenant`. Resolves a request into a `TenantContext`
   and derives a key prefix. Default is a no-op bypass.
3. **Storage Engines** — `pkg/storage` (KV cache, LSM, replication) and
   `pkg/memory` (the memory data model). This is the core and is unchanged by
   the tenant/cloud layers.
4. **Cloud Vendors** — `pkg/storage/local` (disk) and `pkg/storage/s3` (AWS S3 /
   S3-compatible), both implementing `storage.PersistentStorage`.

Dependencies point downward only. `pkg/arch` contains a test that fails if the
storage engine ever imports the tenant or UI layers.

## Tenant Service (Layer 2)

Every memory operation passes through `tenant.Service`:

```go
type Service interface {
    Resolve(req Request) (TenantContext, error)
}
```

- `BypassService` (default) resolves to an empty key prefix — keys stay under
  `/_memory/...`, identical to single-tenant behavior.
- `StaticService` pins to one tenant id and namespaces keys under
  `tenant/<id>/_memory/...`.
- An external multi-tenant project implements `Service` (e.g. resolving the
  tenant from an RPC header) without touching the memory or storage layers.

The memory store applies the prefix transparently via `NewStoreWithTenant`:

```go
store := memory.NewStoreWithTenant(kv, ctx.KeyPrefix()) // "" prefix == NewStore
```

CLI usage (the `--tenant` flag is persistent across memory subcommands):

```bash
go run ./cmd/cli --tenant acme memory basic put --json '{"kind":"preference","scope":"user","content":"x"}'
go run ./cmd/cli --tenant acme memory basic list --filter '{"scope":"user"}'
# omitting --tenant => bypass => original un-prefixed keys
```

## Cloud Storage: AWS S3 (Layer 4)

`pkg/storage/s3` implements `storage.PersistentStorage` and writes the encoded
snapshot to an S3 object. It uses a small built-in AWS SigV4 signer over the
standard library (no AWS SDK dependency), so the build stays hermetic. A custom
`endpoint` targets S3-compatible stores (MinIO, etc.).

Config block (read only when `storage: s3`):

```yaml
storage: s3
s3:
  bucket: my-yazi-memory
  region: us-east-1
  key: yazi.data
  # endpoint: http://127.0.0.1:9000   # for MinIO / S3-compatible
  # accessKey / secretKey, or AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY
```

Missing required S3 config (bucket/region/key/credentials) fails startup with a
clear error — it does **not** silently fall back to local storage.

### Manual MinIO verification

```bash
# 1. run MinIO and create the bucket
docker run -p 9000:9000 -e MINIO_ROOT_USER=minio -e MINIO_ROOT_PASSWORD=minio12345 minio/minio server /data
# 2. point config/cloud.yml at it (endpoint: http://127.0.0.1:9000, matching creds)
# 3. start yazi, write a memory, wait for the scheduled flush, restart, list again
```

The automated equivalent is `pkg/storage/s3` `TestWriteReadRoundTrip`, which
signs real SigV4 requests against an in-process S3-compatible endpoint.

## Deployment Modes

The same binary runs locally or in the cloud, selected by config:

- **Local** (`config/local.yml`): `storage: local`, `engine: lsm`, no tenant
  block (bypass). Offline, single-tenant.
- **Cloud** (`config/cloud.yml`): `storage: s3`, tenant service enabled.
  Company-wide shared memory.

> Note: `config/**` is gitignored (including `default.yml`), so `local.yml` and
> `cloud.yml` are local working examples; the canonical content lives here.

## Extension Point: Vector Storage

A future AWS vector-storage backend (or any other cloud vendor) plugs in by
implementing the same `storage.PersistentStorage` contract used by the local and
S3 backends:

```go
type PersistentStorage interface {
    Write(data []byte) (int, error)
    Read(data []byte) (int, error)
}
```

Add the implementation under `pkg/storage/<vendor>`, give it a `StorageClass`
value, and wire it into the storage factory in `pkg/server`. The memory data
model and the tenant layer require no changes. If a vector backend needs
per-record (rather than snapshot) semantics, define a sibling interface beside
`PersistentStorage` and select it the same way — the upper layers stay intact.

## Recommendations

- use `grpc` for memory testing
- use `engine: lsm` for durable local memory
- keep memory and future KVCache in separate logical prefixes
- if you plan to store large binary memory artifacts later, evolve the RPC layer from string payloads to bytes
