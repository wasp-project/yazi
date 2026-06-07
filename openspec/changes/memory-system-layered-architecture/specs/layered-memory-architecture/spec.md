## ADDED Requirements

### Requirement: Four-Layer Memory Architecture

The memory system SHALL be structured as four layers with stable contracts between them: (1) User Interface, (2) Tenant Service, (3) Storage Engines, and (4) Cloud Vendors. Each layer SHALL depend only on the layer directly beneath it through a defined interface, so that any layer can be replaced without modifying the others. The Storage Engines layer (the existing `memory.Store` + `storage.KVStore`) SHALL remain the reusable core and SHALL NOT depend on the User Interface or Tenant Service layers.

#### Scenario: A memory write traverses all four layers

- **WHEN** a client issues `memory basic put` through the User Interface
- **THEN** the request passes through the Tenant Service layer, then the Storage Engines layer persists it, and when cloud persistence is configured the Cloud Vendors layer receives the encoded snapshot
- **AND** no layer reaches past the layer directly beneath it

#### Scenario: Replacing the cloud backend does not affect upper layers

- **WHEN** the configured Cloud Vendors backend changes from local disk to S3
- **THEN** the User Interface, Tenant Service, and Storage Engines layers operate unchanged with no code modification

### Requirement: Layer Boundary Interfaces

Each inter-layer boundary SHALL be expressed as a Go interface. The Tenant Service boundary SHALL expose a tenant-resolution interface; the Storage Engines boundary SHALL reuse the existing `storage.KVStore` and `memory` KV interfaces; the Cloud Vendors boundary SHALL reuse the existing `storage.PersistentStorage` interface. Adding a new implementation at any boundary SHALL require only implementing the corresponding interface and registering it via configuration.

#### Scenario: New backend added by interface implementation

- **WHEN** a developer adds a new cloud persistence backend
- **THEN** they implement `storage.PersistentStorage` and make it selectable by a `storage` config value, without changing the memory or tenant layers

### Requirement: Per-Tenant Key Namespacing

When a tenant is present, the architecture SHALL namespace all memory keys under a per-tenant prefix so that data for different tenants is isolated within the same storage engine. When no tenant is present (bypass mode), keys SHALL use the existing un-prefixed `/_memory/...` layout unchanged. The namespacing SHALL be applied at the Tenant Service boundary and SHALL be transparent to the Storage Engines layer.

#### Scenario: Two tenants store the same logical key

- **WHEN** tenant `acme` and tenant `globex` each store a basic memory with identical content
- **THEN** the records are written under distinct tenant-prefixed keys and neither tenant can read the other's record

#### Scenario: Bypass mode preserves existing layout

- **WHEN** the tenant service is in bypass mode
- **THEN** memory keys are stored under `/_memory/...` with no tenant prefix, identical to the pre-change layout

### Requirement: Local and Cloud Deployment Modes

The same binary SHALL support two deployment modes selected purely by configuration: a **local** mode (single node, local-disk or LSM persistence, tenant service bypassed) and a **cloud** mode (S3 persistence, tenant service enabled). Switching modes SHALL NOT require code changes or recompilation.

#### Scenario: Local mode reproduces current behavior

- **WHEN** the server starts with `storage: local` and no tenant block configured
- **THEN** it behaves identically to the current single-node memory store

#### Scenario: Cloud mode enables shared company-wide use

- **WHEN** the server starts with `storage: s3` and a tenant service configured
- **THEN** memory is persisted to the configured S3 bucket and access is namespaced per tenant
