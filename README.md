# yazi

[![GitHub release](https://img.shields.io/github/release/wasp-project/yazi.svg)](https://github.com/wasp-project/yazi/releases)
[![codecov](https://codecov.io/gh/wasp-project/yazi/branch/main/graph/badge.svg)](https://codecov.io/gh/wasp-project/yazi)
[![GoReport](https://goreportcard.com/badge/github.com/wasp-project/yazi)](https://goreportcard.com/badge/github.com/wasp-project/yazi)
[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

## Overview

yazi is a lightweight KV server with configurable protocol, persistence and storage engine layers. It now includes a memory-oriented data model that can be used as a local memory store for OpenClaw or similar LLM applications.

Current storage capabilities include:

- in-memory KV with local snapshot persistence
- single-node LSM with WAL, SSTable and simple compaction
- optional async replication with read/write quorum
- memory data model built on top of the existing KV/storage layer

## Project Layout

- `cmd/yazi`: start the server
- `cmd/cli`: local CLI client
- `pkg/storage`: storage abstraction, persistence policy and engines
- `pkg/storage/lsm`: WAL + SSTable based single-node engine
- `pkg/replication`: async replication + quorum wrapper
- `pkg/memory`: basic/advanced/policy memory model built on KV

## Quick Start

### Start the server

The server reads [default.yml](file:///Users/bytedance/mworks/goprojects/yazi/config/default.yml) by default:

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

Detailed usage is documented in [MEMORY.md](file:///Users/bytedance/mworks/goprojects/yazi/MEMORY.md).

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
