## ADDED Requirements

### Requirement: Tenant Resolution Interface

The system SHALL define a tenant-service interface that every memory operation passes through. The interface SHALL, at minimum, resolve a request into a `TenantContext` (carrying a tenant identifier) and derive a tenant-scoped key prefix from that context. The memory and storage layers SHALL depend only on this interface, never on a concrete tenant implementation.

#### Scenario: Memory operation resolves a tenant context

- **WHEN** a memory operation begins
- **THEN** the tenant service is invoked to produce a `TenantContext` before any key is read or written

#### Scenario: External implementation plugs into the interface

- **WHEN** a separate project provides a multi-tenant implementation
- **THEN** it satisfies the tenant-service interface and is wired in by configuration, without changes to the memory or storage layers

### Requirement: Default Bypass Implementation

The system SHALL provide a default bypass (no-op pass-through) implementation of the tenant-service interface. In bypass mode, the resolved `TenantContext` SHALL carry no tenant identifier and the derived key prefix SHALL be empty, so memory keys are stored under the existing `/_memory/...` layout. Bypass SHALL be the default when no tenant configuration is present.

#### Scenario: No tenant configuration falls back to bypass

- **WHEN** the server starts with no `tenant:` configuration block
- **THEN** the bypass implementation is used and memory behavior is identical to the pre-change system

#### Scenario: Bypass produces an empty key prefix

- **WHEN** the bypass tenant service resolves any request
- **THEN** the derived key prefix is empty and no tenant namespacing is applied

### Requirement: Tenant-Scoped Key Derivation

When a non-bypass tenant service resolves a `TenantContext` with a tenant identifier, the derived key prefix SHALL be deterministic and unique per tenant, and the memory layer SHALL prepend it to every memory and index key. Reads and writes SHALL only ever see keys within the resolving tenant's prefix.

#### Scenario: Tenant identifier yields a stable prefix

- **WHEN** the same tenant identifier is resolved on two separate requests
- **THEN** the derived key prefix is identical on both requests

#### Scenario: Tenant isolation on read

- **WHEN** tenant `acme` lists basic memories
- **THEN** only records written under the `acme` prefix are returned, and no record belonging to another tenant is visible

### Requirement: Stable Interface Contract

The tenant-service interface SHALL remain stable so an external multi-tenant project can implement it without coupling to internal storage details. The interface SHALL NOT expose storage-engine types; it SHALL operate only on request/identity inputs and produce a `TenantContext` and key prefix.

#### Scenario: Interface independent of storage engine

- **WHEN** the underlying storage engine changes between mem, LSM, and S3-backed
- **THEN** the tenant-service interface and its implementations require no changes
