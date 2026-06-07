## Context

yazi is a Go KV server (`cmd/yazi`) with a CLI client (`cmd/cli`) and a memory data model in `pkg/memory` built on the `storage.KVStore` cache + `PersistentStorage` persistence path. The memory store already defines a narrow `KV` interface and a hierarchical `/_memory/...` key layout with secondary indexes. Persistence today is local-only: `storage.go` declares `StorageClassLocal` and `StorageClassS3`, but only the local `DiskWriter`/`DiskReader` (implementing `PersistentStorage`) exist. There is no notion of a tenant — every caller shares one global key space.

The goal is to grow this into a company-wide LLM-agent memory backend without rewriting the core. The existing engine is the reusable center; we add two thin seams around it: a tenant service above the memory store, and a cloud persistence backend below the storage engine. Constraints: keep the default (local, single-tenant) behavior byte-for-byte identical; do not change the memory data model or key layout in bypass mode; keep new boundaries as plain Go interfaces so an external multi-tenant project can plug in later.

## Goals / Non-Goals

**Goals:**
- Formalize a four-layer architecture (UI → Tenant Service → Storage Engines → Cloud Vendors) over existing code.
- Add a tenant-resolution interface with a default bypass implementation; namespace keys per tenant when present.
- Implement an AWS S3 `PersistentStorage` backend selectable via `storage: s3`.
- Run the same binary in local mode and cloud mode by configuration only.
- Preserve full backward compatibility for the existing local/single-tenant path.

**Non-Goals:**
- Building the real multi-tenant implementation (a future external project supplies it; we ship only the interface + bypass).
- Implementing AWS vector storage now (only leave a documented extension point).
- Authentication/authorization, billing, or tenant-onboarding flows.
- Changing the memory data model, RPC schema (still string KV), or the LSM/replication internals.
- Per-key remote object storage (S3 stores the encoded snapshot, matching the current local-disk persistence granularity).

## Decisions

### Decision 1: Tenant seam as a key-prefix resolver, not a storage wrapper
Introduce `pkg/tenant` with `type Service interface { Resolve(ctx) (TenantContext, error) }` and `TenantContext.KeyPrefix() string`. The memory layer asks the tenant service for a prefix and prepends it to every memory/index key. The default `BypassService` returns an empty prefix.
- **Why**: Keeps tenancy orthogonal to the storage engine. The engine never learns about tenants; isolation is purely a key-namespacing concern, which matches yazi's existing prefix-based design (`/_memory/...`, `/_meta`, `/_raft`). An external project implements only `Service` — no storage coupling, satisfying the "stable interface" requirement.
- **Alternatives considered**: (a) A per-tenant `KVStore` wrapper that filters keys — heavier, leaks storage types into the tenant interface, and complicates the shared cache/persistence path. (b) Separate physical stores per tenant — operationally expensive and breaks the single-snapshot persistence model. Prefixing is the lightest reuse of what exists.

### Decision 2: Pass tenant prefix into `memory.Store` rather than mutating global key constants
`memory.Store` gains a prefix (e.g. constructed via `NewStoreWithTenant(kv, prefix)` or a per-call prefix), applied where it currently builds `/_memory/...` keys. Empty prefix == today's behavior.
- **Why**: Localizes the change to key construction in `pkg/memory/store.go`; indexes automatically inherit the prefix, preserving filter/list correctness within a tenant.
- **Alternative**: Wrap the `KV` interface with a prefixing decorator. Rejected because index scans in `ListBasic` rely on key shapes the store itself constructs; doing it inside the store keeps reads and writes consistent.

