## Why

yazi already ships a working memory data model (basic / advanced / cognitive-policy) layered on top of its KV and storage engines, but it has no formal architecture for growing into a company-wide memory backend for LLM agents. Two gaps block that: there is no tenant isolation seam (every caller shares one global key space), and the only durable backend is the local disk — `StorageClassS3` is declared but unimplemented, so the same binary cannot be deployed to the cloud for shared use. This change defines a clean four-layer architecture so the existing engine stays the reusable core while tenancy and cloud storage become pluggable.

## What Changes

- Define a four-layer architecture for the memory system, reusing the current code as the core:
  1. **User Interface** — existing CLI (`cmd/cli memory ...`) and gRPC/naive RPC client paths; no behavior change required.
  2. **Tenant Service** — a NEW pass-through (bypass) interface that all memory access flows through. Default implementation is a no-op that preserves today's behavior; a future external project supplies a real multi-tenant implementation by implementing this interface.
  3. **Storage Engines** — the existing `memory.Store` + `storage.KVStore` (mem / LSM, WAL+SSTable, replication). This is the reusable core; left functionally unchanged.
  4. **Cloud Vendors** — a NEW pluggable persistence backend behind the existing `storage.PersistentStorage` contract, starting with AWS S3 object storage and leaving room for AWS vector storage.
- Introduce a `TenantContext` / tenant-resolver seam that namespaces memory keys per tenant when a tenant is present, and is a transparent no-op when bypassed.
- Implement an S3-backed `PersistentStorage` so `storage: s3` becomes a real, selectable config value alongside `storage: local`.
- Support two deployment modes from one codebase: **local** (single-node, local disk/LSM, tenant bypass) and **cloud** (company-wide, S3 persistence, tenant service enabled).

## Capabilities

### New Capabilities
- `layered-memory-architecture`: Defines the four-layer structure, the contract/interface boundaries between layers, the per-tenant key-namespacing scheme, and the two supported deployment modes (local vs cloud).
- `tenant-service`: A tenant-resolution and isolation seam that every memory operation passes through, with a default bypass (no-op pass-through) implementation and a stable interface for an external multi-tenant implementation to plug into later.
- `cloud-storage-backend`: A pluggable cloud-vendor persistence backend implementing the existing `storage.PersistentStorage` contract, with an AWS S3 object-storage implementation selectable via `storage: s3`, plus an extension point for AWS vector storage.

### Modified Capabilities
<!-- No existing specs under openspec/specs/; all behavior is introduced as new capabilities. -->

## Impact

- **New code**: `pkg/tenant/` (tenant interface + bypass impl), `pkg/storage/s3/` (S3 `PersistentStorage`), wiring in `pkg/server` and `cmd/cli` to route memory access through the tenant seam.
- **Modified code**: `pkg/storage/storage.go` (storage-class factory for `s3`), `pkg/config/config.go` (S3 + tenant config blocks), `pkg/memory/store.go` (accept tenant-scoped key prefix), `pkg/server/grpc` wiring.
- **Config**: new `storage: s3` option with an `s3:` block (bucket/region/credentials/endpoint); new optional `tenant:` block (defaults to bypass).
- **Dependencies**: AWS SDK for Go (S3 client).
- **Compatibility**: Backward compatible — default config (local storage, tenant bypass) reproduces current behavior; the memory data model and key layout are unchanged.
