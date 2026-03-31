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

## Recommendations

- use `grpc` for memory testing
- use `engine: lsm` for durable local memory
- keep memory and future KVCache in separate logical prefixes
- if you plan to store large binary memory artifacts later, evolve the RPC layer from string payloads to bytes
