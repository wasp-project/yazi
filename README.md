# yazi

[![GitHub release](https://img.shields.io/github/release/wasp-project/yazi.svg)](https://github.com/wasp-project/yazi/releases)
[![codecov](https://codecov.io/gh/wasp-project/yazi/branch/main/graph/badge.svg)](https://codecov.io/gh/wasp-project/yazi)
[![GoReport](https://goreportcard.com/badge/github.com/wasp-project/yazi)](https://goreportcard.com/badge/github.com/wasp-project/yazi)
[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

## Overview

yazi is a lightweight KV server with configurable protocol, persistence and storage engine layers. It now includes a memory-oriented data model that can be used as a local memory store for OpenClaw or similar LLM applications.

## Goal

The current goal is to grow yazi into a **layered memory system for LLM agents**
that runs the same binary in two modes:

- **Locally** — single node, offline, durable on disk (LSM), single-tenant — as
  a personal/agent memory store.
- **On the cloud** — company-wide shared memory, persisted to AWS S3 (or an
  S3-compatible store), with per-tenant isolation.

It is built as four layers so the existing KV/storage engine stays the reusable
core: **User Interface → Tenant Service → Storage Engines → Cloud Vendors**.
The tenant layer is a no-op by default and an interface point for a future
multi-tenant project; the cloud layer is a pluggable persistence backend. See
**[ARCHITECTURE.md](./ARCHITECTURE.md)** for the full design and diagrams.

The long-term direction and core value is to be a **cost-aware agent memory
system**: keep a cheap deterministic core (key/index retrieval at ~zero token
cost, tiered RAM → disk → S3 storage) and make expensive features — semantic
search, LLM-based extraction, temporal reasoning — *opt-in, metered, and
budgeted* rather than mandatory. See **[STATE-OF-ART.md](./STATE-OF-ART.md)** for
how this compares to mem0, Zep, Milvus, Letta, MemOS, and mem9.

The storage capabilities below back this goal:

- in-memory KV with local snapshot persistence
- single-node LSM with WAL, SSTable and simple compaction
- optional async replication with read/write quorum
- memory data model built on top of the existing KV/storage layer
- AWS S3 (and S3-compatible) snapshot persistence for cloud deployment
- per-tenant memory namespacing (default: single-tenant bypass)

## Architecture

The memory system is organized as four layers. The storage engine is the
reusable core; tenancy and cloud storage are thin, pluggable seams around it.

| Layer | Responsibility | Code |
| --- | --- | --- |
| 1. User Interface | CLI and RPC entry points | `cmd/cli`, `pkg/server`, `pkg/client` |
| 2. Tenant Service | resolve tenant, namespace keys (bypass by default) | `pkg/tenant` |
| 3. Storage Engines | KV cache + memory model + LSM/replication (the core) | `pkg/storage`, `pkg/memory` |
| 4. Cloud Vendors | snapshot persistence backends (local disk, AWS S3) | `pkg/storage/local`, `pkg/storage/s3` |

Each layer depends only on the layer beneath it through a Go interface, so any
layer can be replaced without touching the others. An architecture test
(`pkg/arch`) enforces that the storage engine never imports the tenant or UI
layers.

## Project Layout

- `cmd/yazi`: start the server
- `cmd/cli`: local CLI client
- `pkg/tenant`: tenant-service layer (bypass + static; interface for external multi-tenant impl)
- `pkg/storage`: storage abstraction, persistence policy and engines
- `pkg/storage/lsm`: WAL + SSTable based single-node engine
- `pkg/storage/s3`: AWS S3 / S3-compatible snapshot persistence backend
- `pkg/replication`: async replication + quorum wrapper
- `pkg/memory`: basic/advanced/policy memory model built on KV
- `pkg/arch`: architecture-conformance tests (layer dependency direction)

## Deployment Modes

The same binary runs in two modes selected purely by configuration:

- **Local** (`config/local.yml`): single node, local-disk/LSM persistence,
  tenant bypass. Reproduces the original single-tenant behavior.
- **Cloud** (`config/cloud.yml`): AWS S3 snapshot persistence and an enabled
  tenant service, for company-wide shared use.

### Tenant namespacing

Memory access flows through the tenant-service layer. With no `tenant:` config
(or `--tenant` unset on the CLI) the bypass service applies no namespacing and
keys keep the original `/_memory/...` layout. With a tenant set, keys are
namespaced under `tenant/<id>/...` so tenants are isolated within one store:

```bash
go run ./cmd/cli --tenant acme   memory basic put --json '{"kind":"preference","scope":"user","content":"x"}'
go run ./cmd/cli --tenant globex memory basic list --filter '{"scope":"user"}'   # never sees acme's records
```

A separate project can supply a real multi-tenant implementation of the
`tenant.Service` interface without changing the memory or storage layers.

### Cloud storage (AWS S3)

Set `storage: s3` with an `s3:` block (bucket/region/key, optional `endpoint`
for S3-compatible stores; credentials via config or `AWS_ACCESS_KEY_ID` /
`AWS_SECRET_ACCESS_KEY`). The snapshot is written to the object via the existing
persistence path. Missing required S3 config fails startup instead of silently
falling back to local. A future AWS vector-storage backend can be added by
implementing the same `storage.PersistentStorage` contract — see MEMORY.md.

## Running

### Build

```bash
go build ./...

# Optional: produce binaries
go build -o bin/yazi   ./cmd/yazi    # server
go build -o bin/yazictl ./cmd/cli    # CLI
```

> On some macOS + Go toolchain combinations, building/testing binaries that pull
> the net resolver aborts with a `missing LC_UUID` linker error. Prefix commands
> with `CGO_ENABLED=0` if you hit it: `CGO_ENABLED=0 go build ./...`.

### Run locally (single-tenant, on-disk)

The server reads [`config/default.yml`](./config/default.yml) by default, so to
run a specific mode point that file at the example config you want:

```bash
cp config/local.yml config/default.yml   # local mode: storage=local, engine=lsm, tenant bypass
go run ./cmd/yazi
```

In another terminal, use the memory commands (no `--tenant` ⇒ bypass):

```bash
go run ./cmd/cli memory basic put --json '{"kind":"preference","scope":"user","subject":"ui","content":"prefers dark theme","tags":["theme","ui"]}'
go run ./cmd/cli memory basic list --filter '{"tag":"ui","limit":10}'
go run ./cmd/cli memory basic get <memory-id>
```

Per-tenant namespacing is opt-in via `--tenant`:

```bash
go run ./cmd/cli --tenant acme   memory basic put --json '{"kind":"preference","scope":"user","content":"x"}'
go run ./cmd/cli --tenant globex memory basic list --filter '{"scope":"user"}'   # never sees acme's records
```

### Run on the cloud (multi-tenant, S3-backed)

```bash
cp config/cloud.yml config/default.yml    # cloud mode: storage=s3, tenant service enabled
# edit config/default.yml: set s3.bucket/region (and s3.endpoint for MinIO)
export AWS_ACCESS_KEY_ID=...              # or set s3.accessKey / s3.secretKey in the config
export AWS_SECRET_ACCESS_KEY=...
go run ./cmd/yazi
```

The server fails fast at startup if required S3 config (bucket/region/key/
credentials) is missing — it does not silently fall back to local storage.

### Test

```bash
CGO_ENABLED=0 go test ./...
```

## Quick Start

### Start the server

The server reads [`config/default.yml`](./config/default.yml) by default:

```bash
go run ./cmd/yazi
```

With the current default configuration:

- protocol is `grpc`
- persistence target is local file storage
- persistence policy is `scheduled`
- memory engine is `mem`

This means writes are persisted to the local snapshot file managed by the existing storage layer.

### Basic KV test

In another terminal:

```bash
go run ./cmd/cli set hello world
go run ./cmd/cli get hello
go run ./cmd/cli keys
```

## Persistence Modes

### Local snapshot persistence

The current default config uses the original persistent layer:

```yaml
protocol: grpc
storage: local
engine: mem
persistent: scheduled
```

This path writes the whole in-memory dataset to local storage on schedule or on append, depending on `persistent`.

### LSM persistence

For memory-heavy workloads, LLM memory and future KVCache-style scenarios, LSM is recommended:

```yaml
protocol: grpc
storage: local
engine: lsm
lsm:
  dir: data/lsm
  memtableMaxEntries: 1024
  compactionMaxTables: 4
  walMaxSegmentEntries: 0
```

Then start the server:

```bash
go run ./cmd/yazi
```

Data is persisted under `lsm.dir` with WAL and SSTable files.

## OpenClaw Memory Store

yazi now provides a memory data model on top of the existing KV/storage layer. It does not bypass the storage engine. Instead:

- CLI talks to yazi through the existing client protocol
- server writes memory data into the selected KV/storage engine
- persistence depends on the configured engine/persistence mode

The memory model has three layers:

- basic memory: preference, history, decision, context
- advanced memory: structured experience and cognitive template
- policy memory: role/domain/scenario specific cognitive policy

Detailed usage is documented in [MEMORY.md](./MEMORY.md) and the design in [ARCHITECTURE.md](./ARCHITECTURE.md).

## Local Memory Test

### 1. Start yazi with persistence

Recommended for local verification:

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

Then run:

```bash
go run ./cmd/yazi
```

### 2. Write a basic memory

```bash
go run ./cmd/cli memory basic put --json '{"kind":"preference","scope":"user","subject":"ui","content":"prefers dark theme","tags":["theme","ui"]}'
```

The command returns a generated memory id.

### 3. Read and list

```bash
go run ./cmd/cli memory basic get <memory-id>
go run ./cmd/cli memory basic list --filter '{"tag":"ui","limit":10}'
```

### 4. Verify persistence

- stop the server
- restart `go run ./cmd/yazi`
- run the same `get` or `list` command again

If the memory is still available, the selected storage engine has persisted it correctly.

## Testing

Run relevant tests:

```bash
go test ./pkg/storage/lsm
go test ./pkg/replication
go test ./pkg/memory
go test ./cmd/cli
```

## Current Notes

- `grpc` is the recommended protocol for local memory testing
- memory commands currently operate through the existing CLI client path
- memory payloads are JSON documents stored inside the selected KV engine
- if you want OpenClaw to call yazi by RPC instead of CLI, the next step is to expose a dedicated memory RPC service
