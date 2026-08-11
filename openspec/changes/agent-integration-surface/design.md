## Context

Today the four-layer architecture (UI → tenant → storage → cloud) is real in the
code, but Layer 1 is thinner than it looks. `cmd/cli/yazi_ctl.go:620`
(`newMemoryStore`) builds a `memory.Store` over a `client.Client`, resolves the
tenant, and applies the key prefix — **client-side**. Only `Get/Set/Del/MGet/
MSet/Keys` cross the wire. `pkg/server/server.go:47` constructs a
`tenant.Service` and then never uses it, with a comment marking it as "the choke
point for a future server-side memory RPC". This change builds that RPC.

Consequences of the current placement, all verified against the running server:

- A recall does N round-trips: `Keys()`/index reads, then `MGet` of the matched
  ids, then client-side ranking. The server cannot optimize what it cannot see.
- Tenant isolation is enforced by a **cooperative client**. A raw `yazictl set`
  can write `tenant/acme/...` keys directly. This is acceptable for a
  single-user local store and unacceptable for the "company-wide shared memory"
  mode `config/cloud.yml` advertises.
- Non-Go agents have no path in. mem9 — the closest neighbor — ships TS plugins;
  we ship a Go binary you must `exec`.

Constraints that must survive this change: single static binary, no new module
dependencies, `lite` profile stays $0, on-disk key layout unchanged, and gRPC KV
remains available (the benchmark's `yazi_adapter.py` drives `yazictl`).

## Goals / Non-Goals

**Goals:**
- One transport-agnostic memory service, owned by the server, used by every
  protocol (gRPC, HTTP, MCP, CLI).
- An HTTP/JSON API that any language can call, returning the same
  `hits + usage + costUSD` shape the CLI prints today.
- An MCP stdio server so agent harnesses attach Yazi via configuration only.
- A server that starts correctly from a fresh `git clone` with no config file.
- Failure modes that read like a tool, not like a crash.

**Non-Goals:**
- Authentication / authorization. The HTTP API binds to loopback by default and
  carries no auth; multi-user hardening belongs with the external multi-tenant
  implementation that `tenant.Service` exists for.
- Changing memory schemas, filters, key layout, or profile semantics.
- Deprecating the gRPC KV surface.
- Streaming, pagination cursors, or bulk import APIs — deferred until a consumer
  needs them.

## Decisions

### Decision 1: A `pkg/memory/service` layer, not handlers-with-logic

Introduce `service.Memory` — a struct holding the tenant service and a factory
that produces a tenant-scoped `*memory.Store` over the server's `KVStore`. It
exposes one method per operation (`PutBasic`, `ListBasic`, `Recall`, …), each
taking an explicit `tenant string` argument resolved from the transport.

- **Why**: Three transports (HTTP, MCP, CLI-over-gRPC) must not each re-derive
  key prefixes or re-implement recall. Putting the logic in a service keeps the
  handlers as pure marshalling, and keeps `pkg/memory` itself transport-free so
  the `pkg/arch` layering test still passes.
- **Alternatives**: (a) Put the logic in the HTTP handlers and have MCP call
  HTTP over loopback — rejected: an in-process call should not need a socket.
  (b) Extend the gRPC service with memory RPCs and have HTTP proxy to it —
  rejected: same objection, plus it forces protobuf changes for every schema
  tweak.

### Decision 2: Server-side store construction over the *local* KVStore

The service wraps `storage.KVStore` directly (the same instance the gRPC handler
serves), not a loopback client. Tenant prefixing uses the existing `prefixKV`
decorator, unchanged.

- **Why**: Removes the N-round-trip recall and makes the tenant prefix
  authoritative rather than advisory. It also means `Recall` can later push
  ranking down next to the data (a prerequisite for the M1/M2 changes).
- **Trade-off**: The CLI must now round-trip to the server for operations it
  previously computed locally — but the count of round-trips drops from O(hits)
  to 1.

### Decision 3: HTTP/JSON as the integration surface, MCP as a thin adapter

`pkg/server/http` implements `/v1/memory/{basic,advanced,policy}` (POST/GET/
DELETE), `/v1/memory/{class}:list` (POST with a filter body), `/v1/memory:recall`
(POST), and `/healthz`. Tenant comes from the `X-Yazi-Tenant` header, falling
back to the configured tenant service. `pkg/mcp` implements JSON-RPC 2.0 over
stdio and maps four tools onto the same `service.Memory`.

- **Why HTTP first**: it is the lowest common denominator for SDK-less
  integration (curl, Python, TS) and is what every competitor exposes. MCP is
  then ~200 lines of JSON-RPC framing on top, with no new dependency.
- **Why stdlib `net/http` and hand-rolled JSON-RPC**: the same reasoning that
  produced the hand-rolled SigV4 signer in `pkg/storage/s3` — `go.mod` stays at
  six direct dependencies and the binary stays hermetic.
- **Alternatives**: gRPC-gateway (adds protoc plugins and a large dependency
  tree), or an off-the-shelf MCP SDK (adds a dependency and pins us to its
  release cadence).

### Decision 4: MCP tool surface is four verbs, deliberately small

`remember(content, kind?, scope?, subject?, tags?)`,
`recall(query, top_k?, max_context_tokens?, tag?)`,
`list_memories(filter)`, `forget(id)`.

- **Why**: Tool descriptions are prompt tokens paid on **every** agent turn. A
  12-tool surface would undercut the cost thesis at the integration layer. Four
  verbs cover the recall→act→remember loop that `LOCAL-DEPLOYMENT.md` describes.
- **Note**: `recall` returns hits **and** the `usage`/`costUSD` block, so a
  cost-aware agent can see what memory is charging it.

### Decision 5: Config resolution order, with a real zero-config default

`--config <path>` → `$YAZI_CONFIG` → `./config/default.yml` → built-in defaults.
A missing file at the *default* path is an INFO log and a normal start; a
missing file at an *explicitly requested* path is a fatal error. `.gitignore`
changes from `config/**` to ignoring only `config/default.yml`, and
`config/local.yml` / `config/cloud.yml` become tracked examples.

- **Why**: The current `config.Load` logs `Read config file error` and silently
  continues with defaults — which is right for the no-config case and wrong when
  the operator asked for a specific file. Distinguishing the two is the whole
  fix. Tracking the examples closes the "fresh clone has no config" hole that
  forced `QUICKSTART.md` to inline a heredoc.

### Decision 6: Errors — a `run() error` main, no panics

`cmd/cli` command bodies become functions returning `error`; cobra's
`RunE` + `SilenceUsage` prints `Error: <msg>` and exits 1. `provider.BuildPipeline`
already returns a descriptive "requires: …" error for unavailable providers; the
CLI must surface it as text rather than re-panicking (`yazi_ctl.go:489`).

- **Why**: Panics leak internal paths and make the CLI unusable in scripts,
  which is exactly how the benchmark adapter and the `remember`/`recall` shell
  helpers drive it.

## Risks / Trade-offs

- **[Behavior drift between CLI and server after the move]** → The existing
  `pkg/memory` tests already cover key layout and tenant isolation; add a
  service-level test asserting the HTTP path produces byte-identical keys to the
  pre-change client path (`TestEmptyPrefixPreservesLayout` is the reference).
- **[Unauthenticated HTTP port on a shared host]** → Bind `127.0.0.1` by default,
  require an explicit `http.address` to expose it, and document that cloud mode
  needs a fronting proxy until auth lands.
- **[MCP protocol churn]** → Keep the JSON-RPC framing in one file behind the
  four tool definitions; a protocol revision touches framing only.
- **[Removing `pkg/storage/memory_store.go` breaks an unknown consumer]** → It is
  unreferenced outside its own test (verified by grep); removal is safe and is
  called out in the proposal as a deliberate deletion.
- **[Two ports to operate]** → Accepted: gRPC KV and HTTP memory have different
  audiences. A single-port multiplexer is more magic than it is worth.

## Migration Plan

1. Land `service.Memory` and switch `cmd/cli` to it via gRPC — no user-visible
   change, all existing tests must pass unmodified.
2. Add the HTTP server behind a config block that defaults to enabled on
   loopback; add MCP as a separate `yazi mcp` subcommand (no port).
3. Fix config resolution and `.gitignore`; regenerate `QUICKSTART.md` step 3 to
   `cp config/local.yml config/default.yml`.
4. Delete the duplicate memory model.

Rollback: each step is independent; the HTTP/MCP surfaces can be disabled by
config without affecting the gRPC path.

## Open Questions

- Should `recall` over HTTP accept a `profile` override per request, or only
  read the server's configured profile? (Leaning: allow per-request override,
  since the CLI already has `--profile` and per-tenant profiles are a
  `memory-profiles` requirement.)
- Does the MCP server need to expose the advanced/policy memory classes as
  separate tools, or is `kind` on `remember` enough for now? (Leaning: enough —
  see Decision 4.)