### Decision 3: S3 backend implements the existing `PersistentStorage` interface
Add `pkg/storage/s3` with a type implementing `Write([]byte) (int, error)` / `Read([]byte) (int, error)`, modeled on `pkg/storage/local`. `Write` puts the encoded snapshot to `s3://<bucket>/<key>`; `Read` gets it. Wire it into the storage factory so `StorageClassS3` constructs it, used by the existing `Manager` persistence loop (append/scheduled policies unchanged).
- **Why**: Zero change to memory/storage logic — the `Manager` already calls `PersistentStorage.Write(cache.Encode())`. S3 becomes a drop-in sibling of local disk. Honors the declared-but-unimplemented `StorageClassS3`.
- **Alternatives considered**: A new bytes-oriented per-key object API. Rejected for now — it would change the persistence granularity and RPC payload model (out of scope), and the snapshot approach reuses the existing encode/decode path. Documented as the future direction (vector storage / per-key) via an extension point.
- **Dependency**: A self-contained S3 client built on the standard library (`net/http` + a small AWS SigV4 signer), NOT the AWS SDK for Go v2. Rationale discovered during implementation: `aws-sdk-go-v2` now requires Go ≥1.24 (this project is on 1.21.6) and pulls a large transitive tree, which would force a project-wide toolchain bump. The hand-rolled signer is ~100 lines, keeps `go.mod` minimal and the build hermetic, supports bucket/region/credentials and a custom endpoint (S3-compatible stores / `httptest`), and is testable fully offline. It can be swapped for the official SDK later behind the unchanged `PersistentStorage` interface.

### Decision 4: Configuration-driven mode selection
Extend `ServerConfig` with an `s3:` block (bucket, region, key/prefix, endpoint, credential source) read when `storage: s3`, and an optional `tenant:` block (defaults to bypass). The storage factory and server wiring choose implementations from config. Missing required S3 fields fail startup with a clear error (no silent fallback to local).
- **Why**: Satisfies the "same binary, two modes" requirement and the fail-fast S3 config requirement. Mirrors how `engine`, `persistent`, and `replication` are already configured.

### Decision 5: Wire the tenant seam at the request boundary
The server (gRPC/naive handlers) and CLI construct/resolve a `TenantContext` once per request and hand the derived prefix to the memory store. In bypass mode this is a no-op returning "".
- **Why**: Single choke point keeps every memory path tenant-aware without scattering logic; matches the existing `DataKeyPrefix` handling location in `pkg/server/grpc`.

## Risks / Trade-offs

- **Snapshot-granularity S3 writes are heavy under high write rates** → Reuse the existing `scheduled` persistence policy (batched flush) for S3; document that `append` policy is unsuitable for S3 at volume. Per-key remote storage is deferred, not precluded.
- **Tenant prefix added in only some call sites leaks data across tenants** → Centralize prefix application inside `memory.Store` (Decision 2) and resolve tenant once at the request boundary (Decision 5); add tests asserting cross-tenant isolation on read.
- **S3 outage or credential failure blocks startup/persistence** → Fail fast on startup with clear errors; for runtime write failures, surface the error through the existing `Manager` path (same as local-disk write errors today). Local mode remains fully offline-capable.
- **AWS SDK is a new dependency increasing binary size / build surface** → Isolate it in `pkg/storage/s3`; it is only compiled into the same binary but inactive unless `storage: s3`. Endpoint override allows testing against MinIO without real AWS.
- **Interface churn could break the future external tenant project** → Keep `tenant.Service` minimal and storage-agnostic (Decision 1); version it conservatively.

## Migration Plan

1. Add `pkg/tenant` (interface + `BypassService`) — no behavior change; default wiring uses bypass.
2. Thread tenant prefix through `memory.Store` and request handlers; verify bypass reproduces current keys via tests.
3. Add `pkg/storage/s3` + config block + factory wiring; keep `storage: local` default.
4. Roll out: existing deployments keep `storage: local` / no tenant block → unchanged. Cloud deployment sets `storage: s3` and configures a tenant service when the external project is ready.
5. **Rollback**: revert config to `storage: local` and bypass tenant; data already written to S3 stays as a snapshot, local mode resumes from local persistence. No schema migration is required because the memory data model is unchanged.

## Open Questions

- Should the per-tenant prefix be derived from the tenant ID directly or via a stable hash (to bound key length / avoid unsafe characters)? Default plan: sanitized `tenant/<id>/` prefix.
- For cloud mode, is one shared snapshot object acceptable across tenants (isolation via key prefix within the snapshot), or is a per-tenant S3 object/prefix desired? Default plan: single snapshot, intra-snapshot prefixing; revisit if tenant counts are large.
- Which credential source(s) to support first — IAM role / instance profile vs. static keys? Default plan: standard AWS default credential chain plus optional static config.
- Does the future vector-storage backend belong behind `PersistentStorage` or a new sibling interface? Leave as an extension point; decide when that work starts.
