## Why

Yazi's memory engine works and is free at the margin, but almost nobody can
actually plug an agent into it. Three concrete blockers: (1) `config/**` is
gitignored and the config path is hardcoded to `./config/default.yml`, so a fresh
clone has **no config file and no way to point at one** — the very first command
in the README fails; (2) the memory model is built **client-side** in `cmd/cli`
(keys, indexes and all), and the server is a generic KV server that knows nothing
about memory, so any non-Go agent must shell out to `yazictl` — there is no HTTP
API and no MCP server; (3) every CLI error path is a `panic()` that prints a Go
stack trace. A cost advantage that cannot be adopted is not an advantage; this
change makes Yazi integrable before we make it cheaper.

## What Changes

- Move the `memory.Store` construction from `cmd/cli` to the **server**, so
  memory keys, secondary indexes, and the tenant seam are enforced server-side
  and every client speaks the same memory semantics. `yazictl` becomes a thin
  client over the same API. **BREAKING** for anyone depending on client-side key
  construction (no released client exists; the wire keys are unchanged).
- Add an **HTTP/JSON memory API** (`/v1/memory/...`) alongside the existing gRPC
  KV service: put/get/list/delete for basic/advanced/policy, plus recall with the
  usage/cost report. Any language can now use Yazi.
- Add an **MCP server** (`yazi mcp`) exposing `remember` / `recall` /
  `list_memories` / `forget` tools over stdio, so MCP-capable agents (Claude
  Code, OpenClaw, and others) can attach Yazi with a config entry and no glue
  code.
- Add a `--config` flag (plus `YAZI_CONFIG` env) and make the server start with
  built-in defaults when no config file exists. Track `config/*.example.yml` in
  git instead of ignoring the whole directory.
- Replace CLI `panic()` calls with printed errors and non-zero exit codes;
  unavailable providers (e.g. `--profile pro`) report a clear message instead of
  a stack trace.
- Ship release artifacts: a `make build` that works (`CGO_ENABLED=0`, current
  `LOCLABIN` typo fixed), a multi-stage Dockerfile, and a GitHub release workflow
  producing static binaries for linux/darwin × amd64/arm64.
- Remove the dead, duplicate memory model in `pkg/storage/memory_store.go`
  (`session:turn:` OpenClaw types) that shadows `pkg/memory`.

## Capabilities

### New Capabilities
- `server-side-memory-service`: the memory model, indexes and tenant namespacing
  live behind a server-owned service that all protocols share; clients no longer
  construct memory keys.
- `http-memory-api`: a versioned HTTP/JSON API for memory write, read, filtered
  list, delete and recall, including the per-request cost report.
- `mcp-memory-server`: an MCP stdio server exposing memory as agent tools.
- `runtime-configuration`: explicit config path selection, tracked example
  configs, and a working zero-config default.
- `cli-diagnostics`: predictable CLI failure behavior — human-readable errors and
  meaningful exit codes instead of panics.

### Modified Capabilities
<!-- No requirement changes to the existing memory-profiles /
     memory-provider-interface / cost-metering-and-budgets specs: profiles keep
     their current semantics, they simply become reachable over new transports. -->

## Impact

- **New code**: `pkg/memory/service` (transport-agnostic memory service),
  `pkg/server/http` (REST handlers), `pkg/mcp` (MCP stdio server),
  `cmd/yazi` flag parsing.
- **Modified code**: `pkg/server/server.go` (construct and wire the memory
  service; today it only holds `tenant.Service` as a placeholder),
  `cmd/cli/yazi_ctl.go` (thin client, no `newMemoryStore`, no panics),
  `pkg/config/config.go` (config path resolution + HTTP port), `Makefile`,
  `.gitignore`.
- **Removed code**: `pkg/storage/memory_store.go` and its test.
- **APIs**: new HTTP surface on a separate port (default `3457`); gRPC KV
  surface unchanged and still available.
- **Dependencies**: none new — MCP is JSON-RPC over stdio and HTTP is stdlib
  `net/http`, keeping the hermetic single-binary property.
- **Compatibility**: the on-disk key layout (`/_memory/...`, `tenant/<id>/...`)
  is unchanged, so existing data files keep working.
