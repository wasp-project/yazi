## 1. Tenant Service Layer (bypass seam)

- [x] 1.1 Create `pkg/tenant/tenant.go` defining `TenantContext` (carries tenant id) with a `KeyPrefix() string` method, and a `Service` interface with `Resolve(...) (TenantContext, error)`
- [x] 1.2 Implement `BypassService` in `pkg/tenant/bypass.go` returning a `TenantContext` with empty id and empty key prefix
- [x] 1.3 Implement deterministic, sanitized prefix derivation (e.g. `tenant/<sanitized-id>/`) for non-empty tenant ids, with a unit test asserting the same id yields the same prefix
- [x] 1.4 Add unit tests for the tenant package: bypass returns empty prefix; named tenant returns stable, unique prefix

## 2. Tenant-Scoped Memory Keys

- [x] 2.1 Add a tenant key prefix to `memory.Store` (`NewStoreWithTenant(kv, prefix)`), defaulting to empty for backward compatibility
- [x] 2.2 Apply the prefix transparently via a `prefixKV` decorator (namespaces every key; strips the prefix in `Keys()` so the store's `/_memory/...` index scans keep working) — covers basic, advanced, policy + their indexes
- [x] 2.3 Add tests: empty prefix reproduces the exact existing `/_memory/...` keys; non-empty prefix isolates two tenants on `Put`/`Get`/`List` (including the no-filter scan path) so neither sees the other's records

## 3. Wire Tenant Seam at Request Boundary

- [x] 3.1 Add optional `tenant:` config block to `ServerConfig` in `pkg/config/config.go` (defaults to bypass when absent)
- [x] 3.2 Construct the configured tenant `Service` during server startup in `pkg/server` (bypass by default). Note: the memory store is built client-side (`cmd/cli`), so namespacing is enforced at that seam today; the server holds the service as the choke point for a future server-side memory RPC
- [x] 3.3 Resolve tenant in `cmd/cli` memory subcommands via a persistent `--tenant` flag so CLI calls flow through the same seam (bypass by default)
- [x] 3.4 Verify default config (no tenant block) produces identical behavior — covered by `TestEmptyPrefixPreservesLayout` and `TestNewDefaultsToBypass`

## 4. Cloud Storage Backend (AWS S3)

- [x] 4.1 Implement a self-contained S3 client (stdlib `net/http` + AWS SigV4 signer) in `pkg/storage/s3` — no new module dependency (AWS SDK v2 requires Go ≥1.24; project is on 1.21.6), keeping `go.mod` minimal and the build hermetic
- [x] 4.1 Implement a self-contained S3 client (stdlib `net/http` + AWS SigV4 signer) in `pkg/storage/s3` — no new module dependency
- [x] 4.2 Create `pkg/storage/s3/s3.go` implementing `storage.PersistentStorage` (`Write([]byte)`/`Read([]byte)`) that puts/gets the encoded snapshot to a configured bucket+key, supporting region, credentials, and an optional custom endpoint (path-style)
- [x] 4.3 Add an `s3:` config block (bucket, region, key, endpoint, credentials) to `pkg/config/config.go` with defaults + `AWS_*` env fallback, read when `storage: s3`
- [x] 4.4 Extend the storage-class factory (server wiring in `pkg/server`) so `StorageClassS3` constructs the S3 backend as the active `PersistentStorage`
- [x] 4.5 Fail startup with a clear error when `storage: s3` is selected but required S3 config (bucket/region/key/creds) is missing — no silent fallback to local
- [x] 4.6 Add tests for the S3 backend against an in-process `httptest` S3-compatible endpoint: `Write`→`Read` round-trips the snapshot bytes; missing object reads empty; custom endpoint override and required-field validation covered

## 5. Deployment Modes & Configuration

- [x] 5.1 Provide example configs for local mode (`config/local.yml`) and cloud mode (`config/cloud.yml`). Note: `config/**` is gitignored (even `default.yml` is local-only), so the canonical examples are also embedded in the tracked docs (MEMORY.md)
- [x] 5.2 Verify local mode end-to-end: started server, `memory basic put`/`list`/`get` in bypass + per-tenant; confirmed cross-tenant isolation and `tenant/<id>/...` key namespacing; confirmed persistence survives restart with the LSM engine
- [x] 5.3 Verify cloud mode: S3 `Write`→`Read` round-trip with real SigV4 signing against an in-process S3-compatible (`httptest`) endpoint (unit test); server cloud-mode wiring fail-fast on missing S3 config verified end-to-end. (Real MinIO/AWS not available in this environment; documented manual steps in MEMORY.md)

## 6. Architecture Validation & Docs

- [x] 6.1 Add an architecture test (`pkg/arch`) confirming the storage-engine layer has no import dependency on the tenant or UI layers (and that the tenant layer is storage-agnostic)
- [x] 6.2 Update `README.md` and `MEMORY.md` to document the four-layer architecture, the tenant bypass interface, the `storage: s3` cloud backend, and the local-vs-cloud deployment modes
- [x] 6.3 Document the vector-storage extension point (how a future AWS vector backend implements the cloud persistence contract without touching the memory/tenant layers)
- [x] 6.4 Run `go build ./...` and `go test ./...`; confirm the full suite passes
